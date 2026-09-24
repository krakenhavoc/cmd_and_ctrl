package ws

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestAttackRequirementRefusalsReachTheWireAsIllegalAttack — #1571. A
// declaration or pass refused by a CR 508.1d requirement arrives as
// `illegal_attack` / `attack_requirement`, with the creature in card_id
// and a sentence addressed to its reader, on both the hub's structured
// path and the plain classifier.
func TestAttackRequirementRefusalsReachTheWireAsIllegalAttack(t *testing.T) {
	attacker, goader := uuid.New(), uuid.New()
	refusal := &game.AttackRequirementError{
		Attacker: attacker, AttackerName: "Grizzly Bears",
		Requirement: game.AttackRequirement{GoadedBy: goader, OtherThan: goader},
		GoaderName:  "P2", OtherThanName: "P2",
	}
	for _, err := range []error{refusal, fmt.Errorf("dispatch: %w", refusal)} {
		body, ok := attackRequirementPayload(err, goader)
		if !ok {
			t.Fatal("not recognised as a requirement refusal")
		}
		if body.Code != protocol.CodeIllegalAttack || body.Reason != protocol.AttackRefusalRequirement || body.CardID != attacker.String() {
			t.Errorf("payload = %+v", body)
		}
		if body.Message != "Grizzly Bears is goaded by you and must attack a player other than you if able." {
			t.Errorf("the goader reads %q", body.Message)
		}
		code, msg := classifyActionError(err)
		if code != protocol.CodeIllegalAttack || msg != "Grizzly Bears is goaded by P2 and must attack a player other than P2 if able." {
			t.Errorf("classifyActionError = (%q, %q)", code, msg)
		}
	}
	if _, ok := attackRequirementPayload(game.ErrAttackLimit, goader); ok {
		t.Error("an attack-limit refusal was sent as attack_requirement")
	}
}
