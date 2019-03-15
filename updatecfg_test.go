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

	// Test command line input with only value file.
	Context("When input file is nil", func() {

		It("GenerateUpdatedValues will exit if --set-value is also nil", func() {
			_, err := updatec.GenerateUpdatedValues(nil, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("GenerateUpdatedValues will set key/value from --set-value", func() {
			values := []string{"foo=bar"}
			uValues, err := updatec.GenerateUpdatedValues(nil, values)
			Expect(err).NotTo(HaveOccurred())

			Expect(uValues["badvalue"]).Should(BeNil())
			Expect(uValues["foo"]).To(Equal("bar"))
		})
	})

	Context("When input file is not nil but invalid", func() {
		// Test command line input with only --set-value.
		It("GenerateUpdatedValues will set key/value from --set-value", func() {
			values := []string{"foo.bar=10", "setvalue=1"}
			vf = updatec.ValueFiles{"non_exist_file"}
			uValues, err := updatec.GenerateUpdatedValues(vf, values)
			Expect(err).To(HaveOccurred())
			Expect(uValues["setvalue"]).Should(BeNil())
		})
	})

	Context("When input file is not nil and is a valid file path", func() {
		It("GenerateUpdatedValues will set values from input file if --set-values is nil", func() {
			uValues, err := updatec.GenerateUpdatedValues(vf, nil)
			Expect(err).NotTo(HaveOccurred())

			uValuesNext := uValues["foo"]
			result := map[string]interface{}{}
			for k, v := range uValuesNext.(map[interface{}]interface{}) {
				result[k.(string)] = v
			}

			Expect(result["bar"]).To(Equal(3))
			Expect(uValues["teststr"]).To(Equal("origion"))
			Expect(uValues["addmore"]).To(Equal(10))
		})

		It("GenerateUpdatedValues will set values from both input file and --set-value, the latter will override conflict values", func() {
			values := []string{"addmore=9", "newvalue=hello"}
			uValues, err := updatec.GenerateUpdatedValues(vf, values)
			Expect(err).NotTo(HaveOccurred())

			Expect(uValues["testint"]).To(Equal(1))
			Expect(uValues["addmore"]).To(Equal(int64(9)))
			Expect(uValues["newvalue"]).To(Equal("hello"))
		})
	})

})
