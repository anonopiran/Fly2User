package v2ray_test

import (
	"fmt"

	"github.com/anonopiran/Fly2User/internal/v2ray"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GrpcError", func() {
	It("IsUserExistsError", func() {
		err := &v2ray.GrpcError{fmt.Errorf("already exists.")}
		Expect(err.IsUserExistsError()).To(BeTrue())
	})

	It("IsUserExistsError false", func() {
		err := &v2ray.GrpcError{fmt.Errorf("unknown error")}
		Expect(err.IsUserExistsError()).To(BeFalse())
	})

	It("IsUserNotFoundError", func() {
		err := &v2ray.GrpcError{fmt.Errorf("not found.")}
		Expect(err.IsUserNotFoundError()).To(BeTrue())
	})

	It("IsUserNotFoundError false", func() {
		err := &v2ray.GrpcError{fmt.Errorf("unknown error")}
		Expect(err.IsUserNotFoundError()).To(BeFalse())
	})
})
