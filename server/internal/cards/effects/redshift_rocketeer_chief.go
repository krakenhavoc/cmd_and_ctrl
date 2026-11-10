package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Redshift, Rocketeer Chief — Legendary Creature — Goblin Pilot {R}{G},
// 2/3 (Aetherdrift, #4464):
//
//	"Vigilance
//	 {T}: Add X mana of any one color, where X is Redshift's power. Spend
//	 this mana only to activate abilities.
//	 Exhaust — {10}{R}{G}: Put any number of permanent cards from your
//	 hand onto the battlefield. (Activate each exhaust ability only once.)"
//
// The mana is Avatar Kyoshi's count-driven one-colour mana ability, read
// off Redshift's power as the ability resolves (layers brought current),
// with Pit Automaton's activate-only restriction: it pays for any
// activated ability, a mana ability's cost included, and refuses a spell.
// A Redshift with power 0 adds nothing. The exhaust is Ugin's put from the
// hand with no limit and no "may" beyond choosing none: the cards enter
// together as one batch (CR 614), so every entry sees the board as it
// stood before any arrived.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "582c7e54-d64c-4ffb-ae5f-d2752b3ed10b",
		Name:            "Redshift, Rocketeer Chief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(func(g *game.Game, _, source uuid.UUID) int {
				return b31CurrentPowerOf(g, source)
			}),
			Label:        "{T}: Add X mana of any one color, where X is Redshift's power. Spend this mana only to activate abilities.",
			Restrictions: []string{game.ManaRestrictActivate},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {10}{R}{G}: Put any number of permanent cards from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Exhaust: true,
			Cost:    ManaCost("{10}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Optional: true,
					Max:      1000,
					Label:    "Redshift, Rocketeer Chief — put any number of permanent cards from your hand onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
