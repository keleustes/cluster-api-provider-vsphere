/*
Copyright 2019 The Kubernetes Authors.

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

package util_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/cluster-api-provider-vsphere/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-vsphere/pkg/cloud/vsphere/util"
)

func Test_GetMachinePreferredIPAddress(t *testing.T) {
	testCases := []struct {
		name        string
		machine     *v1alpha3.VSphereMachine
		ipAddr      string
		expectedErr error
	}{
		{
			name: "single IPv4 address, no preferred CIDR",
			machine: &v1alpha3.VSphereMachine{
				Status: v1alpha3.VSphereMachineStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "192.168.0.1",
						},
					},
				},
			},
			ipAddr:      "192.168.0.1",
			expectedErr: nil,
		},
		{
			name: "single IPv6 address, no preferred CIDR",
			machine: &v1alpha3.VSphereMachine{
				Status: v1alpha3.VSphereMachineStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "fdf3:35b5:9dad:6e09::0001",
						},
					},
				},
			},
			ipAddr:      "fdf3:35b5:9dad:6e09::0001",
			expectedErr: nil,
		},
		{
			name: "multiple IPv4 addresses, only 1 internal, no preferred CIDR",
			machine: &v1alpha3.VSphereMachine{
				Status: v1alpha3.VSphereMachineStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "192.168.0.1",
						},
						{
							Type:    corev1.NodeExternalIP,
							Address: "1.1.1.1",
						},
						{
							Type:    corev1.NodeExternalIP,
							Address: "2.2.2.2",
						},
					},
				},
			},
			ipAddr:      "192.168.0.1",
			expectedErr: nil,
		},
		{
			name: "multiple IPv4 addresses, preferred CIDR set to v4",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						PreferredAPIServerCIDR: "192.168.0.0/16",
					},
				},
				Status: v1alpha3.VSphereMachineStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "192.168.0.1",
						},
						{
							Type:    corev1.NodeInternalIP,
							Address: "172.17.0.1",
						},
					},
				},
			},
			ipAddr:      "192.168.0.1",
			expectedErr: nil,
		},
		{
			name: "multiple IPv4 and IPv6 addresses, preferred CIDR set to v4",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						PreferredAPIServerCIDR: "192.168.0.0/16",
					},
				},
				Status: v1alpha3.VSphereMachineStatus{
					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "192.168.0.1",
						},
						{
							Type:    corev1.NodeInternalIP,
							Address: "fdf3:35b5:9dad:6e09::0001",
						},
					},
				},
			},
			ipAddr:      "192.168.0.1",
			expectedErr: nil,
		},
		{
			name: "multiple IPv4 and IPv6 addresses, preferred CIDR set to v6",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						PreferredAPIServerCIDR: "fdf3:35b5:9dad:6e09::/64",
					},
				},
				Status: v1alpha3.VSphereMachineStatus{

					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "192.168.0.1",
						},
						{
							Type:    corev1.NodeInternalIP,
							Address: "fdf3:35b5:9dad:6e09::0001",
						},
					},
				},
			},
			ipAddr:      "fdf3:35b5:9dad:6e09::0001",
			expectedErr: nil,
		},
		{
			name: "no addresses found",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						PreferredAPIServerCIDR: "fdf3:35b5:9dad:6e09::/64",
					},
				},
				Status: v1alpha3.VSphereMachineStatus{
					Addresses: []corev1.NodeAddress{},
				},
			},
			ipAddr:      "",
			expectedErr: util.ErrNoMachineIPAddr,
		},
		{
			name: "no addresses found with preferred CIDR",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						PreferredAPIServerCIDR: "192.168.0.0/16",
					},
				},
				Status: v1alpha3.VSphereMachineStatus{

					Addresses: []corev1.NodeAddress{
						{
							Type:    corev1.NodeInternalIP,
							Address: "10.0.0.1",
						},
					},
				},
			},
			ipAddr:      "",
			expectedErr: util.ErrNoMachineIPAddr,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ipAddr, err := util.GetMachinePreferredIPAddress(tc.machine)
			if err != tc.expectedErr {
				t.Logf("expected err: %q", tc.expectedErr)
				t.Logf("actual err: %q", err)
				t.Errorf("unexpected error")
			}

			if ipAddr != tc.ipAddr {
				t.Logf("expected IP addr: %q", tc.ipAddr)
				t.Logf("actual IP addr: %q", ipAddr)
				t.Error("unexpected IP addr from machine context")
			}
		})
	}
}

func Test_GetMachineMetadata(t *testing.T) {
	testCases := []struct {
		name    string
		machine *v1alpha3.VSphereMachine
	}{
		{
			name: "dhcp4",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName: "network1",
								MACAddr:     "00:00:00:00:00",
								DHCP4:       true,
							},
						},
					},
				},
			},
		},
		{
			name: "dhcp6",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName: "network1",
								MACAddr:     "00:00:00:00:00",
								DHCP6:       true,
							},
						},
					},
				},
			},
		},
		{
			name: "dhcp4+dhcp6",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName: "network1",
								MACAddr:     "00:00:00:00:00",
								DHCP4:       true,
								DHCP6:       true,
							},
						},
					},
				},
			},
		},
		{
			name: "static4+dhcp6",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName: "network1",
								MACAddr:     "00:00:00:00:00",
								DHCP6:       true,
								IPAddrs:     []string{"192.168.4.21"},
								Gateway4:    "192.168.4.1",
							},
						},
					},
				},
			},
		},
		{
			name: "static4+dhcp6+static-routes",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName: "network1",
								MACAddr:     "00:00:00:00:00",
								DHCP6:       true,
								IPAddrs:     []string{"192.168.4.21"},
								Gateway4:    "192.168.4.1",
							},
						},
						Routes: []v1alpha3.NetworkRouteSpec{
							{
								To:     "192.168.5.1/24",
								Via:    "192.168.4.254",
								Metric: 3,
							},
						},
					},
				},
			},
		},
		{
			name: "2nets",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName: "network1",
								MACAddr:     "00:00:00:00:00",
								DHCP4:       true,
								Routes: []v1alpha3.NetworkRouteSpec{
									{
										To:     "192.168.5.1/24",
										Via:    "192.168.4.254",
										Metric: 3,
									},
								},
							},
							{
								NetworkName: "network12",
								MACAddr:     "00:00:00:00:01",
								DHCP6:       true,
								MTU:         mtu(100),
							},
						},
					},
				},
			},
		},
		{
			name: "2nets-static+dhcp",
			machine: &v1alpha3.VSphereMachine{
				Spec: v1alpha3.VSphereMachineSpec{
					Network: v1alpha3.NetworkSpec{
						Devices: []v1alpha3.NetworkDeviceSpec{
							{
								NetworkName:   "network1",
								MACAddr:       "00:00:00:00:00",
								IPAddrs:       []string{"192.168.4.21"},
								Gateway4:      "192.168.4.1",
								MTU:           mtu(0),
								Nameservers:   []string{"1.1.1.1"},
								SearchDomains: []string{"vmware.ci"},
							},
							{
								NetworkName:   "network12",
								MACAddr:       "00:00:00:00:01",
								DHCP6:         true,
								SearchDomains: []string{"vmware6.ci"},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.machine.Name = tc.name
			actVal, err := util.GetMachineMetadata(*tc.machine)
			if err != nil {
				t.Fatal(err)
			}
			t.Log(string(actVal))
		})
	}
}

func mtu(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}
