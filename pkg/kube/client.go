package kube

import (
	"context"
	"os"

	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha3.AddToScheme(scheme)
}

func NewClient() (client.Client, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return client.New(config, client.Options{
		Scheme: scheme,
	})
}

// CreateOrUpdate ...
func CreateOrUpdate(ctx context.Context, cr client.Client, obj client.Object) error {
	exist := obj.DeepCopyObject().(client.Object)

	err := cr.Get(context.TODO(), client.ObjectKeyFromObject(obj), exist)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = cr.Create(context.TODO(), obj)
			if apierrors.IsAlreadyExists(err) {
				return nil
			}
		}
		return err
	}
	return cr.Patch(ctx, obj, client.MergeFromWithOptions(exist, client.MergeFromWithOptimisticLock{}))
}
