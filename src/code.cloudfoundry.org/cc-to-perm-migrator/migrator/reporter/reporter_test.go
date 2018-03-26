package reporter_test

import (
	"errors"

	. "code.cloudfoundry.org/cc-to-perm-migrator/migrator/reporter"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = ginkgo.Describe("Reporter", func() {
	var c chan models.RoleAssignment
	var e chan error
	ginkgo.BeforeEach(func() {
		c = make(chan models.RoleAssignment, 1000)
		e = make(chan error, 1000)
	})

	ginkgo.Describe("#GenerateReport", func() {
		var (
			b       *Buffer
			subject *Reporter
		)

		ginkgo.BeforeEach(func() {
			b = NewBuffer()
		})

		ginkgo.Context("when the channel is sent 2 elements", func() {
			It1Second("gives success and error counts", func() {
				c <- models.RoleAssignment{}
				c <- models.RoleAssignment{}
				e <- errors.New("There has been a problem")
				close(c)
				close(e)

				subject.GenerateReport(b, c, e)
				Expect(b).To(Say("Number of role assignments: 2\\."))
				Expect(b).To(Say("Total errors: 1\\."))
			})
		})
		ginkgo.Context("when the channel receives errors and ErrorEvents", func() {
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
				subject.GenerateReport(b, c, e)
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

	ginkgo.Describe(".ComputeNumberAssignments", func() {
		ginkgo.Context("when the channel is closed without sending anything onto it", func() {
			It1Second("returns 0", func() {
				close(c)
				count := ComputeNumberAssignments(c)

				Expect(count).To(BeZero())
			})
		})

		ginkgo.Context("when the channel is sent one element", func() {
			It1Second("returns 1", func() {
				c <- models.RoleAssignment{}
				close(c)

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(1))
			})
		})

		ginkgo.Context("when the channel is sent 2 elements", func() {
			It1Second("returns 1", func() {
				c <- models.RoleAssignment{}
				c <- models.RoleAssignment{}
				close(c)

				count := ComputeNumberAssignments(c)

				Expect(count).To(Equal(2))
			})
		})

		ginkgo.Context("when the channel is sent more elements", func() {
			var (
				n int
			)

			ginkgo.BeforeEach(func() {
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
	ginkgo.Describe("ErrorSummary", func() {
		var summary ErrorSummary
		ginkgo.BeforeEach(func() {
			summary = NewErrorSummary()
		})
		ginkgo.It("maintains summary and counts", func() {
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
	ginkgo.It(text, func(done ginkgo.Done) {
		defer close(done)
		f()
	}, 1)
}
