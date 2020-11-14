/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	bmh "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"github.com/metal3-io/baremetal-operator/pkg/bmc"
	redfishv1alpha1 "github.com/metal3-io/redfish-event-controller/apis/redfish/v1alpha1"
)

// EventSubscriptionReconciler reconciles a EventSubscription object
type EventSubscriptionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=redfish.metal3.io,resources=eventsubscriptions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=redfish.metal3.io,resources=eventsubscriptions/status,verbs=get;update;patch

func (r *EventSubscriptionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("eventsubscription", req.NamespacedName)

	log.Info("reconciling")

	subscription := &redfishv1alpha1.EventSubscription{}
	err := r.Get(context.TODO(), req.NamespacedName, subscription)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrap(err, "could not load subscription data")
	}

	hostKey := client.ObjectKey{
		Name:      subscription.Spec.HostRef.Name,
		Namespace: subscription.Spec.HostRef.Namespace,
	}
	if hostKey.Namespace == "" {
		hostKey.Namespace = req.Namespace
	}

	log.Info("loading host", "host", hostKey)
	host := &bmh.BareMetalHost{}
	err = r.Get(context.TODO(), hostKey, host)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("failed to load host", "host", hostKey)
			msg := fmt.Sprintf("could not load host %s", hostKey)
			err := r.setErrorMessage(subscription, msg)
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, errors.Wrap(err, "could not load subscription data")
	}

	// TODO: add finalizer to the host and handle deleting the subscription

	log.Info("verifying BMC access", "address", host.Spec.BMC.Address)
	if host.Spec.BMC.Address == "" {
		err := fmt.Errorf("host %s/%s has no BMC address", host.Namespace, host.Name)
		return ctrl.Result{}, err
	}
	accessDetails, err := bmc.NewAccessDetails(
		host.Spec.BMC.Address,
		host.Spec.BMC.DisableCertificateVerification,
	)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "could not understand BMC address: %s")
	}
	if !strings.Contains(accessDetails.Type(), "redfish") {
		return ctrl.Result{}, fmt.Errorf("BMC does not use Redfish protocol")
	}

	username, password, err := r.getBMCCredentials(req, subscription, host)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "could not find credentials for host BMC")
	}

	parsedURL, err := url.Parse(host.Spec.BMC.Address)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "failed to parse BMC address")
	}
	address := getRedfishAddress(parsedURL.Scheme, parsedURL.Host)
	log.Info("connecting to BMC", "address", address)
	gofishConfig := gofish.ClientConfig{
		Endpoint:  address,
		Username:  username,
		Password:  password,
		BasicAuth: true,
		Insecure:  host.Spec.BMC.DisableCertificateVerification,
	}
	conn, err := gofish.Connect(gofishConfig)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "could not connect to BMC")
	}

	eventService, err := conn.Service.EventService()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "could not access event service on BMC")
	}

	// The Redfish API does not return all of the details passed when
	// the subscription is registered, so we cannot simply ask the BMC
	// for the existing subscription and then see if the fields
	// match. Instead, we have to encode a signature in a field that
	// is returned on a later fetch and the compare that with our
	// current signature. We search through all of the existing
	// subscriptions instead of looking directly for the one with the
	// matching UID in our status field in case we successfully
	// created the subscription but failed to record its ID.
	signature := getSignatureForSubscription(subscription)
	contextPrefix := fmt.Sprintf("metal3:%s", subscription.UID)
	expectedContext := fmt.Sprintf("%s:%s", contextPrefix, signature)
	var existingSub *redfish.EventDestination
	existingSubscriptions, err := eventService.GetEventSubscriptions()
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "could not access existing subscriptions")
	}
	log.Info("looking for existing subscription",
		"key", contextPrefix,
		"count", len(existingSubscriptions),
	)
	for _, candidate := range existingSubscriptions {
		if strings.HasPrefix(candidate.Context, contextPrefix) {
			existingSub = candidate
			break
		}
	}
	if existingSub != nil && existingSub.Context == expectedContext {
		log.Info("existing subscription matches signature", "uri", existingSub.ODataID)
		subscription.Status.ODataID = existingSub.ODataID
		subscription.Status.ErrorMessage = ""
		err := r.Status().Update(context.TODO(), subscription)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to update subscription status")
		}
		return ctrl.Result{}, nil
	}

	// create the new subscription
	eventTypes := []redfish.EventType{}
	for _, et := range subscription.Spec.EventTypes {
		eventTypes = append(eventTypes, redfish.EventType(et))
	}
	// FIXME: This conversion is lossy when compared to the way the
	// value is used in the signature, since we only use the first
	// value for each header here.
	httpHeaders := getHeadersForSubscription(subscription)
	headers := map[string]string{}
	for k := range httpHeaders {
		headers[k] = httpHeaders.Get(k)
	}
	subscriptionURI, err := eventService.CreateEventSubscription(
		subscription.Spec.DestinationURL,
		eventTypes,
		headers,
		redfish.RedfishEventDestinationProtocol,
		expectedContext,
		nil,
	)
	log.Info("created new subscription", "uri", subscriptionURI)
	subscription.Status.ODataID = subscriptionURI
	err = r.Status().Update(context.TODO(), subscription)
	if err != nil {
		return ctrl.Result{}, errors.Wrap(err, "could not update status after creating subscription")
	}

	// The API for subscriptions does not allow changes, so we have
	// already made a new subscription. Now we delete the existing
	// subscription, since we won't miss any events. We do this in a
	// somewhat brute-force way, because if we encountered an error
	// previously when we tried to update the status we could have an
	// extra subscription.
	for _, candidate := range existingSubscriptions {
		if !strings.HasPrefix(candidate.Context, contextPrefix) {
			continue
		}
		if candidate.Context == expectedContext {
			continue
		}
		err := eventService.DeleteEventSubscription(candidate.ODataID)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err,
				fmt.Sprintf("failed to remove old subscription %s", candidate.ODataID))
		}
	}

	return ctrl.Result{}, nil
}

