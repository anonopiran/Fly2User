package v2ray_test

import (
	"github.com/anonopiran/Fly2User/internal/config"
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	. "github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Xray server", Ordered, func() {
	logrus.SetLevel(logrus.PanicLevel)
	realTest(v2ray.NewServer, config.XRAY_SRV, test_setup.NewXrayServerDeployment())
	staticTest(&v2ray.XrayServer{})
})
