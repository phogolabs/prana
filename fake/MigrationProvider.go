// Code generated by counterfeiter. DO NOT EDIT.
package fake

import (
	"sync"

	"github.com/phogolabs/prana/sqlmigr"
)

type MigrationProvider struct {
	MigrationsStub        func() ([]*sqlmigr.Migration, error)
	migrationsMutex       sync.RWMutex
	migrationsArgsForCall []struct{}
	migrationsReturns     struct {
		result1 []*sqlmigr.Migration
		result2 error
	}
	migrationsReturnsOnCall map[int]struct {
		result1 []*sqlmigr.Migration
		result2 error
	}
	InsertStub        func(item *sqlmigr.Migration) error
	insertMutex       sync.RWMutex
	insertArgsForCall []struct {
		item *sqlmigr.Migration
	}
	insertReturns struct {
		result1 error
	}
	insertReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteStub        func(item *sqlmigr.Migration) error
	deleteMutex       sync.RWMutex
	deleteArgsForCall []struct {
		item *sqlmigr.Migration
	}
	deleteReturns struct {
		result1 error
	}
	deleteReturnsOnCall map[int]struct {
		result1 error
	}
	ExistsStub        func(item *sqlmigr.Migration) bool
	existsMutex       sync.RWMutex
	existsArgsForCall []struct {
		item *sqlmigr.Migration
	}
	existsReturns struct {
		result1 bool
	}
	existsReturnsOnCall map[int]struct {
		result1 bool
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *MigrationProvider) Migrations() ([]*sqlmigr.Migration, error) {
	fake.migrationsMutex.Lock()
	ret, specificReturn := fake.migrationsReturnsOnCall[len(fake.migrationsArgsForCall)]
	fake.migrationsArgsForCall = append(fake.migrationsArgsForCall, struct{}{})
	fake.recordInvocation("Migrations", []interface{}{})
	fake.migrationsMutex.Unlock()
	if fake.MigrationsStub != nil {
		return fake.MigrationsStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.migrationsReturns.result1, fake.migrationsReturns.result2
}

func (fake *MigrationProvider) MigrationsCallCount() int {
	fake.migrationsMutex.RLock()
	defer fake.migrationsMutex.RUnlock()
	return len(fake.migrationsArgsForCall)
}

func (fake *MigrationProvider) MigrationsReturns(result1 []*sqlmigr.Migration, result2 error) {
	fake.MigrationsStub = nil
	fake.migrationsReturns = struct {
		result1 []*sqlmigr.Migration
		result2 error
	}{result1, result2}
}

func (fake *MigrationProvider) MigrationsReturnsOnCall(i int, result1 []*sqlmigr.Migration, result2 error) {
	fake.MigrationsStub = nil
	if fake.migrationsReturnsOnCall == nil {
		fake.migrationsReturnsOnCall = make(map[int]struct {
			result1 []*sqlmigr.Migration
			result2 error
		})
	}
	fake.migrationsReturnsOnCall[i] = struct {
		result1 []*sqlmigr.Migration
		result2 error
	}{result1, result2}
}

func (fake *MigrationProvider) Insert(item *sqlmigr.Migration) error {
	fake.insertMutex.Lock()
	ret, specificReturn := fake.insertReturnsOnCall[len(fake.insertArgsForCall)]
	fake.insertArgsForCall = append(fake.insertArgsForCall, struct {
		item *sqlmigr.Migration
	}{item})
	fake.recordInvocation("Insert", []interface{}{item})
	fake.insertMutex.Unlock()
	if fake.InsertStub != nil {
		return fake.InsertStub(item)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.insertReturns.result1
}

func (fake *MigrationProvider) InsertCallCount() int {
	fake.insertMutex.RLock()
	defer fake.insertMutex.RUnlock()
	return len(fake.insertArgsForCall)
}

func (fake *MigrationProvider) InsertArgsForCall(i int) *sqlmigr.Migration {
	fake.insertMutex.RLock()
	defer fake.insertMutex.RUnlock()
	return fake.insertArgsForCall[i].item
}

func (fake *MigrationProvider) InsertReturns(result1 error) {
	fake.InsertStub = nil
	fake.insertReturns = struct {
		result1 error
	}{result1}
}

func (fake *MigrationProvider) InsertReturnsOnCall(i int, result1 error) {
	fake.InsertStub = nil
	if fake.insertReturnsOnCall == nil {
		fake.insertReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.insertReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MigrationProvider) Delete(item *sqlmigr.Migration) error {
	fake.deleteMutex.Lock()
	ret, specificReturn := fake.deleteReturnsOnCall[len(fake.deleteArgsForCall)]
	fake.deleteArgsForCall = append(fake.deleteArgsForCall, struct {
		item *sqlmigr.Migration
	}{item})
	fake.recordInvocation("Delete", []interface{}{item})
	fake.deleteMutex.Unlock()
	if fake.DeleteStub != nil {
		return fake.DeleteStub(item)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.deleteReturns.result1
}

func (fake *MigrationProvider) DeleteCallCount() int {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return len(fake.deleteArgsForCall)
}

func (fake *MigrationProvider) DeleteArgsForCall(i int) *sqlmigr.Migration {
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	return fake.deleteArgsForCall[i].item
}

func (fake *MigrationProvider) DeleteReturns(result1 error) {
	fake.DeleteStub = nil
	fake.deleteReturns = struct {
		result1 error
	}{result1}
}

func (fake *MigrationProvider) DeleteReturnsOnCall(i int, result1 error) {
	fake.DeleteStub = nil
	if fake.deleteReturnsOnCall == nil {
		fake.deleteReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *MigrationProvider) Exists(item *sqlmigr.Migration) bool {
	fake.existsMutex.Lock()
	ret, specificReturn := fake.existsReturnsOnCall[len(fake.existsArgsForCall)]
	fake.existsArgsForCall = append(fake.existsArgsForCall, struct {
		item *sqlmigr.Migration
	}{item})
	fake.recordInvocation("Exists", []interface{}{item})
	fake.existsMutex.Unlock()
	if fake.ExistsStub != nil {
		return fake.ExistsStub(item)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.existsReturns.result1
}

func (fake *MigrationProvider) ExistsCallCount() int {
	fake.existsMutex.RLock()
	defer fake.existsMutex.RUnlock()
	return len(fake.existsArgsForCall)
}

func (fake *MigrationProvider) ExistsArgsForCall(i int) *sqlmigr.Migration {
	fake.existsMutex.RLock()
	defer fake.existsMutex.RUnlock()
	return fake.existsArgsForCall[i].item
}

func (fake *MigrationProvider) ExistsReturns(result1 bool) {
	fake.ExistsStub = nil
	fake.existsReturns = struct {
		result1 bool
	}{result1}
}

func (fake *MigrationProvider) ExistsReturnsOnCall(i int, result1 bool) {
	fake.ExistsStub = nil
	if fake.existsReturnsOnCall == nil {
		fake.existsReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.existsReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *MigrationProvider) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.migrationsMutex.RLock()
	defer fake.migrationsMutex.RUnlock()
	fake.insertMutex.RLock()
	defer fake.insertMutex.RUnlock()
	fake.deleteMutex.RLock()
	defer fake.deleteMutex.RUnlock()
	fake.existsMutex.RLock()
	defer fake.existsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *MigrationProvider) recordInvocation(key string, args []interface{}) {
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

var _ sqlmigr.MigrationProvider = new(MigrationProvider)
