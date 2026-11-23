package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twice Upon a Time // Unlikely Meeting — Sorcery with an Adventure (Doctor
// Who):
//
//	Twice Upon a Time {4}{U}{U}: "Cast this spell only if you control two or
//	 more Doctors.
//	 Take an extra turn after this one. Exile Twice Upon a Time."
//	Unlikely Meeting {2}{U} — Sorcery — Adventure: "Search your library for a
//	 Doctor card, reveal it, put it into your hand, then shuffle."
//
// Face 0 is the card's oracle id, the Adventure is "#1" (Bonecrusher Giant), and
// the engine's CR 715.3d branch exiles the card "on an adventure" after Unlikely
// Meeting resolves. The cast condition counts the Doctors its caster controls
// (a permanent with the Doctor subtype, read through the layers). Taking the
// extra turn and exiling itself is Temporal Mastery's body
// (takeAnExtraTurnThenExileThisSpell, CR 500.7): the spell moves itself to
// exile instead of the graveyard.
//
// No simplification.
func init() {
	const oracle = "04dc2e9c-312c-4af8-950c-6a7ed859c9be"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Twice Upon a Time",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraTurns: 1},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			doctors := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == controller && c.HasSubtypeWord("Doctor") {
					doctors++
				}
			}
			return doctors >= 2
		},
		CastConditionLabel: "Cast this spell only if you control two or more Doctors.",
		OnResolve:          effects.TakeAnExtraTurnThenExileThisSpell,
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Unlikely Meeting",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Doctor") },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Unlikely Meeting — a Doctor card",
				Source:    item.SourceCardID,
			}.Apply(ctx)
		},
	})
}
