package tester_test

import (
	"github.com/anonopiran/Fly2User/internal/v2ray"
	test_setup "github.com/anonopiran/Fly2User/test/setup"
	gomega "github.com/onsi/gomega"
)

func CheckConn(usrs []v2ray.UserType, notUsrs []v2ray.UserType, dpl *test_setup.ClientStruct, targetSrvIp string, u string) {
	for _, usr := range usrs {
		err := dpl.SetOutbound(targetSrvIp, usr.Secret)
		gomega.Expect(err).To(gomega.BeNil())
		err = dpl.TestURL(false, u)
		gomega.Expect(err).To(gomega.BeNil())
	}
	for _, usr := range notUsrs {
		err := dpl.SetOutbound(targetSrvIp, usr.Secret)
		gomega.Expect(err).To(gomega.BeNil())
		err = dpl.TestURL(true, u)
		gomega.Expect(err).To(gomega.BeNil())
	}
}
