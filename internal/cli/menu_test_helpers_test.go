package cli

import "testing"

func assertMenuDestination(t *testing.T, got menuTransition, want menuDestination) {
	t.Helper()
	if got.destination != want {
		t.Fatalf("destination = %v, want %v", got.destination, want)
	}
}

func assertMenuTransition(t *testing.T, want menuDestination, run func() menuTransition) menuTransition {
	t.Helper()
	got := run()
	assertMenuDestination(t, got, want)
	return got
}

func assertMenuTransitionWithPrompts(t *testing.T, want menuDestination, run func() menuTransition, responses ...string) menuTransition {
	t.Helper()
	restore := withPromptResponses(t, responses...)
	defer restore()
	return assertMenuTransition(t, want, run)
}

func assertResetAllRequested(t *testing.T, picker *stubRuntime) {
	t.Helper()
	if picker.commands.resetAllCalls != 1 {
		t.Fatalf("reset all requests = %d, want 1", picker.commands.resetAllCalls)
	}
}

func assertNoResetAllRequested(t *testing.T, picker *stubRuntime) {
	t.Helper()
	if picker.commands.resetAllCalls != 0 {
		t.Fatalf("reset all requests = %d, want 0", picker.commands.resetAllCalls)
	}
}

func assertCategoryResetRequested(t *testing.T, picker *stubRuntime, category string) {
	t.Helper()
	if len(picker.commands.resetCategoryCalls) != 1 || picker.commands.resetCategoryCalls[0] != category {
		t.Fatalf("category reset requests = %#v, want [%s]", picker.commands.resetCategoryCalls, category)
	}
}

func assertNoCategoryResetRequested(t *testing.T, picker *stubRuntime) {
	t.Helper()
	if len(picker.commands.resetCategoryCalls) != 0 {
		t.Fatalf("category reset requests = %#v, want none", picker.commands.resetCategoryCalls)
	}
}

func assertFactoryResetRequested(t *testing.T, picker *stubRuntime) {
	t.Helper()
	if picker.commands.factoryResetCalls != 1 {
		t.Fatalf("factory reset requests = %d, want 1", picker.commands.factoryResetCalls)
	}
}

func assertNoFactoryResetRequested(t *testing.T, picker *stubRuntime) {
	t.Helper()
	if picker.commands.factoryResetCalls != 0 {
		t.Fatalf("factory reset requests = %d, want 0", picker.commands.factoryResetCalls)
	}
}

func assertWearRequested(t *testing.T, picker *stubRuntime, fileName string) {
	t.Helper()
	if len(picker.commands.wearCalls) != 1 || picker.commands.wearCalls[0].FileName != fileName {
		t.Fatalf("wear requests = %#v, want %s worn", picker.commands.wearCalls, fileName)
	}
}

func assertNoCategoryRandomRequested(t *testing.T, picker *stubRuntime) {
	t.Helper()
	if picker.random.categoryCalls != 0 {
		t.Fatalf("category random requests = %d, want 0", picker.random.categoryCalls)
	}
}

func assertCategoryRandomRequestCount(t *testing.T, picker *stubRuntime, want int) {
	t.Helper()
	if picker.random.categoryCalls != want {
		t.Fatalf("category random requests = %d, want %d", picker.random.categoryCalls, want)
	}
}
