package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fae of Wishes // Granted — Creature — Faerie Wizard {1}{U}, 1/1,
// with Adventure Sorcery — Granted {3}{U} (CR 715):
//
//	Fae of Wishes — "Flying
//	 {1}{U}, Discard two cards: Return this creature to its owner's hand."
//	Granted — "You may reveal a noncreature card you own from outside
//	 the game and put it into your hand."
//
// Both faces register under one oracle ID, the creature keyed bare and
// the Adventure at "#1" (the Foulmire Knight shape, ADR 0034 §5). The
// Adventure half is TakeFromOutside over noncreature cards, revealed;
// the card then goes to exile and the creature is castable from there
// (the engine's CR 715.3d branch, not this file).
//
// The creature's return is a discard cost paid at announce (CR 602.2b),
// the same shape as Leashling's put-a-card cost.
func init() {
	const oracleID = "c0abbed2-d213-47ef-8d6c-4a21efb9a55f"

	// Face 0 — the creature.
	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Fae of Wishes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, Discard two cards: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.DiscardN(2, "two cards")),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})

	// Face 1 — Granted, an Adventure sorcery (CR 715.3a).
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Granted",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TakeFromOutside{
				Pick:     grantedTake,
				Question: "Granted — reveal a noncreature card you own from outside the game, or none",
			}.Apply(ctx)
		},
	})
}

// grantedTake is Granted's keyed pick (#4256), so a table waiting on it
// is a restore point.
var grantedTake = effects.RegisterOutsideTake("choose-cards/granted-take",
	func(c game.Card) bool { return !c.IsCreature() }, true)
