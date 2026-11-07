package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baron Bertram Graywater — Legendary Creature — Vampire Noble
// {2}{W}{B}, 3/4 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever one or more tokens you control enter, create a 1/1 black
//	 Vampire Rogue creature token with lifelink. This ability triggers
//	 only once each turn.
//	 {1}{B}, Sacrifice another creature or artifact: Draw a card."
//
// The token trigger is Caretaker's Talent's: a batch trigger (CR 603.2c)
// over tokens entering under your control, held to once each turn
// (TriggersOnlyOnceEachTurn). The Vampire Rogue it makes is a token
// entering too, but the turn's one trigger is already spent, so it does
// not loop. The draw is Ahriman's sacrifice outlet at a different price.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e25b516c-bf95-42bc-8b1e-04617a3d28df",
		Name:         "Baron Bertram Graywater",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventETB,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsToken()
				},
				"Baron Bertram Graywater — create a 1/1 black Vampire Rogue with lifelink (tokens entered)",
				Do(CreateToken{Template: TokenCard("1/1 black Vampire Rogue with lifelink"), N: 1})))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature or artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{1}{B}"), game.AbilityCost{
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			}),
			Effect: Do(DrawCards{N: 1}),
		}},
	})
}
