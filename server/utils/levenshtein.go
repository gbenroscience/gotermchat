package utils

//StringCompare - Comparer of strings using Levenshtein's algorithm
type StringCompare struct {
	Source string
	Target string
}

//ComputeDistance - The Levenshtein algorithm for finding the distance or the number of steps
// needed to convert one string into another
func (compare *StringCompare) ComputeDistance() int {

	sourceLen := len(compare.Source)
	targetLen := len(compare.Target)

	if compare.Source == compare.Target {
		return 0
	}

	if sourceLen == 0 {
		return targetLen
	}

	if targetLen == 0 {
		return sourceLen
	}

	rows := (sourceLen + 1)
	cols := (targetLen + 1)
	distance := make([][]int, rows)

	//create the columns on the rows
	for i := 0; i < rows; i++ {
		distance[i] = make([]int, cols)
	}

	for i := 0; i <= sourceLen; i++ {
		distance[i][0] = i
	}

	for j := 0; j <= targetLen; j++ {
		distance[0][j] = j
	}

	for i := 1; i <= sourceLen; i++ {
		for j := 1; j <= targetLen; j++ {
			var cost int
			if compare.Target[j-1:j] == compare.Source[i-1:i] {
				cost = 0
			} else {
				cost = 1
			}
			distance[i][j] = min(min(distance[i-1][j]+1, distance[i][j-1]+1), distance[i-1][j-1]+cost)
		}
	}
	return distance[sourceLen][targetLen]
}

//IsSimilarUpTo - The tolerance between 0.0 and 1.0...where 0.0 means they are not similar at all and 1.0 means they are same
// It returns true if the strings are similar such that their percent similarity is greater than or equal to tolerance * 100
func (compare *StringCompare) IsSimilarUpTo(similarityFraction float64) bool {
	sourceLen := len(compare.Source)
	targetLen := len(compare.Target)

	if sourceLen == 0 || targetLen == 0 {
		return false
	}

	if compare.Source == compare.Target {
		return true
	}

	distance := compare.ComputeDistance()

	return (1.0 - (float64(distance) / float64(max(sourceLen, targetLen)))) >= similarityFraction

}

//IsSimilarTo -  derives from `IsSimilarUpTo` except that the similarity scale is measured in percentages
// instead of being measured as a decimal between 0.0 and 1.0
// This function returns true if the strings compared are at least `similarityPercent` like one another
func (compare *StringCompare) IsSimilarTo(similarityPercent float64) bool {
	return compare.IsSimilarUpTo(similarityPercent / 100.0)

}

func min(a int, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func max(a int, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}
