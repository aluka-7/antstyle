package antstyle

import (
	"strings"

	"github.com/aluka-7/utils"
)

// PatternInfo
/**
 *包含有关模式信息的值类，例如 “ *”，“ **”和“ {”模式元素的出现次数
 */
type PatternInfo struct {
	pattern         string
	uriVars         int
	singleWildcards int
	doubleWildcards int
	catchAllPattern bool
	prefixPattern   bool
	length          int
}

func NewDefaultPatternInfo(pattern string) *PatternInfo {
	hasText := utils.HasText(pattern)
	// 实例化
	pi := &PatternInfo{}
	pi.pattern = pattern
	if hasText {
		pi.initCounters()
		pi.catchAllPattern = strings.EqualFold("/**", pattern)
		pi.prefixPattern = !pi.catchAllPattern && strings.HasSuffix(pi.pattern, "/**")
	}
	if pi.uriVars == 0 {
		if hasText {
			pi.length = len(pattern)
		} else {
			pi.length = 0
		}
	}
	return pi
}

func (pi *PatternInfo) initCounters() {
	pos := 0
	if utils.HasText(pi.pattern) {
		for {
			if pos < len(pi.pattern) {
				if rune(pi.pattern[pos]) == Brackets {
					pi.uriVars++
					pos++
				} else if rune(pi.pattern[pos]) == Asterisk {
					if pos+1 < len(pi.pattern) && rune(pi.pattern[pos+1]) == Asterisk {
						pi.doubleWildcards++
						pos += 2
					} else if pos > 0 && !strings.EqualFold(".*", pi.pattern[pos-1:]) {
						pi.singleWildcards++
						pos++
					} else {
						pos++
					}
				} else {
					pos++
				}
			} else {
				break
			}
		}
	}
}

func (pi *PatternInfo) GetUriVars() int {
	return pi.uriVars
}

func (pi *PatternInfo) GetSingleWildcards() int {
	return pi.singleWildcards
}

func (pi *PatternInfo) GetDoubleWildcards() int {
	return pi.doubleWildcards
}

func (pi *PatternInfo) IsLeastSpecific() bool {
	return utils.IsBlank(pi.pattern) || pi.catchAllPattern
}

func (pi *PatternInfo) IsPrefixPattern() bool {
	return pi.prefixPattern
}

func (pi *PatternInfo) GetTotalCount() int {
	return pi.uriVars + pi.singleWildcards + (2 * pi.doubleWildcards)
}

// 返回给定模式的长度，其中模板变量被认为是1长。
func (pi *PatternInfo) GetLength() int {
	if pi.length == 0 {
		if utils.HasText(pi.pattern) {
			target := VariablePattern.ReplaceAllString(pi.pattern, "#")
			pi.length = len(target)
		}
	}
	return pi.length
}
