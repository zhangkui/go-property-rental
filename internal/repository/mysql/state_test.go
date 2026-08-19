package mysql

import "testing"

func TestLeaseTransitions(t *testing.T) {
	if !validLeaseTransition("draft", "pending") {
		t.Fatal("draft to pending must be valid")
	}
	if validLeaseTransition("draft", "closed") {
		t.Fatal("draft to closed must be rejected")
	}
}
func TestWorkOrderTransitions(t *testing.T) {
	if !validWorkTransition("assigned", "repairing") {
		t.Fatal("assigned to repairing must be valid")
	}
	if validWorkTransition("reported", "completed") {
		t.Fatal("reported to completed must be rejected")
	}
}
