package server

import (
	"context"
	"net/http"
	"slices"

	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/fake"
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

// newResourceGroupsServer creates a fake server for the armresources.ResourceGroupsClient.
func newResourceGroupsServer(f *fakeResourceCalls) fake.ResourceGroupsServer {
	return fake.ResourceGroupsServer{
		BeginDelete:    f.BeginDelete,
		CreateOrUpdate: f.CreateOrUpdate,
		Get:            f.Get,
		Update:         f.Update,
		NewListPager:   f.NewListPager,
	}
}

// fakeResourceCalls is used to build the responses for an armresources.ResourceGroupsClient that has
// a fake server attached to it. You specify the responses that you wish to receive in the order that they will return.
// For beginDelete and list, these are the set of responses that will be returned by the poller or the pager.
// If beginDeleteErr is set, then the call to BeginDelete will return an immediate error. For both pooller and pager,
// if the last entry is an error, it is a terminal error. Otherwise it is transient.
// Use newResourceGroupsServer() to create a fake server from this.
type fakeResourceCalls struct {
	createOrUpdate []any //  armresources.ResourceGroupsClientCreateOrUpdateResponse or error
	beginDelete    []any // armresources.ResourceGroupsClientDeleteResponse or error
	beginDeleteErr error
	get            []any // armresources.ResourceGroupsClientGetResponse or error
	list           []any // armresources.ResourceGroupsClientListResponse or error
	update         []any // armresources.ResourceGroupsClientUpdateResponse or error
}

func (f *fakeResourceCalls) CreateOrUpdate(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (resp azfake.Responder[armresources.ResourceGroupsClientCreateOrUpdateResponse], errResp azfake.ErrorResponder) {
	if len(f.createOrUpdate) == 0 {
		panic("unexpected call")
	}
	defer func() {
		f.createOrUpdate = slices.Delete(f.createOrUpdate, 0, 1)
	}()
	switch t := f.createOrUpdate[0].(type) {
	case armresources.ResourceGroupsClientCreateOrUpdateResponse:
		resp.SetResponse(http.StatusOK, t, nil)
		return resp, azfake.ErrorResponder{}
	case error:
		errResp.SetError(t)
		return azfake.Responder[armresources.ResourceGroupsClientCreateOrUpdateResponse]{}, errResp
	}
	panic("unknown return type")
}

func (f *fakeResourceCalls) BeginDelete(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientBeginDeleteOptions) (azfake.PollerResponder[armresources.ResourceGroupsClientDeleteResponse], azfake.ErrorResponder) {
	if f.beginDeleteErr != nil {
		e := azfake.ErrorResponder{}
		e.SetError(f.beginDeleteErr)
		return azfake.PollerResponder[armresources.ResourceGroupsClientDeleteResponse]{}, e
	}

	poller := azfake.PollerResponder[armresources.ResourceGroupsClientDeleteResponse]{}
	for i, r := range f.beginDelete {
		switch t := r.(type) {
		case armresources.ResourceGroupsClientDeleteResponse:
			if i == len(f.beginDelete)-1 {
				poller.SetTerminalResponse(http.StatusOK, t, nil)
				continue
			}
			poller.AddNonTerminalResponse(http.StatusAccepted, nil)
		case error:
			poller.SetTerminalError(http.StatusInternalServerError, t.Error())
		default:
			panic("unknown return type")
		}
	}
	return poller, azfake.ErrorResponder{}
}

func (f *fakeResourceCalls) Get(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (resp azfake.Responder[armresources.ResourceGroupsClientGetResponse], errResp azfake.ErrorResponder) {
	if len(f.get) == 0 {
		panic("unexpected call")
	}

	defer func() {
		f.get = slices.Delete(f.get, 0, 1)
	}()

	switch t := f.get[0].(type) {
	case armresources.ResourceGroupsClientGetResponse:
		resp.SetResponse(http.StatusOK, t, nil)
		return resp, azfake.ErrorResponder{}
	case error:
		errResp.SetError(t)
		return azfake.Responder[armresources.ResourceGroupsClientGetResponse]{}, errResp
	}
	panic("unknown return type")
}

func (f *fakeResourceCalls) Update(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroupPatchable, options *armresources.ResourceGroupsClientUpdateOptions) (resp azfake.Responder[armresources.ResourceGroupsClientUpdateResponse], errResp azfake.ErrorResponder) {
	if len(f.update) == 0 {
		panic("unexpected call")
	}

	defer func() {
		f.update = slices.Delete(f.update, 0, 1)
	}()

	switch t := f.update[0].(type) {
	case armresources.ResourceGroupsClientUpdateResponse:
		resp.SetResponse(http.StatusOK, t, nil)
		return resp, azfake.ErrorResponder{}
	case error:
		errResp.SetError(t)
		return azfake.Responder[armresources.ResourceGroupsClientUpdateResponse]{}, errResp
	}
	panic("unknown return type")
}

func (f *fakeResourceCalls) NewListPager(options *armresources.ResourceGroupsClientListOptions) (resp azfake.PagerResponder[armresources.ResourceGroupsClientListResponse]) {
	pager := azfake.PagerResponder[armresources.ResourceGroupsClientListResponse]{}

	for i, item := range f.list {
		switch t := item.(type) {
		case armresources.ResourceGroupsClientListResponse:
			pager.AddPage(http.StatusOK, t, nil)
		case error:
			if i == len(f.list)-1 {
				pager.AddError(t)
				continue
			}
			pager.AddResponseError(http.StatusRequestTimeout, t.Error())
		default:
			panic("unknown return type")
		}
	}
	return pager
}
