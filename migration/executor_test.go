package migration_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/oak/fake"
	"github.com/phogolabs/oak/migration"
)

var _ = Describe("Executor", func() {
	var (
		executor  *migration.Executor
		provider  *fake.MigrationProvider
		generator *fake.MigrationGenerator
		runner    *fake.MigrationRunner
	)

	BeforeEach(func() {
		provider = &fake.MigrationProvider{}
		generator = &fake.MigrationGenerator{}
		runner = &fake.MigrationRunner{}

		executor = &migration.Executor{
			Provider:  provider,
			Generator: generator,
			Runner:    runner,
		}
	})

	Describe("Setup", func() {
		It("setups the migrations successfully", func() {
			format := "20060102150405"
			min := time.Date(1, time.January, 1970, 0, 0, 0, 0, time.UTC)

			Expect(executor.Setup()).To(Succeed())

			Expect(generator.WriteCallCount()).To(Equal(1))

			item, content := generator.WriteArgsForCall(0)
			Expect(item.Id).To(Equal(min.Format(format)))
			Expect(item.Description).To(Equal("setup"))

			data, err := ioutil.ReadAll(content.UpCommand)
			Expect(err).NotTo(HaveOccurred())

			up := &bytes.Buffer{}
			fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS migrations (")
			fmt.Fprintln(up, " id          TEXT      NOT NULL PRIMARY KEY,")
			fmt.Fprintln(up, " description TEXT      NOT NULL,")
			fmt.Fprintln(up, " created_at  TIMESTAMP NOT NULL")
			fmt.Fprintln(up, ");")
			Expect(string(data)).To(Equal(up.String()))

			data, err = ioutil.ReadAll(content.DownCommand)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("DROP TABLE IF EXISTS migrations;"))

			Expect(runner.RunCallCount()).To(Equal(1))
			item = runner.RunArgsForCall(0)

			Expect(item.Id).To(Equal(min.Format(format)))
			Expect(item.Description).To(Equal("setup"))
		})

		Context("when the generator fails", func() {
			It("returns the error", func() {
				generator.WriteReturns(fmt.Errorf("oh no!"))
				Expect(executor.Setup()).To(MatchError("oh no!"))
			})
		})

		Context("when the runner fails", func() {
			It("return the error", func() {
				runner.RunReturns(fmt.Errorf("oh no!"))
				Expect(executor.Setup()).To(MatchError("oh no!"))
			})
		})
	})

	Describe("Create", func() {
		It("creates migration successfully", func() {
			format := "20060102150405"

			migration, err := executor.Create("schema")
			Expect(err).NotTo(HaveOccurred())
			Expect(generator.CreateCallCount()).To(Equal(1))

			item := generator.CreateArgsForCall(0)
			Expect(item.Id).To(Equal(item.CreatedAt.Format(format)))
			Expect(item.Description).To(Equal("schema"))
			Expect(item).To(Equal(migration))
		})

		Context("when the generator fails", func() {
			It("returns the error", func() {
				generator.WriteReturns(fmt.Errorf("oh no!"))
				Expect(executor.Setup()).To(MatchError("oh no!"))
			})
		})
	})

	Describe("Migrations", func() {
		It("returns the migrations successfully", func() {
			provider.MigrationsReturns([]migration.Item{migration.Item{Id: "id-123"}}, nil)
			migrations, err := executor.Migrations()
			Expect(err).To(BeNil())
			Expect(migrations).To(HaveLen(1))
			Expect(migrations[0].Id).To(Equal("id-123"))
			Expect(provider.MigrationsCallCount()).To(Equal(1))
		})

		Context("when the provider fails", func() {
			It("returns the error", func() {
				provider.MigrationsReturns([]migration.Item{}, fmt.Errorf("oh no!"))
				migrations, err := executor.Migrations()
				Expect(err).To(MatchError("oh no!"))
				Expect(migrations).To(BeEmpty())
			})
		})
	})

	Describe("Run", func() {
		Context("when there are no migrations", func() {
			It("does not run any migration", func() {
				cnt, err := executor.Run(1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(0))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RunCallCount()).To(BeZero())
			})
		})

		Context("when there are applied migrations", func() {
			It("does not run any of the applied migrations", func() {
				migrations := []migration.Item{
					migration.Item{
						Id:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					migration.Item{
						Id:          "20070102150405",
						Description: "Second",
					},
					migration.Item{
						Id:          "20080102150405",
						Description: "Third",
					},
				}

				provider.MigrationsReturns(migrations, nil)
				cnt, err := executor.Run(1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(1))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RunCallCount()).To(Equal(1))

				item := runner.RunArgsForCall(0)
				Expect(*item).To(Equal(migrations[1]))

				Expect(provider.InsertCallCount()).To(Equal(1))
				item = provider.InsertArgsForCall(0)
				Expect(*item).To(Equal(migrations[1]))
			})
		})

		Context("when the step is negative number", func() {
			var migrations []migration.Item

			BeforeEach(func() {
				migrations = []migration.Item{
					migration.Item{
						Id:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					migration.Item{
						Id:          "20070102150405",
						Description: "Second",
					},
					migration.Item{
						Id:          "20080102150405",
						Description: "Third",
					},
				}

				provider.MigrationsReturns(migrations, nil)
			})

			It("runs all pending migrations", func() {
				cnt, err := executor.Run(-1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(2))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RunCallCount()).To(Equal(2))

				for i := 0; i < runner.RunCallCount(); i++ {
					item := runner.RunArgsForCall(i)
					Expect(*item).To(Equal(migrations[i+1]))
				}
			})

			Context("when the runner fails", func() {
				It("returns the error", func() {
					runner.RunReturns(fmt.Errorf("Oh no!"))

					cnt, err := executor.Run(-1)
					Expect(err).To(MatchError("Oh no!"))
					Expect(cnt).To(Equal(0))

					Expect(runner.RunCallCount()).To(Equal(1))
				})
			})

			Context("when the provider fails", func() {
				Context("when the insert fails", func() {
					It("returns the error", func() {
						provider.InsertReturns(fmt.Errorf("Oh no!"))

						cnt, err := executor.Run(1)
						Expect(err).To(MatchError("Oh no!"))
						Expect(cnt).To(Equal(0))
					})
				})
			})
		})

		Context("when the provider fails", func() {
			It("returns the error", func() {
				provider.MigrationsReturns([]migration.Item{}, fmt.Errorf("Oh no!"))

				cnt, err := executor.Run(1)
				Expect(err).To(MatchError("Oh no!"))
				Expect(cnt).To(Equal(0))
			})
		})
	})

	Describe("Revert", func() {
		Context("when there are no migrations", func() {
			It("does not revert any migration", func() {
				cnt, err := executor.Revert(1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(0))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RevertCallCount()).To(BeZero())
			})
		})

		Context("when there are pending migrations", func() {
			It("does not revert any of the pending migrations", func() {
				migrations := []migration.Item{
					migration.Item{
						Id:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					migration.Item{
						Id:          "20070102150405",
						Description: "Second",
						CreatedAt:   time.Now(),
					},
					migration.Item{
						Id:          "20080102150405",
						Description: "Third",
					},
				}

				provider.MigrationsReturns(migrations, nil)

				cnt, err := executor.Revert(1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(1))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RevertCallCount()).To(Equal(1))

				item := runner.RevertArgsForCall(0)
				Expect(*item).To(Equal(migrations[1]))

				Expect(provider.DeleteCallCount()).To(Equal(1))
				item = provider.DeleteArgsForCall(0)
				Expect(*item).To(Equal(migrations[1]))
			})
		})

		Context("when the step is negative number", func() {
			var migrations []migration.Item

			BeforeEach(func() {
				migrations = []migration.Item{
					migration.Item{
						Id:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					migration.Item{
						Id:          "20070102150405",
						Description: "Second",
						CreatedAt:   time.Now(),
					},
					migration.Item{
						Id:          "20080102150405",
						Description: "Third",
					},
				}

				provider.MigrationsReturns(migrations, nil)
			})

			It("reverts all applied migrations", func() {
				cnt, err := executor.Revert(-1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(2))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RevertCallCount()).To(Equal(2))

				for i := 0; i < runner.RunCallCount(); i++ {
					item := runner.RevertArgsForCall(i)
					Expect(*item).To(Equal(migrations[i+1]))
				}
			})

			Context("when the runner fails", func() {
				It("returns the error", func() {
					runner.RevertReturns(fmt.Errorf("Oh no!"))

					cnt, err := executor.Revert(1)
					Expect(err).To(MatchError("Oh no!"))
					Expect(cnt).To(Equal(0))

					Expect(runner.RevertCallCount()).To(Equal(1))
				})
			})
		})

		Context("when the provider fails", func() {
			It("returns the error", func() {
				provider.MigrationsReturns([]migration.Item{}, fmt.Errorf("Oh no!"))

				cnt, err := executor.Revert(1)
				Expect(err).To(MatchError("Oh no!"))
				Expect(cnt).To(Equal(0))
			})

			Context("when the delete fails", func() {
				It("returns the error", func() {
					migrations := []migration.Item{
						migration.Item{
							Id:          "20060102150405",
							Description: "First",
							CreatedAt:   time.Now(),
						},
						migration.Item{
							Id:          "20070102150405",
							Description: "Second",
						},
						migration.Item{
							Id:          "20080102150405",
							Description: "Third",
						},
					}

					provider.MigrationsReturns(migrations, nil)
					provider.DeleteReturns(fmt.Errorf("Oh no!"))

					cnt, err := executor.Revert(1)
					Expect(err).To(MatchError("Oh no!"))
					Expect(cnt).To(Equal(0))
				})
			})
		})
	})
})
