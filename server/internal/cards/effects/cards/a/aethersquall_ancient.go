package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aethersquall Ancient — Creature — Leviathan {5}{U}{U}, 6/6:
//
//	"Flying
//	 At the beginning of your upkeep, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay eight {E}: Return all other creatures to their owners' hands.
//	 Activate only as a sorcery."
//
// ADR 0129 PR 1 (#1995). "All other creatures" is every creature on the
// battlefield but the Ancient, whoever controls them, read as the
// ability resolves. The row declares a bounce sweep, partial because
// it spares the Ancient.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47c685b1-6e1e-4e08-8a52-9da6902df113",
		Name:            "Aethersquall Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Aethersquall Ancient — you get {E}{E}{E}", effects.EbYouGetEnergy(3)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "Pay eight {E}: Return all other creatures to their owners' hands. Activate only as a sorcery.",
			Cost:         effects.PayEnergy(8),
			SorcerySpeed: true,
			Purpose: game.Purpose{Sweep: game.Sweep{
				Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BounceAllMatching{Match: effects.And(effects.Creature(), effects.OtherThan(item.SourceCardID))}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
