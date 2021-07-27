/*
Copyright 2021.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type JVBStrategyType string

const (
	JVBStrategyStatic     JVBStrategyType = "static"
	JVBStrategyDeamon     JVBStrategyType = "deamonset"
	JVBStrategyAutoScaled JVBStrategyType = "autoscaled"
)

type JVBStrategy struct {
	//+kubebuilder:validation:Enum=static;deamonset;autoscaled
	Type JVBStrategyType `json:"type,omitempty"`
	//+optional
	Replicas *int32 `json:"replicas,omitempty"`
	//+optional
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
}

type JVBPorts struct {
	//+optional
	UDP *int32 `json:"udp,omitempty"`
	//+optional
	TCP *int32 `json:"tcp,omitempty"`
}

type JVB struct {
	//+optional
	Strategy JVBStrategy `json:"strategy,omitempty"`
	//+optional
	*ContainerRuntime `json:",inline"`
	//+optional
	Ports JVBPorts `json:"ports,omitempty"`
}

type Prosody struct {
	*ContainerRuntime `json:",inline"`
}

type ContainerRuntime struct {
	//+optional
	Image string `json:"image,omitempty"`
	//+optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	//+optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type Jicofo struct {
	*ContainerRuntime `json:",inline"`
}

type Jibri struct {
	//+optional
	Enabled bool `json:"enabled,omitempty"`
	//+optional
	*ContainerRuntime `json:",inline"`
	//+optional
	Replicas *int32 `json:"replicas,omitempty"`
}

type Web struct {
	*ContainerRuntime `json:"inline,"`
	//+optional
	Replicas *int32 `json:"replicas,omitempty"`
}

type VersionChannel string

const (
	VersionUnstable VersionChannel = "stable"
	VersionStable   VersionChannel = "unstable"
)

type Version struct {
	Channel VersionChannel `json:"channel,omitempty"`
	Tag     string         `json:"tag,omitempty"`
}

// JitsiSpec defines the desired state of Jitsi
type JitsiSpec struct {
	//+optional
	JVB JVB `json:"jvb,omitempty"`
	//+optional
	Prosody Prosody `json:"prosody,omitempty"`
	//+optional
	Jicofo Jicofo `json:"jicofo,omitempty"`
	//+optional
	Jibri Jibri `json:"jibri,omitempty"`
	//+optional
	Web Web `json:"web,omitempty"`
	//+optional
	Domain string `json:"domain"`
	//+optional
	Region string `json:"region,omitempty"`
	//+optional
	Timezone string `json:"timezone,omitempty"`
	//+optional
	Version Version `json:"version,omitempty"`
}

// JitsiStatus defines the observed state of Jitsi
type JitsiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	JVBStatus JVBStatus `json:"jvb,omitempty"`
}

type JVBStatus struct {
	Replicas int `json:"jvb,omitempty"`
	// Condition
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Jitsi is the Schema for the jitsis API
type Jitsi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JitsiSpec   `json:"spec,omitempty"`
	Status JitsiStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JitsiList contains a list of Jitsi
type JitsiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Jitsi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Jitsi{}, &JitsiList{})
}
