package test_test

import (
	"strconv"
	"testing"
)

func TestSuccess(t *testing.T) {
	t.Log("good")
	for j := 0; j < 3; j++ {
		t.Run("success"+strconv.Itoa(j), func(t *testing.T) {
			t.Log("good")
		})
	}
}

func TestFail(t *testing.T) {
	t.Error("fail")
	for i := 0; i < 3; i++ {
		t.Run("fail"+strconv.Itoa(i), func(t *testing.T) {
			t.Error("fail")
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
