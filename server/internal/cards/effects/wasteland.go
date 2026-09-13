package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wasteland — Land (EDHREC rank 1204):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice this land: Destroy target nonbasic land."
//
// Strip Mine with the basics off the menu — the whole difference
// between the two cards is the word "nonbasic", and it is the whole
// reason one is banned in Legacy and the other is not. Same
// tap-and-sacrifice cost paid at announce, so the Wasteland is gone
// before the ability resolves. "Nonbasic" reads the Basic supertype
// off the printed type line: a land something else has turned into a
// Swamp is still a nonbasic Swamp, and a Wastes is still basic.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "09a70ae8-3859-4a09-901d-dce063fa3b5f",
		Name:         "Wasteland",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Destroy target nonbasic land.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target nonbasic land", Land(), b10NonbasicLand()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
