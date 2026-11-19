package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sage of the Unknowable — Creature — Human Wizard {1}{U}, 0/4 (Modern
// Horizons 3):
//
//	"{T}: Add {C}. Spend this mana only to cast a colorless spell or to
//	 activate an ability."
//
// Ambulatory Reactor's one-tag restriction (CR 106.6) as an OR of two
// clauses: a cast of a colorless spell (Shrine of the Forsaken Gods'
// ManaRestrictCast plus ManaRestrictColorless) or any activation (an
// activated ability of any permanent, which includes equip, crew and
// cycling). The same tag is enforced on spend and stamped on production,
// so it cannot pay for a coloured spell.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "359a570e-92a7-4954-9a05-89cc75df57e7",
		Name:         "Sage of the Unknowable",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C} (spend only to cast a colorless spell or to activate an ability)",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictColorless},
				[]string{game.ManaRestrictActivate},
			)},
		}},
	})
}
