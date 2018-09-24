package cmd

import (
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Resources struct {
	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface
}

type UnstructuredResourceNsItemsPair struct {
	Items []unstructured.Unstructured
	schema.GroupVersionResource
	Namespace string
}

type UnstructuredResourceNsPair struct {
	Item unstructured.Unstructured
	schema.GroupVersionResource
	Namespace string
}

type ResourcesAllOpts struct {
	IncludeNonNamespaced bool
	IncludeAllNamespaces bool
	IncludeNamespaces    []string
}

func NewResources(coreClient kubernetes.Interface, dynamicClient dynamic.Interface) Resources {
	return Resources{coreClient, dynamicClient}
}

func (c Resources) All(resTypePairs []GroupVersionResourceNsPair, opts ResourcesAllOpts) ([]UnstructuredResourceNsPair, error) {
	namespaces, err := c.allNamespaces()
	if err != nil {
		return nil, err
	}

	items := make(chan UnstructuredResourceNsItemsPair, len(resTypePairs)*len(namespaces))
	var itemsDone sync.WaitGroup

	for _, thing := range resTypePairs {
		thing := thing
		errStr := fmt.Sprintf("%#v, namespaced: %t", thing.GroupVersionResource, thing.Namespaced())

		if thing.Namespaced() {
			for _, ns := range namespaces {
				ns := ns
				if opts.IncludeAllNamespaces || opts.IncludesNamespace(ns) {
					itemsDone.Add(1)
					go func() {
						list2, err := c.dynamicClient.Resource(thing.GroupVersionResource).Namespace(ns).List(metav1.ListOptions{})
						if err != nil {
							fmt.Printf("%s: %s\n", errStr, err) // todo
						} else {
							items <- UnstructuredResourceNsItemsPair{list2.Items, thing.GroupVersionResource, ns}
						}
						itemsDone.Done()
					}()
				}
			}
		} else {
			if opts.IncludeNonNamespaced {
				itemsDone.Add(1)
				go func() {
					list2, err := c.dynamicClient.Resource(thing.GroupVersionResource).List(metav1.ListOptions{})
					if err != nil {
						fmt.Printf("%s: %s\n", errStr, err) // todo
					} else {
						items <- UnstructuredResourceNsItemsPair{list2.Items, thing.GroupVersionResource, ""}
					}
					itemsDone.Done()
				}()
			}
		}
	}

	itemsDone.Wait()

	close(items)

	var resPairs []UnstructuredResourceNsPair

	for itemNs := range items {
		for _, item := range itemNs.Items {
			resPairs = append(resPairs, UnstructuredResourceNsPair{item, itemNs.GroupVersionResource, itemNs.Namespace})
		}
	}

	return resPairs, nil
}

func (c Resources) Get(ref ObjectRef, ns string) (UnstructuredResourceNsPair, error) {
	item, err := c.dynamicClient.Resource(ref.GroupVersionResource).Namespace(ns).Get(ref.Name, metav1.GetOptions{})
	if err != nil {
		return UnstructuredResourceNsPair{}, err
	}

	// todo item pointer

	return UnstructuredResourceNsPair{*item, ref.GroupVersionResource, ns}, nil
}

func (c Resources) Delete(ref ObjectRef, ns string) error {
	return c.dynamicClient.Resource(ref.GroupVersionResource).Namespace(ns).Delete(ref.Name, &metav1.DeleteOptions{})
}

func (c Resources) allNamespaces() ([]string, error) {
	result, err := c.coreClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var names []string

	for _, ns := range result.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

func (o ResourcesAllOpts) IncludesNamespace(ns string) bool {
	for _, a := range o.IncludeNamespaces {
		if a == ns {
			return true
		}
	}
	return false
}
