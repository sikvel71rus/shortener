package shortenerpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	ShortenerService_ShortenURL_FullMethodName   = "/ShortenerService/ShortenURL"
	ShortenerService_ExpandURL_FullMethodName    = "/ShortenerService/ExpandURL"
	ShortenerService_ListUserURLs_FullMethodName = "/ShortenerService/ListUserURLs"
)

type URLShortenRequest struct {
	Url string
}

type URLShortenResponse struct {
	Result string
}

type URLExpandRequest struct {
	Id string
}

type URLExpandResponse struct {
	Result string
}

type UserURLsResponse struct {
	Url []*URLData
}

type URLData struct {
	ShortUrl    string
	OriginalUrl string
}

type ShortenerServiceClient interface {
	ShortenURL(ctx context.Context, in *URLShortenRequest, opts ...grpc.CallOption) (*URLShortenResponse, error)
	ExpandURL(ctx context.Context, in *URLExpandRequest, opts ...grpc.CallOption) (*URLExpandResponse, error)
	ListUserURLs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserURLsResponse, error)
}

type shortenerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewShortenerServiceClient(cc grpc.ClientConnInterface) ShortenerServiceClient {
	return &shortenerServiceClient{cc}
}

func (c *shortenerServiceClient) ShortenURL(ctx context.Context, in *URLShortenRequest, opts ...grpc.CallOption) (*URLShortenResponse, error) {
	out := new(URLShortenResponse)
	err := c.cc.Invoke(ctx, ShortenerService_ShortenURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ExpandURL(ctx context.Context, in *URLExpandRequest, opts ...grpc.CallOption) (*URLExpandResponse, error) {
	out := new(URLExpandResponse)
	err := c.cc.Invoke(ctx, ShortenerService_ExpandURL_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ListUserURLs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserURLsResponse, error) {
	out := new(UserURLsResponse)
	err := c.cc.Invoke(ctx, ShortenerService_ListUserURLs_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type ShortenerServiceServer interface {
	ShortenURL(context.Context, *URLShortenRequest) (*URLShortenResponse, error)
	ExpandURL(context.Context, *URLExpandRequest) (*URLExpandResponse, error)
	ListUserURLs(context.Context, *emptypb.Empty) (*UserURLsResponse, error)
}

func RegisterShortenerServiceServer(s grpc.ServiceRegistrar, srv ShortenerServiceServer) {
	s.RegisterService(&ShortenerService_ServiceDesc, srv)
}

func _ShortenerService_ShortenURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLShortenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ShortenURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_ShortenURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ShortenURL(ctx, req.(*URLShortenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_ExpandURL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLExpandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ExpandURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_ExpandURL_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ExpandURL(ctx, req.(*URLExpandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ShortenerService_ListUserURLs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ListUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerService_ListUserURLs_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ListUserURLs(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var ShortenerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ShortenerService",
	HandlerType: (*ShortenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ShortenURL",
			Handler:    _ShortenerService_ShortenURL_Handler,
		},
		{
			MethodName: "ExpandURL",
			Handler:    _ShortenerService_ExpandURL_Handler,
		},
		{
			MethodName: "ListUserURLs",
			Handler:    _ShortenerService_ListUserURLs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/shortener.proto",
}
