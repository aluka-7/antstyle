package antstyle

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/aluka-7/utils"
)

const (
	DefaultVariablePattern = "(.*)"
	MaxFindCount           = 1 << 5 // MaxFindCount默认值= 32
)

var GlobPattern *regexp.Regexp

func init() {
	reg, err := regexp.Compile("\\?|\\*|\\{((?:\\{[^/]+?\\}|[^/{}]|\\\\[{}])+?)\\}")
	if err == nil {
		GlobPattern = reg
	}

}

// AntPathMatcher 实现了接口 PathMatcher
type AntPathMatcher struct {
	pathSeparator         string
	tokenizedPatternCache *utils.SyncMap // 标记化模式缓存（线程安全）
	stringMatcherCache    *utils.SyncMap // 字符串匹配器缓存（线程安全）

	pathSeparatorPatternCache *PathSeparatorPatternCache
	caseSensitive             bool // 区分大小写,默认值为true
	trimTokens                bool // 默认值为false
	cachePatterns             bool // 默认值为true
}

func New() *AntPathMatcher {
	ant := NewS(DefaultPathSeparator)
	return ant
}
func NewS(separator string) *AntPathMatcher {
	if strings.EqualFold(utils.EmptyString, separator) {
		separator = DefaultPathSeparator
	}
	ant := &AntPathMatcher{}
	//
	ant.pathSeparator = separator
	ant.tokenizedPatternCache = new(utils.SyncMap)
	ant.stringMatcherCache = new(utils.SyncMap)
	ant.pathSeparatorPatternCache = NewDefaultPathSeparatorPatternCache(separator)

	// filed
	ant.caseSensitive = true
	ant.trimTokens = false
	ant.cachePatterns = true
	return ant
}

// @Override
// IsPattern
func (ant *AntPathMatcher) IsPattern(path string) bool {
	return strings.Index(path, "*") != -1 || strings.Index(path, "?") != -1
}

// @Override
// Match
func (ant *AntPathMatcher) Match(pattern, path string) bool {
	return ant.doMatch(pattern, path, true, nil)
}

// @Override
// MatchStart
func (ant *AntPathMatcher) MatchStart(pattern, path string) bool {
	return ant.doMatch(pattern, path, false, nil)
}

// @Override
// ExtractPathWithinPattern
func (ant *AntPathMatcher) ExtractPathWithinPattern(pattern, path string) string {
	patternParts := utils.TokenizeToStringArray(pattern, ant.pathSeparator, ant.trimTokens, true)
	pathParts := utils.TokenizeToStringArray(path, ant.pathSeparator, ant.trimTokens, true)
	builder := utils.EmptyString
	pathStarted := false
	for segment := 0; segment < len(patternParts); segment++ {
		patternPart := patternParts[segment]
		if strings.Index(*patternPart, "*") > -1 || strings.Index(*patternPart, "?") > -1 {
			for ; segment < len(pathParts); segment++ {
				if pathStarted || (segment == 0 && !strings.HasPrefix(pattern, ant.pathSeparator)) {
					builder += ant.pathSeparator
				}
				builder += *pathParts[segment]
				pathStarted = true
			}
		}
	}

	return builder
}

// @Override
// ExtractUriTemplateVariables
func (ant *AntPathMatcher) ExtractUriTemplateVariables(pattern, path string) *map[string]string {
	variables := make(map[string]string)
	result := ant.doMatch(pattern, path, true, &variables)
	if !result {
		panic("Pattern \"" + pattern + "\" is not a match for \"" + path + "\"")
	}
	return &variables
}

// @Override
// GetPatternComparator
func (ant *AntPathMatcher) GetPatternComparator(path string) *AntPatternComparator {
	return NewDefaultAntPatternComparator(path)
}

