package attribute

func getAttrsByCategory(category attrCategory) []attrType {
	var attrTypesToTest []attrType
	for k, v := range attrCategoryByType {
		if v == category {
			attrTypesToTest = append(attrTypesToTest, k)
		}
	}
	return attrTypesToTest
}
