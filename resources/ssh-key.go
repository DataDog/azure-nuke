package resources

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2021-04-01/compute" //nolint:staticcheck

	"github.com/ekristen/libnuke/pkg/registry"
	"github.com/ekristen/libnuke/pkg/resource"
	"github.com/ekristen/libnuke/pkg/types"

	"github.com/ekristen/azure-nuke/pkg/nuke"
)

const SSHPublicKeyResource = "SSHPublicKey"

func init() {
	registry.Register(&registry.Registration{
		Name:   SSHPublicKeyResource,
		Lister: &SSHPublicKeyLister{},
		Scope:  nuke.Subscription,
	})
}

type SSHPublicKey struct {
	client compute.SSHPublicKeysClient

	Region         *string
	SubscriptionID *string
	ResourceGroup  *string
	Name           *string
	Tags           map[string]*string
}

func (r *SSHPublicKey) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *SSHPublicKey) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *SSHPublicKey) String() string {
	return *r.Name
}

// --------------------------------------

type SSHPublicKeyLister struct {
}

func (l SSHPublicKeyLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", SSHPublicKeyResource).WithField("s", opts.SubscriptionID)

	client := compute.NewSSHPublicKeysClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list ssh key")

	list, err := client.ListBySubscription(ctx)
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &SSHPublicKey{
				client:        client,
				Name:          g.Name,
				ResourceGroup: nuke.GetResourceGroupFromID(*g.ID),
				Region:        g.Location,
				Tags:          g.Tags,
			})
		}

		if err := list.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}

	log.Trace("done")

	return resources, nil
}
