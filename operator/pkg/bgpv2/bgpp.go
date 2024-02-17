// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package bgpv2

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strings"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	cilium_api_v2alpha1 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2alpha1"
	"github.com/cilium/cilium/pkg/k8s/resource"
	slim_core_v1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/api/core/v1"
	slim_meta_v1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
)

const (
	// bgpPPAdvertisementLabel is the label key used in BGP Advertisements.
	// This label is used in autogenerated BGP Advertisements from CiliumBGPPeeringPolicy.
	bgpPPAdvertisementLabel = "bgpPeeringPolicy.advertise"
)

// reconcileBGPPeeringPolicies updates various BGPv2 resource based on CiliumBGPPeeringPolicy.
// This method reconciles all CiliumBGPPeeringPolicy resources by fetching them from the local store.
func (b *BGPResourceManager) reconcileBGPPeeringPolicies(ctx context.Context) error {
	var err error
	for _, bgpp := range b.peeringPolicyStore.List() {
		rpErr := b.reconcileBGPPeeringPolicy(ctx, bgpp)
		if rpErr != nil {
			err = errors.Join(err, rpErr)
		}
	}
	return err
}

func (b *BGPResourceManager) reconcileBGPPeeringPolicy(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy) error {
	var err error

	// set defaults for the policy
	bgpp.SetDefaults()

	// update BGP advertisements from the policy
	raErr := b.reconcileBGPPAdvertisement(ctx, bgpp)
	if raErr != nil {
		err = errors.Join(err, raErr)
	}

	// update peer configuration from the policy
	rpErr := b.reconcileBGPPPeerConfig(ctx, bgpp)
	if rpErr != nil {
		err = errors.Join(err, rpErr)
	}

	// update node configuration
	rnErr := b.reconcileBGPPNodeConfig(ctx, bgpp)
	if rnErr != nil {
		err = errors.Join(err, rnErr)
	}

	return err
}

// reconcileBGPPAdvertisement updates BGP Advertisements based on CiliumBGPPeeringPolicy.
func (b *BGPResourceManager) reconcileBGPPAdvertisement(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy) error {
	var err error
	// expectedAdvertisements is a map of all expected BGP Advertisements for this policy.
	// This is used to make sure we cleanup any other stale BGP Advertisements which may have been previously created.
	expectedAdvertisements := make(map[string]struct{})

	for _, vr := range bgpp.Spec.VirtualRouters {
		// For each neighbor, advertisement is created.
		// Each neighbor can have different path attributes, so we need to create different advertisements for each neighbor.
		for _, neigh := range vr.Neighbors {
			var advertisements []cilium_api_v2alpha1.Advertisement

			// if export pod cidr is enabled, add pod cidr advertisement.
			// selector field is nil in this case.
			if *vr.ExportPodCIDR {
				advertisements = append(advertisements, cilium_api_v2alpha1.Advertisement{
					AdvertisementType: cilium_api_v2alpha1.PodCIDRAdvert,
					Selector:          nil,
					Attributes:        getAttributes(neigh, cilium_api_v2alpha1.PodCIDRSelectorName),
				})
			}

			// if service selector is not nil, add service selector advertisement.
			// selector field is copied over from vr.ServiceSelector.
			// attributes are taken from neigh.AdvertisedPathAttributes.
			if vr.ServiceSelector != nil {
				advertisements = append(advertisements, cilium_api_v2alpha1.Advertisement{
					AdvertisementType: cilium_api_v2alpha1.CiliumLoadBalancerIPAdvert,
					Selector:          vr.ServiceSelector,
					Attributes:        getAttributes(neigh, cilium_api_v2alpha1.CiliumLoadBalancerIPPoolSelectorName),
				})
			}

			// if pod ip pool selector is not nil, add pod ip pool advertisement.
			// selector field is copied over from vr.PodIPPoolSelector.
			// attributes are taken from neigh.AdvertisedPathAttributes.
			if vr.PodIPPoolSelector != nil {
				advertisements = append(advertisements, cilium_api_v2alpha1.Advertisement{
					AdvertisementType: cilium_api_v2alpha1.CiliumPodIPPoolAdvert,
					Selector:          vr.PodIPPoolSelector,
					Attributes:        getAttributes(neigh, cilium_api_v2alpha1.CiliumPodIPPoolSelectorName),
				})
			}

			// key for advertisement is generated from bgpp.Name, vr.LocalASN and neigh.PeerAddress.
			peerKey := peerKeyFromBGPP(bgpp.Name, vr.LocalASN, neigh.PeerAddress)
			expectedAdvertisements[peerKey] = struct{}{}

			uErr := b.updateAdvertisement(ctx, bgpp, peerKey, advertisements...)
			if uErr != nil {
				err = errors.Join(err, uErr)
			}
		}
	}

	// cleanup stale advertisements
	cleanupErr := b.deleteStaleAdvertisement(ctx, expectedAdvertisements, bgpp.Name)
	if cleanupErr != nil {
		err = errors.Join(err, cleanupErr)
	}

	return err
}

