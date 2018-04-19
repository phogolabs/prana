package model_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phogolabs/oak/model"
)

var _ = Describe("Model", func() {
	Describe("TypeDef", func() {
		It("returns the type name", func() {
			def := &model.TypeDef{Type: "int"}
			Expect(def.As(false)).To(Equal("int"))
		})

		Context("when the type is nullable", func() {
			It("returns the nullable type name", func() {
				def := &model.TypeDef{
					Type:         "int",
					NullableType: "null.int",
				}
				Expect(def.As(true)).To(Equal("null.int"))
			})
		})
	})

	Describe("ColumnType", func() {
		var columnType model.ColumnType

		BeforeEach(func() {
			columnType = model.ColumnType{
				Name:          "varchar",
				IsPrimaryKey:  true,
				IsNullable:    true,
				CharMaxLength: 200,
			}
		})

		It("returns the column type as string correctly", func() {
			Expect(columnType.String()).To(Equal("VARCHAR(200) PRIMARY KEY NULL"))
		})

		Context("when the type is not null", func() {
			BeforeEach(func() {
				columnType.IsNullable = false
			})

			It("returns the column type as string correctly", func() {
				Expect(columnType.String()).To(Equal("VARCHAR(200) PRIMARY KEY NOT NULL"))
			})
		})

		Context("when the type has precision and scale", func() {
			BeforeEach(func() {
				columnType.CharMaxLength = 0
				columnType.Name = "numeric"
				columnType.Precision = 10
				columnType.PrecisionScale = 20
			})

			It("returns the column type as string correctly", func() {
				Expect(columnType.String()).To(Equal("NUMERIC(10, 20) PRIMARY KEY NULL"))
			})
		})

		Context("when the type has precision only", func() {
			BeforeEach(func() {
				columnType.CharMaxLength = 0
				columnType.Name = "numeric"
				columnType.Precision = 10
				columnType.PrecisionScale = 0
			})

			It("returns the column type as string correctly", func() {
				Expect(columnType.String()).To(Equal("NUMERIC(10) PRIMARY KEY NULL"))
			})
		})
	})
})
