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
package cnimanager

//+kubebuilder:rbac:groups=egressgateway.kubernetes.azure.com,resources=staticgatewayconfigurations,verbs=get;list;watch
//+kubebuilder:rbac:groups=egressgateway.kubernetes.azure.com,resources=staticgatewayconfigurations/status,verbs=get;
//+kubebuilder:rbac:groups=egressgateway.kubernetes.azure.com,resources=podwireguardendpoints,verbs=list;watch;create;update;patch;delete;

import (
	"context"

	current "github.com/Azure/kube-egress-gateway/api/v1alpha1"
	cniprotocol "github.com/Azure/kube-egress-gateway/pkg/cniprotocol/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type NicService struct {
	k8sClient client.Client
	cniprotocol.UnimplementedNicServiceServer
}

func NewNicService(k8sClient client.Client) *NicService {
	return &NicService{k8sClient: k8sClient}
}

// NicAdd add nic

func (s *NicService) NicAdd(ctx context.Context, in *cniprotocol.NicAddRequest) (*cniprotocol.NicAddResponse, error) {
	gwConfig := &current.StaticGatewayConfiguration{}
	if err := s.k8sClient.Get(ctx, client.ObjectKey{Name: in.GetGatewayName(), Namespace: in.GetPodConfig().GetPodNamespace()}, gwConfig); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to retrieve staticgatewayconfiguration %s/%s: %s", in.GetPodConfig().GetPodNamespace(), in.GetPodConfig().GetPodName(), err)
	}
	podEndpoint := &current.PodWireguardEndpoint{ObjectMeta: metav1.ObjectMeta{Name: in.GetPodConfig().GetPodName(), Namespace: in.GetPodConfig().GetPodNamespace()}}
	if _, err := controllerutil.CreateOrUpdate(ctx, s.k8sClient, podEndpoint, func() error {
		if err := controllerutil.SetControllerReference(gwConfig, podEndpoint, s.k8sClient.Scheme()); err != nil {
			return err
		}
		podEndpoint.Spec.PodIpAddress = in.GetAllowedIp()
		podEndpoint.Spec.StaticGatewayConfiguration = in.GetGatewayName()
		podEndpoint.Spec.PodWireguardPublicKey = in.PublicKey
		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to update staticgatewayconfiguration %s/%s: %s", in.GetPodConfig().GetPodNamespace(), in.GetPodConfig().GetPodName(), err)
	}
	return &cniprotocol.NicAddResponse{
		EndpointIp: gwConfig.Status.WireguardServerIP,
		ListenPort: gwConfig.Status.WireguardServerPort,
		PublicKey:  gwConfig.Status.WireguardPublicKey,
	}, nil
}

func (s *NicService) NicDel(ctx context.Context, in *cniprotocol.NicDelRequest) (*cniprotocol.NicDelResponse, error) {
	podEndpoint := &current.PodWireguardEndpoint{ObjectMeta: metav1.ObjectMeta{Name: in.GetPodConfig().GetPodName(), Namespace: in.GetPodConfig().GetPodNamespace()}}
	if err := s.k8sClient.Delete(ctx, podEndpoint); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, status.Errorf(codes.Unknown, "failed to delete PodWireguardEndpoint %s/%s: %s", in.GetPodConfig().GetPodNamespace(), in.GetPodConfig().GetPodName(), err)
		}
	}
	return &cniprotocol.NicDelResponse{}, nil
}