package attribute

func getAttrsByCategory(category attrCategory) []AttrType {
	var attrTypesToTest []AttrType
	for k, v := range attrCategoryByType {
		if v == category {
			attrTypesToTest = append(attrTypesToTest, k)
		}
	}
	return attrTypesToTest
}
