package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mauhúr, Uruk-hai Captain — Legendary Creature — Orc Soldier {B}{R},
// 2/2:
//
//	"Menace
//	 If one or more +1/+1 counters would be put on an Army, Goblin, or
//	 Orc you control, that many plus one +1/+1 counters are put on it
//	 instead."
//
// Hardened Scales' replacement narrowed from "a creature you control"
// to a tribe: the placement is read through the same
// plusOneCounterPlacementOnYourCreature test, then the receiver must be
// an Army, a Goblin or an Orc. An amass on a fresh Army reads the
// token, which is an Army from the moment it exists, so the first
// counters are bumped too. Mauhúr is itself an Orc and bumps counters
// put on himself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9966cac0-331f-4627-be5a-5060a6ac5a32",
		Name:            "Mauhúr, Uruk-hai Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if !plusOneCounterPlacementOnYourCreature(ev, g, src) {
					return false
				}
				target, ok := g.LookupCardForEffect(ev.CounterTarget)
				return ok && (target.HasSubtype(game.ArmySubtype) || target.HasSubtype("Goblin") || target.HasSubtype("Orc"))
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Mauhúr: one more +1/+1 counter on an Army, Goblin or Orc",
		}},
	})
}
