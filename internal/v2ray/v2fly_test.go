package v2ray_test

import (
	"context"
	"fmt"
	"net/url"

	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ = Describe("V2fly server", Ordered, func() {
	var dpl *test_setup.DeploymentStruct
	var srv *v2ray.UpServer
	var conn *grpc.ClientConn
	ctx := context.Background()
	usrs := []v2ray.UserType{{
		Email:  "test@user.com",
		Secret: uuid.NewString(),
		Level:  0,
	}, {
		Email:  "test@user2.com",
		Secret: uuid.NewString(),
		Level:  0,
	},
	}
	dupUserEmail := v2ray.UserType{
		Email:  usrs[0].Email,
		Secret: uuid.NewString(),
		Level:  0,
	}
	addUser := func(_usr *v2ray.UserType, expectErr bool, userExistErr bool) {
		for _, inbd := range test_setup.INBOUNDS {
			err := srv.AddUser(ctx, &inbd, _usr, conn)
			if !expectErr {
				Expect(err).To(BeNil())
			}
			if expectErr {
				Expect(err).ToNot(BeNil())
				if userExistErr {
					grpcErr, ok := err.(*v2ray.GrpcError)
					Expect(ok).To(BeTrue())
					Expect(grpcErr.IsUserExistsError()).To(BeTrue())
					Expect(grpcErr.IsUserNotFoundError()).To(BeFalse())

				}
			}
		}
	}
	rmUser := func(_usr *v2ray.UserType, expectErr bool, userNotExistErr bool) {
		for _, inbd := range test_setup.INBOUNDS {
			err := srv.RmUser(ctx, &inbd, _usr, conn)
			if !expectErr {
				Expect(err).To(BeNil())
			}
			if expectErr {
				Expect(err).ToNot(BeNil())
				if userNotExistErr {
					grpcErr, ok := err.(*v2ray.GrpcError)
					Expect(ok).To(BeTrue())
					Expect(grpcErr.IsUserNotFoundError()).To(BeTrue())
					Expect(grpcErr.IsUserExistsError()).To(BeFalse())

				}
			}
		}
	}
	discover := func(expectedResult mapset.Set[string], expectErr bool) {
		res, err := srv.Discover(ctx)
		if expectErr {
			Expect(res).To(BeNil())
			Expect(err).ToNot(BeNil())
		} else {
			Expect(res).ToNot(BeNil())
			Expect(err).To(BeNil())
			Expect(res).To(BeEquivalentTo(expectedResult))
		}
	}
	Context("When service not available", func() {
		BeforeAll(func() {
			url, err := url.Parse("nil.local:8080")
			Expect(err).To(BeNil())
			// ...
			srv, err = v2ray.NewServer(config.UpstreamUrlType{URL: *url, ServerType: config.V2FLY_SRV})
			Expect(err).To(BeNil())
		})
		It("not panic discover", func() {
			discover(nil, true)
		})
		It("not panic add user", func() {
			addUser(&usrs[0], true, false)
		})
		It("not panic rm user", func() {
			rmUser(&usrs[0], true, false)
		})
	})
	Context("When service is up", Ordered, func() {
		BeforeAll(func() {
			dpl = test_setup.NewV2FlyServerDeployment()
			dpl.Deploy(1)
			url, err := url.Parse(fmt.Sprintf("grpc://%s:8080", dpl.HostName))
			Expect(err).To(BeNil())
			// ...
			srv, err = v2ray.NewServer(config.UpstreamUrlType{URL: *url, ServerType: config.V2FLY_SRV})
			Expect(err).To(BeNil())
			// ...
			opt := grpc.WithTransportCredentials(insecure.NewCredentials())
			conn, err = grpc.NewClient(fmt.Sprintf("%s:%s", srv.Address.Hostname(), srv.Address.Port()), opt)
			Expect(err).To(BeNil())

		})
		AfterAll(func() {
			conn.Close()
			dpl.UnDeploy()
		})
		It("can discover", func() {
			discover(dpl.Ips, false)
		})
		It("adds user", func() {
			addUser(&usrs[0], false, false)
		})
		It("adds another user", func() {
			addUser(&usrs[1], false, false)
		})
		It("not add duplicated email user", func() {
			addUser(&dupUserEmail, true, true)
		})
		It("rm user", func() {
			rmUser(&usrs[0], false, false)
		})
		It("rm another user", func() {
			rmUser(&usrs[1], false, false)
		})
		It("not rm unknown user", func() {
			rmUser(&usrs[1], true, true)
		})
		It("not add to unknown inbount", func() {
			err := srv.AddUser(ctx, &config.InboundConfigType{Proto: config.TROJAN_PROTO, Tag: "invalid"}, &usrs[0], conn)
			Expect(err).NotTo(BeNil())
		})
		It("not rm to unknown inbount", func() {
			err := srv.RmUser(ctx, &config.InboundConfigType{Proto: config.TROJAN_PROTO, Tag: "invalid"}, &usrs[0], conn)
			Expect(err).NotTo(BeNil())
		})
	})
	Context("When multiple service is up", Ordered, func() {
		BeforeAll(func() {
			dpl = test_setup.NewV2FlyServerDeployment()
			dpl.Deploy(5)
			url, err := url.Parse(fmt.Sprintf("grpc://%s:8080", dpl.HostName))
			Expect(err).To(BeNil())
			// ...
			srv, err = v2ray.NewServer(config.UpstreamUrlType{URL: *url, ServerType: config.V2FLY_SRV})
			Expect(err).To(BeNil())
			// ...
			opt := grpc.WithTransportCredentials(insecure.NewCredentials())
			conn, err = grpc.NewClient(fmt.Sprintf("%s:%s", srv.Address.Hostname(), srv.Address.Port()), opt)
			Expect(err).To(BeNil())
		})
		AfterAll(func() {
			conn.Close()
			dpl.UnDeploy()
		})
		It("can discover", func() {
			discover(dpl.Ips, false)
		})
	})
})

var _ = DescribeTable("V2fly new user request error", func(inbound *config.InboundConfigType, user *v2ray.UserType, expErr error) {
	v := &v2ray.V2flyServer{}
	_, err := v.NewAddUserReq(inbound, user)
	Expect(err).NotTo(BeNil())
	Expect(err).To(BeEquivalentTo(expErr))
},
	Entry("user is nill", &config.InboundConfigType{}, nil, v2ray.ErrUserNotDefined),
	Entry("inbound is nill", nil, &v2ray.UserType{}, v2ray.ErrInboundNotDefined),
)
var _ = DescribeTable("V2fly rm user request error", func(inbound *config.InboundConfigType, user *v2ray.UserType, expErr error) {
	v := &v2ray.V2flyServer{}
	_, err := v.NewRmUserReq(inbound, user)
	Expect(err).NotTo(BeNil())
	Expect(err).To(BeEquivalentTo(expErr))
},
	Entry("user is nill", &config.InboundConfigType{}, nil, v2ray.ErrUserNotDefined),
	Entry("inbound is nill", nil, &v2ray.UserType{}, v2ray.ErrInboundNotDefined),
)
