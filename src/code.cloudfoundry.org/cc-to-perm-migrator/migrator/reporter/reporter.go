package reporter

import (
	"fmt"
	"io"

	"sort"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
)

//go:generate counterfeiter io.Writer

type Reporter struct{}

func (r *Reporter) GenerateReport(w io.Writer, orgs []models.Organization, spaces []models.Space, errs []error) {
	numAssignments := countNumAssignments(orgs, spaces)

	fmt.Fprint(w, "Report\n==========================================\n")
	fmt.Fprintf(w, "Number of role assignments: %d\n", numAssignments)
	fmt.Fprintf(w, "Total errors: %d\n\n", len(errs))
	fmt.Fprint(w, "Summary\n==========================================\n")

	errorSummary := computeErrors(errs)
	var perTypeKeys []string
	for key := range errorSummary.perType {
		perTypeKeys = append(perTypeKeys, key)
	}
	sort.Strings(perTypeKeys)

	for _, endpoint := range perTypeKeys {
		messageCount := errorSummary.perType[endpoint]
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

	if len(errorSummary.other) > 0 {
		fmt.Fprint(w, "Other errors:\n")
		var otherMessageKeys []string
		for messageKey := range errorSummary.other {
			otherMessageKeys = append(otherMessageKeys, messageKey)
		}
		sort.Strings(otherMessageKeys)
		for _, messageKey := range otherMessageKeys {
			messageCount := errorSummary.other[messageKey]
			fmt.Fprintf(w, "- %3d %s", messageCount, messageKey)
		}
	}
}

type errorSummary struct {
	other   map[string]int
	perType map[string]map[string]int
}

func (e *errorSummary) addPerTypeError(entity, errorMessage string) {
	if _, ok := e.perType[entity]; !ok {
		e.perType[entity] = make(map[string]int)
	}
	e.perType[entity][errorMessage] += 1
}

func countNumAssignments(orgs []models.Organization, spaces []models.Space) int {
	var numAssignments int

	for _, org := range orgs {
		numAssignments += len(org.Assignments)
	}

	for _, space := range spaces {
		numAssignments += len(space.Assignments)
	}

	return numAssignments
}

func computeErrors(errs []error) errorSummary {
	summary := errorSummary{
		other:   make(map[string]int),
		perType: make(map[string]map[string]int),
	}

	for _, errorItem := range errs {
		switch errorEvent := errorItem.(type) {
		case *models.ErrorEvent:
			summary.addPerTypeError(errorEvent.EntityType, errorEvent.Error())
		default:
			summary.other[errorItem.Error()] += 1
		}
	}
	return summary
}
