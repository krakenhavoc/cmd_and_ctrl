package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilt-Leaf Ambush — Kindred Instant — Elf {2}{G} (Lorwyn, #5593):
//
//	"Create two 1/1 green Elf Warrior creature tokens. Clash with an
//	 opponent. If you win, those creatures gain deathtouch until end of
//	 turn. (Each clashing player reveals the top card of their library, then
//	 puts that card on their choice of the top or bottom. A player wins if
//	 their card had a greater mana value. Any amount of damage a creature
//	 with deathtouch deals to a creature is enough to destroy it.)"
//
// The tokens are made first and their ids carried through the creation's own
// tail into the clash, whose tail gives deathtouch to those two creatures
// (the ones still on the battlefield) on a win, until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5a06062-0513-48af-b92a-420ebbeb1695",
		Name:         "Gilt-Leaf Ambush",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups: []game.TokenGroup{{
					Template: ctx.TextToken(effects.TokenCard("1/1 green Elf Warrior")),
					Count:    2,
				}},
			}, ctx.ItemTail(giltLeafAmbushMade))
		},
	})
}

// giltLeafAmbushMade is "Clash with an opponent" once the tokens exist: they
// ride the clash's tail. An on-disk key.
var giltLeafAmbushMade = effects.CardTail("then/gilt-leaf-ambush-made", func(ctx *effects.Context, o game.TailOutcome) error {
	then := ctx.ItemTail(giltLeafAmbushClashed)
	then.Carry = append([]uuid.UUID(nil), o.IDs...)
	return effects.ClashWithAnOpponent{Then: then}.Apply(ctx)
})

// giltLeafAmbushClashed is "If you win, those creatures gain deathtouch
// until end of turn". An on-disk key.
var giltLeafAmbushClashed = effects.CardTail("then/gilt-leaf-ambush-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) {
		return nil
	}
	for _, id := range o.Tail.Carry {
		if !effects.OnBattlefield(ctx.Game, id) {
			continue
		}
		if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Gilt-Leaf Ambush — deathtouch"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
