package server

import (
	"context"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"google.golang.org/grpc"

	gpb "github.com/element-of-surprise/examples/testing/servwithclients/proto/greeter/proto"
)

type fakeGreeter struct {
	shResps []any // *gpb.HelloReply or error
}

func (f *fakeGreeter) SayHello(ctx context.Context, req *gpb.HelloRequest, opts ...grpc.CallOption) (*gpb.HelloReply, error) {
	if len(f.shResps) == 0 {
		panic("unexpected call")
	}
	defer func() {
		f.shResps = slices.Delete(f.shResps, 0, 1)
	}()
	switch t := f.shResps[0].(type) {
	case *gpb.HelloReply:
		return t, nil
	case error:
		return nil, t
	}
	panic("unknown return type")
}

type fakeResourceClient struct {
	createOrUpdate []any // armresources.ResourceGroupsClientCreateOrUpdateResponse or error
	beginDelete    []any // *runtime.Poller[ResourceGroupsClientDeleteResponse] or error
	get            []any // armresources.ResourceGroupsClientGetResponse or error
	list           []*runtime.Pager[armresources.ResourceGroupsClientListResponse]
	update         []any // armresources.ResourceGroupsClientUpdateResponse or error
}

func (f *fakeResourceClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error) {
	defer func() {
		f.createOrUpdate = slices.Delete(f.createOrUpdate, 0, 1)
	}()

	switch t := f.createOrUpdate[0].(type) {
	case armresources.ResourceGroupsClientCreateOrUpdateResponse:
		return t, nil
	case error:
		return armresources.ResourceGroupsClientCreateOrUpdateResponse{}, t
	}
	panic("unknown return type")
}
func (f *fakeResourceClient) BeginDelete(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error) {
	defer func() {
		f.beginDelete = slices.Delete(f.beginDelete, 0, 1)
	}()
	switch t := f.beginDelete[0].(type) {
	case *runtime.Poller[armresources.ResourceGroupsClientDeleteResponse]:
		return t, nil
	case error:
		return nil, t
	}
	panic("unknown return type")
}
func (f *fakeResourceClient) Get(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (armresources.ResourceGroupsClientGetResponse, error) {
	defer func() {
		f.get = slices.Delete(f.get, 0, 1)
	}()
	switch t := f.get[0].(type) {
	case armresources.ResourceGroupsClientGetResponse:
		return t, nil
	case error:
		return armresources.ResourceGroupsClientGetResponse{}, t
	}
	panic("unknown return type")
}
func (f *fakeResourceClient) NewListPager(options *armresources.ResourceGroupsClientListOptions) *runtime.Pager[armresources.ResourceGroupsClientListResponse] {
	defer func() {
		f.list = slices.Delete(f.list, 0, 1)
	}()
	return f.list[0]
}

func (f *fakeResourceClient) Update(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroupPatchable, options *armresources.ResourceGroupsClientUpdateOptions) (armresources.ResourceGroupsClientUpdateResponse, error) {
	defer func() {
		f.update = slices.Delete(f.update, 0, 1)
	}()
	switch t := f.update[0].(type) {
	case armresources.ResourceGroupsClientUpdateResponse:
		return t, nil
	case error:
		return armresources.ResourceGroupsClientUpdateResponse{}, t
	}
	panic("unknown return type")
}
