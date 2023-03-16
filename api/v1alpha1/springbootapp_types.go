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
	"reflect"
	"strconv"

	"github.com/hashicorp/go-version"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

var (
	AppKind = reflect.TypeOf(SpringBootApp{}).Name()
)

// SpringBootAppSpec defines the desired state of SpringBootApp
type SpringBootAppSpec struct {
	// Server is the discovered server name or ip address of SpringBootApp.
	AppName                   string      `json:"appName,omitempty"`
	ArtifactName              string      `json:"artifactName,omitempty"`
	Checksum                  string      `json:"checksum,omitempty"`
	SiteName                  string      `json:"siteName,omitempty"`
	SpringBootVersion         string      `json:"springBootVersion,omitempty"`
	AppType                   string      `json:"appType,omitempty"`
	RuntimeJdkVersion         string      `json:"runtimeJdkVersion,omitempty"`
	BuildJdkVersion           string      `json:"buildJdkVersion,omitempty"`
	Environments              []string    `json:"environments,omitempty"`
	JvmOptions                []string    `json:"jvmOptions,omitempty"`
	Dependencies              []string    `json:"dependencies,omitempty"`
	ApplicationConfigurations []*KV       `json:"applicationConfigurations,omitempty"`
	Certificates              []string    `json:"certificates,omitempty"`
	JarFileLocation           string      `json:"jarFileLocation,omitempty"`
	StaticContentLocations    []string    `json:"staticContentLocations,omitempty"`
	JvmMemoryInMB             int         `json:"jvmMemoryInMB,omitempty"`
	AppPort                   int         `json:"appPort,omitempty"`
	BindingPorts              []int       `json:"bindingPorts,omitempty"`
	Miscs                     []*KV       `json:"miscs,omitempty"`
	InstanceCount             int         `json:"instanceCount,omitempty"`
	LastModifiedTime          metav1.Time `json:"lastModifiedTime,omitempty"`
	// +optional
	LastUpdatedTime metav1.Time `json:"lastUpdatedTime,omitempty"`
}

// SpringBootAppStatus defines the observed state of SpringBootApp
type SpringBootAppStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Created",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="App",type="string",JSONPath=".spec.appName"
//+kubebuilder:printcolumn:name="RuntimeJdk",type="string",JSONPath=".spec.runtimeJdkVersion"
//+kubebuilder:printcolumn:name="SpringBoot",type="string",JSONPath=".spec.springBootVersion"
//+kubebuilder:printcolumn:name="Location",type="string",JSONPath=".spec.jarFileLocation"
//+kubebuilder:printcolumn:name="Port",type="string",JSONPath=".spec.appPort"
//+kubebuilder:printcolumn:name="InstanceCount",type="string",JSONPath=".spec.instanceCount"
//+kubebuilder:printcolumn:name="LastModified",type="string",JSONPath=".spec.lastModifiedTime"
//+kubebuilder:printcolumn:name="LastUpdated",type="string",JSONPath=".spec.lastUpdatedTime"

// SpringBootApp is the Schema for the springbootapps API
type SpringBootApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SpringBootAppSpec   `json:"spec,omitempty"`
	Status SpringBootAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SpringBootAppList contains a list of SpringBootApp
type SpringBootAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpringBootApp `json:"items"`
}

