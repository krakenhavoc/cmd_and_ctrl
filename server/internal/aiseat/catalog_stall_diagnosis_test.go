package aiseat_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDiagnoseStallNamesTheLoopBreaker pins #2609: a catalog-soak
// stall with nothing pending used to print kinds=[] first="" reason=""
// and read as a hang, when the reproducing seeds were the CR 732
// breaker parking a bot-only table (ADR 0055 section 5).
func TestDiagnoseStallNamesTheLoopBreaker(t *testing.T) {
	g := newRoom(t, 4, 7).Game

	_, d := diagnoseStall(g)
	if !strings.HasPrefix(d.cause, "UNEXPLAINED") {
		t.Fatalf("a quiet table with no notice must be reported as unexplained, got %q", d.cause)
	}

	g.LoopNotice = &game.LoopNotice{Source: uuid.New(), Label: "Nomads en-Kor - redirect", Count: 25}
	_, d = diagnoseStall(g)
	if !strings.Contains(d.cause, "loop breaker") || !strings.Contains(d.cause, "Nomads en-Kor") || !strings.Contains(d.cause, "25") {
		t.Fatalf("a standing loop notice must be named in the cause, got %q", d.cause)
	}
}
