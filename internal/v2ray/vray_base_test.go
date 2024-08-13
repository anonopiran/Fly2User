package v2ray_test

import (
	"context"
	"fmt"
	"net/url"

	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	test_tester "github.com/anonopiran/Fly2User/test/tester"
	mapset "github.com/deckarep/golang-set/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TstAddUser(_usr *v2ray.UserType, srv *v2ray.UpServer, expectErr bool, userExistErr bool, conn *grpc.ClientConn, ctx context.Context) {
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
func TstRmUser(_usr *v2ray.UserType, srv *v2ray.UpServer, expectErr bool, userNotExistErr bool, conn *grpc.ClientConn, ctx context.Context) {
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
func TstDiscover(expectedResult mapset.Set[string], srv *v2ray.UpServer, ctx context.Context) {
	res := srv.Discover(ctx)
	Expect(res).ToNot(BeNil())
	Expect(res).To(BeEquivalentTo(expectedResult))
}

func realTest(getSrvFunc func(srv config.UpstreamUrlType) (*v2ray.UpServer, error), srvType config.ServerTypeEnumType, dpl *test_setup.DeploymentStruct) {
	var srv *v2ray.UpServer
	var ctx = context.Background()
	var dSeed = test_setup.NewUserSeed()
	var conn *grpc.ClientConn
	var clientDpl *test_setup.ClientStruct
	var connDpl *test_setup.ConnTestStruct

	BeforeAll(func() {
		clientDpl = test_setup.NewV2FlyClientDeployment()
		connDpl = test_setup.NewConnTestDeployment()
		clientDpl.Deploy()
		connDpl.Deploy()
	})
	AfterAll(func() {
		clientDpl.UnDeploy()
		connDpl.UnDeploy()
	})
	Context("When no service available", func() {
		BeforeAll(func() {
			url, err := url.Parse("nil.local:8080")
			Expect(err).To(BeNil())
			// ...
			srv, err = getSrvFunc(config.UpstreamUrlType{URL: *url, ServerType: srvType})
			Expect(err).To(BeNil())
		})
		It("can discover", func() {
			TstDiscover(dpl.Ips, srv, ctx)
		})
		It("error add user with nill conn", func() {
			TstAddUser(&dSeed.Usrs[0], srv, true, false, nil, ctx)
		})
		It("error rm user with nill conn", func() {
			TstRmUser(&dSeed.Usrs[0], srv, true, false, nil, ctx)
		})
	})
	Context("When service is up", Ordered, func() {
		BeforeAll(func() {
			dpl.Deploy(1)
			url, err := url.Parse(fmt.Sprintf("grpc://%s:8080", dpl.HostName))
			Expect(err).To(BeNil())
			// ...
			srv, err = getSrvFunc(config.UpstreamUrlType{URL: *url, ServerType: srvType})
			Expect(err).To(BeNil())
			// ...
			conn, err = grpc.NewClient(fmt.Sprintf("%s:%s", srv.Address.Hostname(), srv.Address.Port()), grpc.WithTransportCredentials(insecure.NewCredentials()))
			Expect(err).To(BeNil())

		})
		AfterAll(func() {
			conn.Close()
			dpl.UnDeploy()
		})
		It("can discover", func() {
			TstDiscover(dpl.Ips, srv, ctx)
		})
		It("can add user", func() {
			TstAddUser(&dSeed.Usrs[0], srv, false, false, conn, ctx)
			test_tester.CheckConn(dSeed.Usrs[:1], dSeed.Usrs[1:], clientDpl, dpl.Ips.ToSlice()[0], fmt.Sprintf("http://%s:80", connDpl.Ips.ToSlice()[0]))
		})
		It("can add another user", func() {
			TstAddUser(&dSeed.Usrs[1], srv, false, false, conn, ctx)
			test_tester.CheckConn(dSeed.Usrs, []v2ray.UserType{}, clientDpl, dpl.Ips.ToSlice()[0], fmt.Sprintf("http://%s:80", connDpl.Ips.ToSlice()[0]))
		})
		It("error add duplicated email user", func() {
			TstAddUser(&dSeed.Usrs[1], srv, true, true, conn, ctx)
			test_tester.CheckConn(dSeed.Usrs, []v2ray.UserType{}, clientDpl, dpl.Ips.ToSlice()[0], fmt.Sprintf("http://%s:80", connDpl.Ips.ToSlice()[0]))
		})
		It("error add conflict email user", func() {
			TstAddUser(&dSeed.DupUserEmail, srv, true, true, conn, ctx)
			test_tester.CheckConn(dSeed.Usrs, []v2ray.UserType{}, clientDpl, dpl.Ips.ToSlice()[0], fmt.Sprintf("http://%s:80", connDpl.Ips.ToSlice()[0]))
		})
		It("can rm user", func() {
			TstRmUser(&dSeed.Usrs[0], srv, false, false, conn, ctx)
			test_tester.CheckConn(dSeed.Usrs[1:], dSeed.Usrs[:1], clientDpl, dpl.Ips.ToSlice()[0], fmt.Sprintf("http://%s:80", connDpl.Ips.ToSlice()[0]))
		})
		It("can rm another user", func() {
			TstRmUser(&dSeed.Usrs[1], srv, false, false, conn, ctx)
			test_tester.CheckConn([]v2ray.UserType{}, dSeed.Usrs, clientDpl, dpl.Ips.ToSlice()[0], fmt.Sprintf("http://%s:80", connDpl.Ips.ToSlice()[0]))

		})
		It("can rm unknown user", func() {
			TstRmUser(&dSeed.Usrs[0], srv, true, true, conn, ctx)
		})
		It("not add to unknown inbount", func() {
			err := srv.AddUser(ctx, &config.InboundConfigType{Proto: config.TROJAN_PROTO, Tag: "invalid"}, &dSeed.Usrs[0], conn)
			Expect(err).NotTo(BeNil())
		})
		It("not rm to unknown inbount", func() {
			err := srv.RmUser(ctx, &config.InboundConfigType{Proto: config.TROJAN_PROTO, Tag: "invalid"}, &dSeed.Usrs[0], conn)
			Expect(err).NotTo(BeNil())
		})
	})

	Context("When multiple service is up", Ordered, func() {
		BeforeAll(func() {
			dpl.Deploy(5)
			url, err := url.Parse(fmt.Sprintf("grpc://%s:8080", dpl.HostName))
			Expect(err).To(BeNil())
			// ...
			srv, err = getSrvFunc(config.UpstreamUrlType{URL: *url, ServerType: srvType})
			Expect(err).To(BeNil())
			// ...
			conn, err = grpc.NewClient(fmt.Sprintf("%s:%s", srv.Address.Hostname(), srv.Address.Port()), grpc.WithTransportCredentials(insecure.NewCredentials()))
			Expect(err).To(BeNil())

		})
		AfterAll(func() {
			conn.Close()
			dpl.UnDeploy()
		})
		It("can discover", func() {
			TstDiscover(dpl.Ips, srv, ctx)
		})
	})

}
func staticTestDescribe(v v2ray.IServer, inbound *config.InboundConfigType, user *v2ray.UserType, expErr error) {
	_, err := v.NewRmUserReq(inbound, user)
	Expect(err).NotTo(BeNil())
	Expect(err).To(BeEquivalentTo(expErr))
}
func staticTest(v v2ray.IServer) {
	DescribeTable("new user request error", staticTestDescribe, Entry("user is nill", v, &config.InboundConfigType{}, nil, v2ray.ErrUserNotDefined), Entry("inbound is nill", v, nil, &v2ray.UserType{}, v2ray.ErrInboundNotDefined))
	DescribeTable("rm user request error", staticTestDescribe, Entry("user is nill", v, &config.InboundConfigType{}, nil, v2ray.ErrUserNotDefined), Entry("inbound is nill", v, nil, &v2ray.UserType{}, v2ray.ErrInboundNotDefined))
}