// @Override
// Combine 将pattern1和pattern2联合成一个新的pattern
func (ant *AntPathMatcher) Combine(pattern1, pattern2 string) string {
	if !utils.HasText(pattern1) && !utils.HasText(pattern2) {
		return ""
	}
	if !utils.HasText(pattern1) {
		return pattern2
	}
	if !utils.HasText(pattern2) {
		return pattern1
	}
	// 处理pattern
	pattern1ContainsUriVar := strings.Index(pattern1, "{") != -1
	if !strings.EqualFold(pattern1, pattern2) && !pattern1ContainsUriVar && ant.Match(pattern1, pattern2) {
		// /* + /hotel -> /hotel ; "/*.*" + "/*.html" -> /*.html
		// However /user + /user -> /usr/user ; /{foo} + /bar -> /{foo}/bar
		return pattern2
	}
	// /hotels/* + /booking -> /hotels/booking
	// /hotels/* + booking -> /hotels/booking
	if strings.HasSuffix(pattern1, ant.pathSeparatorPatternCache.GetEndsOnWildCard()) {
		return ant.concat(pattern1[0:len(pattern1)-2], pattern2)
	}

	// /hotels/** + /booking -> /hotels/**/booking
	// /hotels/** + booking -> /hotels/**/booking
	if strings.HasSuffix(pattern1, ant.pathSeparatorPatternCache.GetEndsOnDoubleWildCard()) {
		return ant.concat(pattern1, pattern2)
	}

	starDotPos1 := strings.Index(pattern1, "*.")
	if pattern1ContainsUriVar || starDotPos1 == -1 || strings.EqualFold(".", ant.pathSeparator) {
		// simply concatenate the two patterns
		return ant.concat(pattern1, pattern2)
	}

	ext1 := pattern1[starDotPos1+1:]
	dotPos2 := strings.Index(pattern2, ".")
	file2 := utils.EmptyString
	ext2 := utils.EmptyString
	if dotPos2 == -1 {
		file2 = pattern2
		ext2 = ""
	} else {
		file2 = pattern2[0:dotPos2]
		ext2 = pattern2[dotPos2:]
	}
	ext1All := strings.EqualFold(".*", ext1) || strings.EqualFold(utils.EmptyString, ext1)
	ext2All := strings.EqualFold(".*", ext2) || strings.EqualFold(utils.EmptyString, ext2)
	if !ext1All && !ext2All {
		panic("Cannot combine patterns: " + pattern1 + " vs " + pattern2)
	}
	//
	ext := utils.EmptyString
	if ext1All {
		ext = ext2
	} else {
		ext = ext1
	}
	return file2 + ext
}

func (ant *AntPathMatcher) PatternCacheSize() int64 {
	return ant.stringMatcherCache.MyLen()
}

// SetPathSeparator The default is "/",as in ant.
/**
 *Set the path separator to use for pattern parsing.
 */
func (ant *AntPathMatcher) SetPathSeparator(pathSeparator string) {
	if !strings.EqualFold(utils.EmptyString, pathSeparator) {
		ant.pathSeparator = pathSeparator
		ant.pathSeparatorPatternCache = NewDefaultPathSeparatorPatternCache(pathSeparator)
	}
}

// SetCaseSensitive 区分大小写 The default is false
/*
 * Specify whether to perform pattern matching in a case-sensitive fashion.
 * Default is {@code true}. Switch this to {@code false} for case-insensitive matching.
 */
func (ant *AntPathMatcher) SetCaseSensitive(caseSensitive bool) {
	ant.caseSensitive = caseSensitive
}

// SetTrimTokens 是否去除空格 The default is false
/**
 *Specify whether to trim tokenized paths and patterns.
 */
func (ant *AntPathMatcher) SetTrimTokens(trimTokens bool) {
	ant.trimTokens = trimTokens
}

// SetCachePatterns
/**
 * Specify whether to cache parsed pattern metadata for patterns passed
 * into this matcher's {@link #match} method. A value of {@code true}
 * activates an unlimited pattern cache; a value of {@code false} turns
 * the pattern cache off completely.
 * <p>Default is for the cache to be on, but with the variant to automatically
 * turn it off when encountering too many patterns to cache at runtime
 * (the threshold is 65536), assuming that arbitrary permutations of patterns
 * are coming in, with little chance for encountering a recurring pattern.
 */
func (ant *AntPathMatcher) SetCachePatterns(cachePatterns bool) {
	ant.cachePatterns = cachePatterns
}

/**
 *实际上将给定的{@code path}与给定的{@code pattern}相匹配。
 *@param pattern要匹配的模式
 *@param path要测试的路径字符串
 *@param fullMatch是否需要完整的模式匹配（否则为模式匹配只要给定的基本路径就足够了）
 *@return {@code true}（如果提供的{@code path}匹配，{@ code false}，如果不匹配）
 */
