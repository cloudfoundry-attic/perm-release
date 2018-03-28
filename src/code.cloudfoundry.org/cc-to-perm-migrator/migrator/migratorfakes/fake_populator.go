// Code generated by counterfeiter. DO NOT EDIT.
package migratorfakes

import (
	"sync"

	"code.cloudfoundry.org/cc-to-perm-migrator/migrator"
	"code.cloudfoundry.org/cc-to-perm-migrator/migrator/models"
	"code.cloudfoundry.org/lager"
)

type FakePopulator struct {
	PopulateOrganizationStub        func(logger lager.Logger, org models.Organization, namespace string) []error
	populateOrganizationMutex       sync.RWMutex
	populateOrganizationArgsForCall []struct {
		logger    lager.Logger
		org       models.Organization
		namespace string
	}
	populateOrganizationReturns struct {
		result1 []error
	}
	populateOrganizationReturnsOnCall map[int]struct {
		result1 []error
	}
	PopulateSpaceStub        func(logger lager.Logger, space models.Space, namespace string) []error
	populateSpaceMutex       sync.RWMutex
	populateSpaceArgsForCall []struct {
		logger    lager.Logger
		space     models.Space
		namespace string
	}
	populateSpaceReturns struct {
		result1 []error
	}
	populateSpaceReturnsOnCall map[int]struct {
		result1 []error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakePopulator) PopulateOrganization(logger lager.Logger, org models.Organization, namespace string) []error {
	fake.populateOrganizationMutex.Lock()
	ret, specificReturn := fake.populateOrganizationReturnsOnCall[len(fake.populateOrganizationArgsForCall)]
	fake.populateOrganizationArgsForCall = append(fake.populateOrganizationArgsForCall, struct {
		logger    lager.Logger
		org       models.Organization
		namespace string
	}{logger, org, namespace})
	fake.recordInvocation("PopulateOrganization", []interface{}{logger, org, namespace})
	fake.populateOrganizationMutex.Unlock()
	if fake.PopulateOrganizationStub != nil {
		return fake.PopulateOrganizationStub(logger, org, namespace)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.populateOrganizationReturns.result1
}

func (fake *FakePopulator) PopulateOrganizationCallCount() int {
	fake.populateOrganizationMutex.RLock()
	defer fake.populateOrganizationMutex.RUnlock()
	return len(fake.populateOrganizationArgsForCall)
}

func (fake *FakePopulator) PopulateOrganizationArgsForCall(i int) (lager.Logger, models.Organization, string) {
	fake.populateOrganizationMutex.RLock()
	defer fake.populateOrganizationMutex.RUnlock()
	return fake.populateOrganizationArgsForCall[i].logger, fake.populateOrganizationArgsForCall[i].org, fake.populateOrganizationArgsForCall[i].namespace
}

func (fake *FakePopulator) PopulateOrganizationReturns(result1 []error) {
	fake.PopulateOrganizationStub = nil
	fake.populateOrganizationReturns = struct {
		result1 []error
	}{result1}
}

func (fake *FakePopulator) PopulateOrganizationReturnsOnCall(i int, result1 []error) {
	fake.PopulateOrganizationStub = nil
	if fake.populateOrganizationReturnsOnCall == nil {
		fake.populateOrganizationReturnsOnCall = make(map[int]struct {
			result1 []error
		})
	}
	fake.populateOrganizationReturnsOnCall[i] = struct {
		result1 []error
	}{result1}
}

func (fake *FakePopulator) PopulateSpace(logger lager.Logger, space models.Space, namespace string) []error {
	fake.populateSpaceMutex.Lock()
	ret, specificReturn := fake.populateSpaceReturnsOnCall[len(fake.populateSpaceArgsForCall)]
	fake.populateSpaceArgsForCall = append(fake.populateSpaceArgsForCall, struct {
		logger    lager.Logger
		space     models.Space
		namespace string
	}{logger, space, namespace})
	fake.recordInvocation("PopulateSpace", []interface{}{logger, space, namespace})
	fake.populateSpaceMutex.Unlock()
	if fake.PopulateSpaceStub != nil {
		return fake.PopulateSpaceStub(logger, space, namespace)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.populateSpaceReturns.result1
}

func (fake *FakePopulator) PopulateSpaceCallCount() int {
	fake.populateSpaceMutex.RLock()
	defer fake.populateSpaceMutex.RUnlock()
	return len(fake.populateSpaceArgsForCall)
}

func (fake *FakePopulator) PopulateSpaceArgsForCall(i int) (lager.Logger, models.Space, string) {
	fake.populateSpaceMutex.RLock()
	defer fake.populateSpaceMutex.RUnlock()
	return fake.populateSpaceArgsForCall[i].logger, fake.populateSpaceArgsForCall[i].space, fake.populateSpaceArgsForCall[i].namespace
}

func (fake *FakePopulator) PopulateSpaceReturns(result1 []error) {
	fake.PopulateSpaceStub = nil
	fake.populateSpaceReturns = struct {
		result1 []error
	}{result1}
}

func (fake *FakePopulator) PopulateSpaceReturnsOnCall(i int, result1 []error) {
	fake.PopulateSpaceStub = nil
	if fake.populateSpaceReturnsOnCall == nil {
		fake.populateSpaceReturnsOnCall = make(map[int]struct {
			result1 []error
		})
	}
	fake.populateSpaceReturnsOnCall[i] = struct {
		result1 []error
	}{result1}
}

func (fake *FakePopulator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.populateOrganizationMutex.RLock()
	defer fake.populateOrganizationMutex.RUnlock()
	fake.populateSpaceMutex.RLock()
	defer fake.populateSpaceMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakePopulator) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ migrator.Populator = new(FakePopulator)
