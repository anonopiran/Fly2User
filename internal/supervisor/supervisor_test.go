package supervisor_test

import (
	"fmt"
	"os"

	mapset "github.com/deckarep/golang-set/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/supervisor"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	test_tester "github.com/anonopiran/Fly2User/test/tester"
)

type userSeed struct {
	us           test_setup.UserSeed
	Usrs         []supervisor.UserRecord
	DupUserEmail supervisor.UserRecord
	DupUserUUID  supervisor.UserRecord
}

func newUserSeed() *userSeed {
	us := userSeed{us: test_setup.NewUserSeed()}
	for _, u := range us.us.Usrs {
		us.Usrs = append(us.Usrs, supervisor.UserRecord{
			UUID:  u.Secret,
			Email: u.Email,
			Level: u.Level,
		})
	}
	us.DupUserEmail = supervisor.UserRecord{
		UUID:  us.us.DupUserEmail.Secret,
		Email: us.us.DupUserEmail.Email,
		Level: us.us.DupUserEmail.Level,
	}
	us.DupUserUUID = supervisor.UserRecord{
		UUID:  us.us.DupUserUUID.Secret,
		Email: us.us.DupUserUUID.Email,
		Level: us.us.DupUserUUID.Level,
	}
	return &us
}

var _ = Describe("Supervisor", Ordered, func() {
	var dplXray *test_setup.DeploymentStruct
	var dplV2fly *test_setup.DeploymentStruct
	var suprvsr *supervisor.Supervisor
	var tmpDir string
	var v2flyClient *test_setup.ClientStruct
	var connClient *test_setup.ConnTestStruct
	var userSeed *userSeed
	var testURL string
	addUsers := func(_usr supervisor.UserRecord, expectErr bool) {
		err := suprvsr.AddUser(&_usr)
		if expectErr {
			Expect(err).ToNot(BeNil())
		} else {
			Expect(err).To(BeNil())
		}
	}
	rmUsers := func(_usr supervisor.UserRecord, expectErr bool) {
		err := suprvsr.RmUser(&_usr)
		if expectErr {
			Expect(err).ToNot(BeNil())
		} else {
			Expect(err).To(BeNil())
		}
	}
	expectDBUserState := func(_usr []supervisor.UserRecord) {
		var dbUsrs []supervisor.UserRecord
		err := suprvsr.UserDB.Select("uuid", "email", "level").Find(&dbUsrs).Error
		Expect(err).To(BeNil())
		Expect(dbUsrs).To(BeEquivalentTo(_usr))
	}
	expectDownServerState := func() {
		dbDS := []supervisor.DownServerType{}
		err := suprvsr.DownServerDB.Select("UpSrvId", "IpAddress").Find(&dbDS).Error
		Expect(err).To(BeNil())
		exp := []supervisor.DownServerType{}
		for c, vv := range []*test_setup.DeploymentStruct{dplV2fly, dplXray} {
			for v := range vv.Ips.Iterator().C {
				exp = append(exp, supervisor.DownServerType{
					ID:        0,
					UpSrvId:   uint(c),
					IpAddress: v,
				})
			}
		}
		Expect(mapset.NewSet(dbDS...)).To(BeEquivalentTo(mapset.NewSet(exp...)))
	}
	expectConn := func(usrs []supervisor.UserRecord, notUsrs []supervisor.UserRecord) {
		dbDS := []supervisor.DownServerType{}
		err := suprvsr.DownServerDB.Select("IpAddress").Find(&dbDS).Error
		Expect(err).To(BeNil())

		for _, srv := range dbDS {
			usrsV2ray := []v2ray.UserType{}
			notUsrsV2ray := []v2ray.UserType{}
			for _, u := range usrs {
				usrsV2ray = append(usrsV2ray, *u.AsV2ray())
			}
			for _, u := range notUsrs {
				notUsrsV2ray = append(notUsrsV2ray, *u.AsV2ray())
			}
			test_tester.CheckConn(usrsV2ray, notUsrsV2ray, v2flyClient, srv.IpAddress, testURL)
		}
	}
	// ...
	discoverSuit := func() {
		By("discovering servers")
		suprvsr.ServiceDiscovery()
		expectDownServerState()
	}
	addUsrSuit := func() {
		By("adding users")
		for c, u := range userSeed.Usrs {
			addUsers(u, false)
			expectDBUserState(userSeed.Usrs[:c+1])
			expectConn(userSeed.Usrs[:c+1], userSeed.Usrs[c+1:])
		}
		// ...
		By("skiping add duplicated user")
		addUsers(userSeed.Usrs[0], false)
		expectDBUserState(userSeed.Usrs)
		expectConn(userSeed.Usrs, []supervisor.UserRecord{})
		// ...
		By("not adding duplicated email user")
		addUsers(userSeed.DupUserEmail, true)
		expectDBUserState(userSeed.Usrs)
		expectConn(userSeed.Usrs, []supervisor.UserRecord{})
		// ...
		By("not adding duplicated uuid user")
		addUsers(userSeed.DupUserUUID, true)
		expectDBUserState(userSeed.Usrs)
		expectConn(userSeed.Usrs, []supervisor.UserRecord{})
	}
	rmUserSuit := func() {
		By("removing users")
		for c, u := range userSeed.Usrs {
			rmUsers(u, false)
			expectDBUserState(userSeed.Usrs[c+1:])
			expectConn(userSeed.Usrs[c+1:], userSeed.Usrs[:c+1])
		}
		// ...
		By("skiping rm unknown user")
		rmUsers(userSeed.Usrs[1], false)
		expectDBUserState([]supervisor.UserRecord{})
		expectConn([]supervisor.UserRecord{}, userSeed.Usrs)
	}
	flushUserSuit := func() {
		By("flusing user")
		err := suprvsr.FlushUser()
		Expect(err).To(BeNil())
		expectDBUserState([]supervisor.UserRecord{})
		expectConn([]supervisor.UserRecord{}, userSeed.Usrs)
	}
	restartSuit := func() {
		By("adding db users after restart")
		dplV2fly.Restart()
		dplXray.Restart()
		discoverSuit()
		suprvsr.RestartHandler()
		var dbUsrs []supervisor.UserRecord
		err := suprvsr.UserDB.Select("uuid", "email", "level").Find(&dbUsrs).Error
		Expect(err).To(BeNil())
		expectConn(dbUsrs, []supervisor.UserRecord{})
	}
	BeforeAll(func() {
		logrus.SetLevel(logrus.ErrorLevel)
		// ...
		var err error
		// ...
		tmpDir, err = os.MkdirTemp("", "sample")
		Expect(err).To(BeNil())
		// ...
		dplXray = test_setup.NewXrayServerDeployment()
		dplV2fly = test_setup.NewV2FlyServerDeployment()
		// ...
		v2flyClient = test_setup.NewV2FlyClientDeployment()
		v2flyClient.Deploy()
		// ...
		connClient = test_setup.NewConnTestDeployment()
		connClient.Deploy()
		testURL = fmt.Sprintf("http://%s:80", connClient.Ips.ToSlice()[0])
		// ...
		userSeed = newUserSeed()
		// ...
		url1 := config.UpstreamUrlType{}
		url2 := config.UpstreamUrlType{}
		err = url1.UnmarshalText([]byte(fmt.Sprintf("grpc://v2fly@%s:8080", dplV2fly.HostName)))
		Expect(err).To(BeNil())
		err = url2.UnmarshalText([]byte(fmt.Sprintf("grpc://xray@%s:8080", dplXray.HostName)))
		Expect(err).To(BeNil())
		suprvsr, err = supervisor.NewSupervisor(config.SupervisorConfigType{
			UserDB: fmt.Sprintf("%s/db.sqlite", tmpDir), Interval: 5},
			config.UpstreamConfigType{
				Address:     []config.UpstreamUrlType{url1, url2},
				InboundList: test_setup.INBOUNDS,
			})
		Expect(err).To(BeNil())
		// ...
	})
	AfterAll(func() {
		os.RemoveAll(tmpDir)
		dplV2fly.UnDeploy()
		dplXray.UnDeploy()
		v2flyClient.UnDeploy()
		connClient.UnDeploy()
	})
	Context("no service available", Ordered, func() {
		AfterAll(func() {
			flushUserSuit()
		})
		It("can discover servers", discoverSuit)
		It("can handle restart", restartSuit)
		It("can add users", addUsrSuit)
		It("can rm users", rmUserSuit)
		It("can flush users", flushUserSuit)
		It("can flush empty users", flushUserSuit)
	})
	Context("many service start from zero", Ordered, func() {
		BeforeAll(func() {
			dplXray.Deploy(5)
			dplV2fly.Deploy(5)
		})
		AfterAll(func() {
			flushUserSuit()
		})
		It("can discover servers", discoverSuit)
		It("can add users", addUsrSuit)
		It("can rm users", rmUserSuit)
		It("can flush users", flushUserSuit)
		It("can flush empty users", flushUserSuit)
		It("can handler restart [when there are no users]", restartSuit)
		It("can add users [to check restart handler]", addUsrSuit)
		It("can handler restart [when there are users]", restartSuit)
	})
	Context("one service down another scale", Ordered, func() {
		BeforeAll(func() {
			dplXray.Deploy(5)
			dplV2fly.UnDeploy()
		})
		AfterAll(func() {
			dplV2fly.UnDeploy()
			dplXray.UnDeploy()
		})
		It("can discover servers", discoverSuit)
		It("can add users", addUsrSuit)
		It("can rm users", rmUserSuit)
		It("can flush users", flushUserSuit)
		It("can flush empty users", flushUserSuit)
		It("can handler restart [when there are no users]", restartSuit)
		It("can add users [to check restart handler]", addUsrSuit)
		It("can handler restart [when there are users]", restartSuit)
	})
	Context("When server deployment change", Ordered, func() {
		BeforeAll(func() {
			dplV2fly.UnDeploy()
			dplXray.UnDeploy()
			discoverSuit()
			restartSuit()
			suprvsr.FlushUser()
		})
		It("can add users", addUsrSuit)
		// ...
		When("scale one service from zero", Ordered, func() {
			BeforeAll(func() {
				dplXray.Deploy(1)
			})
			It("can discover servers", discoverSuit)
			It("can handler restart", restartSuit)
			It("User can connect", func() {
				expectConn(userSeed.Usrs, []supervisor.UserRecord{})
			})
		})

		// ...
		When("scale other service from zero", Ordered, func() {
			BeforeAll(func() {
				dplV2fly.Deploy(1)
			})
			It("can discover servers", discoverSuit)
			It("can handler restart", restartSuit)
			It("User can connect", func() {
				expectConn(userSeed.Usrs, []supervisor.UserRecord{})
			})
		})
		// ...
		When("scale all service up", Ordered, func() {
			BeforeAll(func() {
				dplXray.Deploy(1)
				dplV2fly.Deploy(1)
			})
			It("can discover servers", discoverSuit)
			It("can handler restart", restartSuit)
			It("User can connect", func() {
				expectConn(userSeed.Usrs, []supervisor.UserRecord{})
			})
		})

	})
})
