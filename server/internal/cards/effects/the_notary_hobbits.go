package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Notary Hobbits — Legendary Creature — Halfling Advisor {3}{G}{G},
// 1/1 (The Hobbit):
//
//	"When The Notary Hobbits enter, if they're not a token, create two
//	 tokens that are copies of them, except the tokens aren't legendary.
//	 {T}: Add {C} for each Halfling you control."
//
// The trigger is Miirym's token copy (CreateTokenCopy, removeLegendaryFromTypeLine)
// made twice; the copy is read from wherever the card sits when the trigger
// resolves, so it still works if they left in response. "If they're not a
// token" is an intervening if (CR 603.4) on the permanent's own, unchanging
// token status, so it is judged as the ability triggers, and a copy token never
// triggers a copy of itself. Each token, being a Halfling, adds to the count of
// the mana ability, which is Howlsquad Heavy's ProducedPerPermanent over the
// Halflings you control (read through the source's text, CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc20b29b-856e-4d45-99cd-5e12cfc2324d",
		Name:         "The Notary Hobbits",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && !source.IsToken()
			}, "The Notary Hobbits — create two tokens that are copies of them, except the tokens aren't legendary",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{
						Controller: item.Controller,
						Copy:       item.SourceCardID,
						N:          2,
						Except: func(t *game.Card) {
							t.TypeLine = removeLegendaryFromTypeLine(t.TypeLine)
						},
					}.Apply(NewContext(g, item))
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedPerPermanent("C", MatchCreatureSubtype("Halfling")),
			Label:          "Add {C} for each Halfling you control",
		}},
	})
}
