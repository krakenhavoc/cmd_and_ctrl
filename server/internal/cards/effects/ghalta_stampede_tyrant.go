package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghalta, Stampede Tyrant — Legendary Creature — Elder Dinosaur
// {5}{G}{G}{G}, 12/12 (The Lost Caverns of Ixalan, #4470):
//
//	"Trample
//	 When Ghalta enters, put any number of creature cards from your hand
//	 onto the battlefield."
//
// Smuggler's Surprise's second bullet without its ceiling: one prompt
// over the creature cards in your hand as the trigger resolves, with a
// floor of zero ("any number" includes none, Earth's Mightiest Heroes'
// reading) and a ceiling of every creature card you hold. The picks enter
// as ONE simultaneous batch (PutFromHandOntoBattlefield), so each sees the
// others arrive (CR 603.6a). They are put, not cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b8f16311-be5a-4648-b024-849a61dff0bc",
		Name:            "Ghalta, Stampede Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ghalta, Stampede Tyrant — put any number of creature cards from your hand onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					return PutFromHandOntoBattlefield{
						Player:   item.Controller,
						Match:    Creature(),
						Optional: true,
						Max:      len(handCardsMatching(g, item.Controller, Creature())),
						Label:    "Ghalta, Stampede Tyrant — put any number of creature cards from your hand onto the battlefield",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
