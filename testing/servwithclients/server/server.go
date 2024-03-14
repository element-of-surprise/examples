package server

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gpb "github.com/element-of-surprise/examples/testing/servwithclients/proto/greeter/proto"
	pb "github.com/element-of-surprise/examples/testing/servwithclients/server/proto"
)

// resourceClient represents a client for the Azure Resource Manager API.
// Usually implemented with github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources.ResourceGroupsClient{}.
type resourceClient interface {
	CreateOrUpdate(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroup, options *armresources.ResourceGroupsClientCreateOrUpdateOptions) (armresources.ResourceGroupsClientCreateOrUpdateResponse, error)
	BeginDelete(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientBeginDeleteOptions) (*runtime.Poller[armresources.ResourceGroupsClientDeleteResponse], error)
	Get(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientGetOptions) (armresources.ResourceGroupsClientGetResponse, error)
	NewListPager(options *armresources.ResourceGroupsClientListOptions) *runtime.Pager[armresources.ResourceGroupsClientListResponse]
	Update(ctx context.Context, resourceGroupName string, parameters armresources.ResourceGroupPatchable, options *armresources.ResourceGroupsClientUpdateOptions) (armresources.ResourceGroupsClientUpdateResponse, error)
}

// Server implements a gRPC server that acts as a proxy for the greeter service and the Azure Resource Manager.
type Server struct {
	pb.UnimplementedRPCServer

	greeterClient  gpb.GreeterClient
	resourceClient resourceClient
}

// New is the constructore for Server.
func New(greeter gpb.GreeterClient, resources resourceClient) (*Server, error) {
	if greeter == nil {
		return nil, errors.New("greeter is required")
	}
	if resources == nil {
		return nil, errors.New("resources is required")
	}
	return &Server{
		greeterClient:  greeter,
		resourceClient: resources,
	}, nil
}

// SayHello implements gpb.GreeterClient.SayHello().
func (s *Server) SayHello(ctx context.Context, in *gpb.HelloRequest) (*gpb.HelloReply, error) {
	const maxTries = 3

	var resp *gpb.HelloReply
	var err error
	for i := 0; i < maxTries; i++ {
		resp, err = s.greeterClient.SayHello(ctx, in)
		if err == nil {
			return resp, nil
		}

		switch status.Code(err) {
		case codes.Unavailable:
			continue
		default:
			return nil, err
		}
	}
	return nil, err
}

func (s *Server) CreateResourceGroup(ctx context.Context, in *pb.CreateResourceGroupRequest) (*pb.CreateResourceGroupReply, error) {
	_, err := s.resourceClient.CreateOrUpdate(ctx, in.GetName(), armresources.ResourceGroup{Location: &in.Region}, nil)
	if err != nil {
		return nil, err
	}

	return &pb.CreateResourceGroupReply{Status: "Success"}, nil
}

func (s *Server) ReadResourceGroup(ctx context.Context, in *pb.ReadResourceGroupRequest) (*pb.ReadResourceGroupReply, error) {
	_, err := s.resourceClient.Get(ctx, in.GetId(), nil)
	if err != nil {
		return nil, err
	}

	return &pb.ReadResourceGroupReply{Status: "Success"}, nil
}

func (s *Server) UpdateResourceGroup(ctx context.Context, in *pb.UpdateResourceGroupRequest) (*pb.UpdateResourceGroupReply, error) {
	_, err := s.resourceClient.Update(ctx, in.GetName(), armresources.ResourceGroupPatchable{ManagedBy: &in.Id}, nil)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateResourceGroupReply{Status: "Success"}, nil
}

func (s *Server) DeleteResourceGroup(ctx context.Context, in *pb.DeleteResourceGroupRequest) (*pb.DeleteResourceGroupReply, error) {
	poll, err := s.resourceClient.BeginDelete(ctx, in.GetId(), nil)
	if err != nil {
		return nil, err
	}

	_, err = poll.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteResourceGroupReply{Status: "Success"}, nil
}

func (s *Server) ListResourceGroups(ctx context.Context, in *pb.ListResourceGroupsRequest) (*pb.ListResourceGroupsReply, error) {
	pager := s.resourceClient.NewListPager(nil)
	groups := []*pb.ResourceGroup{}
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, group := range page.Value {
			if group.Name == nil {
				continue
			}
			groups = append(groups, &pb.ResourceGroup{Name: *group.Name})
		}
	}
	return &pb.ListResourceGroupsReply{ResourceGroups: groups}, nil
}
