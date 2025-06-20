package sign

import (
	"context"
	"fmt"

	"github.com/loft-sh/magpie/pkg/signing"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const signatureAnnotation = "magpie.loft.sh/signature"

var _ client.Client = (*Client)(nil)

type Client struct {
	client.Client
}

func NewClient(client client.Client) Client {
	return Client{Client: client}
}

func (c Client) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	err := c.Client.Get(ctx, key, obj, opts...)
	if err != nil {
		return err
	}

	signature := obj.GetAnnotations()[signatureAnnotation]
	revertValuesFunc := removeNonComparableValues(obj)
	defer revertValuesFunc()

	err = signing.Verify(obj, signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

func (c Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	signature, err := signing.Sign(obj)
	if err != nil {
		return fmt.Errorf("failed to sign object: %w", err)
	}

	obj.SetAnnotations(map[string]string{signatureAnnotation: signature})

	err = c.Client.Create(ctx, obj, opts...)
	if err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}
	return nil
}

func (c Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	revertValuesFunc := removeNonComparableValues(obj)

	signature, err := signing.Sign(obj)
	if err != nil {
		revertValuesFunc()
		return fmt.Errorf("failed to sign object: %w", err)
	}

	revertValuesFunc()
	obj.SetAnnotations(map[string]string{signatureAnnotation: string(signature)})

	err = c.Client.Update(ctx, obj, opts...)
	if err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}
	return nil
}

func removeNonComparableValues(obj client.Object) func() {
	signature := obj.GetAnnotations()[signatureAnnotation]
	resourceVersion := obj.GetResourceVersion()
	creationTimestamp := obj.GetCreationTimestamp()
	uid := obj.GetUID()
	gvk := obj.GetObjectKind().GroupVersionKind()
	managedFields := obj.GetManagedFields()

	delete(obj.GetAnnotations(), signatureAnnotation)
	obj.SetResourceVersion("")
	obj.SetCreationTimestamp(v1.Time{})
	obj.SetUID("")
	obj.SetManagedFields(nil)

	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		// with normal types (any type that is not unstructured), setting the GVK is optional. Usually created will be
		// missing it and updates will contain it. We just remove it here as it is not needed.
		obj.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{})
	}

	return func() {
		obj.SetAnnotations(map[string]string{signatureAnnotation: signature})
		obj.SetResourceVersion(resourceVersion)
		obj.SetCreationTimestamp(creationTimestamp)
		obj.SetUID(uid)
		obj.SetManagedFields(managedFields)
		obj.GetObjectKind().SetGroupVersionKind(gvk)
	}
}
