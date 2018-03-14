package cmd_test

import (
	"errors"

	. "code.cloudfoundry.org/cc-to-perm-migrator/cmd"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Generating a report", func() {
	var c chan migrator.RoleAssignment
	var e chan error
	BeforeEach(func() {
		c = make(chan migrator.RoleAssignment, 1000)
		e = make(chan error, 1000)
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
				c <- migrator.RoleAssignment{}
				c <- migrator.RoleAssignment{}
				e <- errors.New("There has been a problem")
				close(c)
				close(e)

				GenerateReport(b, c, e)

				Expect(b).To(Say("Number of role assignments: 2\\."))
				Expect(b).To(Say("Number of errors: 1\\."))
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
				c <- migrator.RoleAssignment{}
				close(c)

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(1))
			})
		})

		Context("when the channel is sent 2 elements", func() {
			It1Second("returns 1", func() {
				c <- migrator.RoleAssignment{}
				c <- migrator.RoleAssignment{}
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
						c <- migrator.RoleAssignment{}
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
