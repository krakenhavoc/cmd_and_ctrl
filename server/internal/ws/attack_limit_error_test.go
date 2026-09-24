package ws

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestAttackLimitRefusalsReachTheWireAsIllegalAttack — #1507. An
// attack refused by a CR 508.1c count limit arrives as
// `illegal_attack` / `attack_limit` with a creature from the refused
// declaration in card_id and a sentence addressed to its reader,
// whether the hub's structured path or the plain classifier sees it,
// and however deep it is wrapped.
func TestAttackLimitRefusalsReachTheWireAsIllegalAttack(t *testing.T) {
	attacker, source, defender := uuid.New(), uuid.New(), uuid.New()
	refusal := &game.AttackLimitError{
		Scope: game.AttackLimitAttackingYou, Max: 2,
		Source: source, SourceName: "Crawlspace",
		Defender: defender, DefenderName: "P2",
		Attacker: attacker, AttackerName: "Grizzly Bears",
	}
	for _, err := range []error{refusal, fmt.Errorf("dispatch: %w", refusal)} {
		body, ok := attackLimitPayload(err, defender)
		if !ok {
			t.Fatal("not recognised as an attack-limit refusal")
		}
		if body.Code != protocol.CodeIllegalAttack || body.Reason != protocol.AttackRefusalLimit || body.CardID != attacker.String() {
			t.Errorf("payload = %+v", body)
		}
		if body.Message != "No more than two creatures can attack you each combat (Crawlspace)." {
			t.Errorf("the protected seat reads %q", body.Message)
		}
		code, msg := classifyActionError(err)
		if code != protocol.CodeIllegalAttack || msg != "No more than two creatures can attack P2 each combat (Crawlspace)." {
			t.Errorf("classifyActionError = (%q, %q)", code, msg)
		}
	}
	if body, ok := attackLimitPayload(game.ErrAttackLimit, defender); !ok || body.Code != protocol.CodeIllegalAttack || body.Message == "" {
		t.Errorf("bare ErrAttackLimit: payload = %+v, ok = %v", body, ok)
	}
	if _, ok := attackLimitPayload(game.ErrWrongStep, defender); ok {
		t.Error("ErrWrongStep was sent as illegal_attack")
	}
	// A combat-wide limit has no "you".
	all := &game.AttackLimitError{Scope: game.AttackLimitEachCombat, Max: 1, SourceName: "Silent Arbiter"}
	if got := all.Sentence(defender); got != "No more than one creature can attack each combat (Silent Arbiter)." {
		t.Errorf("combat-wide sentence %q", got)
	}
}
