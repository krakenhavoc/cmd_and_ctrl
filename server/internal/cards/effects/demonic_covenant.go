package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demonic Covenant — Kindred Enchantment — Demon {4}{B}{B} (Duskmourn
// Commander, #5368):
//
//	"Whenever one or more Demons you control attack a player, you draw a
//	 card and lose 1 life.
//	 At the beginning of your end step, create a 5/5 black Demon creature
//	 token with flying, then mill two cards. If two cards that share all
//	 their card types were milled this way, sacrifice this enchantment."
//
// The attack trigger is Attuma's "attack a player" guard (CR 603.2c, 508.3e):
// one draw and one life per player attacked, however many Demons go at that
// player, and none for an attack at a planeswalker or battle. "Demon" is a
// word of the source's text, read on the attacker as it is now.
//
// The end-step ability is a chain of keyed tails (#4253): the token, then
// the mill of two (the "moved" form, so a card a replacement exiled instead
// still counts as milled, CR 701.17c), then the check over the cards that
// were actually milled. It needs exactly two milled cards whose card types
// are identical sets (a Kindred Enchantment pairs only with another
// Kindred Enchantment, a Land with a Land), and sacrifices the enchantment
// when they are. Library of one card: only one is milled, so no pair, no
// sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "393b9d9e-9d71-43e7-aede-eac7f85d2365",
		Name:         "Demonic Covenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(On(game.EventAttack, cmdr2S50DemonAttackAPlayer,
				"Demonic Covenant — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, Instruction(func(ctx *Context) error {
					return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
				})))),
			AtYourEndStep("Demonic Covenant — create a 5/5 black Demon with flying, then mill two cards; sacrifice this if they share all card types",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.CreateTokensKeyedForEffect(game.TokenCreation{
						Controller: ctx.Controller(),
						Source:     item.SourceCardID,
						Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("5/5 black Demon with flying")), Count: 1}},
					}, ctx.ItemTail(demonicCovenantMill))
				}),
		},
	})
}

// cmdr2S50DemonAttackAPlayer is the When of "whenever one or more Demons you
// control attack a player": a Demon creature the source's controller controls
// was declared attacking a player.
func cmdr2S50DemonAttackAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b16YouAttackedAPlayer(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Demon"))
}

// demonicCovenantMill is "then mill two cards", after the token. An on-disk
// key: never renamed, never reused.
var demonicCovenantMill = CardTail("then/demonic-covenant-mill", func(ctx *Context, _ game.TailOutcome) error {
	return ctx.Game.MillToZoneKeyedMovedForEffect(ctx.Controller(), 2, game.ZoneGraveyard, ctx.ItemTail(demonicCovenantMilled))
})

// demonicCovenantMilled is "if two cards that share all their card types were
// milled this way, sacrifice this enchantment", read over the cards milled. An
// on-disk key: never renamed, never reused.
var demonicCovenantMilled = CardTail("then/demonic-covenant-milled", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) != 2 {
		return nil
	}
	a, okA := ctx.Game.LookupCardForEffect(o.IDs[0])
	b, okB := ctx.Game.LookupCardForEffect(o.IDs[1])
	if !okA || !okB || !cmdr2S50ShareAllCardTypes(a, b) {
		return nil
	}
	return SacrificePermanent{Target: ctx.Item.SourceCardID}.Apply(ctx)
})
