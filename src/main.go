package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

var (
	kubeconfig string
)

type Controller struct {
	clientset   *kubernetes.Clientset
	queue       workqueue.RateLimitingInterface
	informer    cache.SharedIndexInformer
}

func NewController(clientset *kubernetes.Clientset, informer cache.SharedIndexInformer) *Controller {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	controller := &Controller{
		clientset: clientset,
		queue:     queue,
		informer:  informer,
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.enqueueDeployment,
		UpdateFunc: func(old, new interface{}) { controller.enqueueDeployment(new) },
	})

	return controller
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer c.queue.ShutDown()

	klog.Info("Starting Deployment controller")
	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		klog.Error("Failed to sync cache")
		return
	}

	klog.Info("Deployment controller synced and ready")

	go c.runWorker()

	<-stopCh
	klog.Info("Stopping Deployment controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncDeployment(key.(string))
	c.handleErr(err, key)

	return true
}

func (c *Controller) syncDeployment(key string) error {
	obj, exists, err := c.informer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	deployment := obj.(*appsv1.Deployment)
	fmt.Printf("Syncing Deployment %s\n", deployment.Name)

	return nil
}

func (c *Controller) enqueueDeployment(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		klog.Errorf("Failed to get key for object: %v", err)
		return
	}
	c.queue.Add(key)
}

func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing Deployment %v: %v", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	klog.Infof("Dropping Deployment %q out of the queue: %v", key, err)
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	stopCh := make(chan struct{})
	defer close(stopCh)

	cfg, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building clientset: %v", err)
	}

	deploymentInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return clientset.AppsV1().Deployments(metav1.NamespaceAll).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return clientset.AppsV1().Deployments(metav1.NamespaceAll).Watch(context.TODO(), options)
			},
		},
		&appsv1.Deployment{},
		0, // Skip resync
		cache.Indexers{},
	)

	controller := NewController(clientset, deploymentInformer)
	go controller.Run(stopCh)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
}
