package jobs

import (
	"fmt"
	"my_toolbox/entities"
	"my_toolbox/helpers"
	"my_toolbox/library/log"
	"strings"
)

type AttributeResult struct {
	PropertyType     string `gorm:"column:property_type"`
	PropertyRequired bool   `gorm:"column:property_required"`
	PropertyFiltered bool   `gorm:"column:property_filtered"`
	PropertyName     string `gorm:"column:property_name"`
}

type AttributeExporter struct {
	Job
}

func (ae *AttributeExporter) Run() {
	categories := ae.getCategories()
	log.GetLogger().Info("active product updates started")

	for _, category := range categories {
		log.GetLogger().Info(fmt.Sprintf("category: %d", category.ID))
		attributes := ae.getAttributes(category.ID)

		var data [][]string
		for _, attribute := range attributes {
			isFilterable := "Hayır"
			if attribute.PropertyFiltered {
				isFilterable = "Evet"
			}

			isRequired := "Hayır"
			if attribute.PropertyRequired {
				isRequired = "Evet"
			}

			attribute.PropertyName = strings.Replace(attribute.PropertyName, ";", "-", -1)
			data = append(data, []string{fmt.Sprintf("%d", category.ID), category.Name, attribute.PropertyName, attribute.PropertyType, isRequired, isFilterable})
		}

		helpers.Write("assets/category_attributes.csv", data)
	}
	log.GetLogger().Info(fmt.Sprintf("active product updates ended. Total Product Count: %d", pc))

	log.GetLogger().Info("bitti")
}

func (ae *AttributeExporter) getCategories() []entities.EvoCategories {
	var categories []entities.EvoCategories
	ae.DB.Find(&categories)

	return categories
}

func (ae *AttributeExporter) getAttributes(categoryId int) []AttributeResult {
	if categoryId == 53 {
		fmt.Println("ahada geldik")
	}
	sql := fmt.Sprintf(`SELECT 
        e_property.property_type property_type,
        IF(e_property.property_required = 1, TRUE, FALSE) AS property_required,
    IF(e_property.property_filtered = 1, TRUE, FALSE) AS property_filtered,
        e_property_content.property_nome property_name
    FROM e_property_group_subcat 
    INNER JOIN e_property ON (e_property_group_subcat.group_id = e_property.group_id) 
    INNER JOIN e_property_detail ON (e_property.property_id = e_property_detail.property_id) 
    INNER JOIN e_property_prodotto ON (e_property_detail.property_id = e_property_prodotto.detail_id) 
    INNER JOIN e_property_content ON (e_property.property_id = e_property_content.property_id) 
    WHERE e_property_group_subcat.evo_category_id = %d 
    GROUP BY e_property_detail.property_id`, categoryId)

	var attributes []AttributeResult

	ae.DB.Raw(sql).Scan(&attributes)

	return attributes
}
