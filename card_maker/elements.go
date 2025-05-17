package card_maker

import "fmt"

var AllElements = []string{ELEM_AIR_ZH, ELEM_DARK_ZH, ELEM_EARTH_ZH, ELEM_FIRE_ZH, ELEM_LIGHT_ZH, ELEM_WATER_ZH, ELEM_NONE_ZH, ELEM_MANA_ZH}

type Elements map[string]int

func (ele Elements) Equals(other Elements) bool {
	if len(ele) != len(other) {
		return false
	}
	for key, value := range ele {
		if otherVal, ok := other[key]; !ok || otherVal != value {
			return false
		}
	}
	return true
}
func NewElements(elementDic map[string]int) Elements {
	elem := Elements{}

	for key, value := range elementDic {
		elem[key] = value
	}
	for _, element := range AllElements {
		if _, ok := elementDic[element]; !ok {
			elem[element] = 0
		}
	}
	return elem
}

func (elem Elements) TotalCost() int {
	cost := 0
	for _, val := range elem {
		cost += val
	}
	return cost
}

func (elem Elements) Get(elemString string) (int, bool) {
	val, ok := elem[elemString]
	return val, ok
}

func (elem Elements) Set(elemString string, value int) {
	elem[elemString] = value
}

func (elem Elements) String() string {
	str := ""
	for key, value := range elem {
		if value != 0 {
			str += fmt.Sprintf("%s:%d ", key, value)
		}
	}
	return str
}
