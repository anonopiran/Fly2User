package v2ray_test

import (
	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	. "github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
)

var _ = Describe("V2fly server", Ordered, func() {
	logrus.SetLevel(logrus.PanicLevel)
	realTest(v2ray.NewServer, config.V2FLY_SRV, test_setup.NewV2FlyServerDeployment())
	staticTest(&v2ray.V2flyServer{})
})
