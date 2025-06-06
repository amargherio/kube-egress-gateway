// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package azmanager

import (
	"context"
	"fmt"
	"time"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v6"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/interfaceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/loadbalancerclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/publicipprefixclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/subnetclient"
	_ "sigs.k8s.io/cloud-provider-azure/pkg/azclient/trace"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient"

	"github.com/Azure/kube-egress-gateway/pkg/config"
	"github.com/Azure/kube-egress-gateway/pkg/consts"
	"github.com/Azure/kube-egress-gateway/pkg/utils/to"
)

const (
	// LB frontendIPConfiguration ID template
	LBFrontendIPConfigTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/frontendIPConfigurations/%s"
	// LB backendAddressPool ID template
	LBBackendPoolIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/backendAddressPools/%s"
	// LB probe ID template
	LBProbeIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/probes/%s"
)

type AzureManager struct {
	*config.CloudConfig

	LoadBalancerClient   loadbalancerclient.Interface
	VmssClient           virtualmachinescalesetclient.Interface
	VmssVMClient         virtualmachinescalesetvmclient.Interface
	PublicIPPrefixClient publicipprefixclient.Interface
	InterfaceClient      interfaceclient.Interface
	SubnetClient         subnetclient.Interface
}

func CreateAzureManager(cloud *config.CloudConfig, factory azclient.ClientFactory) (*AzureManager, error) {
	az := AzureManager{
		CloudConfig: cloud,
	}

	az.LoadBalancerClient = factory.GetLoadBalancerClient()
	az.VmssClient = factory.GetVirtualMachineScaleSetClient()
	az.PublicIPPrefixClient = factory.GetPublicIPPrefixClient()
	az.VmssVMClient = factory.GetVirtualMachineScaleSetVMClient()
	az.InterfaceClient = factory.GetInterfaceClient()
	az.SubnetClient = factory.GetSubnetClient()

	return &az, nil
}

func (az *AzureManager) SubscriptionID() string {
	return az.CloudConfig.SubscriptionID
}

func (az *AzureManager) Location() string {
	return az.CloudConfig.Location
}

func (az *AzureManager) LoadBalancerName() string {
	if az.CloudConfig.LoadBalancerName == "" {
		return consts.DefaultGatewayLBName
	}
	return az.CloudConfig.LoadBalancerName
}

func (az *AzureManager) GetLBFrontendIPConfigurationID(name string) *string {
	return to.Ptr(fmt.Sprintf(LBFrontendIPConfigTemplate, az.SubscriptionID(), az.LoadBalancerResourceGroup, az.LoadBalancerName(), name))
}

func (az *AzureManager) GetLBBackendAddressPoolID(name string) *string {
	return to.Ptr(fmt.Sprintf(LBBackendPoolIDTemplate, az.SubscriptionID(), az.LoadBalancerResourceGroup, az.LoadBalancerName(), name))
}

func (az *AzureManager) GetLBProbeID(name string) *string {
	return to.Ptr(fmt.Sprintf(LBProbeIDTemplate, az.SubscriptionID(), az.LoadBalancerResourceGroup, az.LoadBalancerName(), name))
}

func (az *AzureManager) GetLB(ctx context.Context) (*network.LoadBalancer, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	lb, err := az.LoadBalancerClient.Get(ctx, az.LoadBalancerResourceGroup, az.LoadBalancerName(), nil)
	if err != nil {
		return nil, err
	}
	return lb, nil
}

func (az *AzureManager) CreateOrUpdateLB(ctx context.Context, lb network.LoadBalancer) (*network.LoadBalancer, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	ret, err := az.LoadBalancerClient.CreateOrUpdate(ctx, az.LoadBalancerResourceGroup, to.Val(lb.Name), lb)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (az *AzureManager) DeleteLB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if err := az.LoadBalancerClient.Delete(ctx, az.LoadBalancerResourceGroup, az.LoadBalancerName()); err != nil {
		return err
	}
	return nil
}

func (az *AzureManager) ListVMSS(ctx context.Context) ([]*compute.VirtualMachineScaleSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	vmssList, err := az.VmssClient.List(ctx, az.ResourceGroup)
	if err != nil {
		return nil, err
	}
	return vmssList, nil
}

func (az *AzureManager) GetVMSS(ctx context.Context, resourceGroup, vmssName string) (*compute.VirtualMachineScaleSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if vmssName == "" {
		return nil, fmt.Errorf("vmss name is empty")
	}
	vmss, err := az.VmssClient.Get(ctx, resourceGroup, vmssName, nil)
	if err != nil {
		return nil, err
	}
	return vmss, nil
}

