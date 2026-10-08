package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martyr of Bones — Creature — Human Wizard {B}, 1/1:
//
//	"{1}, Reveal X black cards from your hand, Sacrifice this creature:
//	 Exile up to X target cards from a single graveyard."
//
// The reveal is a cost component (#2598, ADR 0020's 2026-10-08
// amendment): effects.RevealX, whose count IS the X announced with the
// activation (CR 602.2b). The activator names that many black cards in
// hand (`reveal_ids`); they are shown to the table and stay in the hand
// (CR 701.20b). There is no {X} in the mana cost, so it is {1} whatever
// X is, and X = 0 is a legal announcement that exiles nothing.
//
// The target clause is "up to X target cards from a single graveyard":
// the single-graveyard rule (ADR 0106 §5) with a count that follows the
// same announced X (TargetSpec.CountFromX + UpToX). Targets that left
// before resolution are skipped and the rest still go (CR 608.2b).
//
// No simplification.
func init() {
	targets := UpToCardsFromASingleGraveyard("up to X target cards from a single graveyard", 0)
	targets.CountFromX = true
	targets.UpToX = true
	Register(Spec{
		OracleID:     "a099d8fa-d51e-4dbc-a03f-6f912097d41e",
		Name:         "Martyr of Bones",
		Completeness: CompletenessFull,
		XMatters:     true,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Reveal X black cards from your hand, Sacrifice this creature: Exile up to X target cards from a single graveyard.",
			Cost:    Plus(ManaCost("{1}"), RevealX("X black cards", "B"), SacrificeThis()),
			Targets: targets,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if ctx.X() <= 0 {
					return nil
				}
				return ExileTargetCards(g, item)
			},
		}},
	})
}
