package shortenerpb

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	ShortenerServiceShortenURLFullMethodName   = "/ShortenerService/ShortenURL"
	ShortenerServiceExpandURLFullMethodName    = "/ShortenerService/ExpandURL"
	ShortenerServiceListUserURLsFullMethodName = "/ShortenerService/ListUserURLs"
)

type URLShortenRequest struct {
	URL string
}

type URLShortenResponse struct {
	Result string
}

type URLExpandRequest struct {
	ID string
}

type URLExpandResponse struct {
	Result string
}

type UserURLsResponse struct {
	URL []*URLData
}

type URLData struct {
	ShortURL    string
	OriginalURL string
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
	err := c.cc.Invoke(ctx, ShortenerServiceShortenURLFullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ExpandURL(ctx context.Context, in *URLExpandRequest, opts ...grpc.CallOption) (*URLExpandResponse, error) {
	out := new(URLExpandResponse)
	err := c.cc.Invoke(ctx, ShortenerServiceExpandURLFullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *shortenerServiceClient) ListUserURLs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*UserURLsResponse, error) {
	out := new(UserURLsResponse)
	err := c.cc.Invoke(ctx, ShortenerServiceListUserURLsFullMethodName, in, out, opts...)
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
	s.RegisterService(&ShortenerServiceServiceDesc, srv)
}

func shortenerServiceShortenURLHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLShortenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ShortenURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerServiceShortenURLFullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ShortenURL(ctx, req.(*URLShortenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func shortenerServiceExpandURLHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLExpandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ExpandURL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerServiceExpandURLFullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ExpandURL(ctx, req.(*URLExpandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func shortenerServiceListUserURLsHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ShortenerServiceServer).ListUserURLs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ShortenerServiceListUserURLsFullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ShortenerServiceServer).ListUserURLs(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var ShortenerServiceServiceDesc = grpc.ServiceDesc{
	ServiceName: "ShortenerService",
	HandlerType: (*ShortenerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ShortenURL",
			Handler:    shortenerServiceShortenURLHandler,
		},
		{
			MethodName: "ExpandURL",
			Handler:    shortenerServiceExpandURLHandler,
		},
		{
			MethodName: "ListUserURLs",
			Handler:    shortenerServiceListUserURLsHandler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/shortener.proto",
}