// deleteStaleAdvertisementBGPP deletes stale BGP advertisements which are not present in expectedAdverts list.
func (b *BGPResourceManager) deleteStaleAdvertisement(ctx context.Context, expectedAdverts map[string]struct{}, bgppName string) error {
	var err error
	for _, advert := range b.advertStore.List() {
		_, exists := expectedAdverts[advert.Name]
		if !exists && IsOwner(advert.GetOwnerReferences(), bgppName) {
			b.logger.Infof("Deleting BGP Advertisement %s", advert.Name)

			dErr := b.advertClient.Delete(ctx, advert.Name, meta_v1.DeleteOptions{})
			if dErr != nil && !k8s_errors.IsNotFound(dErr) {
				err = errors.Join(err, dErr)
			}
		}
	}
	return err
}

func (b *BGPResourceManager) updateAdvertisement(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy, name string, adverts ...cilium_api_v2alpha1.Advertisement) error {
	prev, exists, err := b.advertStore.GetByKey(resource.Key{Name: name})
	if err != nil {
		return err
	}

	// delete advertisement if there are no advertisements to be created.
	if len(adverts) == 0 {
		if !exists {
			return nil
		}

		b.logger.WithField("advertisement", name).Debug("Deleting BGP Advertisement")
		err = b.advertClient.Delete(ctx, name, meta_v1.DeleteOptions{})
		if err != nil && !k8s_errors.IsNotFound(err) {
			return err
		}
		return nil
	}

	advert := &cilium_api_v2alpha1.CiliumBGPAdvertisement{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
			// owner reference is set to the BGP Peering Policy.
			OwnerReferences: []meta_v1.OwnerReference{
				{
					APIVersion: slim_core_v1.SchemeGroupVersion.String(),
					Kind:       cilium_api_v2alpha1.BGPPKindDefinition,
					Name:       bgpp.GetName(),
					UID:        bgpp.GetUID(),
				},
			},
			// label is set to bgpPeeringPolicy.advertise=<name> to identify the advertisement.
			// Advertisements are created independently and are not tied to any other resource.
			// CiliumBGPPeerConfig resource will refer to advertisements using this label.
			Labels: map[string]string{
				bgpPPAdvertisementLabel: name,
			},
		},
		Spec: cilium_api_v2alpha1.CiliumBGPAdvertisementSpec{
			Advertisements: adverts,
		},
	}

	switch {
	case exists && prev.Spec.DeepEqual(&advert.Spec):
		return nil
	case exists:
		// reinitialize spec
		prev.Spec = advert.Spec
		_, err = b.advertClient.Update(ctx, prev, meta_v1.UpdateOptions{})
	default:
		_, err = b.advertClient.Create(ctx, advert, meta_v1.CreateOptions{})
		if err != nil && k8s_errors.IsAlreadyExists(err) {
			// local store is not yet updated, but API server has the resource. Get resource from API server
			// to compare spec.
			prev, err = b.advertClient.Get(ctx, advert.Name, meta_v1.GetOptions{})
			if err != nil {
				return err
			}

			// if prev already exist and spec is different, update it
			if !prev.Spec.DeepEqual(&advert.Spec) {
				prev.Spec = advert.Spec
				_, err = b.advertClient.Update(ctx, prev, meta_v1.UpdateOptions{})
			} else {
				return nil
			}
		}
	}

	b.logger.WithField("advertisement", name).Debug("Upserting BGP Advertisement")
	return err
}

