package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pako, Arcane Retriever — Legendary Creature — Elemental Dog {3}{R}{G}, 3/3
// (Commander 2020, S82 OLD, #5513):
//
//	"Partner with Haldan, Avid Arcanist
//	 Haste
//	 Whenever Pako attacks, exile the top card of each player's library and
//	 put a fetch counter on each of them. Put a +1/+1 counter on Pako for
//	 each noncreature card exiled this way."
//
// Evelyn, the Covetous's exile-and-mark chain: the top cards go one library at
// a time in APNAP order (an empty library gives none), and once every library
// is done each card that reached exile gets its fetch counter, an ordinary
// counter on the exiled card (CR 122.1) that MoveCard clears when the card
// leaves exile. The +1/+1 counters follow, one for each exiled card that is
// not a creature card (a land counts); a Pako that has left gets none.
// "Partner with" is the CR 702.124j entry search; Haste rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "680a3e17-5144-4e97-bc59-9f5fc3ad6514",
		Name:            "Pako, Arcane Retriever",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Pako, Arcane Retriever", "Haldan, Avid Arcanist"),
			effects.On(game.EventAttack, effects.Self,
				"Pako, Arcane Retriever — exile the top card of each player's library with a fetch counter on it; put a +1/+1 counter on Pako for each noncreature card exiled this way",
				pakoExileTopOfEachLibrary),
		},
	})
}

func pakoExileTopOfEachLibrary(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.EachPlayerInTurn(g, ctx.ItemTail(pakoExileStep), effects.ApnapPlayers(g), nil,
		ctx.ItemTail(pakoCollect), pakoExileAct)
}

// pakoExileStep is one library of the chain. An on-disk key.
var pakoExileStep = effects.RegisterEachPlayerStep("pako-arcane-retriever/exile-step", pakoExileAct)

func pakoExileAct(ctx *effects.Context, player uuid.UUID, _ int, step game.KeyedTail) error {
	return ctx.Game.ExileTopKeyedForEffect(player, ctx.Controller(), 1, step)
}

// pakoCollect marks each exiled card still in exile with a fetch counter,
// then counts the noncreature ones for Pako's +1/+1 counters. An on-disk key.
var pakoCollect = effects.CardTail("pako-arcane-retriever/collect", func(ctx *effects.Context, o game.TailOutcome) error {
	noncreature := 0
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if err := ctx.Game.AddCounterForEffect(id, effects.FetchCounter, 1); err != nil {
			return err
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsCreature() {
			noncreature++
		}
	}
	if noncreature == 0 || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: noncreature}.Apply(ctx)
})
