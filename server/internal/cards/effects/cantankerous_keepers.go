package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cantankerous Keepers — Creature — Elf Soldier {5}{G}, 4/3 (The Hobbit,
// HOB):
//
//	"Affinity for Elves (This spell costs {1} less to cast for each Elf you
//	 control.)
//	 When this creature enters, mill four cards, then put all Elf cards from
//	 among them into your hand."
//
// Affinity is Emry's self cost modifier over Elves you control (CR 702.41a).
// The mill is Tetzin's: the clause after it runs on the cards that actually
// landed in the graveyard (#4253), and, unlike Tetzin's "may", takes every
// Elf among them with no prompt. A milled card that has left the graveyard
// by then is not taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77c97dad-17cc-4a28-a48b-907c044d9034",
		Name:         "Cantankerous Keepers",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Elves", HasSubtype("Elf")),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Cantankerous Keepers — mill four cards, then put all Elf cards from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 4, ThenTail: ctx.ItemTail(cantankerousKeepersMilled)}.Apply(ctx)
				}),
		},
	})
}

// cantankerousKeepersMilled is the clause after the mill: every milled Elf
// card still in the graveyard goes to its owner's hand.
var cantankerousKeepersMilled = CardTail("cantankerous-keepers/after-mill", func(ctx *Context, o game.TailOutcome) error {
	return hobA11AllMilledToHand(ctx, o, HasSubtype(ctx.TextSubtype("Elf")))
})
