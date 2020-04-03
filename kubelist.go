package kubelist

import (
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type Lister interface {
	ListAllGroupVersionResources() ([]*Resoure, error)
	ListAllResources(opts metav1.ListOptions, includeControlled bool) ([]unstructured.Unstructured, error)
}

type Resoure struct {
	Kind     schema.GroupVersionResource
	Resource metav1.APIResource
}

func (r *Resoure) HasListVerb() bool {
	for _, v := range r.Resource.Verbs {
		if v == "list" {
			return true
		}
	}
	return false
}

type lister struct {
	discoveryClient *discovery.DiscoveryClient
	dynamicClient   dynamic.Interface
}

func NewForConfig(c *rest.Config) (Lister, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(c)
	if err != nil {
		return nil, err
	}
	dyn, err := dynamic.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	return &lister{
		discoveryClient: disc,
		dynamicClient:   dyn,
	}, nil
}

func (l *lister) ListAllGroupVersionResources() ([]*Resoure, error) {
	all := []*Resoure{}
	start := time.Now()
	_, resources, err := l.discoveryClient.ServerGroupsAndResources()
	log.Printf("ServerGroupsAndResources: %s", time.Now().Sub(start))
	if err != nil {
		return nil, err
	}
	for _, group := range resources {
		for _, resource := range group.APIResources {
			gv, err := schema.ParseGroupVersion(group.GroupVersion)
			if err != nil {
				return nil, err
			}
			if resource.Group != "" {
				gv.Group = resource.Group
			}
			if resource.Version != "" {
				gv.Version = resource.Version
			}
			gvr := gv.WithResource(resource.Name)
			all = append(all, &Resoure{Kind: gvr, Resource: resource})
		}
	}
	return all, nil
}

func (l *lister) ListAllResources(opts metav1.ListOptions, includeControlled bool) ([]unstructured.Unstructured, error) {
	all := []unstructured.Unstructured{}
	resources, err := l.ListAllGroupVersionResources()
	if err != nil {
		return nil, err
	}
	for _, r := range resources {
		if !r.HasListVerb() {
			continue
		}
		list, err := l.getObjectsOfType(r.Kind, opts)
		if err != nil {
			log.Println(err)
			continue
		}
		if list != nil {
			all = append(all, list...)
		}
	}
	if includeControlled {
		return all, nil
	}
	// filter out anything with an owner reference
	standalone := make([]unstructured.Unstructured, 0, len(all))
	for _, u := range all {
		if len(u.GetOwnerReferences()) == 0 {
			standalone = append(standalone, u)
		}
	}
	return standalone, nil
}

func (l *lister) getObjectsOfType(gvr schema.GroupVersionResource, opts metav1.ListOptions) ([]unstructured.Unstructured, error) {
	client := l.dynamicClient.Resource(gvr)
	start := time.Now()
	list, err := client.List(opts)
	log.Printf("List %s: %s", gvr, time.Now().Sub(start))
	if err != nil {
		return nil, err
	}
	if list != nil {
		return list.Items, nil
	}
	return nil, nil
}
