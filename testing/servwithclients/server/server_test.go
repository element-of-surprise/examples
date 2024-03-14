package server

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	azfake "github.com/Azure/azure-sdk-for-go/sdk/azcore/fake"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/fake"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/testing/protocmp"

	gpb "github.com/element-of-surprise/examples/testing/servwithclients/proto/greeter/proto"
	pb "github.com/element-of-surprise/examples/testing/servwithclients/server/proto"
)

func TestSayHello(t *testing.T) {
	t.Parallel()

	unavailable := status.Error(codes.Unavailable, "unavailable")

	req := &gpb.HelloRequest{
		Name: "Bob",
		Age:  53,
		Address: &gpb.Address{
			Street:  "123 Main St",
			City:    "Seattle",
			State:   "WA",
			Zipcode: 98012,
		},
	}
	resp := &gpb.HelloReply{Message: "Hello Bob"}

	tests := []struct {
		name    string
		req     *gpb.HelloRequest
		greeter gpb.GreeterClient
		wantErr bool
		want    *gpb.HelloReply
	}{
		{
			name:    "Success",
			req:     req,
			greeter: &fakeGreeter{shResps: []any{resp}},
			want:    resp,
		},
		{
			name:    "Success: 1 retry needed",
			req:     req,
			greeter: &fakeGreeter{shResps: []any{unavailable, resp}},
			want:    resp,
		},
		{
			name:    "Error: too many retries",
			req:     req,
			greeter: &fakeGreeter{shResps: []any{unavailable, unavailable, unavailable, resp}},
			wantErr: true,
		},
		{
			name:    "Error: should not be retried because it is not a retriable error",
			req:     req,
			greeter: &fakeGreeter{shResps: []any{fmt.Errorf("error"), resp}},
			wantErr: true,
		},
	}

	for _, test := range tests {
		s := &Server{greeterClient: test.greeter}
		got, err := s.SayHello(context.Background(), test.req)
		switch {
		case err == nil && test.wantErr:
			t.Errorf("TestSayHello(%s): got err == nil, want err != nil", test.name)
			continue
		case err != nil && !test.wantErr:
			t.Errorf("TestSayHello(%s): got err == %s, want err == nil", test.name, err)
			continue
		case err != nil:
			continue
		}
		if diff := cmp.Diff(test.want, got, protocmp.Transform()); diff != "" {
			t.Errorf("TestSayHello(%s): -want/+got:\n%s", test.name, diff)
		}
	}
}

func mustFakeResourceGroupClient(calls *fakeResourceCalls) resourceClient {
	fs := newResourceGroupsServer(calls)
	client, err := armresources.NewResourceGroupsClient(
		"subscriptionID",
		&azfake.TokenCredential{},
		&arm.ClientOptions{
			ClientOptions: azcore.ClientOptions{
				Transport: fake.NewResourceGroupsServerTransport(&fs),
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return client
}

func TestCreateResourceGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fakeCalls *fakeResourceCalls
		client    resourceClient
		wantErr   bool
	}{
		{
			name:      "Error: client returned an error",
			fakeCalls: &fakeResourceCalls{createOrUpdate: []any{errors.New("error")}},
			wantErr:   true,
		},
		{
			name: "Success",
			fakeCalls: &fakeResourceCalls{
				createOrUpdate: []any{
					armresources.ResourceGroupsClientCreateOrUpdateResponse{
						ResourceGroup: armresources.ResourceGroup{ID: toPtr("id"), Name: toPtr("name"), Location: toPtr("westus")},
					},
				},
			},
		},
	}

	for _, test := range tests {
		fakeClient := mustFakeResourceGroupClient(test.fakeCalls)
		s := &Server{resourceClient: fakeClient}
		_, err := s.CreateResourceGroup(context.Background(), &pb.CreateResourceGroupRequest{Name: "name"})
		switch {
		case err == nil && test.wantErr:
			t.Errorf("TestCreateResourceGroup(%s): got err == nil, want err != nil", test.name)
			continue
		case err != nil && !test.wantErr:
			t.Errorf("TestCreateResourceGroup(%s): got err == %s, want err == nil", test.name, err)
			continue
		}
	}
}

func TestDeleteResourceGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fakeCalls *fakeResourceCalls
		client    resourceClient
		wantErr   bool
	}{
		{
			name:      "Error: client returned an error",
			fakeCalls: &fakeResourceCalls{beginDeleteErr: errors.New("error")},
			wantErr:   true,
		},
		{
			name: "Error: polling error",
			fakeCalls: &fakeResourceCalls{
				beginDelete: []any{
					armresources.ResourceGroupsClientDeleteResponse{},
					fmt.Errorf("error"),
				},
			},
			wantErr: true,
		},
		{
			name: "Success",
			fakeCalls: &fakeResourceCalls{
				beginDelete: []any{
					armresources.ResourceGroupsClientDeleteResponse{},
					armresources.ResourceGroupsClientDeleteResponse{},
				},
			},
		},
	}

	for _, test := range tests {
		fakeClient := mustFakeResourceGroupClient(test.fakeCalls)
		s := &Server{resourceClient: fakeClient}
		_, err := s.DeleteResourceGroup(context.Background(), &pb.DeleteResourceGroupRequest{Id: "id"})
		switch {
		case err == nil && test.wantErr:
			t.Errorf("TestCreateResourceGroup(%s): got err == nil, want err != nil", test.name)
			continue
		case err != nil && !test.wantErr:
			t.Errorf("TestCreateResourceGroup(%s): got err == %s, want err == nil", test.name, err)
			continue
		}
	}
}

// toPtr will make any value of T become *T. If T is already a pointer, it will return a pointer to the pointer.
func toPtr[T any](v T) *T {
	return &v
}
