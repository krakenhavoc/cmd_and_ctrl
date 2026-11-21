package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activation_restriction_cards_test.go — #1210's card proofs.
//
// The engine test (game/activation_gate_test.go) pins the gate
// against stub restrictions; these drive the four printed cards
// through the real catalog, because the half of this issue that can
// go wrong per card is which clause the card file reached for: a
// Cursed Totem written with the mana exemption is a blank against the
// deck it exists to beat, and a Linvala written against the activator
// instead of its own controller breaks the moment somebody steals it.

const (
	actKrenkoOracle         = "68418069-f615-40ef-ae0d-764192acae00"
	ActSolRingOracle        = "6ad8011d-3471-4369-9d68-b264cc027487"
	ActBirdsOracle          = "d3a0b660-358c-41bd-9cd2-41fbf3491b1a"
	ActLlanowarElvesOracle  = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	ActCursedTotemOracle    = "6225a704-430a-4f56-ad87-0e8d87f285f5"
	ActLinvalaOracle        = "88dadc31-dfac-41b3-bf2a-65fa89e3c16d"
	ActCollectorOupheOracle = "0c4bc9ea-a5fd-4f44-96a1-5448eee228c4"
	ActPithingNeedleOracle  = "a188fe7e-68de-4c7c-806c-bfe8fc7b44bf"
	ActHarshMentorOracle    = "168336e2-d795-4b75-bf21-ce128b0dd7b0"
)

// actPushKrenko seeds a Krenko, Mob Boss — a creature with a NON-mana
// {T} ability, which is the thing an activation restriction is really
// about.
func ActPushKrenko(g *game.Game, owner uuid.UUID) uuid.UUID {
	return B21Push(g, owner, "Krenko, Mob Boss", "Legendary Creature — Goblin Warrior", actKrenkoOracle, 3, 3, "R")
}

// actNameFor answers the open choose_card_name prompt addressed to
// `chooser` with `name`.
func ActNameFor(t *testing.T, g *game.Game, chooser uuid.UUID, name string) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoiceCardName || c.Chooser != chooser {
			continue
		}
		if err := g.ResolveCardNameChoice(c.ID, chooser, name); err != nil {
			t.Fatalf("ResolveCardNameChoice(%q): %v", name, err)
		}
		return
	}
	t.Fatalf("no open choose_card_name prompt for %s", chooser)
}