func getHeadersForSubscription(subscription *redfishv1alpha1.EventSubscription) http.Header {
	headers := http.Header{}
	for _, header := range subscription.Spec.HTTPHeaders {
		headers.Add(header.Name, header.Value)
	}
	return headers
}

func getSignatureForSubscription(subscription *redfishv1alpha1.EventSubscription) string {
	h := sha256.New()
	h.Write([]byte(subscription.Namespace))
	h.Write([]byte(subscription.Name))
	h.Write([]byte(subscription.Spec.DestinationURL))

	eventTypes := []string{}
	for _, et := range subscription.Spec.EventTypes {
		eventTypes = append(eventTypes, string(et))
	}
	sort.Strings(eventTypes)
	for _, et := range eventTypes {
		h.Write([]byte(et))
	}

	// http.Header writes the headers in order based on the key names,
	// so we should get a consistent result this way
	headers := getHeadersForSubscription(subscription)
	headers.Write(h)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// taken from github.com/metal3-io/baremetal-operator/pkg/bmc/redfish.go
func getRedfishAddress(bmcType, host string) string {
	redfishAddress := []string{}
	schemes := strings.Split(bmcType, "+")
	if len(schemes) > 1 {
		redfishAddress = append(redfishAddress, schemes[1])
	} else {
		redfishAddress = append(redfishAddress, "https")
	}
	redfishAddress = append(redfishAddress, "://")
	redfishAddress = append(redfishAddress, host)
	return strings.Join(redfishAddress, "")
}

func (r *EventSubscriptionReconciler) setErrorMessage(subscription *redfishv1alpha1.EventSubscription, message string) error {
	if subscription.Status.ErrorMessage == message {
		// we've already recorded this error
		return nil
	}
	subscription.Status.ErrorMessage = message
	err := r.Status().Update(context.TODO(), subscription)
	if err != nil {
		return errors.Wrap(err, "failed to update status")
	}
	return nil
}

func (r *EventSubscriptionReconciler) getBMCCredentials(req ctrl.Request, subscription *redfishv1alpha1.EventSubscription, host *bmh.BareMetalHost) (username, password string, err error) {

	if host.Spec.BMC.CredentialsName == "" {
		err = fmt.Errorf("host %s/%s has no credentials secret", host.Namespace, host.Name)
		return
	}

	log := r.Log.WithValues("eventsubscription", req.NamespacedName)
	secretKey := host.CredentialsKey()
	log.Info("loading credentials", "secret", secretKey)
	bmcCredsSecret := &corev1.Secret{}
	err = r.Get(context.TODO(), secretKey, bmcCredsSecret)
	if err != nil {
		return
	}

	username = string(bmcCredsSecret.Data["username"])
	password = string(bmcCredsSecret.Data["password"])
	return
}

func (r *EventSubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redfishv1alpha1.EventSubscription{}).
		Complete(r)
}
