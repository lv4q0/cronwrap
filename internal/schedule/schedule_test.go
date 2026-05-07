package schedule

import (
	"testing"
	"time"
)

func TestParseValidExpression(t *testing.T) {
	_, err := Parse("*/5 * * * *")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestParseDescriptor(t *testing.T) {
	_, err := Parse("@hourly")
	if err != nil {
		t.Fatalf("expected no error for @hourly, got %v", err)
	}
}

func TestParseInvalidExpression(t *testing.T) {
	_, err := Parse("not-a-cron")
	if err == nil {
		t.Fatal("expected error for invalid expression, got nil")
	}
}

func TestNext(t *testing.T) {
	e, err := Parse("0 * * * *") // top of every hour
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	base := time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)
	next := e.Next(base)
	expected := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("Next() = %v, want %v", next, expected)
	}
}

func TestIsDue(t *testing.T) {
	e, _ := Parse("30 10 * * *") // 10:30 daily
	tolerance := 30 * time.Second

	due := time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC)
	if !e.IsDue(due, tolerance) {
		t.Error("expected IsDue=true at exact scheduled time")
	}

	notDue := time.Date(2024, 1, 1, 10, 31, 45, 0, time.UTC)
	if e.IsDue(notDue, tolerance) {
		t.Error("expected IsDue=false well past scheduled time")
	}
}

func TestUntilNext(t *testing.T) {
	e, _ := Parse("0 12 * * *") // noon daily
	now := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	d := e.UntilNext(now)
	if d != time.Hour {
		t.Errorf("UntilNext() = %v, want 1h", d)
	}
}