func (ant *AntPathMatcher) doMatch(pattern, path string, fullMatch bool, uriTemplateVariables *map[string]string) bool {
	if strings.HasPrefix(path, ant.pathSeparator) != strings.HasPrefix(pattern, ant.pathSeparator) {
		return false
	}
	pattDirs := ant.tokenizePattern(pattern)
	if fullMatch && ant.caseSensitive && !ant.isPotentialMatch(path, pattDirs) {
		return false
	}

	pathDirs := ant.tokenizePath(path)
	// define variable
	pattIdxStart := 0
	pattIdxEnd := len(pattDirs) - 1
	pathIdxStart := 0
	pathIdxEnd := len(pathDirs) - 1

	// Match all elements up to the first **
	for {
		if pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
			pattDir := pattDirs[pattIdxStart]
			if strings.EqualFold("**", *pattDir) {
				break
			}
			if !ant.matchStrings(*pattDir, *pathDirs[pathIdxStart], uriTemplateVariables) {
				return false
			}
			pattIdxStart++
			pathIdxStart++
		} else {
			// jump out of
			break
		}
	}

	if pathIdxStart > pathIdxEnd {
		// Path is exhausted, only match if rest of pattern is * or **'s
		if pattIdxStart > pattIdxEnd {
			return strings.HasSuffix(pattern, ant.pathSeparator) == strings.HasSuffix(path, ant.pathSeparator)
		}
		if !fullMatch {
			return true
		}
		if pattIdxStart == pattIdxEnd && strings.EqualFold("*", *pattDirs[pattIdxStart]) && strings.HasSuffix(path, ant.pathSeparator) {
			return true
		}
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if !strings.EqualFold("**", *pattDirs[i]) {
				return false
			}
		}
		return true
	} else if pattIdxStart > pattIdxEnd {
		// String not exhausted, but pattern is. Failure.
		return false
	} else if !fullMatch && strings.EqualFold("**", *pattDirs[pattIdxStart]) {
		// Path start definitely matches due to "**" part in pattern.
		return true
	}

	// up to last '**'
	for {
		if pattIdxStart <= pattIdxEnd && pathIdxStart <= pathIdxEnd {
			pattDir := pattDirs[pattIdxEnd]
			if strings.EqualFold("**", *pattDir) {
				break
			}
			if !ant.matchStrings(*pattDir, *pathDirs[pathIdxEnd], uriTemplateVariables) {
				return false
			}
			pattIdxEnd--
			pathIdxEnd--
		} else {
			break
		}
	}
	if pathIdxStart > pathIdxEnd {
		// String is exhausted
		for i := pattIdxStart; i <= pattIdxEnd; i++ {
			if !strings.EqualFold("**", *pattDirs[i]) {
				return false
			}
		}
		return true
	}

	for {
		if pattIdxStart != pattIdxEnd && pathIdxStart <= pathIdxEnd {
			patIdxTmp := -1
			for i := pattIdxStart + 1; i <= pattIdxEnd; i++ {
				if strings.EqualFold("**", *pattDirs[i]) {
					patIdxTmp = i
					break
				}
			}
			if patIdxTmp == pattIdxStart+1 {
				// '**/**' situation, so skip one
				pattIdxStart++
				continue
			}
			// Find the pattern between padIdxStart & padIdxTmp in str between
			// strIdxStart & strIdxEnd
			patLength := patIdxTmp - pattIdxStart - 1
			strLength := pathIdxEnd - pathIdxStart + 1
			foundIdx := -1

		strLoop:
			for i := 0; i <= strLength-patLength; i++ {
				for j := 0; j < patLength; j++ {
					subPat := pattDirs[pattIdxStart+j+1]
					subStr := pathDirs[pathIdxStart+i+j]
					if !ant.matchStrings(*subPat, *subStr, uriTemplateVariables) {
						continue strLoop
					}
				}
				foundIdx = pathIdxStart + i
				break
			}

			if foundIdx == -1 {
				return false
			}

			pattIdxStart = patIdxTmp
			pathIdxStart = foundIdx + patLength
		} else {
			break
		}
	}

	for i := pattIdxStart; i <= pattIdxEnd; i++ {
		if !strings.EqualFold("**", *pattDirs[i]) {
			return false
		}
	}
	return true
}

