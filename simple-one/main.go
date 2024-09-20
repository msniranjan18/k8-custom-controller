package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Step 1: Set up the kubeconfig path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")

	// Step 2: Load the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		panic(err.Error())
	}

	// Step 3: Create a clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Step 4: Create an informer factory (watches Pods by default in all namespaces)
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Minute*10)
	podInformer := informerFactory.Core().V1().Pods().Informer()

	// Step 5: Add event handlers for Pod lifecycle events (Add, Update, Delete)
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// Add Event
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("New Pod added: %s/%s\n", pod.Namespace, pod.Name)
		},

		// Update Event
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)
			fmt.Printf("Pod updated: %s/%s (Old Version: %s, New Version: %s)\n",
				newPod.Namespace, newPod.Name, oldPod.ResourceVersion, newPod.ResourceVersion)
		},

		// Delete Event
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			fmt.Printf("Pod deleted: %s/%s\n", pod.Namespace, pod.Name)
		},
	})

	// Step 6: Start the informer
	stopCh := make(chan struct{})
	defer close(stopCh)

	// Start the informer factory
	informerFactory.Start(stopCh)

	// Wait until the informer cache is synced
	informerFactory.WaitForCacheSync(stopCh)

	// Let the controller run indefinitely
	<-context.Background().Done()
}
