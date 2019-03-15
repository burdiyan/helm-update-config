package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	updatec "github.com/bluebosh/helm-update-config"
)

var _ = Describe("Updatecfg", func() {

	var vf updatec.ValueFiles
	BeforeEach(func() {
		vf = updatec.ValueFiles{"test.yaml"}
	})
	// Test command line input with only --set-value.
	It("GenerateUpdatedValues", func() {
		values := []string{"setvalue=1", "foo.bar=10"}
		uValues, _ := updatec.GenerateUpdatedValues(nil, values)

		Expect(uValues["setvalue"]).To(Equal(int64(1)))

	})

	// Test command line input with only value file.
	It("GenerateUpdatedValues", func() {
		uValues, _ := updatec.GenerateUpdatedValues(vf, nil)
		uValuesNext := uValues["foo"]
		result := map[string]interface{}{}
		for k, v := range uValuesNext.(map[interface{}]interface{}) {
			result[k.(string)] = v
		}

		Expect(result["bar"]).To(Equal(3))

		Expect(uValues["teststr"]).To(Equal("origion"))
		Expect(uValues["addmore"]).To(Equal(10))
	})

	// Test command line input with both --set-value and value file.
	It("GenerateUpdatedValues", func() {
		values := []string{"addmore=9", "newvalue=hello"}
		uValues, _ := updatec.GenerateUpdatedValues(vf, values)

		Expect(uValues["testint"]).To(Equal(1))
		Expect(uValues["addmore"]).To(Equal(int64(9)))
		Expect(uValues["newvalue"]).To(Equal("hello"))
	})

	// Input parameter test.
	It("GenerateUpdatedValues", func() {
		values := []string{"setvalue=1"}
		uValues, _ := updatec.GenerateUpdatedValues(nil, values)

		Expect(uValues["badvalue"]).Should(BeNil())
	})
})
