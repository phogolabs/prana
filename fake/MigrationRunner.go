// This file was generated by counterfeiter
package fake

import (
	"sync"

	"github.com/phogolabs/prana/sqlmigr"
)

type MigrationRunner struct {
	RunStub        func(item *sqlmigr.Migration) error
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		item *sqlmigr.Migration
	}
	runReturns struct {
		result1 error
	}
	RevertStub        func(item *sqlmigr.Migration) error
	revertMutex       sync.RWMutex
	revertArgsForCall []struct {
		item *sqlmigr.Migration
	}
	revertReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *MigrationRunner) Run(item *sqlmigr.Migration) error {
	fake.runMutex.Lock()
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		item *sqlmigr.Migration
	}{item})
	fake.recordInvocation("Run", []interface{}{item})
	fake.runMutex.Unlock()
	if fake.RunStub != nil {
		return fake.RunStub(item)
	}
	return fake.runReturns.result1
}

func (fake *MigrationRunner) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *MigrationRunner) RunArgsForCall(i int) *sqlmigr.Migration {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return fake.runArgsForCall[i].item
}

func (fake *MigrationRunner) RunReturns(result1 error) {
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *MigrationRunner) Revert(item *sqlmigr.Migration) error {
	fake.revertMutex.Lock()
	fake.revertArgsForCall = append(fake.revertArgsForCall, struct {
		item *sqlmigr.Migration
	}{item})
	fake.recordInvocation("Revert", []interface{}{item})
	fake.revertMutex.Unlock()
	if fake.RevertStub != nil {
		return fake.RevertStub(item)
	}
	return fake.revertReturns.result1
}

func (fake *MigrationRunner) RevertCallCount() int {
	fake.revertMutex.RLock()
	defer fake.revertMutex.RUnlock()
	return len(fake.revertArgsForCall)
}

func (fake *MigrationRunner) RevertArgsForCall(i int) *sqlmigr.Migration {
	fake.revertMutex.RLock()
	defer fake.revertMutex.RUnlock()
	return fake.revertArgsForCall[i].item
}

func (fake *MigrationRunner) RevertReturns(result1 error) {
	fake.RevertStub = nil
	fake.revertReturns = struct {
		result1 error
	}{result1}
}

func (fake *MigrationRunner) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	fake.revertMutex.RLock()
	defer fake.revertMutex.RUnlock()
	return fake.invocations
}

func (fake *MigrationRunner) recordInvocation(key string, args []interface{}) {
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

var _ sqlmigr.MigrationRunner = new(MigrationRunner)
