package sqlmigr_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmigr"
)

var _ = Describe("Executor", func() {
	var (
		executor  *sqlmigr.Executor
		provider  *fake.MigrationProvider
		generator *fake.MigrationGenerator
		runner    *fake.MigrationRunner
		logger    *fake.Logger
	)

	BeforeEach(func() {
		provider = &fake.MigrationProvider{}
		generator = &fake.MigrationGenerator{}
		runner = &fake.MigrationRunner{}
		logger = &fake.Logger{}

		executor = &sqlmigr.Executor{
			Logger:    logger,
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
			Expect(item.ID).To(Equal(min.Format(format)))
			Expect(item.Description).To(Equal("setup"))

			data, err := ioutil.ReadAll(content.UpCommand)
			Expect(err).NotTo(HaveOccurred())

			up := &bytes.Buffer{}
			fmt.Fprintln(up, "CREATE TABLE IF NOT EXISTS migrations (")
			fmt.Fprintln(up, " id          VARCHAR(15) NOT NULL PRIMARY KEY,")
			fmt.Fprintln(up, " description TEXT        NOT NULL,")
			fmt.Fprintln(up, " created_at  TIMESTAMP   NOT NULL")
			fmt.Fprintln(up, ");")
			fmt.Fprintln(up)
			Expect(string(data)).To(Equal(up.String()))

			data, err = ioutil.ReadAll(content.DownCommand)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal("DROP TABLE IF EXISTS migrations;\n"))
		})

		Context("when the migration exists", func() {
			It("does not setup the project", func() {
				provider.ExistsReturns(true)
				Expect(executor.Setup()).To(Succeed())
				Expect(runner.RunCallCount()).To(Equal(0))
			})
		})

		Context("when the generator fails", func() {
			It("returns the error", func() {
				generator.WriteReturns(fmt.Errorf("oh no!"))
				Expect(executor.Setup()).To(MatchError("oh no!"))
			})
		})
	})

	Describe("Create", func() {
		It("creates migration successfully", func() {
			migration, err := executor.Create("schema")
			Expect(err).NotTo(HaveOccurred())
			Expect(generator.CreateCallCount()).To(Equal(1))

			item := generator.CreateArgsForCall(0)
			Expect(item.Description).To(Equal("schema"))
			Expect(item.Drivers).To(ContainElement("sql"))
			Expect(item).To(Equal(migration))
		})

		Context("when the generator fails", func() {
			It("returns the error", func() {
				generator.CreateReturns(fmt.Errorf("oh no!"))
				item, err := executor.Create("test")
				Expect(err).To(MatchError("oh no!"))
				Expect(item).To(BeNil())
			})
		})
	})

	Describe("Migrations", func() {
		It("returns the migrations successfully", func() {
			provider.MigrationsReturns([]*sqlmigr.Migration{{ID: "id-123"}}, nil)
			migrations, err := executor.Migrations()
			Expect(err).To(BeNil())
			Expect(migrations).To(HaveLen(1))
			Expect(migrations[0].ID).To(Equal("id-123"))
			Expect(provider.MigrationsCallCount()).To(Equal(1))
		})

		Context("when the provider fails", func() {
			It("returns the error", func() {
				provider.MigrationsReturns([]*sqlmigr.Migration{}, fmt.Errorf("oh no!"))
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
				migrations := []*sqlmigr.Migration{
					{
						ID:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					{
						ID:          "20070102150405",
						Description: "Second",
					},
					{
						ID:          "20080102150405",
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
				Expect(item).To(Equal(migrations[1]))

				Expect(provider.InsertCallCount()).To(Equal(1))
				item = provider.InsertArgsForCall(0)
				Expect(item).To(Equal(migrations[1]))
			})

			It("runs all migrations", func() {
				migrations := []*sqlmigr.Migration{
					{
						ID:          "20060102150405",
						Description: "First",
					},
					{
						ID:          "20070102150405",
						Description: "Second",
					},
					{
						ID:          "20080102150405",
						Description: "Third",
					},
				}

				provider.MigrationsReturns(migrations, nil)
				cnt, err := executor.RunAll()
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(3))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RunCallCount()).To(Equal(3))
				Expect(provider.InsertCallCount()).To(Equal(3))

				for i := 0; i < 3; i++ {
					item := runner.RunArgsForCall(i)
					Expect(item).To(Equal(migrations[i]))

					item = provider.InsertArgsForCall(i)
					Expect(item).To(Equal(migrations[i]))
				}
			})
		})

		Context("when the step is negative number", func() {
			var migrations []*sqlmigr.Migration

			BeforeEach(func() {
				migrations = []*sqlmigr.Migration{
					{
						ID:          "20060102150405",
						Description: "First",
						Drivers:     []string{"sql"},
						CreatedAt:   time.Now(),
					},
					{
						ID:          "20070102150405",
						Description: "Second",
						Drivers:     []string{"sql"},
					},
					{
						ID:          "20070102150405",
						Description: "Second",
						Drivers:     []string{"sql", "sqlite3"},
					},
					{
						ID:          "20080102150405",
						Drivers:     []string{"sql"},
						Description: "Third",
					},
				}

				provider.MigrationsReturns(migrations, nil)
			})

			It("runs all pending migrations", func() {
				cnt, err := executor.Run(-1)
				Expect(err).To(Succeed())
				Expect(cnt).To(Equal(3))

				Expect(provider.MigrationsCallCount()).To(Equal(1))
				Expect(runner.RunCallCount()).To(Equal(3))
				Expect(logger.InfofCallCount()).To(Equal(3))

				for i := 0; i < runner.RunCallCount(); i++ {
					item := runner.RunArgsForCall(i)
					Expect(item).To(Equal(migrations[i+1]))
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
				provider.MigrationsReturns([]*sqlmigr.Migration{}, fmt.Errorf("Oh no!"))

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

		It("revert all migrations", func() {
			migrations := []*sqlmigr.Migration{
				{
					ID:          "20060102150405",
					Description: "First",
					Drivers:     []string{"sql"},
					CreatedAt:   time.Now(),
				},
				{
					ID:          "20070102150405",
					Description: "Second",
					Drivers:     []string{"sql", "sqlite3"},
					CreatedAt:   time.Now(),
				},
				{
					ID:          "20080102150405",
					Description: "Third",
					Drivers:     []string{"sql"},
					CreatedAt:   time.Now(),
				},
			}

			provider.MigrationsReturns(migrations, nil)
			cnt, err := executor.RevertAll()
			Expect(err).To(Succeed())
			Expect(cnt).To(Equal(3))

			Expect(provider.MigrationsCallCount()).To(Equal(1))
			Expect(runner.RevertCallCount()).To(Equal(3))
			Expect(provider.DeleteCallCount()).To(Equal(3))

			item := runner.RevertArgsForCall(0)
			Expect(item).To(Equal(migrations[2]))

			item = provider.DeleteArgsForCall(0)
			Expect(item).To(Equal(migrations[2]))

			item = runner.RevertArgsForCall(1)
			Expect(item).To(Equal(migrations[1]))

			item = provider.DeleteArgsForCall(1)
			Expect(item).To(Equal(migrations[1]))

			item = runner.RevertArgsForCall(2)
			Expect(item).To(Equal(migrations[0]))

			item = provider.DeleteArgsForCall(2)
			Expect(item).To(Equal(migrations[0]))
		})

		Context("when there are pending migrations", func() {
			It("does not revert any of the pending migrations", func() {
				migrations := []*sqlmigr.Migration{
					{
						ID:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					{
						ID:          "20070102150405",
						Description: "Second",
						CreatedAt:   time.Now(),
					},
					{
						ID:          "20080102150405",
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
				Expect(item).To(Equal(migrations[1]))

				Expect(provider.DeleteCallCount()).To(Equal(1))
				item = provider.DeleteArgsForCall(0)
				Expect(item).To(Equal(migrations[1]))
			})
		})

		Context("when the step is negative number", func() {
			var migrations []*sqlmigr.Migration

			BeforeEach(func() {
				migrations = []*sqlmigr.Migration{
					{
						ID:          "20060102150405",
						Description: "First",
						CreatedAt:   time.Now(),
					},
					{
						ID:          "20070102150405",
						Description: "Second",
						CreatedAt:   time.Now(),
					},
					{
						ID:          "20080102150405",
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
				provider.MigrationsReturns([]*sqlmigr.Migration{}, fmt.Errorf("Oh no!"))

				cnt, err := executor.Revert(1)
				Expect(err).To(MatchError("Oh no!"))
				Expect(cnt).To(Equal(0))
			})

			Context("when the delete fails", func() {
				var migrations []*sqlmigr.Migration

				BeforeEach(func() {
					migrations = []*sqlmigr.Migration{
						{
							ID:          "20060102150405",
							Description: "First",
							CreatedAt:   time.Now(),
						},
						{
							ID:          "20070102150405",
							Description: "Second",
						},
						{
							ID:          "20080102150405",
							Description: "Third",
						},
					}
				})

				Context("when the error is not exist", func() {
					It("does not return the error", func() {
						provider.MigrationsReturns(migrations, nil)
						provider.DeleteReturns(fmt.Errorf("no such table: migrations"))

						cnt, err := executor.Revert(1)
						Expect(err).To(BeNil())
						Expect(cnt).To(Equal(0))
					})
				})

				It("returns the error", func() {
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
