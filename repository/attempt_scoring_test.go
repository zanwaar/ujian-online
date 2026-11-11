package repository

import "testing"

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name           string
		correctCount   int64
		totalQuestions int64
		expectedResult float64
	}{
		{name: "empty attempt", correctCount: 0, totalQuestions: 0, expectedResult: 0},
		{name: "no correct answers", correctCount: 0, totalQuestions: 4, expectedResult: 0},
		{name: "half correct", correctCount: 1, totalQuestions: 2, expectedResult: 50},
		{name: "all correct", correctCount: 4, totalQuestions: 4, expectedResult: 100},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if result := calculateScore(testCase.correctCount, testCase.totalQuestions); result != testCase.expectedResult {
				t.Fatalf("calculateScore(%d, %d) = %v, want %v", testCase.correctCount, testCase.totalQuestions, result, testCase.expectedResult)
			}
		})
	}
}
