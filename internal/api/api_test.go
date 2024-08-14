package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"

	"github.com/anonopiran/Fly2User/internal/api"
	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/supervisor"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	test_tester "github.com/anonopiran/Fly2User/test/tester"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

type userSeed struct {
	AddUsrs       []api.AddUserReq
	RmUsers       []api.RmUserReq
	DupUserEmail  api.AddUserReq
	DupUserUUID   api.AddUserReq
	NullUserEmail api.AddUserReq
	NullUserUUID  api.AddUserReq
}

func newUserSeed(n int) *userSeed {
	us := userSeed{}
	for c := range n {
		_u := api.AddUserReq{
			UUID:  uuid.NewString(),
			Level: 0,
			Email: fmt.Sprintf("u%d@v2ray.cc", c),
		}
		us.AddUsrs = append(us.AddUsrs, _u)
		us.RmUsers = append(us.RmUsers, api.RmUserReq{Email: _u.Email})
	}
	us.DupUserEmail = api.AddUserReq{
		UUID:  uuid.NewString(),
		Level: 0,
		Email: us.AddUsrs[0].Email,
	}
	us.DupUserUUID = api.AddUserReq{
		UUID:  us.AddUsrs[0].UUID,
		Level: 0,
		Email: "dup@v2ray.cc",
	}
	us.NullUserEmail = api.AddUserReq{
		UUID:  uuid.NewString(),
		Level: 0,
	}
	us.NullUserUUID = api.AddUserReq{
		Level: 0,
		Email: "noid@v2ray.cc",
	}
	return &us
}

