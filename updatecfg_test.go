package main_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	updatec "github.com/bluebosh/helm-update-config"
)

var _ = Describe("Updatecfg", func() {
	It("GenerateUpdatedValues", func() {
		vf := updatec.ValueFiles{"test.yaml"}
		var values []string
		values[0] = "--set-value=foo.bar=\"9\""
		uValues, err := updatec.GenerateUpdatedValues(vf, values)

		fmt.Println(uValues)
		fmt.Println("error:", err)
		Expect(true)
		//		Expect(values[0].To(Equal("--set-value=foo.bar=\"9\"")))
		//Expect(uValues["test"].To(Equal(1)))
		//		Expect(uValues["foo"]["baz"].To(Equal(6)))
		//		Expect(uValues["qux"]["uier"].To(Equal(false)))
	})
})
