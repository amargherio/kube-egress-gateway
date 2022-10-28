package vmssvmclient

import (
	"context"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
)

type Interface interface {
	// List() gets a list of VMs in VM scale sets
	List(ctx context.Context, resourceGroupName, vmScaleSetName string) ([]*compute.VirtualMachineScaleSetVM, error)

	// Update() updates a VM instance in a VM scale set
	Update(ctx context.Context, resourceGroupName, vmScaleSetName, instanceID string, vmssVM compute.VirtualMachineScaleSetVM) (*compute.VirtualMachineScaleSetVM, error)
}