// tokenizePattern default use cache
/**
 * Tokenize the given path pattern into parts, based on this matcher's settings.
 * <p>Performs caching based on {@link #setCachePatterns}, delegating to
 * {@link #tokenizePath(String)} for the actual tokenization algorithm.
 * @param pattern the pattern to tokenize
 * @return the tokenized pattern parts
 */
func (ant *AntPathMatcher) tokenizePattern(pattern string) []*string {
	tokenized := make([]*string, 0)
	// The first step is to fetch from the cache map.
	value, ok := ant.tokenizedPatternCache.MyLoad(pattern)
	if ok {
		tokenized = value.([]*string)
	} else {
		// No records was fetched from the cache map.
		tokenized = ant.tokenizePath(pattern)
		// add
		if tokenized != nil {
			ant.tokenizedPatternCache.MyStore(pattern, tokenized)
		}
	}
	return tokenized
}

// tokenizePath
func (ant *AntPathMatcher) tokenizePath(path string) []*string {
	return utils.TokenizeToStringArray(path, ant.pathSeparator, ant.trimTokens, true)
}

// isPotentialMatch
func (ant *AntPathMatcher) isPotentialMatch(path string, pattDirs []*string) bool {
	if !ant.trimTokens {
		pos := 0
		for _, pattDir := range pattDirs {
			skipped := ant.skipSeparator(path, pos, ant.pathSeparator)
			pos += skipped
			skipped = ant.skipSegment(path, pos, *pattDir)
			if skipped < utf8.RuneCountInString(*pattDir) {
				tempPattDir := rune((*pattDir)[0])
				return skipped > 0 || utf8.RuneCountInString(*pattDir) > 0 && ant.isWildcardChar(tempPattDir)
			}
			pos += skipped
		}
	}
	return true
}

// skipSegment
func (ant *AntPathMatcher) skipSegment(path string, pos int, prefix string) int {
	skipped := 0
	for i := 0; i < utf8.RuneCountInString(prefix); i++ {
		c := rune(prefix[i])
		if ant.isWildcardChar(c) {
			return skipped
		}
		currPos := pos + skipped
		if currPos >= utf8.RuneCountInString(path) {
			return 0
		}
		if c == rune(path[currPos]) {
			skipped++
		}
	}
	return skipped
}

// skipSeparator
func (ant *AntPathMatcher) skipSeparator(path string, pos int, separator string) int {
	skipped := 0
	for {
		if utils.StartsWith(path, separator, pos+skipped) {
			skipped += utf8.RuneCountInString(separator)
		} else {
			break
		}
	}
	return skipped
}

// isWildcardChar
func (ant *AntPathMatcher) isWildcardChar(c rune) bool {
	for _, candidate := range WildcardChars {
		if c == candidate {
			return true
		}
	}
	return false
}

/**
* Test whether or not a string matches against a pattern.
*
* @param pattern the pattern to match against (never {@code null})
* @param str     the String which must be matched against the pattern (never {@code null})
* @return {@code true} if the string matches against the pattern, or {@code false} otherwise
 */
// MatchStrings
func (ant *AntPathMatcher) matchStrings(pattern, str string, uriTemplateVariables *map[string]string) bool {
	return ant.getStringMatcher(pattern).MatchStrings(str, uriTemplateVariables)
}

/**
*为给定模式构建或检索{@link AntPathStringMatcher}。
*默认实现检查此AntPathMatcher的内部缓存（请参阅{@link #setCachePatterns}），如果未找到任何缓存副本，则创建一个新的AntPathStringMatcher实例。
*当遇到太多无法在运行时进行缓存的模式时（阈值为65536），它会关闭默认的缓存，并假设模式的任意排列即将到来，而遇到重复模式的机会很小。
*可以重写此方法以实现自定义缓存策略。
*@param pattern要匹配的模式（永远{@code null}）
*@返回相应的AntPathStringMatcher（从不{@code null}）
 */
