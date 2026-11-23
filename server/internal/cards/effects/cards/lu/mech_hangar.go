package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mech Hangar — Land (Neon Dynasty):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Pilot
//	 or Vehicle spell.
//	 {3}, {T}: Target Vehicle becomes an artifact creature until end of
//	 turn."
//
// Access Tunnel's land shape with Intrepid Stablemaster's spend restriction
// (CR 106.6) on one coloured mana, and Alacrian Armory's animation of a
// Vehicle (BecomeCreatureUntilEOT, CR 301.7b).
//
// No simplification.
func init() {
	restricted := effects.Cmdr2B01TapForAnyColor()
	restricted.Label = "{T}: Add one mana of any color. Spend this mana only to cast a Pilot or Vehicle spell."
	restricted.Restrictions = []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Pilot|Vehicle")}
	effects.Register(effects.Spec{
		OracleID:     "abc04775-171d-41f3-83ea-4b4eb72723d5",
		Name:         "Mech Hangar",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}, restricted},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target Vehicle becomes an artifact creature until end of turn",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Vehicle", effects.Subtype("Vehicle")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.BecomeCreatureUntilEOT{Target: t.ID, Label: "Mech Hangar — becomes an artifact creature"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
