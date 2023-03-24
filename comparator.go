package antstyle

import "strings"

// AntPatternComparator
/**
 *由返回的默认{@link Comparator}实现
 *按顺序，最“通用”的模式由以下内容确定:
 *如果它为null或捕获所有模式(即等于“ / **”)
 *如果其他模式是实际匹配项
 *如果它是一个包罗万象的模式\(即以“ **”结尾)
 *如果它的"*"比其他模式多
 *如果"{foo}"比其他模式多
 *如果比其他格式短
 */
type AntPatternComparator struct {
	path string
}

func NewDefaultAntPatternComparator(path string) *AntPatternComparator {
	comparator := &AntPatternComparator{}
	comparator.path = path
	return comparator
}

func (comparator *AntPatternComparator) Compare(pattern1, pattern2 string) int {
	info1 := NewDefaultPatternInfo(pattern1)
	info2 := NewDefaultPatternInfo(pattern2)

	if info1.IsLeastSpecific() && info2.IsLeastSpecific() {
		return 0
	} else if info1.IsLeastSpecific() {
		return 1
	} else if info2.IsLeastSpecific() {
		return -1
	}

	pattern1EqualsPath := strings.EqualFold(comparator.path, pattern1)
	pattern2EqualsPath := strings.EqualFold(comparator.path, pattern2)
	if pattern1EqualsPath && pattern2EqualsPath {
		return 0
	} else if pattern1EqualsPath {
		return -1
	} else if pattern2EqualsPath {
		return 1
	}

	if info1.IsPrefixPattern() && info2.GetDoubleWildcards() == 0 {
		return 1
	} else if info2.IsPrefixPattern() && info1.GetDoubleWildcards() == 0 {
		return -1
	}

	if info1.GetTotalCount() != info2.GetTotalCount() {
		return info1.GetTotalCount() - info2.GetTotalCount()
	}

	if info1.GetLength() != info2.GetLength() {
		return info2.GetLength() - info1.GetLength()
	}

	if info1.GetSingleWildcards() < info2.GetSingleWildcards() {
		return -1
	} else if info2.GetSingleWildcards() < info1.GetSingleWildcards() {
		return 1
	}

	if info1.GetUriVars() < info2.GetUriVars() {
		return -1
	} else if info2.GetUriVars() < info1.GetUriVars() {
		return 1
	}

	return 0
}
