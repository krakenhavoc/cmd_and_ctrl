package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trenchpost — Land — Locus (Modern Horizons 3 Commander, #5468):
//
//	"{T}: Add {C}.
//	 {3}, {T}: Target player mills a card for each Locus you control."
//
// The mill counts the Loci its controller controls as the ability
// resolves, Trenchpost itself included if it is still there. "Locus" is a
// word of the land's text (CR 612.2) and is read through it, and the
// count reads the layered subtypes, so a land made a Locus by an effect
// counts (Cloudpost's reading). Any player may be the target, the
// activator included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42f1ccb8-eda0-4828-ac07-82d4e950d7e1",
		Name:         "Trenchpost",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target player mills a card for each Locus you control.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := effects.TrekPermanentsWithSubtypeYouControl(g, item, item.Controller, "Locus")
				return effects.TargetPlayerMills(n)(g, item)
			},
		}},
	})
}
