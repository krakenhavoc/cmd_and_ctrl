package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malakir Bloodwitch — Creature — Vampire Shaman {3}{B}{B}, 4/4:
//
//	"Flying, protection from white
//	 When this creature enters, each opponent loses life equal to the number
//	 of Vampires you control. You gain life equal to the life lost this way."
//
// The Vampires are counted as the trigger resolves, the Bloodwitch itself
// among them if it is still there, each read through the layered view. The
// gain is what the opponents really lost, told once every loss has settled
// (Gray Merchant's shared drain).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "870e199e-1d54-44ec-8e24-f24af93ae9c5",
		Name:            "Malakir Bloodwitch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("W")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Malakir Bloodwitch — each opponent loses life equal to the number of Vampires you control and you gain the life lost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					word := "Vampire"
					if src, ok := g.LookupCardForEffect(ctx.Source()); ok {
						word = src.TextSubtype("Vampire") // a word of the source's text (CR 612.2)
					}
					return b21DrainEachOpponentAndGainTheTotal(g, item, vowB05VampiresControlledBy(g, ctx.Controller(), word))
				}),
		},
	})
}
