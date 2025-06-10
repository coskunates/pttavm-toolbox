package jobs

import (
	"fmt"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"os"
)

type AttributeFromCSV struct {
	CategoryId    string
	CategoryName  string
	AttributeName string
	AttributeType string
	IsRequired    string
	IsFilterable  string
}

type AttributeNameFix struct {
	Job
}

func (anf *AttributeNameFix) Run() {
	attributesFromCsv, err := helpers.ReadFromCSV("assets/category_attributes.csv")
	if err != nil {
		os.Exit(1)
	}

	var attributes []AttributeFromCSV
	for _, attribute := range attributesFromCsv {
		attributes = append(attributes, AttributeFromCSV{
			CategoryId:    attribute[0],
			CategoryName:  attribute[1],
			AttributeName: attribute[2],
			AttributeType: attribute[3],
			IsRequired:    attribute[4],
			IsFilterable:  attribute[5],
		})
	}
	var data [][]string
	for _, attribute := range attributes {
		snakeCase := helpers.ToSnakeCase(helpers.Normalize(attribute.AttributeName))
		data = append(data, []string{attribute.CategoryId, attribute.CategoryName, attribute.AttributeName, snakeCase, attribute.AttributeType, attribute.IsRequired, attribute.IsFilterable})
		fmt.Println(attribute.AttributeName, snakeCase)
	}

	helpers.Write("assets/category_attributes_formatted.csv", data)

	log.GetLogger().Info(fmt.Sprintf("active product updates ended. Total Product Count: %d", pc))

	log.GetLogger().Info("bitti")
}
