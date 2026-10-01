package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c9b82110-7dfd-4617-9399-9510be449043",
		Name:         "Dark Depths",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(ice, 10, "Dark Depths: enters with ten ice counters"),
		},
		Activated: []ActivatedAbility{{
			Label:  "{3}: Remove an ice counter from Dark Depths.",
			Cost:   ManaCost("{3}"),
			Effect: removeACounterFromThis(ice),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisHasNo(ice, "Dark Depths — sacrifice it and create Marit Lage",
				SacrificeThisThen(func(ctx *Context) error {
					return CreateToken{Template: TokenCard("20/20 black Marit Lage"), N: 1}.Apply(ctx)
				})),
		},
	})
}
