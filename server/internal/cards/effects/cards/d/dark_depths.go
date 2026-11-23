package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Depths — Legendary Snow Land:
//
//	"Dark Depths enters with ten ice counters on it.
//	 {3}: Remove an ice counter from Dark Depths.
//	 When Dark Depths has no ice counters on it, sacrifice it. If you do,
//	 create Marit Lage, a legendary 20/20 black Avatar creature token
//	 with flying and indestructible."
//
// ADR 0107 §1 (#1858). The ten counters are an entry replacement, so
// they are on the land before anything can look at it. Removing one is
// the {3} row's effect (CR 609.3: with none left it does nothing). The
// sacrifice is a CR 603.8 state trigger, so it triggers however the last
// counter went — the {3} row, Vampire Hexmage's "remove all counters", or
// a copy of Dark Depths that entered with none — and Marit Lage is made
// only if the land was really sacrificed.
//
// No simplification.
func init() {
	const ice = "ice"
	effects.Register(effects.Spec{
		OracleID:     "c9b82110-7dfd-4617-9399-9510be449043",
		Name:         "Dark Depths",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(ice, 10, "Dark Depths: enters with ten ice counters"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Remove an ice counter from Dark Depths.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}"),
			Effect:  effects.RemoveACounterFromThis(ice),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisHasNo(ice, "Dark Depths — sacrifice it and create Marit Lage",
				effects.SacrificeThisThen("then/dark-depths-marit-lage", func(ctx *effects.Context) error {
					return effects.CreateToken{Template: effects.TokenCard("20/20 black Marit Lage"), N: 1}.Apply(ctx)
				})),
		},
	})
}
