// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/concourse/atc/builds"
	"github.com/concourse/atc/db"
)

type FakeTrackerDB struct {
	GetAllStartedBuildsStub        func() ([]db.Build, error)
	getAllStartedBuildsMutex       sync.RWMutex
	getAllStartedBuildsArgsForCall []struct{}
	getAllStartedBuildsReturns struct {
		result1 []db.Build
		result2 error
	}
	ErrorBuildStub        func(buildID int, err error) error
	errorBuildMutex       sync.RWMutex
	errorBuildArgsForCall []struct {
		buildID int
		err     error
	}
	errorBuildReturns struct {
		result1 error
	}
}

func (fake *FakeTrackerDB) GetAllStartedBuilds() ([]db.Build, error) {
	fake.getAllStartedBuildsMutex.Lock()
	fake.getAllStartedBuildsArgsForCall = append(fake.getAllStartedBuildsArgsForCall, struct{}{})
	fake.getAllStartedBuildsMutex.Unlock()
	if fake.GetAllStartedBuildsStub != nil {
		return fake.GetAllStartedBuildsStub()
	} else {
		return fake.getAllStartedBuildsReturns.result1, fake.getAllStartedBuildsReturns.result2
	}
}

func (fake *FakeTrackerDB) GetAllStartedBuildsCallCount() int {
	fake.getAllStartedBuildsMutex.RLock()
	defer fake.getAllStartedBuildsMutex.RUnlock()
	return len(fake.getAllStartedBuildsArgsForCall)
}

func (fake *FakeTrackerDB) GetAllStartedBuildsReturns(result1 []db.Build, result2 error) {
	fake.GetAllStartedBuildsStub = nil
	fake.getAllStartedBuildsReturns = struct {
		result1 []db.Build
		result2 error
	}{result1, result2}
}

func (fake *FakeTrackerDB) ErrorBuild(buildID int, err error) error {
	fake.errorBuildMutex.Lock()
	fake.errorBuildArgsForCall = append(fake.errorBuildArgsForCall, struct {
		buildID int
		err     error
	}{buildID, err})
	fake.errorBuildMutex.Unlock()
	if fake.ErrorBuildStub != nil {
		return fake.ErrorBuildStub(buildID, err)
	} else {
		return fake.errorBuildReturns.result1
	}
}

func (fake *FakeTrackerDB) ErrorBuildCallCount() int {
	fake.errorBuildMutex.RLock()
	defer fake.errorBuildMutex.RUnlock()
	return len(fake.errorBuildArgsForCall)
}

func (fake *FakeTrackerDB) ErrorBuildArgsForCall(i int) (int, error) {
	fake.errorBuildMutex.RLock()
	defer fake.errorBuildMutex.RUnlock()
	return fake.errorBuildArgsForCall[i].buildID, fake.errorBuildArgsForCall[i].err
}

func (fake *FakeTrackerDB) ErrorBuildReturns(result1 error) {
	fake.ErrorBuildStub = nil
	fake.errorBuildReturns = struct {
		result1 error
	}{result1}
}

var _ builds.TrackerDB = new(FakeTrackerDB)
