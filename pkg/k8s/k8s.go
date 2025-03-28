package k8s

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

var dynamicClient *dynamic.DynamicClient

var ofipResource = schema.GroupVersionResource{
	Group:    "kubeovn.io",
	Version:  "v1",
	Resource: "ovn-fips",
}

func CreateDynamicClient() *dynamic.DynamicClient {
	// singleton
	if dynamicClient != nil {
		return dynamicClient
	}

	config, err := createConfig()
	if err != nil {
		log.Fatalf("Error creating config: %v", err)
	}

	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating dynamicClient: %v", err)
	}
	return dynamicClient
}

func createConfig() (*rest.Config, error) {
	configFile := filepath.Join(homedir.HomeDir(), ".kube", "config")
	_, err := os.Stat(configFile)
	if err != nil {
		return rest.InClusterConfig()
	}
	return clientcmd.BuildConfigFromFlags("", configFile)
}

func GetOFIPByIPv4(ipv4 string) (string, error) {
	clientset := CreateDynamicClient()
	ofip, err := clientset.Resource(ofipResource).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "ovn.kubernetes.io/eip_v4_ip=" + ipv4,
	})
	if err != nil {
		return "", err
	}
	if len(ofip.Items) == 0 {
		return "", fmt.Errorf("no ofip found for ipv4: %s", ipv4)
	}
	return ofip.Items[0].GetName(), nil
}
