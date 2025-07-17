package services

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// StringMatch 字符串匹配结果
type StringMatch struct {
	PatternName string `json:"pattern_name"`
	Offset      int64  `json:"offset"`
	MatchedText string `json:"matched_text"`
}

// StringPattern 字符串模式
type StringPattern struct {
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

// StringMatcher 字符串匹配器
type StringMatcher struct {
	mutex    sync.RWMutex
	patterns map[string]string // name -> pattern
}

// NewStringMatcher 创建字符串匹配器
func NewStringMatcher() *StringMatcher {
	return &StringMatcher{
		patterns: make(map[string]string),
	}
}

// AddPattern 添加模式
func (sm *StringMatcher) AddPattern(name, pattern string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.patterns[name] = pattern
}

// RemovePattern 移除模式
func (sm *StringMatcher) RemovePattern(name string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.patterns, name)
}

// FindMatches 查找匹配
func (sm *StringMatcher) FindMatches(content []byte) []StringMatch {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var matches []StringMatch
	contentStr := string(content)
	contentLower := strings.ToLower(contentStr)

	for name, pattern := range sm.patterns {
		patternLower := strings.ToLower(pattern)

		// 查找所有匹配位置
		offset := 0
		for {
			index := strings.Index(contentLower[offset:], patternLower)
			if index == -1 {
				break
			}

			actualOffset := offset + index
			matches = append(matches, StringMatch{
				PatternName: name,
				Offset:      int64(actualOffset),
				MatchedText: contentStr[actualOffset : actualOffset+len(pattern)],
			})

			offset = actualOffset + 1
		}
	}

	return matches
}

// FindBytesMatches 查找字节匹配
func (sm *StringMatcher) FindBytesMatches(content []byte, patterns map[string][]byte) []StringMatch {
	var matches []StringMatch

	for name, pattern := range patterns {
		offset := 0
		for {
			index := bytes.Index(content[offset:], pattern)
			if index == -1 {
				break
			}

			actualOffset := offset + index
			matches = append(matches, StringMatch{
				PatternName: name,
				Offset:      int64(actualOffset),
				MatchedText: string(pattern),
			})

			offset = actualOffset + 1
		}
	}

	return matches
}

// FuzzyMatcher 模糊匹配器
type FuzzyMatcher struct {
	threshold float64
}

// NewFuzzyMatcher 创建模糊匹配器
func NewFuzzyMatcher(threshold float64) *FuzzyMatcher {
	return &FuzzyMatcher{
		threshold: threshold,
	}
}

// CalculateSimilarity 计算相似度 (使用简化的编辑距离算法)
func (fm *FuzzyMatcher) CalculateSimilarity(str1, str2 string) float64 {
	if len(str1) == 0 && len(str2) == 0 {
		return 1.0
	}

	if len(str1) == 0 || len(str2) == 0 {
		return 0.0
	}

	// 使用Levenshtein距离算法
	distance := fm.levenshteinDistance(str1, str2)
	maxLen := max(len(str1), len(str2))

	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance 计算Levenshtein距离
func (fm *FuzzyMatcher) levenshteinDistance(str1, str2 string) int {
	len1, len2 := len(str1), len(str2)

	// 创建距离矩阵
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// 初始化第一行和第一列
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// 填充矩阵
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if str1[i-1] != str2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // 删除和插入的最小值
				matrix[i-1][j-1]+cost,                   // 替换
			)
		}
	}

	return matrix[len1][len2]
}

// JaccardSimilarity 计算Jaccard相似度 (基于n-gram)
func (fm *FuzzyMatcher) JaccardSimilarity(str1, str2 string, n int) float64 {
	if len(str1) < n || len(str2) < n {
		return 0.0
	}

	ngrams1 := fm.generateNGrams(str1, n)
	ngrams2 := fm.generateNGrams(str2, n)

	intersection := 0
	union := make(map[string]bool)

	// 计算交集
	for ngram := range ngrams1 {
		union[ngram] = true
		if ngrams2[ngram] {
			intersection++
		}
	}

	// 计算并集
	for ngram := range ngrams2 {
		union[ngram] = true
	}

	if len(union) == 0 {
		return 0.0
	}

	return float64(intersection) / float64(len(union))
}

// generateNGrams 生成n-gram
func (fm *FuzzyMatcher) generateNGrams(str string, n int) map[string]bool {
	ngrams := make(map[string]bool)

	if len(str) < n {
		return ngrams
	}

	for i := 0; i <= len(str)-n; i++ {
		ngram := str[i : i+n]
		ngrams[ngram] = true
	}

	return ngrams
}

// CosineSimilarity 计算余弦相似度 (基于字符频率)
func (fm *FuzzyMatcher) CosineSimilarity(str1, str2 string) float64 {
	freq1 := fm.getCharFrequency(str1)
	freq2 := fm.getCharFrequency(str2)

	// 计算点积
	dotProduct := 0.0
	for char, freq := range freq1 {
		if freq2[char] > 0 {
			dotProduct += float64(freq * freq2[char])
		}
	}

	// 计算向量长度
	norm1 := 0.0
	for _, freq := range freq1 {
		norm1 += float64(freq * freq)
	}
	norm1 = sqrt(norm1)

	norm2 := 0.0
	for _, freq := range freq2 {
		norm2 += float64(freq * freq)
	}
	norm2 = sqrt(norm2)

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (norm1 * norm2)
}

// getCharFrequency 获取字符频率
func (fm *FuzzyMatcher) getCharFrequency(str string) map[rune]int {
	freq := make(map[rune]int)
	for _, char := range str {
		freq[char]++
	}
	return freq
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}

	// 使用牛顿法计算平方根
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// HexStringMatcher 十六进制字符串匹配器
type HexStringMatcher struct {
	patterns map[string][]byte
}

// NewHexStringMatcher 创建十六进制字符串匹配器
func NewHexStringMatcher() *HexStringMatcher {
	return &HexStringMatcher{
		patterns: make(map[string][]byte),
	}
}

// AddHexPattern 添加十六进制模式
func (hsm *HexStringMatcher) AddHexPattern(name, hexPattern string) error {
	// 移除空格和分隔符
	hexPattern = strings.ReplaceAll(hexPattern, " ", "")
	hexPattern = strings.ReplaceAll(hexPattern, "-", "")
	hexPattern = strings.ReplaceAll(hexPattern, ":", "")

	// 转换为字节数组
	bytes := make([]byte, 0, len(hexPattern)/2)
	for i := 0; i < len(hexPattern); i += 2 {
		if i+1 >= len(hexPattern) {
			break
		}

		var b byte
		_, err := fmt.Sscanf(hexPattern[i:i+2], "%02x", &b)
		if err != nil {
			return err
		}
		bytes = append(bytes, b)
	}

	hsm.patterns[name] = bytes
	return nil
}

// FindHexMatches 查找十六进制匹配
func (hsm *HexStringMatcher) FindHexMatches(content []byte) []StringMatch {
	var matches []StringMatch

	for name, pattern := range hsm.patterns {
		offset := 0
		for {
			index := bytes.Index(content[offset:], pattern)
			if index == -1 {
				break
			}

			actualOffset := offset + index
			matches = append(matches, StringMatch{
				PatternName: name,
				Offset:      int64(actualOffset),
				MatchedText: fmt.Sprintf("%x", pattern),
			})

			offset = actualOffset + 1
		}
	}

	return matches
}