func (app *SpringBootApp) Merge(other *SpringBootApp) {
	for _, ownRef := range other.OwnerReferences {
		if indexOf(app.OwnerReferences, ownRef, compareOwnerRef) < 0 {
			app.OwnerReferences = append(app.OwnerReferences, ownRef)
		}
	}

	for label, value := range other.Labels {
		if app.Labels == nil {
			app.Labels = make(map[string]string)
		}
		app.Labels[label] = value
	}

	for ann, value := range other.Annotations {
		if app.Annotations == nil {
			app.Annotations = make(map[string]string)
		}
		app.Annotations[ann] = value
	}

	if app.Spec.Checksum != other.Spec.Checksum {
		logf.Log.Info("App with same name but different checksum", "name", app.Name, "new", app.Spec.Checksum, "old", other.Spec.Checksum)
	}

	if app.Spec.RuntimeJdkVersion != other.Spec.RuntimeJdkVersion {
		logf.Log.Info("App with same name but different runtime jdk", "name", app.Name, "new", app.Spec.RuntimeJdkVersion, "old", other.Spec.RuntimeJdkVersion)
		newJdkVersion, err := version.NewVersion(app.Spec.RuntimeJdkVersion)
		if err != nil {
			app.Spec.RuntimeJdkVersion = other.Spec.RuntimeJdkVersion
		} else {
			if oldJdkVersion, err := version.NewVersion(other.Spec.RuntimeJdkVersion); err == nil && newJdkVersion.LessThan(oldJdkVersion) {
				app.Spec.RuntimeJdkVersion = other.Spec.RuntimeJdkVersion
			}
		}
	}

	if app.Spec.JvmMemoryInMB != other.Spec.JvmMemoryInMB {
		logf.Log.Info("App with same name but different heap memory", "name", app.Name, "new", app.Spec.JvmMemoryInMB, "old", other.Spec.JvmMemoryInMB)
		if app.Spec.JvmMemoryInMB < other.Spec.JvmMemoryInMB {
			app.Spec.JvmMemoryInMB = other.Spec.JvmMemoryInMB
			app.Spec.JvmOptions = other.Spec.JvmOptions
		}
	}

	if app.Spec.LastModifiedTime.Before(&other.Spec.LastModifiedTime) {
		app.Spec.LastModifiedTime = other.Spec.LastModifiedTime
	}

	if app.Spec.LastUpdatedTime.Before(&other.Spec.LastUpdatedTime) {
		app.Spec.LastUpdatedTime = other.Spec.LastUpdatedTime
	}

	app.Summarize()
}

func (app *SpringBootApp) Reference(server *SpringBootServer) {
	// copy the site reference from server
	for _, ownRef := range server.OwnerReferences {
		if ownRef.Kind != SiteKind {
			continue
		}
		if indexOf(app.OwnerReferences, ownRef, compareOwnerRef) < 0 {
			app.OwnerReferences = append(app.OwnerReferences, ownRef)
		}
	}
	// set the reference to server
	ownerReference := *newOwnerRef(server, GroupVersion.WithKind(ServerKind))
	if indexOf(app.OwnerReferences, ownerReference, compareOwnerRef) < 0 {
		app.OwnerReferences = append(app.OwnerReferences, ownerReference)
	}
	if app.Labels == nil {
		app.Labels = make(map[string]string)
	}
	if _, exists := app.Labels[server.Name]; exists {
		count, _ := strconv.Atoi(app.Labels[server.Name])
		app.Labels[server.Name] = strconv.Itoa(count + 1)
	} else {
		app.Labels[server.Name] = strconv.Itoa(1)
	}

	app.Summarize()
}

func (app *SpringBootApp) Summarize() {
	var instanceCount int
	for _, label := range app.Labels {
		count, _ := strconv.Atoi(label)
		instanceCount += count
	}
	app.Spec.InstanceCount = instanceCount
}

func (app *SpringBootApp) Dereference(name string) {
	var idxToDelete = -1
	for idx, ownref := range app.OwnerReferences {
		if ownref.Kind == ServerKind && ownref.Name == name {
			idxToDelete = idx
			break
		}
	}
	if idxToDelete >= 0 {
		app.OwnerReferences = removeByIndex(app.OwnerReferences, idxToDelete)
	}

	delete(app.Labels, name)
	app.Summarize()
}

func (app *SpringBootApp) Orphaned() bool {
	orphaned := true
	for _, ownref := range app.OwnerReferences {
		if ownref.Kind == ServerKind {
			orphaned = false
			break
		}
	}
	return orphaned
}

func removeByIndex[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func newOwnerRef(owner client.Object, gvk schema.GroupVersionKind) *metav1.OwnerReference {
	controller := false
	blockOwnerDeletion := true
	return &metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		Controller:         &controller,
		BlockOwnerDeletion: &blockOwnerDeletion,
	}
}

func compareOwnerRef(a, b metav1.OwnerReference) bool {
	return a.Name == b.Name && a.Kind == b.Kind && a.APIVersion == b.APIVersion
}

func indexOf[T any](slices []T, search T, indexFunc func(a, b T) bool) int {
	for index, t := range slices {
		if indexFunc(t, search) {
			return index
		}
	}
	return -1
}