// getAttributes returns CiliumBGPAttributes based on CiliumBGPNeighbor for the given selectorType.
func getAttributes(neigh cilium_api_v2alpha1.CiliumBGPNeighbor, selectorType string) *cilium_api_v2alpha1.CiliumBGPAttributes {
	for _, attr := range neigh.AdvertisedPathAttributes {
		if attr.SelectorType == selectorType {
			return &cilium_api_v2alpha1.CiliumBGPAttributes{
				Community:       attr.Communities,
				LocalPreference: attr.LocalPreference,
			}
		}
	}
	return nil
}

// reconcileBGPPPeerConfig updates BGP Peer Configs based on CiliumBGPPeeringPolicy. This method creates a BGP Peer Config per neighbor
// of the policy.
func (b *BGPResourceManager) reconcileBGPPPeerConfig(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy) error {
	var err error
	// expectedPeerConfig is a map of all expected BGP Peer Configs for this policy.
	expectedPeerConfig := make(map[string]struct{})

	for _, vr := range bgpp.Spec.VirtualRouters {
		for _, neigh := range vr.Neighbors {
			peerSpec := cilium_api_v2alpha1.CiliumBGPPeerConfigSpec{
				Transport: &cilium_api_v2alpha1.CiliumBGPTransport{
					PeerPort: ptr.To[int32](*neigh.PeerPort),
				},
				Timers: &cilium_api_v2alpha1.CiliumBGPTimers{
					ConnectRetryTimeSeconds: ptr.To[int32](*neigh.ConnectRetryTimeSeconds),
					HoldTimeSeconds:         ptr.To[int32](*neigh.HoldTimeSeconds),
					KeepAliveTimeSeconds:    ptr.To[int32](*neigh.KeepAliveTimeSeconds),
				},
				GracefulRestart: neigh.GracefulRestart,
				AuthSecretRef:   neigh.AuthSecretRef,
				EBGPMultihop:    ptr.To[int32](*neigh.EBGPMultihopTTL),
			}

			for _, fam := range neigh.Families {
				peerSpec.Families = append(peerSpec.Families, cilium_api_v2alpha1.CiliumBGPFamilyWithAdverts{
					CiliumBGPFamily: fam,
					// advertisement is configured per peer, so all address families get same advertisement.
					// advertisement is identified by bgpPeeringPolicy.advertise=<name> label.
					// name is unique per neighbor per virtual router in the policy.
					Advertisements: &slim_meta_v1.LabelSelector{
						MatchLabels: map[string]string{
							bgpPPAdvertisementLabel: peerKeyFromBGPP(bgpp.Name, vr.LocalASN, neigh.PeerAddress),
						},
					},
				})
			}

			peerConfigKey := peerKeyFromBGPP(bgpp.Name, vr.LocalASN, neigh.PeerAddress)
			expectedPeerConfig[peerConfigKey] = struct{}{}

			uErr := b.updatePeerConfig(ctx, bgpp, peerConfigKey, peerSpec)
			if uErr != nil {
				err = errors.Join(err, uErr)
			}
		}
	}

	// cleanup peer configs for this BGP peering policy which are not in expectedPeerConfig.
	dErr := b.deleteStalePeerConfig(ctx, expectedPeerConfig, bgpp.Name)
	if dErr != nil {
		err = errors.Join(err, dErr)
	}

	return err
}

