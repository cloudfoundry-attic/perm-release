package reporter

import (
	"fmt"
	"io"

	"sort"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
)

func GenerateReport(w io.Writer, roleAssignments <-chan models.RoleAssignment, errors <-chan error) {
	count := ComputeNumberAssignments(roleAssignments)
	errorSummary := ComputeErrors(errors)

	fmt.Fprintf(w, "\nReport\n==========================================\n")
	fmt.Fprintf(w, "Number of role assignments: %d.\n", count)
	fmt.Fprintf(w, "Total errors: %d.\n", errorSummary.Count())
	fmt.Fprintf(w, "\nSummary\n==========================================\n")

	var perTypeKeys []string
	for key := range errorSummary.PerType {
		perTypeKeys = append(perTypeKeys, key)
	}
	sort.Strings(perTypeKeys)
	for _, endpoint := range perTypeKeys {
		messageCount := errorSummary.PerType[endpoint]
		fmt.Fprintf(w, "For %s:\n", endpoint)
		var messageKeys []string
		for messageKey := range messageCount {
			messageKeys = append(messageKeys, messageKey)
		}
		sort.Strings(messageKeys)
		for _, messageKey := range messageKeys {
			count := messageCount[messageKey]
			fmt.Fprintf(w, "- %3d %s\n", count, messageKey)
		}
	}

	if len(errorSummary.Other) > 0 {
		fmt.Fprint(w, "Other errors:\n")
		var otherMessageKeys []string
		for messageKey := range errorSummary.Other {
			otherMessageKeys = append(otherMessageKeys, messageKey)
		}
		sort.Strings(otherMessageKeys)
		for _, messageKey := range otherMessageKeys {
			messageCount := errorSummary.Other[messageKey]
			fmt.Fprintf(w, "- %3d %s", messageCount, messageKey)
		}
	}
}

func ComputeNumberAssignments(roleAssignments <-chan models.RoleAssignment) int {
	var count int

	for range roleAssignments {
		count++
	}

	return count
}

type ErrorSummary struct {
	Other   map[string]int
	PerType map[string]map[string]int
}

func NewErrorSummary() ErrorSummary {
	summary := ErrorSummary{}
	summary.Other = make(map[string]int)
	summary.PerType = make(map[string]map[string]int)
	return summary
}

func (e *ErrorSummary) AddPerTypeError(entity, errorMessage string) {
	if _, ok := e.PerType[entity]; !ok {
		e.PerType[entity] = make(map[string]int)
	}
	e.PerType[entity][errorMessage] += 1
}

func (e *ErrorSummary) AddOtherError(errorMessage string) {
	e.Other[errorMessage] += 1
}

func (e *ErrorSummary) Count() int {
	total := 0
	for _, subMap := range e.PerType {
		for _, count := range subMap {
			total += count
		}
	}
	total += len(e.Other)
	return total
}

func ComputeErrors(errors <-chan error) ErrorSummary {
	summary := NewErrorSummary()

	for errorItem := range errors {
		switch errorEvent := errorItem.(type) {
		case *models.ErrorEvent:
			summary.AddPerTypeError(errorEvent.EntityType, errorEvent.Error())

		default:
			summary.Other[errorItem.Error()] += 1
		}
	}
	return summary
}