func (az *AzureManager) CreateOrUpdateVMSS(ctx context.Context, resourceGroup, vmssName string, vmss compute.VirtualMachineScaleSet) (*compute.VirtualMachineScaleSet, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if vmssName == "" {
		return nil, fmt.Errorf("vmss name is empty")
	}
	retVmss, err := az.VmssClient.CreateOrUpdate(ctx, resourceGroup, vmssName, vmss)
	if err != nil {
		return nil, err
	}
	return retVmss, nil
}

func (az *AzureManager) ListVMSSInstances(ctx context.Context, resourceGroup, vmssName string) ([]*compute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if vmssName == "" {
		return nil, fmt.Errorf("vmss name is empty")
	}
	vms, err := az.VmssVMClient.List(ctx, resourceGroup, vmssName)
	if err != nil {
		return nil, err
	}
	return vms, nil
}

func (az *AzureManager) GetVMSSInstance(ctx context.Context, resourceGroup, vmssName, instanceID string) (*compute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if vmssName == "" {
		return nil, fmt.Errorf("vmss name is empty")
	}
	if instanceID == "" {
		return nil, fmt.Errorf("vmss instanceID is empty")
	}
	vm, err := az.VmssVMClient.Get(ctx, resourceGroup, vmssName, instanceID)
	if err != nil {
		return nil, err
	}
	return vm, nil
}

func (az *AzureManager) UpdateVMSSInstance(ctx context.Context, resourceGroup, vmssName, instanceID string, vm compute.VirtualMachineScaleSetVM) (*compute.VirtualMachineScaleSetVM, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if vmssName == "" {
		return nil, fmt.Errorf("vmss name is empty")
	}
	if instanceID == "" {
		return nil, fmt.Errorf("vmss instanceID is empty")
	}
	retVM, err := az.VmssVMClient.Update(ctx, resourceGroup, vmssName, instanceID, vm)
	if err != nil {
		return nil, err
	}
	return retVM, nil
}

func (az *AzureManager) GetPublicIPPrefix(ctx context.Context, resourceGroup, prefixName string) (*network.PublicIPPrefix, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if prefixName == "" {
		return nil, fmt.Errorf("public ip prefix name is empty")
	}
	prefix, err := az.PublicIPPrefixClient.Get(ctx, resourceGroup, prefixName, nil)
	if err != nil {
		return nil, err
	}
	return prefix, nil
}

func (az *AzureManager) CreateOrUpdatePublicIPPrefix(ctx context.Context, resourceGroup, prefixName string, ipPrefix network.PublicIPPrefix) (*network.PublicIPPrefix, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if prefixName == "" {
		return nil, fmt.Errorf("public ip prefix name is empty")
	}
	prefix, err := az.PublicIPPrefixClient.CreateOrUpdate(ctx, resourceGroup, prefixName, ipPrefix)
	if err != nil {
		return nil, err
	}
	return prefix, nil
}

func (az *AzureManager) DeletePublicIPPrefix(ctx context.Context, resourceGroup, prefixName string) error {
	ctx, cancel := context.WithTimeout(ctx, 900*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if prefixName == "" {
		return fmt.Errorf("public ip prefix name is empty")
	}
	return az.PublicIPPrefixClient.Delete(ctx, resourceGroup, prefixName)
}

func (az *AzureManager) GetVMSSInterface(ctx context.Context, resourceGroup, vmssName, instanceID, interfaceName string) (*network.Interface, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	if resourceGroup == "" {
		resourceGroup = az.ResourceGroup
	}
	if vmssName == "" {
		return nil, fmt.Errorf("vmss name is empty")
	}
	if instanceID == "" {
		return nil, fmt.Errorf("instanceID is empty")
	}
	if interfaceName == "" {
		return nil, fmt.Errorf("interface name is empty")
	}
	nicResp, err := az.InterfaceClient.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroup, vmssName, instanceID, interfaceName)
	if err != nil {
		return nil, err
	}
	return nicResp, nil
}

func (az *AzureManager) GetSubnet(ctx context.Context) (*network.Subnet, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	subnet, err := az.SubnetClient.Get(ctx, az.VnetResourceGroup, az.VnetName, az.SubnetName, nil)
	if err != nil {
		return nil, err
	}
	return subnet, nil
}
