package scheduler_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager/lagertest"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/atc/db"
	"github.com/concourse/concourse/atc/db/dbfakes"
	. "github.com/concourse/concourse/atc/scheduler"
	"github.com/concourse/concourse/atc/scheduler/algorithm"
	"github.com/concourse/concourse/atc/scheduler/schedulerfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scheduler", func() {
	var (
		fakeAlgorithm    *schedulerfakes.FakeAlgorithm
		fakeBuildStarter *schedulerfakes.FakeBuildStarter

		scheduler *Scheduler

		disaster error
	)

	BeforeEach(func() {
		fakeAlgorithm = new(schedulerfakes.FakeAlgorithm)
		fakeBuildStarter = new(schedulerfakes.FakeBuildStarter)

		scheduler = &Scheduler{
			Algorithm:    fakeAlgorithm,
			BuildStarter: fakeBuildStarter,
		}

		disaster = errors.New("bad thing")
	})

	Describe("Schedule", func() {
		var (
			fakePipeline *dbfakes.FakePipeline
			fakeJob      *dbfakes.FakeJob
			fakeResource *dbfakes.FakeResource
			scheduleErr  error

			expectedResources db.Resources
			expectedJobIDs    algorithm.NameToIDMap
		)

		BeforeEach(func() {
			fakePipeline = new(dbfakes.FakePipeline)
			fakePipeline.NameReturns("fake-pipeline")

			fakeResource = new(dbfakes.FakeResource)
			fakeResource.NameReturns("some-resource")

			expectedResources = db.Resources{fakeResource}
			expectedJobIDs = algorithm.NameToIDMap{"j1": 1}
		})

		JustBeforeEach(func() {
			var waiter interface{ Wait() }

			scheduleErr = scheduler.Schedule(
				lagertest.NewTestLogger("test"),
				fakePipeline,
				fakeJob,
				expectedResources,
				expectedJobIDs,
			)
			if waiter != nil {
				waiter.Wait()
			}
		})

		Context("when the job has no inputs", func() {
			BeforeEach(func() {
				fakeJob = new(dbfakes.FakeJob)
				fakeJob.NameReturns("some-job-1")
			})

			Context("when computing the inputs fails", func() {
				BeforeEach(func() {
					fakeAlgorithm.ComputeReturns(nil, false, false, disaster)
				})

				It("returns the error", func() {
					Expect(scheduleErr).To(Equal(fmt.Errorf("compute inputs: %w", disaster)))
				})
			})

			Context("when computing the inputs succeeds", func() {
				var expectedInputMapping db.InputMapping

				BeforeEach(func() {
					expectedInputMapping = map[string]db.InputResult{
						"input-1": db.InputResult{
							Input: &db.AlgorithmInput{
								AlgorithmVersion: db.AlgorithmVersion{
									ResourceID: 1,
									Version:    db.ResourceVersion("1"),
								},
								FirstOccurrence: true,
							},
						},
					}

					fakeAlgorithm.ComputeReturns(expectedInputMapping, true, false, nil)
				})

				It("computed the inputs", func() {
					Expect(fakeAlgorithm.ComputeCallCount()).To(Equal(1))
					actualJob, resources, relatedJobs := fakeAlgorithm.ComputeArgsForCall(0)
					Expect(actualJob.Name()).To(Equal(fakeJob.Name()))
					Expect(resources).To(Equal(expectedResources))
					Expect(relatedJobs).To(Equal(expectedJobIDs))
				})

				Context("when the algorithm can run again", func() {
					BeforeEach(func() {
						fakeAlgorithm.ComputeReturns(expectedInputMapping, true, true, nil)
					})

					It("requests schedule on the pipeline", func() {
						Expect(fakeJob.RequestScheduleCallCount()).To(Equal(1))
					})
				})

				Context("when the algorithm can not compute a next set of inputs", func() {
					BeforeEach(func() {
						fakeAlgorithm.ComputeReturns(expectedInputMapping, true, false, nil)
					})

					It("does not request schedule on the pipeline", func() {
						Expect(fakeJob.RequestScheduleCallCount()).To(Equal(0))
					})
				})

				Context("when saving the next input mapping fails", func() {
					BeforeEach(func() {
						fakeJob.SaveNextInputMappingReturns(disaster)
					})

					It("returns the error", func() {
						Expect(scheduleErr).To(Equal(fmt.Errorf("save next input mapping: %w", disaster)))
					})
				})

				Context("when saving the next input mapping succeeds", func() {
					BeforeEach(func() {
						fakeJob.SaveNextInputMappingReturns(nil)
					})

					It("saved the next input mapping", func() {
						Expect(fakeJob.SaveNextInputMappingCallCount()).To(Equal(1))
						actualInputMapping, resolved := fakeJob.SaveNextInputMappingArgsForCall(0)
						Expect(actualInputMapping).To(Equal(expectedInputMapping))
						Expect(resolved).To(BeTrue())
					})

					Context("when getting the full next build inputs fails", func() {
						BeforeEach(func() {
							fakeJob.GetFullNextBuildInputsReturns(nil, false, disaster)
						})

						It("returns the error", func() {
							Expect(scheduleErr).To(Equal(fmt.Errorf("get next build inputs: %w", disaster)))
						})
					})

					Context("when getting the full next build inputs succeeds", func() {
						BeforeEach(func() {
							fakeJob.GetFullNextBuildInputsReturns([]db.BuildInput{}, true, nil)
						})

						Context("when starting pending builds for job fails", func() {
							BeforeEach(func() {
								fakeBuildStarter.TryStartPendingBuildsForJobReturns(disaster)
							})

							It("returns the error", func() {
								Expect(scheduleErr).To(Equal(disaster))
							})

							It("started all pending builds", func() {
								Expect(fakeBuildStarter.TryStartPendingBuildsForJobCallCount()).To(Equal(1))
								_, actualPipeline, actualJob, actualResources, relatedJobs := fakeBuildStarter.TryStartPendingBuildsForJobArgsForCall(0)
								Expect(actualPipeline.Name()).To(Equal("fake-pipeline"))
								Expect(actualJob.Name()).To(Equal(fakeJob.Name()))
								Expect(actualResources).To(Equal(db.Resources{fakeResource}))
								Expect(relatedJobs).To(Equal(expectedJobIDs))
							})
						})

						Context("when starting all pending builds succeeds", func() {
							BeforeEach(func() {
								fakeBuildStarter.TryStartPendingBuildsForJobReturns(nil)
							})

							It("returns no error", func() {
								Expect(scheduleErr).NotTo(HaveOccurred())
							})

							It("didn't create a pending build", func() {
								//TODO: create a positive test case for this
								Expect(fakeJob.EnsurePendingBuildExistsCallCount()).To(BeZero())
							})
						})
					})
				})

				It("didn't mark the job as having new inputs", func() {
					Expect(fakeJob.SetHasNewInputsCallCount()).To(BeZero())
				})
			})
		})

		Context("when the job has one trigger: true input", func() {
			BeforeEach(func() {
				fakeJob = new(dbfakes.FakeJob)
				fakeJob.NameReturns("some-job")
				fakeJob.ConfigReturns(atc.JobConfig{
					Plan: atc.PlanSequence{
						{Get: "a", Trigger: true},
						{Get: "b", Trigger: false},
					},
				})

				fakeBuildStarter.TryStartPendingBuildsForJobReturns(nil)
				fakeJob.SaveNextInputMappingReturns(nil)
			})

			Context("when no input mapping is found", func() {
				BeforeEach(func() {
					fakeAlgorithm.ComputeReturns(db.InputMapping{}, false, false, nil)
				})

				It("starts all pending builds and returns no error", func() {
					Expect(fakeBuildStarter.TryStartPendingBuildsForJobCallCount()).To(Equal(1))
					Expect(scheduleErr).NotTo(HaveOccurred())
				})

				It("didn't create a pending build", func() {
					Expect(fakeJob.EnsurePendingBuildExistsCallCount()).To(BeZero())
				})

				It("didn't mark the job as having new inputs", func() {
					Expect(fakeJob.SetHasNewInputsCallCount()).To(BeZero())
				})
			})

			Context("when no first occurrence input has trigger: true", func() {
				BeforeEach(func() {
					fakeJob.GetFullNextBuildInputsReturns([]db.BuildInput{
						{
							Name:            "a",
							Version:         atc.Version{"ref": "v1"},
							ResourceID:      11,
							FirstOccurrence: false,
						},
						{
							Name:            "b",
							Version:         atc.Version{"ref": "v2"},
							ResourceID:      12,
							FirstOccurrence: true,
						},
					}, true, nil)
				})

				It("starts all pending builds and returns no error", func() {
					Expect(fakeBuildStarter.TryStartPendingBuildsForJobCallCount()).To(Equal(1))
					Expect(scheduleErr).NotTo(HaveOccurred())
				})

				It("didn't create a pending build", func() {
					Expect(fakeJob.EnsurePendingBuildExistsCallCount()).To(BeZero())
				})

				Context("when the job does not have new inputs since before", func() {
					BeforeEach(func() {
						fakeJob.HasNewInputsReturns(false)
					})

					Context("when marking job as having new input fails", func() {
						BeforeEach(func() {
							fakeJob.SetHasNewInputsReturns(disaster)
						})

						It("returns the error", func() {
							Expect(scheduleErr).To(Equal(fmt.Errorf("set has new inputs: %w", disaster)))
						})
					})

					Context("when marking job as having new input succeeds", func() {
						BeforeEach(func() {
							fakeJob.SetHasNewInputsReturns(nil)
						})

						It("did the needful", func() {
							Expect(fakeJob.SetHasNewInputsCallCount()).To(Equal(1))
							Expect(fakeJob.SetHasNewInputsArgsForCall(0)).To(Equal(true))
						})
					})
				})

				Context("when the job has new inputs since before", func() {
					BeforeEach(func() {
						fakeJob.HasNewInputsReturns(true)
					})

					It("doesn't mark the job as having new inputs", func() {
						Expect(fakeJob.SetHasNewInputsCallCount()).To(BeZero())
					})
				})
			})

			Context("when a first occurrence input has trigger: true", func() {
				BeforeEach(func() {
					fakeJob.GetFullNextBuildInputsReturns([]db.BuildInput{
						{
							Name:            "a",
							Version:         atc.Version{"ref": "v1"},
							ResourceID:      11,
							FirstOccurrence: true,
						},
						{
							Name:            "b",
							Version:         atc.Version{"ref": "v2"},
							ResourceID:      12,
							FirstOccurrence: false,
						},
					}, true, nil)
				})

				Context("when creating a pending build fails", func() {
					BeforeEach(func() {
						fakeJob.EnsurePendingBuildExistsReturns(disaster)
					})

					It("returns the error", func() {
						Expect(scheduleErr).To(Equal(fmt.Errorf("ensure pending build exists: %w", disaster)))
					})

					It("created a pending build for the right job", func() {
						Expect(fakeJob.EnsurePendingBuildExistsCallCount()).To(Equal(1))
					})
				})

				Context("when creating a pending build succeeds", func() {
					BeforeEach(func() {
						fakeJob.EnsurePendingBuildExistsReturns(nil)
					})

					It("starts all pending builds and returns no error", func() {
						Expect(fakeBuildStarter.TryStartPendingBuildsForJobCallCount()).To(Equal(1))
						Expect(scheduleErr).NotTo(HaveOccurred())
					})
				})
			})

			Context("when no first occurrence", func() {
				BeforeEach(func() {
					fakeJob.GetFullNextBuildInputsReturns([]db.BuildInput{
						{
							Name:            "a",
							Version:         atc.Version{"ref": "v1"},
							ResourceID:      11,
							FirstOccurrence: false,
						},
						{
							Name:            "b",
							Version:         atc.Version{"ref": "v2"},
							ResourceID:      12,
							FirstOccurrence: false,
						},
					}, true, nil)
				})

				Context("when job had new inputs", func() {
					BeforeEach(func() {
						fakeJob.HasNewInputsReturns(true)
					})

					It("marks the job as not having new inputs", func() {
						Expect(fakeJob.SetHasNewInputsCallCount()).To(Equal(1))
						Expect(fakeJob.SetHasNewInputsArgsForCall(0)).To(Equal(false))
					})
				})

				Context("when job did not have new inputs", func() {
					BeforeEach(func() {
						fakeJob.HasNewInputsReturns(false)
					})

					It("doesn't mark the the job as not having new inputs again", func() {
						Expect(fakeJob.SetHasNewInputsCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
