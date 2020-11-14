// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventSubscription) DeepCopyInto(out *EventSubscription) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventSubscription.
func (in *EventSubscription) DeepCopy() *EventSubscription {
	if in == nil {
		return nil
	}
	out := new(EventSubscription)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventSubscription) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventSubscriptionList) DeepCopyInto(out *EventSubscriptionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]EventSubscription, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventSubscriptionList.
func (in *EventSubscriptionList) DeepCopy() *EventSubscriptionList {
	if in == nil {
		return nil
	}
	out := new(EventSubscriptionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *EventSubscriptionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventSubscriptionSpec) DeepCopyInto(out *EventSubscriptionSpec) {
	*out = *in
	out.HostRef = in.HostRef
	if in.EventTypes != nil {
		in, out := &in.EventTypes, &out.EventTypes
		*out = make([]EventType, len(*in))
		copy(*out, *in)
	}
	if in.HTTPHeaders != nil {
		in, out := &in.HTTPHeaders, &out.HTTPHeaders
		*out = make([]HTTPHeader, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventSubscriptionSpec.
func (in *EventSubscriptionSpec) DeepCopy() *EventSubscriptionSpec {
	if in == nil {
		return nil
	}
	out := new(EventSubscriptionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventSubscriptionStatus) DeepCopyInto(out *EventSubscriptionStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventSubscriptionStatus.
func (in *EventSubscriptionStatus) DeepCopy() *EventSubscriptionStatus {
	if in == nil {
		return nil
	}
	out := new(EventSubscriptionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPHeader) DeepCopyInto(out *HTTPHeader) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPHeader.
func (in *HTTPHeader) DeepCopy() *HTTPHeader {
	if in == nil {
		return nil
	}
	out := new(HTTPHeader)
	in.DeepCopyInto(out)
	return out
}