func (ant *AntPathMatcher) getStringMatcher(pattern string) *AntPathStringMatcher {
	var matcher *AntPathStringMatcher
	cachePatterns := ant.cachePatterns
	if cachePatterns {
		value, ok := ant.stringMatcherCache.MyLoad(pattern)
		if ok && value != nil {
			matcher = value.(*AntPathStringMatcher)
		}
	}
	if matcher == nil {
		matcher = NewMatchesStringMatcher(pattern, ant.caseSensitive)
		if cachePatterns && ant.PatternCacheSize() >= CacheTurnoffThreshold {
			// Try to adapt to the runtime situation that we're encountering:
			// There are obviously too many different patterns coming in here...
			// So let's turn off the cache since the patterns are unlikely to be reoccurring.
			ant.deactivatePatternCache()
			return matcher
		}
		if cachePatterns {
			ant.stringMatcherCache.MyStore(pattern, matcher)
		}
	}
	return matcher
}

// concat
func (ant *AntPathMatcher) concat(path1, path2 string) string {
	path1EndsWithSeparator := strings.HasSuffix(path1, ant.pathSeparator)
	path2StartsWithSeparator := strings.HasPrefix(path2, ant.pathSeparator)

	if path1EndsWithSeparator && path2StartsWithSeparator {
		return path1 + path2[1:]
	} else if path1EndsWithSeparator || path2StartsWithSeparator {
		return path1 + path2
	} else {
		return path1 + ant.pathSeparator + path2
	}
}

// deactivatePatternCache
func (ant *AntPathMatcher) deactivatePatternCache() {
	ant.cachePatterns = false
	utils.ClearSyncMap(ant.tokenizedPatternCache)
	utils.ClearSyncMap(ant.stringMatcherCache)
}

/**
// QuoteMeta 将字符串 s 中的“特殊字符”转换为其“转义格式”
// 例如，QuoteMeta（`[foo]`）返回`\[foo\]`。
// 特殊字符有：\.+*?()|[]{}^$
// 这些字符用于实现正则语法，所以当作普通字符使用时需要转换
func QuoteMeta
// 通过 Complite、CompilePOSIX、MustCompile、MustCompilePOSIX
// 四个函数可以创建一个 Regexp 对象
struct Regexp
// Compile 用来解析正则表达式 expr 是否合法，如果合法，则返回一个 Regexp 对象
// Regexp 对象可以在任意文本上执行需要的操作
func Compile

// 在 s 中查找 re 中编译好的正则表达式，并返回第一个匹配的内容
// 同时返回子表达式匹配的内容
// {完整匹配项, 子匹配项, 子匹配项, ...}
func FindStringSubmatch
// 在 b 中查找 re 中编译好的正则表达式，并返回第一个匹配的内容
// 同时返回子表达式匹配的内容
// {{完整匹配项}, {子匹配项}, {子匹配项}, ...}
func FindSubmatch
*/

/**
* Tests whether or not a string matches against a pattern via a {@link Pattern}.
* <p>The pattern may contain special characters: '*' means zero or more characters; '?' means one and
* only one character; '{' and '}' indicate a URI template pattern. For example <tt>/users/{user}</tt>.
 */
// AntPathStringMatcher
type AntPathStringMatcher struct {
	// variableNames
	variableNames []*string
	// pattern
	pattern *regexp.Regexp

	// caseSensitive 区分大小写
	caseSensitive bool

	// capturingGroupCount 内部变量计算匹配个数
	capturingGroupCount int
}

// NewDefaultStringMatcher part match
func NewDefaultStringMatcher(pattern string, caseSensitive bool) *AntPathStringMatcher {
	stringMatcher := &AntPathStringMatcher{}
	stringMatcher.capturingGroupCount = 0
	stringMatcher.variableNames = make([]*string, 0)
	// caseSensitive
	stringMatcher.caseSensitive = caseSensitive
	// 写入表达式
	reg, err := regexp.Compile(*stringMatcher.patternBuilder(pattern, false, caseSensitive))
	if err == nil {
		stringMatcher.pattern = reg
	}
	return stringMatcher
}

// NewMatchesStringMatcher full match
func NewMatchesStringMatcher(pattern string, caseSensitive bool) *AntPathStringMatcher {
	stringMatcher := &AntPathStringMatcher{}
	stringMatcher.capturingGroupCount = 0
	stringMatcher.variableNames = make([]*string, 0)
	// caseSensitive
	stringMatcher.caseSensitive = caseSensitive
	// 写入表达式
	reg, err := regexp.Compile(*stringMatcher.patternBuilder(pattern, true, caseSensitive))
	if err == nil {
		stringMatcher.pattern = reg
	}
	return stringMatcher
}

