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

const DiskResource = "Disk"

func init() {
	registry.Register(&registry.Registration{
		Name:     DiskResource,
		Scope:    nuke.ResourceGroup,
		Lister:   &DiskLister{},
		Resource: &Disk{},
		DependsOn: []string{
			VirtualMachineResource,
		},
	})
}

type Disk struct {
	client        compute.DisksClient
	Region        *string
	ResourceGroup *string
	Name          *string
	Tags          map[string]*string
}

func (r *Disk) Remove(ctx context.Context) error {
	_, err := r.client.Delete(ctx, *r.ResourceGroup, *r.Name)
	return err
}

func (r *Disk) Properties() types.Properties {
	return types.NewPropertiesFromStruct(r)
}

func (r *Disk) String() string {
	return *r.Name
}

type DiskLister struct {
}

func (l DiskLister) List(ctx context.Context, o interface{}) ([]resource.Resource, error) {
	opts := o.(*nuke.ListerOpts)

	log := logrus.WithField("r", DiskResource).WithField("s", opts.SubscriptionID)

	client := compute.NewDisksClient(opts.SubscriptionID)
	client.Authorizer = opts.Authorizers.Management
	client.RetryAttempts = 1
	client.RetryDuration = time.Second * 2

	resources := make([]resource.Resource, 0)

	log.Trace("attempting to list disks")

	list, err := client.ListByResourceGroup(ctx, opts.ResourceGroup)
	if err != nil {
		return nil, err
	}

	log.Trace("listing ....")

	for list.NotDone() {
		log.Trace("list not done")
		for _, g := range list.Values() {
			resources = append(resources, &Disk{
				client:        client,
				Region:        g.Location,
				ResourceGroup: &opts.ResourceGroup,
				Name:          g.Name,
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
