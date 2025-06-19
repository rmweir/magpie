package sign

import (
	"context"
	"fmt"

	"github.com/loft-sh/magpie/pkg/signing"
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

	delete(obj.GetAnnotations(), signatureAnnotation)

	err = signing.Verify(obj, []byte(signature))
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

	obj.SetAnnotations(map[string]string{signatureAnnotation: string(signature)})

	err = c.Client.Create(ctx, obj, opts...)
	if err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}
	return nil
}

func (c Client) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	delete(obj.GetAnnotations(), signatureAnnotation)

	signature, err := signing.Sign(obj)
	if err != nil {
		return fmt.Errorf("failed to sign object: %w", err)
	}

	obj.SetAnnotations(map[string]string{signatureAnnotation: string(signature)})

	err = c.Client.Update(ctx, obj, opts...)
	if err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}
	return nil
}
