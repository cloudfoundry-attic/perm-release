package cmd

import (
	"fmt"
	"io"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator"
)

func GenerateReport(w io.Writer, roleAssignments <-chan migrator.RoleAssignment, errors <-chan error) {
	count := ComputeNumberAssignments(roleAssignments)
	errorCount := ComputeErrors(errors)

	fmt.Fprintf(w, "\nReport\n==========================================\n")
	fmt.Fprintf(w, "Number of role assignments: %d.\n", count)
	fmt.Fprintf(w, "Number of errors: %d.\n", errorCount)
}

func ComputeNumberAssignments(roleAssignments <-chan migrator.RoleAssignment) int {
	var count int

	for range roleAssignments {
		count++
	}

	return count
}

func ComputeErrors(errors <-chan error) int {
	var count int

	for range errors {
		count++
	}

	return count
}
