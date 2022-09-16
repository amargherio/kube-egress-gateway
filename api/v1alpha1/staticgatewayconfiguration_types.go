/*
   MIT License

   Copyright (c) Microsoft Corporation.

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	addonv1alpha1 "sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/apis/v1alpha1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GatewayVMSSProfile finds an existing gateway VMSS (virtual machine scale set).
type GatewayVMSSProfile struct {
	// Resource group of the VMSS. Must be in the same subscription.
	VMSSResourceGroup string `json:"vmssResourceGroup,omitempty"`

	// Name of the VMSS
	VMSSName string `json:"vmssName,omitempty"`
}

// StaticGatewayConfigurationSpec defines the desired state of StaticGatewayConfiguration
type StaticGatewayConfigurationSpec struct {
	addonv1alpha1.CommonSpec `json:",inline"`
	addonv1alpha1.PatchSpec  `json:",inline"`

	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name of the gateway nodepool to apply the gateway configuration.
	GatewayNodepoolName string `json:"gatewayNodepoolName,omitempty"`

	// Profile of the gateway VMSS to apply the gateway configuration.
	GatewayVMSSProfile `json:"gatewayVmssProfile,omitempty"`

	// List of destination cidrs not to be routed via gateway.
	ExceptionCIDRs []string `json:"exceptionCIDRs,omitempty"`

	// BYO Resource ID of public IP prefix to be used as outbound.
	PublicIpPrefixId string `json:"publicIpPrefixId,omitempty"`
}

// GatewayWireguardProfile provides details about gateway side wireguard configuration.
type GatewayWireguardProfile struct {
	// Gateway IP for wireguard connection.
	WireguardServerIP string `json:"wireguardServerIP,omitempty"`

	// Listening port of the gateway side wireguard daemon.
	WireguardServerPort int `json:"wireguardServerPort,omitempty"`

	// Name of the secret that holds gateway side wireguard public key.
	WireguardKeySecret string `json:"wireguardKeySecret,omitempty"`
}

// StaticGatewayConfigurationStatus defines the observed state of StaticGatewayConfiguration
type StaticGatewayConfigurationStatus struct {
	addonv1alpha1.CommonStatus `json:",inline"`

	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// State of the GatewayConfiguration.
	State string `json:"state,omitempty"`

	// Additional message (e.g. error) to explain the state.
	Message string `json:"message,omitempty"`

	// Public IP Prefix CIDR used for this gateway configuration.
	PublicIpPrefix string `json:"publicIpPrefix,omitempty"`

	// Gateway side wireguard profile.
	GatewayWireguardProfile `json:"gatewayWireguardProfile,omitempty"`

	// List of active nodes that have this gateway configuration setup ready.
	ActiveNodes []string `json:"activeNodes,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// StaticGatewayConfiguration is the Schema for the staticgatewayconfigurations API
type StaticGatewayConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaticGatewayConfigurationSpec   `json:"spec,omitempty"`
	Status StaticGatewayConfigurationStatus `json:"status,omitempty"`
}

var _ addonv1alpha1.CommonObject = &StaticGatewayConfiguration{}

func (o *StaticGatewayConfiguration) ComponentName() string {
	return "staticgatewayconfiguration"
}

func (o *StaticGatewayConfiguration) CommonSpec() addonv1alpha1.CommonSpec {
	return o.Spec.CommonSpec
}

func (o *StaticGatewayConfiguration) PatchSpec() addonv1alpha1.PatchSpec {
	return o.Spec.PatchSpec
}

func (o *StaticGatewayConfiguration) GetCommonStatus() addonv1alpha1.CommonStatus {
	return o.Status.CommonStatus
}

func (o *StaticGatewayConfiguration) SetCommonStatus(s addonv1alpha1.CommonStatus) {
	o.Status.CommonStatus = s
}

//+kubebuilder:object:root=true

// StaticGatewayConfigurationList contains a list of StaticGatewayConfiguration
type StaticGatewayConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StaticGatewayConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StaticGatewayConfiguration{}, &StaticGatewayConfigurationList{})
}