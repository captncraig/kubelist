package main

import (
	"log"
	"os"
	"time"

	"github.com/captncraig/kubelist"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// If KUBECONFIG is set we are running locally. Use that. Otherwise do in-cluster
	kc := os.Getenv("KUBECONFIG")
	var config *rest.Config
	var err error
	if kc != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kc)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		log.Fatal(err)
	}
	config.QPS = 100

	lister, err := kubelist.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	list, err := lister.ListAllResources(metav1.ListOptions{LabelSelector: ""}, true)
	if err != nil {
		log.Fatal(err)
	}
	end := time.Now()

	for _, obj := range list {
		log.Println(obj.GetKind(), obj.GetNamespace(), obj.GetName())
	}
	log.Printf("%d Total Objects in %s", len(list), end.Sub(start))
}
