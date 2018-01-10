package cmd_test

import (
	. "code.cloudfoundry.org/cc-to-perm-migrator/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Generating a report", func() {
	var c chan RoleAssignment

	BeforeEach(func() {
		c = make(chan RoleAssignment, 1000)
	})

	Describe(".GenerateReport", func() {
		var (
			b *Buffer
		)

		BeforeEach(func() {
			b = NewBuffer()
		})

		Context("when the channel is sent 2 elements", func() {
			It1Second("", func() {
				c <- RoleAssignment{}
				c <- RoleAssignment{}
				close(c)

				GenerateReport(b, c)

				Expect(b).To(Say("Number of role assignments: 2\\."))
			})
		})
	})

	Describe(".ComputeNumberAssignments", func() {
		Context("when the channel is closed without sending anything onto it", func() {
			It1Second("returns 0", func() {
				close(c)
				count := ComputeNumberAssignments(c)

				Expect(count).To(BeZero())
			})
		})

		Context("when the channel is sent one element", func() {
			It1Second("returns 1", func() {
				c <- RoleAssignment{}
				close(c)

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(1))
			})
		})

		Context("when the channel is sent 2 elements", func() {
			It1Second("returns 1", func() {
				c <- RoleAssignment{}
				c <- RoleAssignment{}
				close(c)

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(2))
			})
		})

		Context("when the channel is sent more elements", func() {
			var (
				n int
			)

			BeforeEach(func() {
				n = 16258
			})

			It1Second("returns 1", func() {
				go func() {
					for i := 0; i < n; i++ {
						c <- RoleAssignment{}
					}
					close(c)
				}()

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(n))
			})
		})

	})
})

func It1Second(text string, f func()) {
	It(text, func(done Done) {
		defer close(done)
		f()
	}, 1)
}
