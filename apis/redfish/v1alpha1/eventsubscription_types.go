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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EventType defines the kinds of events that can trigger notifications.
// +kubebuilder:validation:Enum=Alert;ResourceAdded;ResourceRemoved;ResourceUpdated;StatusChange
type EventType string

const (
	// AlertEventType indicates a condition exists which requires attention.
	AlertEventType EventType = "Alert"
	// ResourceAddedEventType indicates a resource has been added.
	ResourceAddedEventType EventType = "ResourceAdded"
	// ResourceRemovedEventType indicates a resource has been removed.
	ResourceRemovedEventType EventType = "ResourceRemoved"
	// ResourceUpdatedEventType indicates a resource has been updated.
	ResourceUpdatedEventType EventType = "ResourceUpdated"
	// StatusChangeEventType indicates the status of this resource has changed.
	StatusChangeEventType EventType = "StatusChange"
)

// HTTPHeader defines one header value for the HTTP request to the
// callback webhook.
type HTTPHeader struct {
	Name  string `json:"name,required"`
	Value string `json:"value,required"`
}

// EventSubscriptionSpec defines the desired state of EventSubscription
type EventSubscriptionSpec struct {
	// HostRef specifies the host that the subscription is associated
	// with. Only Name and Namespace are required. The named resource
	// must be a BareMetalHost.
	HostRef corev1.ObjectReference `json:"hostRef,required"`

	// DestinationURL defines the endpoint to which event data is
	// POSTed
	DestinationURL string `json:"destinationURL,required"`

	// EventTypes are the kinds of events that should trigger
	// notifications to be sent to this subscription.
	EventTypes []EventType `json:"eventTypes,omitempty"`

	// HTTPHeaders are passed in the POST request to the
	// DestinationURL.
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty"`
}

// EventSubscriptionStatus defines the observed state of EventSubscription
type EventSubscriptionStatus struct {
	// ErrorMessage holds any error encountered when registering the
	// subscription
	ErrorMessage string `json:"errorMessage"`

	// ODataID is taken from the `@odata.id` value of a successful
	// subscription
	ODataID string `json:"odataID"`
}

// +kubebuilder:object:root=true

// EventSubscription is the Schema for the eventsubscriptions API
type EventSubscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventSubscriptionSpec   `json:"spec,omitempty"`
	Status EventSubscriptionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EventSubscriptionList contains a list of EventSubscription
type EventSubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventSubscription `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventSubscription{}, &EventSubscriptionList{})
}
