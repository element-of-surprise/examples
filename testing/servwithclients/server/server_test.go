package server

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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

func TestCreateResourceGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		client  resourceClient
		wantErr bool
	}{
		{
			name:    "Error: client returned an error",
			client:  &fakeResourceClient{createOrUpdate: []any{errors.New("error")}},
			wantErr: true,
		},
		{
			name:   "Success",
			client: &fakeResourceClient{createOrUpdate: []any{armresources.ResourceGroupsClientCreateOrUpdateResponse{}}},
		},
	}

	for _, test := range tests {
		s := &Server{resourceClient: test.client}
		_, err := s.CreateResourceGroup(context.Background(), &pb.CreateResourceGroupRequest{})
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
