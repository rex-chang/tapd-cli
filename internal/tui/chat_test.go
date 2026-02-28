package tui_test

import (
	"testing"

	"github.com/rex-chang/tapd-cli/internal/tui"
)

func TestChatModel_AddMessage(t *testing.T) {
	m := tui.NewChatModel(80, 20)

	m.AddMessage("user", "你好")
	m.AddMessage("assistant", "你好！有什么可以帮助你的？")

	history := m.History()
	if len(history) != 2 {
		t.Errorf("expected 2 messages, got %d", len(history))
	}
	if history[0].Role != "user" {
		t.Errorf("expected user role, got %s", history[0].Role)
	}
	if history[0].Content != "你好" {
		t.Errorf("expected '你好', got %s", history[0].Content)
	}
}

func TestChatModel_Clear(t *testing.T) {
	m := tui.NewChatModel(80, 20)
	m.AddMessage("user", "test")
	m.Clear()

	if len(m.History()) != 0 {
		t.Error("expected empty history after clear")
	}
}

func TestChatModel_History_Empty(t *testing.T) {
	m := tui.NewChatModel(80, 20)
	if m.History() != nil {
		t.Error("expected nil history for new model")
	}
}
