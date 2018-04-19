package schema_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/phogolabs/oak/fake"
	"github.com/phogolabs/oak/schema"
)

var _ = Describe("CompositeTagBuilder", func() {
	It("delegate the build operation to underlying builders", func() {
		builder := schema.CompositeTagBuilder{}

		builder1 := &fake.SchemaTagBuilder{}
		builder1.BuildReturns("tag1")
		builder = append(builder, builder1)

		builder2 := &fake.SchemaTagBuilder{}
		builder2.BuildReturns("tag2")
		builder = append(builder, builder2)

		column := &schema.Column{}
		Expect(builder.Build(column)).To(Equal("`tag1 tag2`"))

		Expect(builder1.BuildCallCount()).To(Equal(1))
		Expect(builder1.BuildArgsForCall(0)).To(Equal(column))

		Expect(builder2.BuildCallCount()).To(Equal(1))
		Expect(builder2.BuildArgsForCall(0)).To(Equal(column))
	})
})

var _ = Describe("SQLXTagBuilder", func() {
	var (
		column  *schema.Column
		builder *schema.SQLXTagBuilder
	)

	BeforeEach(func() {
		builder = &schema.SQLXTagBuilder{}
		column = &schema.Column{
			Name: "id",
			Type: schema.ColumnType{},
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
		column  *schema.Column
		builder *schema.GORMTagBuilder
	)

	BeforeEach(func() {
		builder = &schema.GORMTagBuilder{}
		column = &schema.Column{
			Name: "id",
			Type: schema.ColumnType{
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
		column  *schema.Column
		builder *schema.JSONTagBuilder
	)

	BeforeEach(func() {
		builder = &schema.JSONTagBuilder{}
		column = &schema.Column{
			Name: "id",
		}
	})

	It("creates a json tag", func() {
		Expect(builder.Build(column)).To(Equal("json:\"id\""))
	})
})

var _ = Describe("XMLTagBuilder", func() {
	var (
		column  *schema.Column
		builder *schema.XMLTagBuilder
	)

	BeforeEach(func() {
		builder = &schema.XMLTagBuilder{}
		column = &schema.Column{
			Name: "id",
		}
	})

	It("creates a xml tag", func() {
		Expect(builder.Build(column)).To(Equal("xml:\"id\""))
	})
})
