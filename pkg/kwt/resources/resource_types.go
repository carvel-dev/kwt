package cmd

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

type ResourceTypes struct {
	coreClient kubernetes.Interface
}

type GroupVersionResourceNsPair struct {
	schema.GroupVersionResource
	metav1.APIResource
}

func NewResourceTypes(coreClient kubernetes.Interface) ResourceTypes {
	return ResourceTypes{coreClient}
}

func (g ResourceTypes) All() ([]GroupVersionResourceNsPair, error) {
	serverResources, err := g.coreClient.Discovery().ServerResources()
	if err != nil {
		return nil, err
	}

	var pairs []GroupVersionResourceNsPair

	for _, resList := range serverResources {
		groupVersion, err := schema.ParseGroupVersion(resList.GroupVersion)
		if err != nil {
			return nil, err
		}

		for _, res := range resList.APIResources {
			group := groupVersion.Group
			if len(res.Group) > 0 {
				group = res.Group
			}

			version := groupVersion.Version
			if len(res.Version) > 0 {
				version = res.Version
			}

			gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: res.Name}
			pairs = append(pairs, GroupVersionResourceNsPair{gvr, res})
		}
	}

	return pairs, nil
}

func (p GroupVersionResourceNsPair) Namespaced() bool {
	return p.APIResource.Namespaced
}

func (p GroupVersionResourceNsPair) Listable() bool {
	return p.containsStr(p.APIResource.Verbs, "list")
}

func (p GroupVersionResourceNsPair) containsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Listable(in []GroupVersionResourceNsPair) []GroupVersionResourceNsPair {
	var out []GroupVersionResourceNsPair
	for _, item := range in {
		if item.Listable() {
			out = append(out, item)
		}
	}
	return out
}

func Matching(in []GroupVersionResourceNsPair, ref ResourceRef) []GroupVersionResourceNsPair {
	partResourceRef := PartialResourceRef{ref.GroupVersionResource}
	var out []GroupVersionResourceNsPair
	for _, item := range in {
		if partResourceRef.Matches(item.GroupVersionResource) {
			out = append(out, item)
		}
	}
	return out
}

func NonMatching(in []GroupVersionResourceNsPair, ref ResourceRef) []GroupVersionResourceNsPair {
	partResourceRef := PartialResourceRef{ref.GroupVersionResource}
	var out []GroupVersionResourceNsPair
	for _, item := range in {
		if !partResourceRef.Matches(item.GroupVersionResource) {
			out = append(out, item)
		}
	}
	return out
}
