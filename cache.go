package antstyle

/**
 *  用于依赖配置的路径分隔符的模式的简单缓存。
 */
type PathSeparatorPatternCache struct {
	// "*"
	endsOnWildCard string
	// "**"
	endsOnDoubleWildCard string
}

// NewDefaultPathSeparatorPatternCache 构造函数
func NewDefaultPathSeparatorPatternCache(pathSeparator string) *PathSeparatorPatternCache {
	patternCache := &PathSeparatorPatternCache{}
	patternCache.endsOnWildCard = pathSeparator + "*"
	patternCache.endsOnDoubleWildCard = pathSeparator + "**"
	return patternCache
}

// GetEndsOnWildCard 返回 "*"
func (patternCache *PathSeparatorPatternCache) GetEndsOnWildCard() string {
	return patternCache.endsOnWildCard
}

// GetEndsOnDoubleWildCard 返回 "**"
func (patternCache *PathSeparatorPatternCache) GetEndsOnDoubleWildCard() string {
	return patternCache.endsOnDoubleWildCard
}
