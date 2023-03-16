/*
Copyright (c) Microsoft Corporation. All rights reserved.
Licensed under the MIT License. See License.txt in the project root for
license information.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	SiteKind = reflect.TypeOf(SpringBootSite{}).Name()
)

// SpringBootSiteSpec defines the desired state of SpringBootSite
type SpringBootSiteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is name of SpringBootSite.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Name string `json:"name"`
}

// SpringBootSiteStatus defines the observed state of SpringBootSite
type SpringBootSiteStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	ProvisioningStatus ProvisioningStatus `json:"provisioningStatus,omitempty"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Created",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.provisioningStatus.status"
//+kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.provisioningStatus.error.message"
//+kubebuilder:printcolumn:name="OperationID",type="string",JSONPath=".status.provisioningStatus.operationID"

// SpringBootSite is the Schema for the springbootsites API
type SpringBootSite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpringBootSiteSpec   `json:"spec,omitempty"`
	Status SpringBootSiteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SpringBootSiteList contains a list of SpringBootSite
type SpringBootSiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpringBootSite `json:"items"`
}