// deleteStalePeerConfig deletes peer configs not present in expectedPeerConfigs for the given owner.
func (b *BGPResourceManager) deleteStalePeerConfig(ctx context.Context, expectedPeerConfigs map[string]struct{}, bgppName string) error {
	var err error
	for _, peerConfig := range b.peerConfigStore.List() {
		_, exists := expectedPeerConfigs[peerConfig.Name]
		if !exists && IsOwner(peerConfig.GetOwnerReferences(), bgppName) {
			b.logger.WithField("peer config", peerConfig.Name).Debug("Deleting BGP Peer Config")

			dErr := b.peerConfigClient.Delete(ctx, peerConfig.Name, meta_v1.DeleteOptions{})
			if dErr != nil && !k8s_errors.IsNotFound(dErr) {
				err = errors.Join(err, dErr)
			}
		}
	}
	return err
}

func (b *BGPResourceManager) updatePeerConfig(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy, name string, spec cilium_api_v2alpha1.CiliumBGPPeerConfigSpec) error {
	prev, exists, err := b.peerConfigStore.GetByKey(resource.Key{Name: name})
	if err != nil {
		return err
	}

	peerConfig := &cilium_api_v2alpha1.CiliumBGPPeerConfig{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
			// owner reference is set to the BGP Peering Policy.
			OwnerReferences: []meta_v1.OwnerReference{
				{
					APIVersion: slim_core_v1.SchemeGroupVersion.String(),
					Kind:       cilium_api_v2alpha1.BGPPKindDefinition,
					Name:       bgpp.GetName(),
					UID:        bgpp.GetUID(),
				},
			},
		},
		Spec: spec,
	}

	switch {
	case exists && prev.Spec.DeepEqual(&peerConfig.Spec):
		return nil
	case exists:
		// reinitialize spec
		prev.Spec = peerConfig.Spec
		_, err = b.peerConfigClient.Update(ctx, prev, meta_v1.UpdateOptions{})
	default:
		_, err = b.peerConfigClient.Create(ctx, peerConfig, meta_v1.CreateOptions{})
		if err != nil && k8s_errors.IsAlreadyExists(err) {
			// local store is not yet updated, but API server has the resource. Get resource from API server
			// to compare spec.
			prev, err = b.peerConfigClient.Get(ctx, peerConfig.Name, meta_v1.GetOptions{})
			if err != nil {
				return err
			}

			// if prev already exist and spec is different, update it
			if !prev.Spec.DeepEqual(&peerConfig.Spec) {
				prev.Spec = peerConfig.Spec
				_, err = b.peerConfigClient.Update(ctx, prev, meta_v1.UpdateOptions{})
			} else {
				return nil
			}
		}
	}

	b.logger.WithField("peer config", name).Debug("Upserting BGP Peer Config")
	return err
}

func (b *BGPResourceManager) reconcileBGPPNodeConfig(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy) error {
	matchingNodes, err := b.getMatchingNodes(bgpp.Spec.NodeSelector, bgpp.Name)
	if err != nil {
		return err
	}

	for nodeRef := range matchingNodes {
		uErr := b.upsertBGPPNodeConfig(ctx, bgpp, nodeRef)
		if uErr != nil {
			err = errors.Join(err, uErr)
		}
	}

	// delete node configs for this policy that are not in the matching nodes
	dErr := b.deleteStaleNodeConfigs(ctx, matchingNodes, bgpp.Name)
	if dErr != nil {
		err = errors.Join(err, dErr)
	}

	return err
}

