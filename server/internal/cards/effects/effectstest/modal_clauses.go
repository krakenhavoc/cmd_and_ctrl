package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modal_clauses_test.go — #764 / ADR 0065: the proof set for per-mode
// targets, repeated modes, "choose one or more", and modal triggers.
//
// Every assertion here fails with the machinery backed out: before
// #764 Kolaghan's Command could not be REGISTERED (effects.Register
// panicked at boot on a Max > 1 spec with two targeted options),
// Mystic Confluence could not be ANNOUNCED (validateModes rejected a
// repeated index), and a trigger had no mode slot at all.

const (
	KolaghansCommandOracle = "45f1e957-09f0-4d46-8e32-238f26060a87"
	MysticConfluenceOracle = "cd11c27b-9368-4622-bc5a-2e0b993cc42b"
	SublimeEpiphanyOracle  = "56148ae7-a9df-4771-8d53-d9ffb815c884"
	GlissaSunslayerOracle  = "8859a692-d107-4de2-9110-4681fc2761c7"
)

// modeRef is a target ref that names its mode OCCURRENCE and clause
// slot — what a #764 client sends and what the engine validates
// against the clause of that occurrence.
func ModeRef(kind game.TargetRefKind, id uuid.UUID, mode, slot int) game.TargetRef {
	return game.TargetRef{Kind: kind, ID: id, Mode: mode, Slot: slot}
}

// b64ReadyToAttack clears the summoning-sickness flag so a freshly
// seeded permanent can attack this turn.
func B64ReadyToAttack(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].SummonedThisTurn = false
			return
		}
	}
}
