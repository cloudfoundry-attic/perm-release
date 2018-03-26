package reporter_test

import (
	"errors"

	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator/reporter"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Generating a report", func() {
	var c chan models.RoleAssignment
	var e chan error
	BeforeEach(func() {
		c = make(chan models.RoleAssignment, 1000)
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
			It1Second("gives success and error counts", func() {
				c <- models.RoleAssignment{}
				c <- models.RoleAssignment{}
				e <- errors.New("There has been a problem")
				close(c)
				close(e)

				GenerateReport(b, c, e)
				Expect(b).To(Say("Number of role assignments: 2\\."))
				Expect(b).To(Say("Total errors: 1\\."))
			})
		})
		Context("when the channel receives errors and ErrorEvents", func() {
			It1Second("gives an error summary", func() {
				c <- models.RoleAssignment{}
				c <- models.RoleAssignment{}
				e <- errors.New("There has been a problem")
				for i := 0; i < 3; i++ {
					e <- &models.ErrorEvent{
						Cause:      errors.New("failed-to-fetch-orgs"),
						GUID:       "",
						EntityType: "/v2/organizations",
					}
				}
				for i := 0; i < 4; i++ {
					e <- &models.ErrorEvent{
						Cause:      errors.New("failed-to-decode-response"),
						GUID:       "",
						EntityType: "/v2/organizations/org-guid/user_roles",
					}
				}
				for i := 0; i < 2; i++ {
					e <- &models.ErrorEvent{
						Cause:      errors.New("something-happened"),
						GUID:       "org-guid",
						EntityType: "/v2/organizations/org-guid/user_roles",
					}
				}
				close(c)
				close(e)
				GenerateReport(b, c, e)
				Expect(b).To(Say("Number of role assignments: 2\\."))
				Expect(b).To(Say("Total errors: 10\\."))
				Expect(b).To(Say("Summary"))
				Expect(b).To(Say("For /v2/organizations:"))
				Expect(b).To(Say("3 failed-to-fetch-orgs"))
				Expect(b).To(Say("For /v2/organizations/org-guid/user_roles:"))
				Expect(b).To(Say("4 failed-to-decode-response"))
				Expect(b).To(Say("2 something-happened"))
				Expect(b).To(Say("Other errors"))
				Expect(b).To(Say("1 There has been a problem"))
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
				c <- models.RoleAssignment{}
				close(c)

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(1))
			})
		})

		Context("when the channel is sent 2 elements", func() {
			It1Second("returns 1", func() {
				c <- models.RoleAssignment{}
				c <- models.RoleAssignment{}
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
						c <- models.RoleAssignment{}
					}
					close(c)
				}()

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(n))
			})
		})

	})
	Describe("ErrorSummary", func() {
		var summary ErrorSummary
		BeforeEach(func() {
			summary = NewErrorSummary()
		})
		It("maintains summary and counts", func() {
			summary.AddPerTypeError("one", "two")
			summary.AddPerTypeError("one", "three")
			summary.AddPerTypeError("two", "three")
			summary.AddOtherError("four")
			Expect(len(summary.PerType)).To(Equal(2))
			Expect(len(summary.Other)).To(Equal(1))
			Expect(summary.Count()).To(Equal(4))
		})
	})
})

func It1Second(text string, f func()) {
	It(text, func(done Done) {
		defer close(done)
		f()
	}, 1)
}