var _ = Describe("Api", Ordered, func() {
	var dplXray *test_setup.DeploymentStruct
	var dplV2fly *test_setup.DeploymentStruct
	var suprvsr *supervisor.Supervisor
	var tmpDir string
	var v2flyClient *test_setup.ClientStruct
	var connClient *test_setup.ConnTestStruct
	var userSeed *userSeed
	var testURL string
	var ginApp *gin.Engine
	_statusOk := []interface{}{
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,
		http.StatusMultiStatus,
		http.StatusAlreadyReported,
		http.StatusIMUsed,
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
		dplXray.Deploy(2)
		dplV2fly.Deploy(2)
		// ...
		v2flyClient = test_setup.NewV2FlyClientDeployment()
		v2flyClient.Deploy()
		// ...
		connClient = test_setup.NewConnTestDeployment()
		connClient.Deploy()
		testURL = fmt.Sprintf("http://%s:80", connClient.Ips.ToSlice()[0])
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
		suprvsr.ServiceDiscovery()
		// ...
		ginApp = api.AddRoutes(gin.New(), &config.ServerConfigType{}, suprvsr)
		gin.SetMode("test")
		// ...

	})
	AfterAll(func() {
		os.RemoveAll(tmpDir)
		dplV2fly.UnDeploy()
		dplXray.UnDeploy()
		v2flyClient.UnDeploy()
		connClient.UnDeploy()
	})
	// ...
	addUser := func(u api.AddUserReq, expectErr bool) {
		uJson, err := json.Marshal(u)
		Expect(err).To(BeNil())
		req, err := http.NewRequest("POST", "/user/", strings.NewReader(string(uJson)))
		Expect(err).To(BeNil())
		w := httptest.NewRecorder()
		ginApp.ServeHTTP(w, req)
		defer w.Result().Body.Close()
		if expectErr {
			Expect(w.Result()).NotTo(HaveHTTPStatus(_statusOk...))
		} else {
			Expect(w.Result()).To(HaveHTTPStatus(_statusOk...))
		}
	}
	// ...
	rmUser := func(u api.RmUserReq, expectErr bool) {
		uJson, err := json.Marshal(u)
		Expect(err).To(BeNil())
		req, err := http.NewRequest("DELETE", "/user/", strings.NewReader(string(uJson)))
		Expect(err).To(BeNil())
		w := httptest.NewRecorder()
		ginApp.ServeHTTP(w, req)
		defer w.Result().Body.Close()
		if expectErr {
			Expect(w.Result()).NotTo(HaveHTTPStatus(_statusOk...))
		} else {
			Expect(w.Result()).To(HaveHTTPStatus(_statusOk...))
		}
	}
	// ...
	flushUser := func(expectErr bool) {
		req, err := http.NewRequest("GET", "/user/flush", nil)
		Expect(err).To(BeNil())
		w := httptest.NewRecorder()
		ginApp.ServeHTTP(w, req)
		defer w.Result().Body.Close()
		if expectErr {
			Expect(w.Result()).NotTo(HaveHTTPStatus(_statusOk...))
		} else {
			Expect(w.Result()).To(HaveHTTPStatus(_statusOk...))
		}
	}
	// ...
	expectConn := func(usrs []api.AddUserReq, notUsrs []api.AddUserReq) {
		for _, dpl := range []*test_setup.DeploymentStruct{dplV2fly, dplXray} {
			for srv := range dpl.Ips.Iterator().C {
				usrsV2ray := []v2ray.UserType{}
				notUsrsV2ray := []v2ray.UserType{}
				for _, u := range usrs {
					usrsV2ray = append(usrsV2ray, *u.AsUSer().AsV2ray())
				}
				for _, u := range notUsrs {
					notUsrsV2ray = append(notUsrsV2ray, *u.AsUSer().AsV2ray())
				}
				test_tester.CheckConn(usrsV2ray, notUsrsV2ray, v2flyClient, srv, testURL)
			}
		}
	}
	// ...
	expectCount := func(n int) {
		req, err := http.NewRequest("GET", "/user/count", nil)
		Expect(err).To(BeNil())
		w := httptest.NewRecorder()
		ginApp.ServeHTTP(w, req)
		defer w.Result().Body.Close()
		Expect(w.Result()).To(HaveHTTPStatus(_statusOk...))
		Expect(w.Result()).To(HaveHTTPBody(fmt.Sprint(n)))
	}
	// ...
	Context("Gin api", Ordered, func() {
		BeforeAll(func() {
			userSeed = newUserSeed(10)
		})
		It("can add users", func() {
			By("adding users")
			for c, u := range userSeed.AddUsrs {
				expectCount(c)
				addUser(u, false)
			}
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			// ...
			By("skiping add duplicated user")
			addUser(userSeed.AddUsrs[0], false)
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			expectCount(len(userSeed.AddUsrs))
			// ...
			By("Not Adding duplicated Email user")
			addUser(userSeed.DupUserEmail, true)
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			expectCount(len(userSeed.AddUsrs))
			// ...
			By("Not Adding duplicated UUID user")
			addUser(userSeed.DupUserUUID, true)
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			expectCount(len(userSeed.AddUsrs))
			// ...
			By("Not Adding Nil Email user")
			addUser(userSeed.NullUserEmail, true)
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			expectCount(len(userSeed.AddUsrs))
			// ...
			By("Not Adding Nil UUID user")
			addUser(userSeed.NullUserUUID, true)
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			expectCount(len(userSeed.AddUsrs))
		})
		It("can remove users", func() {
			By("removing single user")
			rmUser(userSeed.RmUsers[0], false)
			expectConn(userSeed.AddUsrs[1:], userSeed.AddUsrs[:1])
			expectCount(len(userSeed.AddUsrs) - 1)
			// ...
			By("removing unknown users")
			rmUser(userSeed.RmUsers[0], false)
			expectConn(userSeed.AddUsrs[1:], userSeed.AddUsrs[:1])
			expectCount(len(userSeed.AddUsrs) - 1)
			// ...
			By("removing users")
			for c, u := range userSeed.RmUsers[1:] {
				expectCount(len(userSeed.AddUsrs) - 1 - c)
				rmUser(u, false)
			}
			expectConn([]api.AddUserReq{}, userSeed.AddUsrs)
		})
		// ...
		It("can flush users", func() {
			By("adding users")
			for _, u := range userSeed.AddUsrs {
				addUser(u, false)
			}
			expectCount(len(userSeed.AddUsrs))
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
			// ...
			By("purging users")
			flushUser(false)
			expectConn([]api.AddUserReq{}, userSeed.AddUsrs)
			expectCount(0)
			// ...
			By("purging users while no user exists")
			flushUser(false)
			expectConn([]api.AddUserReq{}, userSeed.AddUsrs)
			expectCount(0)
		})
	})
	Context("Gin server", Ordered, func() {
		defer GinkgoRecover()
		BeforeAll(func() {
			userSeed = newUserSeed(1000)
		})
		It("can handle many requests", func() {
			By("adding many users concurrently")
			wg := sync.WaitGroup{}
			for _, u := range userSeed.AddUsrs {
				wg.Add(1)
				go func() {
					defer wg.Done()
					uJson, _ := json.Marshal(u)
					req, _ := http.NewRequest("POST", "/user/", strings.NewReader(string(uJson)))
					w := httptest.NewRecorder()
					ginApp.ServeHTTP(w, req)
				}()
			}
			wg.Wait()
			expectCount(len(userSeed.AddUsrs))
			expectConn(userSeed.AddUsrs, []api.AddUserReq{})
		})
	})
})
