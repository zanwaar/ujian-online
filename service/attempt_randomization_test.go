package service

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"golang-backend/entity"
)

func TestShuffleQuestionsIsDeterministicForTheSameSeed(t *testing.T) {
	questions := []entity.Question{
		{Code: "Q1"},
		{Code: "Q2"},
		{Code: "Q3"},
		{Code: "Q4"},
	}
	first := append([]entity.Question(nil), questions...)
	second := append([]entity.Question(nil), questions...)

	shuffleQuestions(rand.New(rand.NewSource(42)), first)
	shuffleQuestions(rand.New(rand.NewSource(42)), second)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same seed produced different question order: first=%v second=%v", first, second)
	}
}

func TestShuffleOptionsKeepsKeysAndContentPaired(t *testing.T) {
	options := []snapshotOption{
		{OptionKey: "A", ContentHTML: "content-a"},
		{OptionKey: "B", ContentHTML: "content-b"},
		{OptionKey: "C", ContentHTML: "content-c"},
		{OptionKey: "D", ContentHTML: "content-d"},
	}
	optionKeys := []string{"A", "B", "C", "D"}

	shuffleOptions(rand.New(rand.NewSource(42)), options, optionKeys)

	if !reflect.DeepEqual(optionKeys, []string{"C", "D", "A", "B"}) {
		t.Fatalf("unexpected deterministic option order: %v", optionKeys)
	}
	for index, option := range options {
		if option.OptionKey != optionKeys[index] || option.ContentHTML != "content-"+strings.ToLower(optionKeys[index]) {
			t.Fatalf("option key/content pairing changed: options=%v keys=%v", options, optionKeys)
		}
	}
}
