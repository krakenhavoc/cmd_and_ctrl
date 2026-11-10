package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Velomachus Lorehold — Legendary Creature — Elder Dragon {5}{R}{W}, 5/5:
//
//	"Flying, vigilance, haste
//	 Whenever Velomachus Lorehold attacks, look at the top seven cards of
//	 your library. You may cast an instant or sorcery spell with mana value
//	 less than or equal to Velomachus Lorehold's power from among them
//	 without paying its mana cost. Put the rest on the bottom of your
//	 library in a random order."
//
// Aetherworks Marvel's look-and-cast (lookAtTopAndOfferFreeCast) with the
// instant-or-sorcery and mana-value clause carried by the free permission and
// judged on the face being cast. The power is Velomachus's as the trigger
// resolves (its last-known power if it has gone, CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "43832745-5d7d-4957-8e30-2072ad012564",
		Name:            "Velomachus Lorehold",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Velomachus Lorehold — look at the top seven cards, you may cast an instant or sorcery with mana value at most its power free, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					power := 0
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						power = c.CurrentPower()
					}
					return lookAtTopAndOfferFreeCast(g, item.Controller, item.SourceCardID, velomachusLoreholdCast, 7,
						freeInstantOrSorcery().withManaValueAtMost(power))
				}),
		},
	})
}

var velomachusLoreholdCast = registerFreeLibraryCastOffer("choose-cards/velomachus-lorehold-cast", "Velomachus Lorehold")
