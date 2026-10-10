package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Karolina Dean, Runaway — Legendary Creature — Alien Hero {3}{R}, 3/3
// (EDHREC rank 19226):
//
//	"Flying
//	 At the beginning of your first main phase, add {W}{U}{B}{R}{G}.
//	 This mana can't be spent to cast spells from your hand."
//
// The mana carries ManaRestrictNotFromHand (#2811): it pays for
// activated abilities and for spells cast from a graveyard, exile, the
// command zone or a library top, and not for a spell cast from hand. It
// empties with the phase like any other mana (CR 500.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "50b5faff-acaa-426b-9f3e-1005c0392e6c",
		Name:            "Karolina Dean, Runaway",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Karolina Dean, Runaway — add {W}{U}{B}{R}{G}",
				func(g *game.Game, item *game.StackItem) error {
					return g.AddManaWithOptionsForEffect(item.Controller, item.SourceCardID, "{W}{U}{B}{R}{G}",
						game.AddManaOptions{Restrictions: []string{ManaRestrictNotFromHand}})
				}),
		},
	})
}
