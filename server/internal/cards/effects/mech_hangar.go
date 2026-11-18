package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	restricted := cmdr2B01TapForAnyColor()
	restricted.Label = "{T}: Add one mana of any color. Spend this mana only to cast a Pilot or Vehicle spell."
	restricted.Restrictions = []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Pilot|Vehicle")}
	Register(Spec{
		OracleID:     "abc04775-171d-41f3-83ea-4b4eb72723d5",
		Name:         "Mech Hangar",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}, restricted},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}: Target Vehicle becomes an artifact creature until end of turn",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    Plus(ManaCost("{3}"), TapCost()),
			Targets: TargetPermanent("target Vehicle", Subtype("Vehicle")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return BecomeCreatureUntilEOT{Target: t.ID, Label: "Mech Hangar — becomes an artifact creature"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
