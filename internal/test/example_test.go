package test_test

import (
	"strconv"
	"strings"
	"testing"
)

func TestFail(t *testing.T) {
	t.Error("fail")
	for i := 0; i < 3; i++ {
		t.Run("fail"+strconv.Itoa(i), func(t *testing.T) {
			t.Error("fail")
		})
	}
}

func TestSuccess(t *testing.T) {
	t.Log("good")
	for j := 0; j < 3; j++ {
		t.Run("success"+strconv.Itoa(j), func(t *testing.T) {
			t.Log("good")
		})
	}
}

func TestSkip(t *testing.T) {
	t.Skip("skip")
	for i := 0; i < 3; i++ {
		t.Run("skip"+strconv.Itoa(i), func(t *testing.T) {
			t.Skip("skip")
		})
	}
}

func BenchmarkStringConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s string
		for j := 0; j < 100; j++ {
			s += "test"
		}
	}
}

func BenchmarkStringBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var builder strings.Builder
		for j := 0; j < 100; j++ {
			builder.WriteString("test")
		}
		_ = builder.String()
	}
}
