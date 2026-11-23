package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvester of Misery — Creature — Spirit {3}{B}{B}, 5/4 (The Big Score,
// #4468):
//
//	"Menace
//	 When this creature enters, other creatures get -2/-2 until end of
//	 turn.
//	 {1}{B}, Discard this card: Target creature gets -2/-2 until end of
//	 turn."
//
// The enters trigger is a sweep over every OTHER creature (the source is
// compared by object, so a re-entered Harvester is not exempt); the hand
// ability is the carnosaur's shape (Zones hand, DiscardThis cost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0d3bb63-2977-41b2-bcc3-2488e6696cb8",
		Name:            "Harvester of Misery",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Harvester of Misery — other creatures get -2/-2",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.TrekB11OtherThanSourceObject(item)),
						Power:     -2,
						Toughness: -2,
						Label:     "Harvester of Misery — -2/-2",
					}.Apply(effects.NewContext(g, item))
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, Partial: true}}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Discard this card: Target creature gets -2/-2 until end of turn",
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -2, -2)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Harvester of Misery — -2/-2", -2, -2),
		}},
	})
}