func (b *BGPResourceManager) upsertBGPPNodeConfig(ctx context.Context, bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy, nodeRef string) error {
	prev, exists, err := b.nodeConfigStore.GetByKey(resource.Key{Name: nodeRef})
	if err != nil {
		return err
	}

	nodeConfig := &cilium_api_v2alpha1.CiliumBGPNodeConfig{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: nodeRef,
			OwnerReferences: []meta_v1.OwnerReference{
				{
					APIVersion: slim_core_v1.SchemeGroupVersion.String(),
					Kind:       cilium_api_v2alpha1.BGPPKindDefinition,
					Name:       bgpp.GetName(),
					UID:        bgpp.GetUID(),
				},
			},
		},
		Spec: cilium_api_v2alpha1.CiliumBGPNodeSpec{
			BGPInstances: nodeBGPInstancesFromBGPP(bgpp),
		},
	}

	switch {
	case exists && prev.Spec.DeepEqual(&nodeConfig.Spec):
		return nil
	case exists:
		prev.Spec = nodeConfig.Spec
		_, err = b.nodeConfigClient.Update(ctx, prev, meta_v1.UpdateOptions{})
	default:
		_, err = b.nodeConfigClient.Create(ctx, nodeConfig, meta_v1.CreateOptions{})
		if err != nil && k8s_errors.IsAlreadyExists(err) {
			// local store is not yet updated, but API server has the resource. Get resource from API server
			// to compare spec.
			prev, err = b.nodeConfigClient.Get(ctx, nodeConfig.Name, meta_v1.GetOptions{})
			if err != nil {
				return err
			}

			// if prev already exist and spec is different, update it
			if !prev.Spec.DeepEqual(&nodeConfig.Spec) {
				prev.Spec = nodeConfig.Spec
				_, err = b.nodeConfigClient.Update(ctx, prev, meta_v1.UpdateOptions{})
			} else {
				return nil
			}
		}
	}

	b.logger.WithField("node config", nodeRef).Debug("Upserting BGP Node Config")

	return err
}

// peerConfigRefFromBGPP returns a PeerConfigReference from the given policyName, localASN and peerAddr.
func peerConfigRefFromBGPP(policyName string, localASN int64, peerAddr string) *cilium_api_v2alpha1.PeerConfigReference {
	return &cilium_api_v2alpha1.PeerConfigReference{
		Group: cilium_api_v2alpha1.CustomResourceDefinitionGroup,
		Kind:  cilium_api_v2alpha1.BGPPCKindDefinition,
		Name:  peerKeyFromBGPP(policyName, localASN, peerAddr),
	}
}

// peerKeyFromBGPP returns a key for the given bgp policy, localASN and peerAddr.
// Note returned string is used as name in various kubernetes resources, so naming convention should be followed.
// Naming should consist of lower case alphanumeric characters, -, and .
func peerKeyFromBGPP(policyName string, localASN int64, peerAddr string) string {
	peerAddr = strings.ReplaceAll(peerAddr, "/", "-")
	peerAddr = strings.ReplaceAll(peerAddr, ":", ".")
	key := fmt.Sprintf("%s-%s", instanceKeyFromBGPP(policyName, localASN), peerAddr)
	return strings.ToLower(key)
}

func instanceKeyFromBGPP(policyName string, localASN int64) string {
	return fmt.Sprintf("%s-%d", policyName, localASN)
}

// nodeBGPInstancesFromBGPP returns a list of CiliumBGPNodeInstance from the given CiliumBGPPeeringPolicy.
func nodeBGPInstancesFromBGPP(bgpp *cilium_api_v2alpha1.CiliumBGPPeeringPolicy) []cilium_api_v2alpha1.CiliumBGPNodeInstance {
	var res []cilium_api_v2alpha1.CiliumBGPNodeInstance

	for _, vr := range bgpp.Spec.VirtualRouters {
		vr := vr
		var instance cilium_api_v2alpha1.CiliumBGPNodeInstance
		instance.Name = instanceKeyFromBGPP(bgpp.Name, vr.LocalASN)
		instance.LocalASN = &vr.LocalASN

		for _, neigh := range vr.Neighbors {
			neigh := neigh

			// BGP Peering Policy peer address is /32 or /128 CIDR. We transform it to IP address.
			// On failure, skip the neighbor.
			peerAddr, err := prefixToAddr(neigh.PeerAddress)
			if err != nil {
				continue
			}

			var peer cilium_api_v2alpha1.CiliumBGPNodePeer
			peer.Name = peerKeyFromBGPP(bgpp.Name, vr.LocalASN, neigh.PeerAddress)
			peer.PeerAddress = &peerAddr
			peer.PeerASN = &neigh.PeerASN
			peer.PeerConfigRef = peerConfigRefFromBGPP(bgpp.Name, vr.LocalASN, neigh.PeerAddress)

			instance.Peers = append(instance.Peers, peer)
		}
		res = append(res, instance)
	}
	return res
}

func prefixToAddr(cidr string) (string, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return "", err
	}

	return prefix.Addr().String(), nil
}
