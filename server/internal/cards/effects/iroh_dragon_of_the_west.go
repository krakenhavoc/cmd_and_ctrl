package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iroh, Dragon of the West — Legendary Creature — Human Noble Ally
// {2}{R}{R}, 4/4 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Haste
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)
//	 At the beginning of combat on your turn, each creature you control
//	 with a counter on it gains firebending 2 until end of turn.
//	 (Whenever it attacks, add {R}{R}. This mana lasts until end of
//	 combat.)"
//
// Mentor is the keyword's trigger (CR 702.134a). Firebending 2 is granted
// as a catalog bundle until end of turn (Sozin's Comet's shape), so it is
// each creature's own triggered ability on the stack, and a creature with
// firebending already triggers once for each instance (the rulings). The
// set is read once as the trigger resolves (CR 611.2c): a counter of any
// kind qualifies, and a creature that gets one later does not gain it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98f9b97b-4c15-4d24-895a-a9bebb68e79b",
		Name:            "Iroh, Dragon of the West",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Grants: []AbilityGrant{{
			Key: irohDragonOfTheWestFirebendingGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisAttacks("Iroh, Dragon of the West — firebending 2", Do(firebending(2))),
			},
			Text: "Firebending 2 (Whenever this creature attacks, add {R}{R}. This mana lasts until end of combat.)",
		}},
		Triggered: []game.TriggeredAbility{
			Mentor("Iroh, Dragon of the West — mentor"),
			AtBeginningOfYourCombat("Iroh, Dragon of the West — each creature you control with a counter on it gains firebending 2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantAbilitiesFor{
						Match: And(Creature(), YouControl(), b64HasAnyCounter()),
						Keys:  []string{irohDragonOfTheWestFirebendingGrant},
						Label: "Iroh, Dragon of the West — firebending 2 until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// irohDragonOfTheWestFirebendingGrant is the grant's catalog key. An
// on-disk key: never renamed, never reused.
const irohDragonOfTheWestFirebendingGrant = "iroh-dragon-of-the-west/firebending-2"
