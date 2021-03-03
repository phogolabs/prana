package sqlmigr_test

import (
	"bytes"
	"time"

	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmigr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Printer", func() {
	var migrations []*sqlmigr.Migration

	BeforeEach(func() {
		migrations = []*sqlmigr.Migration{
			{
				ID:          "20060102150405",
				Description: "First",
				CreatedAt:   time.Now(),
			},
		}
	})

	Context("Flog", func() {
		var logger *fake.Logger

		BeforeEach(func() {
			logger = &fake.Logger{}
			logger.WithFieldsReturns(logger)
		})

		It("logs the migration", func() {
			sqlmigr.Flog(logger, migrations)
			Expect(logger.WithFieldsCallCount()).To(Equal(1))

			fields := logger.WithFieldsArgsForCall(0)
			Expect(fields).To(HaveKeyWithValue("Id", migrations[0].ID))
			Expect(fields).To(HaveKeyWithValue("Description", migrations[0].Description))
			Expect(fields).To(HaveKeyWithValue("Status", "executed"))
		})

		Context("when the migration is not executed", func() {
			BeforeEach(func() {
				migrations[0].CreatedAt = time.Time{}
			})

			It("logs the migration", func() {
				sqlmigr.Flog(logger, migrations)
				Expect(logger.WithFieldsCallCount()).To(Equal(1))

				fields := logger.WithFieldsArgsForCall(0)
				Expect(fields).To(HaveKeyWithValue("Status", "pending"))
			})
		})
	})

	Context("Ftable", func() {
		It("logs the migrations", func() {
			w := &bytes.Buffer{}
			sqlmigr.Ftable(w, migrations)

			content := w.String()
			Expect(content).To(ContainSubstring("Id"))
			Expect(content).To(ContainSubstring("Description"))
			Expect(content).To(ContainSubstring("Status"))
			Expect(content).To(ContainSubstring("Created At"))
			Expect(content).To(ContainSubstring("executed"))
			Expect(content).To(ContainSubstring("20060102150405"))
			Expect(content).To(ContainSubstring("First"))
		})

		Context("when the migration is not applied", func() {
			BeforeEach(func() {
				migrations[0].CreatedAt = time.Time{}
			})

			It("logs the migrations", func() {
				w := &bytes.Buffer{}
				sqlmigr.Ftable(w, migrations)

				content := w.String()
				Expect(content).To(ContainSubstring("pending"))
			})
		})
	})
})
