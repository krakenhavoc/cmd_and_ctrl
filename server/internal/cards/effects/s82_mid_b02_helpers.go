package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_mid_b02_helpers.go — shared bodies for Innistrad: Midnight Hunt's
// plain instants and removal (S82, #5497).

// s82MidThenScry is "then scry N", Tail.N of them, for Tail.Player. An
// on-disk key.
var s82MidThenScry = CardTail("then/s82-mid-scry-n", func(ctx *Context, o game.TailOutcome) error {
	return Scry{Player: o.Tail.Player, N: o.Tail.N}.Apply(ctx)
})

// s82MidSpiritBody is Blessed Defiance's delayed body: a 1/1 white Spirit with flying
// for the controller of the delayed trigger. An on-disk key.
var s82MidSpiritBody = game.SimpleDelayedBody("s82-mid/spirit-1-1-flying", func(g *game.Game, item *game.StackItem) error {
	return CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 1}.Apply(NewContext(g, item))
})

// s82MidDestroyTheTarget is "Destroy target <thing>" as a spell's whole body
// (Defenestrate, Sungold Barrage): the clause on the spec carries the
// restriction, and a target gone by resolution does nothing (CR 608.2b).
func s82MidDestroyTheTarget(_ *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return DestroyTarget{Target: t.ID}.Apply(ctx)
}

// s82MidControlsWolfOrWerewolf is "if you control a Wolf or Werewolf", the
// words read through the spell's text (CR 612.2) and the subtypes through
// the layered view.
func s82MidControlsWolfOrWerewolf(ctx *Context) bool {
	wolf, werewolf := ctx.TextSubtype("Wolf"), ctx.TextSubtype("Werewolf")
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && (c.HasSubtype(wolf) || c.HasSubtype(werewolf)) {
			return true
		}
	}
	return false
}