/**
* Main entry point.
*
* @return {@code true} if the string matches against the pattern, or {@code false} otherwise.
 */
// MatchStrings
func (sm *AntPathStringMatcher) MatchStrings(str string, uriTemplateVariables *map[string]string) bool {
	// 区分大小写
	if !sm.caseSensitive {
		str = strings.ToLower(str)
	}
	// byte
	matchBytes := utils.Str2Bytes(str)
	findIndex := sm.pattern.FindSubmatch(matchBytes)
	if len(findIndex) > 0 {
		if uriTemplateVariables != nil {
			// SPR-8455
			if len(sm.variableNames) != sm.GroupCount() {
				panic("The number of capturing groups in the pattern segment " +
					sm.pattern.String() + " does not match the number of URI template variables it defines, " +
					"which can occur if capturing groups are used in a URI template regex. " +
					"Use non-capturing groups instead.")
			}
			for i := 1; i <= sm.GroupCount(); i++ {
				name := sm.variableNames[i-1]
				// 获取匹配位置
				matched := findIndex[i]
				value := utils.Bytes2Str(matched)
				(*uriTemplateVariables)[*name] = value
			}
		}
		return true
	} else {
		return false
	}
}

// GroupCount
func (sm *AntPathStringMatcher) GroupCount() int {
	return sm.capturingGroupCount
}

// FindSubMatch 子查询
func (sm *AntPathStringMatcher) FindSubMatch(source []byte, index int) *string {
	indexCollection := sm.pattern.FindSubmatch(source)
	result := utils.Bytes2Str(indexCollection[index])
	return &result
}

// takeOffBrackets
func (sm *AntPathStringMatcher) takeOffBrackets(source *string) *string {
	var temp = strings.Trim(*source, "{}")
	return &temp
}

// quote
func (sm *AntPathStringMatcher) quote(s string, start, end int) string {
	if start == end {
		return ""
	}
	return regexp.QuoteMeta(s[start:end])
}

// patternBuilder
func (sm *AntPathStringMatcher) patternBuilder(pattern string, matches, caseSensitive bool) *string {
	// 字符串拼接
	var patternBuilder string
	end := 0
	patternBytes := utils.Str2Bytes(pattern)
	allIndex := GlobPattern.FindAllIndex(patternBytes, MaxFindCount)
	if allIndex != nil && len(allIndex) > 0 {
		for _, matched := range allIndex {
			matchedStart := matched[0]
			matchedEnd := matched[1]
			patternBuilder += sm.quote(pattern, end, matchedStart)
			// matchString
			matchstr := utils.Bytes2Str(patternBytes[matchedStart:matchedEnd])
			if strings.EqualFold("?", matchstr) {
				patternBuilder += "."
			} else if strings.EqualFold("*", matchstr) {
				patternBuilder += ".*"
			} else if strings.HasPrefix(matchstr, "{") && strings.HasSuffix(matchstr, "}") {
				colonIdx := strings.Index(matchstr, ":")
				if colonIdx == -1 {
					patternBuilder += DefaultVariablePattern
					sm.variableNames = append(sm.variableNames, sm.takeOffBrackets(&matchstr))
				} else {
					bytes := utils.Str2Bytes(matchstr)
					variablePattern := utils.Bytes2Str(bytes[colonIdx+1 : len(matchstr)-1])
					patternBuilder += "("
					patternBuilder += variablePattern
					patternBuilder += ")"
					variableName := utils.Bytes2Str(bytes[1:colonIdx])
					sm.variableNames = append(sm.variableNames, &variableName)
				}
				// group
				sm.capturingGroupCount++
			}
			// 向后增加end
			end = matchedEnd
		}
	}
	// patternBuilder
	patternBuilder += sm.quote(pattern, end, len(pattern))
	if !caseSensitive {
		patternBuilder = strings.ToLower(patternBuilder)
	}
	// full match
	if matches {
		patternBuilder = "^" + patternBuilder + "$"
	}

	return &patternBuilder
}
