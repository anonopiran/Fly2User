package v2ray_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestV2ray(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V2ray Suite")
}
