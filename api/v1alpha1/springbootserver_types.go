/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	ServerKind = reflect.TypeOf(SpringBootServer{}).Name()
)

// SpringBootServerSpec defines the desired state of SpringBootServer
type SpringBootServerSpec struct {
	// Server is the target server name or ip address to discover of SpringBootServer.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	Server string `json:"server"`
	// +kubebuilder:validation:Required
	Port int `json:"port"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	SiteName string `json:"siteName"`
}

// SpringBootServerStatus defines the observed state of SpringBootServer
type SpringBootServerStatus struct {
	ObservedGeneration    int64              `json:"observedGeneration,omitempty"`
	ReconcileDurationInMs int64              `json:"reconcileDurationInMs,omitempty"`
	TotalApps             int                `json:"totalApps"`
	SpringBootApps        int                `json:"springBootApps"`
	ProvisioningStatus    ProvisioningStatus `json:"provisioningStatus,omitempty"`
	Errors                []ErrorObj         `json:"errors,omitempty"`
	RunAsAccountId        string             `json:"runAsAccountId,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Created",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Server",type="string",JSONPath=".spec.server"
//+kubebuilder:printcolumn:name="Port",type="integer",JSONPath=".spec.port"
//+kubebuilder:printcolumn:name="TotalApps",type="integer",JSONPath=".status.totalApps"
//+kubebuilder:printcolumn:name="SpringBoot",type="integer",JSONPath=".status.springBootApps"
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.provisioningStatus.status"
//+kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.provisioningStatus.error"

// SpringBootServer is the Schema for the springbootservers API
type SpringBootServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpringBootServerSpec   `json:"spec,omitempty"`
	Status SpringBootServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SpringBootServerList contains a list of SpringBootServer
type SpringBootServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpringBootServer `json:"items"`
}

