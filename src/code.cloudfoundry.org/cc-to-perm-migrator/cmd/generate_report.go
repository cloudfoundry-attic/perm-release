package cmd

import (
	"fmt"
	"io"
)

func GenerateReport(w io.Writer, roleAssignments <-chan RoleAssignment) {
	count := ComputeNumberAssignments(roleAssignments)

	fmt.Fprintf(w, "\nReport\n==========================================\n")
	fmt.Fprintf(w, "Number of role assignments: %d.\n", count)
}

func ComputeNumberAssignments(roleAssignments <-chan RoleAssignment) int {
	var count int

	for range roleAssignments {
		count++
	}

	return count
}
