package sqlmodel_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/prana/fake"
	"github.com/phogolabs/prana/sqlmodel"
)

var _ = Describe("CompositeTagBuilder", func() {
	It("delegate the build operation to underlying builders", func() {
		builder := sqlmodel.CompositeTagBuilder{}

		builder1 := &fake.ModelTagBuilder{}
		builder1.BuildReturns("tag1")
		builder = append(builder, builder1)

		builder2 := &fake.ModelTagBuilder{}
		builder2.BuildReturns("tag2")
		builder = append(builder, builder2)

		column := &sqlmodel.Column{}
		Expect(builder.Build(column)).To(Equal("`tag1 tag2`"))

		Expect(builder1.BuildCallCount()).To(Equal(1))
		Expect(builder1.BuildArgsForCall(0)).To(Equal(column))

		Expect(builder2.BuildCallCount()).To(Equal(1))
		Expect(builder2.BuildArgsForCall(0)).To(Equal(column))
	})

	Context("when some of the builders return an space string", func() {
		It("skips the result", func() {
			builder := sqlmodel.CompositeTagBuilder{}

			builder1 := &fake.ModelTagBuilder{}
			builder1.BuildReturns(" ")
			builder = append(builder, builder1)

			builder2 := &fake.ModelTagBuilder{}
			builder2.BuildReturns(" tag2")
			builder = append(builder, builder2)

			column := &sqlmodel.Column{}
			Expect(builder.Build(column)).To(Equal("`tag2`"))

			Expect(builder1.BuildCallCount()).To(Equal(1))
			Expect(builder1.BuildArgsForCall(0)).To(Equal(column))

			Expect(builder2.BuildCallCount()).To(Equal(1))
			Expect(builder2.BuildArgsForCall(0)).To(Equal(column))
		})
	})
})

var _ = Describe("SQLXTagBuilder", func() {
	var (
		column  *sqlmodel.Column
		builder *sqlmodel.SQLXTagBuilder
	)

	BeforeEach(func() {
		builder = &sqlmodel.SQLXTagBuilder{}
		column = &sqlmodel.Column{
			Name: "id",
			Type: sqlmodel.ColumnType{},
		}
	})

	It("builds the tag correctly", func() {
		Expect(builder.Build(column)).To(Equal("db:\"id,not_null\""))
	})

	Context("when the column is primary key", func() {
		BeforeEach(func() {
			column.Type.IsPrimaryKey = true
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("db:\"id,primary_key,not_null\""))
		})
	})

	Context("when the column allow null", func() {
		BeforeEach(func() {
			column.Type.IsNullable = true
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("db:\"id,null\""))
		})
	})

	Context("when the column has char size", func() {
		BeforeEach(func() {
			column.Type.CharMaxLength = 200
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("db:\"id,not_null,size=200\""))
		})
	})

	Context("when all options are presented", func() {
		BeforeEach(func() {
			column.Type.IsPrimaryKey = true
			column.Type.CharMaxLength = 200
		})
		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("db:\"id,primary_key,not_null,size=200\""))
		})
	})
})

var _ = Describe("GORMTagBuilder", func() {
	var (
		column  *sqlmodel.Column
		builder *sqlmodel.GORMTagBuilder
	)

	BeforeEach(func() {
		builder = &sqlmodel.GORMTagBuilder{}
		column = &sqlmodel.Column{
			Name: "id",
			Type: sqlmodel.ColumnType{
				Name: "db_type",
			},
		}
	})

	It("builds the tag correctly", func() {
		Expect(builder.Build(column)).To(Equal("gorm:\"column:id;type:db_type;not null\""))
	})

	Context("when the column is primary key", func() {
		BeforeEach(func() {
			column.Type.IsPrimaryKey = true
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("gorm:\"column:id;type:db_type;primary_key;not null\""))
		})
	})

	Context("when the column allow null", func() {
		BeforeEach(func() {
			column.Type.IsNullable = true
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("gorm:\"column:id;type:db_type;null\""))
		})
	})

	Context("when the column has char size", func() {
		BeforeEach(func() {
			column.Type.CharMaxLength = 200
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("gorm:\"column:id;type:db_type(200);not null;size:200\""))
		})
	})

	Context("when the column has precision", func() {
		BeforeEach(func() {
			column.Type.Precision = 10
			column.Type.PrecisionScale = 20
		})

		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("gorm:\"column:id;type:db_type(10, 20);not null;precision:10\""))
		})
	})

	Context("when all options are presented", func() {
		BeforeEach(func() {
			column.Type.IsPrimaryKey = true
			column.Type.CharMaxLength = 200
		})
		It("builds the tag correctly", func() {
			Expect(builder.Build(column)).To(Equal("gorm:\"column:id;type:db_type(200);primary_key;not null;size:200\""))
		})
	})
})

var _ = Describe("JSONTagBuilder", func() {
	var (
		column  *sqlmodel.Column
		builder *sqlmodel.JSONTagBuilder
	)

	BeforeEach(func() {
		builder = &sqlmodel.JSONTagBuilder{}
		column = &sqlmodel.Column{
			Name: "id",
		}
	})

	It("creates a json tag", func() {
		Expect(builder.Build(column)).To(Equal("json:\"id\""))
	})
})

var _ = Describe("XMLTagBuilder", func() {
	var (
		column  *sqlmodel.Column
		builder *sqlmodel.XMLTagBuilder
	)

	BeforeEach(func() {
		builder = &sqlmodel.XMLTagBuilder{}
		column = &sqlmodel.Column{
			Name: "id",
		}
	})

	It("creates a xml tag", func() {
		Expect(builder.Build(column)).To(Equal("xml:\"id\""))
	})
})

var _ = Describe("ValidateTagBuilder", func() {
	var (
		column  *sqlmodel.Column
		builder *sqlmodel.ValidateTagBuilder
	)

	BeforeEach(func() {
		builder = &sqlmodel.ValidateTagBuilder{}
		column = &sqlmodel.Column{
			Name:     "id",
			ScanType: "string",
		}
	})

	It("creates a validation tag", func() {
		Expect(builder.Build(column)).To(Equal("validate:\"required,gt=0\""))
	})

	Context("when the value has length", func() {
		BeforeEach(func() {
			column.ScanType = ""
			column.Type.CharMaxLength = 200
		})

		It("creates a validation tag", func() {
			Expect(builder.Build(column)).To(Equal("validate:\"required,max=200\""))
		})
	})

	Context("when the value is not nullable", func() {
		BeforeEach(func() {
			column.Type.IsNullable = true
		})

		Context("when the value has length", func() {
			BeforeEach(func() {
				column.Type.CharMaxLength = 200
			})

			It("creates a validation tag", func() {
				Expect(builder.Build(column)).To(Equal("validate:\"max=200\""))
			})
		})

		It("returns an empty tag", func() {
			Expect(builder.Build(column)).To(Equal("validate:\"-\""))
		})
	})
})
