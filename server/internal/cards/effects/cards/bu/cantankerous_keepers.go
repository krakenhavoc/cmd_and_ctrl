package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "77c97dad-17cc-4a28-a48b-907c044d9034",
		Name:         "Cantankerous Keepers",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Elves", effects.HasSubtype("Elf")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Cantankerous Keepers — mill four cards, then put all Elf cards from among them into your hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(cantankerousKeepersMilled)}.Apply(ctx)
				}),
		},
	})
}

// cantankerousKeepersMilled is the clause after the mill: every milled Elf
// card still in the graveyard goes to its owner's hand.
var cantankerousKeepersMilled = effects.CardTail("cantankerous-keepers/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.HobA11AllMilledToHand(ctx, o, effects.HasSubtype(ctx.TextSubtype("Elf")))
})
