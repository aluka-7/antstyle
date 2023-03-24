package antstyle

import (
	"fmt"
	"regexp"
)

const (
	DefaultPathSeparator  = "/"   // DefaultPathSeparator默认路径分隔符：“ /”
	CacheTurnoffThreshold = 65536 // 缓存关闭阈值
)

var (
	Asterisk        rune           = '\u002a'                                 // *
	QuestionMark    rune           = '\u003f'                                 // ?
	Brackets        rune           = '\u007b'                                 // {
	WildcardChars   []rune         = []rune{Asterisk, QuestionMark, Brackets} // 通配符字符首字母'*'，'？'，'{'
	VariablePattern *regexp.Regexp                                            // pattern
	matcher         PathMatcher
)

func init() {
	fmt.Println("Loading Ant-Style")
	reg, _ := regexp.Compile("{[^/]+?}") // pattern
	VariablePattern = reg
	matcher = New()
}

func Increment(value *int) {
	*value = *value + 1
}

func IsPattern(path string) bool {
	return matcher.IsPattern(path)
}

func Match(pattern, path string) bool {
	return matcher.Match(pattern, path)
}

func MatchStart(pattern, path string) bool {
	return matcher.MatchStart(pattern, path)
}

func SetPathSeparator(pathSeparator string) {
	matcher.SetPathSeparator(pathSeparator)
}

func SetCaseSensitive(caseSensitive bool) {
	matcher.SetCaseSensitive(caseSensitive)
}

func SetTrimTokens(trimTokens bool) {
	matcher.SetTrimTokens(trimTokens)
}

func SetCachePatterns(cachePatterns bool) {
	matcher.SetCachePatterns(cachePatterns)
}

/*
*
  *策略界面，用于基于路径的匹配。
  *
  *默认实现是AntPathMatcher，支持Ant-style风格的模式语法。
*/
type PathMatcher interface {

	/**
	 *给定的{@code path}是否表示可以通过此接口的实现匹配的模式？
	 *如果返回值为{@code false}，则不必使用{@link #match}方法，因为在静态路径Strings上进行直接相等比较将得出相同的结果。
	 *@param path要检查的路径字符串
	 *@return {true}，如果给定的{path}表示一个模式
	 */
	IsPattern(path string) bool

	/**
	 *根据此PathMatcher的匹配策略，将给定的{@code路径}与给定的{@code模式}相匹配。
	 *@param pattern要匹配的模式
	 *@param path要测试的路径字符串
	 *@return {@code true}如果提供的{@code path}匹配，
	 *{@code false}（如果没有）
	 */
	Match(pattern, path string) bool

	/**
	 *根据此PathMatcher的匹配策略，将给定的{@code路径}与给定的{@code模式}的对应部分进行匹配。
	 *确定模式是否至少匹配给定的基本路径，并假设一条完整路径也可以匹配。
	 *@param pattern要匹配的模式
	 *@param path要测试的路径字符串
	 *@return {true}如果提供的{path}匹配，{false}（如果没有）
	 */
	MatchStart(pattern, path string) bool

	/**
	 *给定图案和完整路径，确定图案映射的零件.
	 *该方法应该找出通过实际模式动态匹配路径的哪一部分，即，它从给定的完整路径中剥离静态定义的引导路径，仅返回的实际模式匹配部分路径.
	 *例如:对于"root/*.html"作为模式,"root/file.html"作为完整路径，此方法应返回"file.html".详细的确定规则已指定为此PathMatcher的匹配策略.
	 *如果是实际模式,简单的实现可以按原样返回给定的完整路径,如果模式不包含任何动态部分(即{pattern}参数为静态路径),则为空String则不符合实际的{#isPattern模式}).
	 *复杂的实现将区分给定路径模式的静态部分和动态部分.
	 *@param  pattern 路径模式
	 *@param  path 进行内省的完整路径
	 *@return 给定的{@code path}的模式映射部分
	 *(从不{null})
	 */
	ExtractPathWithinPattern(pattern, path string) string

	/**
	 *给定模式和完整路径,提取URI模板变量.URI模板变量通过大括号("{"和"}")表示.
	 *例如:对于模式"/hotels/{hotel}"和路径"/hotels/1"，此方法将返回包含"hotel"->"1"的地图.
	 *@param pattern string 模式路径模式，可能包含URI模板
	 *@param path string 从中提取模板变量的完整路径
	 *@return map[string]string 其中包含变量名作为键;将变量值作为值
	 */
	ExtractUriTemplateVariables(pattern, path string) *map[string]string

	/**
	 *给定完整路径后，将返回一个{@link Comparator}，适用于按照该路径的显式顺序对模式进行排序。
	 *所使用的完整算法取决于基础实现，但是通常，返回的{AntPatternComparator}一个列表，因此 更具体的模式先于通用模式。
	 *@param path string用于比较的完整路径
	 *@return *AntPatternComparator 能够按显式顺序对模式进行排序的比较器
	 */
	GetPatternComparator(path string) *AntPatternComparator

	/**
	 *将两个模式组合成一个返回的新模式。
	 *用于组合两种模式的完整算法取决于基础实现。
	 *@param pattern1 第一个模式
	 *@param pattern2 第二种模式
	 *@return string 两个模式的组合
	 */
	Combine(pattern1, pattern2 string) string
	SetPathSeparator(pathSeparator string)
	SetCaseSensitive(caseSensitive bool)
	SetTrimTokens(trimTokens bool)
	SetCachePatterns(cachePatterns bool)
	PatternCacheSize() int64
}
