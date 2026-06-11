package cli

import "testing"

func TestOutfitSession_GlobalSkippedTracking(t *testing.T) {
	session := NewOutfitSession()
	if session.IsGloballySkipped("outfit1") {
		t.Fatal("outfit should not be globally skipped yet")
	}

	session.AddSkipped("outfit1")
	if !session.IsGloballySkipped("outfit1") {
		t.Fatal("outfit should be globally skipped")
	}
	if got := session.GlobalSkippedCount(); got != 1 {
		t.Fatalf("GlobalSkippedCount() = %d, want 1", got)
	}

	session.ResetGlobal()
	if session.IsGloballySkipped("outfit1") {
		t.Fatal("outfit should not be globally skipped after reset")
	}
	if got := session.GlobalSkippedCount(); got != 0 {
		t.Fatalf("GlobalSkippedCount() = %d, want 0", got)
	}
}

func TestOutfitSession_CategorySkippedTracking(t *testing.T) {
	session := NewOutfitSession()
	if session.IsCategorySkipped("outfit1", "casual") {
		t.Fatal("outfit should not be skipped in category yet")
	}

	session.AddCategorySkipped("outfit1", "casual")
	if !session.IsCategorySkipped("outfit1", "casual") {
		t.Fatal("outfit should be skipped in category")
	}
	if session.IsCategorySkipped("outfit1", "formal") {
		t.Fatal("outfit should not be skipped in other category")
	}
	if got := session.CategorySkippedCount("casual"); got != 1 {
		t.Fatalf("CategorySkippedCount(casual) = %d, want 1", got)
	}

	session.ResetCategory("casual")
	if session.IsCategorySkipped("outfit1", "casual") {
		t.Fatal("outfit should not be skipped in category after reset")
	}
	if got := session.CategorySkippedCount("casual"); got != 0 {
		t.Fatalf("CategorySkippedCount(casual) = %d, want 0", got)
	}
}
