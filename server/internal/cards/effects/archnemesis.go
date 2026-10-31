package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archnemesis — {1}{U}{B} Enchantment — Aura:
//
//	"Enchant opponent
//	 Whenever you attack enchanted player, that player loses 2 life. You
//	 draw a card and gain 2 life.
//	 Whenever a player attacks you, you may attach this Aura to that
//	 player."
//
// Curse of Opulence's player attachment (Card.AttachedTo is a player ref)
// with "Enchant opponent" as the cast target. Both triggers read an attack
// declaration, once per defending player (OncePerBatchPerPlayer: three
// creatures at one player are one attack, CR 603.2c, the Horizon Explorer
// ruling).
//
//   - Drain: you attack the enchanted player. The player who loses life is
//     the one attacked, which is the enchanted player at the moment it
//     triggered; the draw and the life gain are the Aura controller's.
//   - Move: a player attacks YOU. The "you may" is the CR 603.5 prompt as
//     the trigger resolves, and the Aura attaches to the attacking player
//     only if it is still on the battlefield and that player is still an
//     opponent in the game (CR 303.4f: the new host must be legal; a
//     player who has left is not). An attack that also names the
//     enchanted player does not move it: it already is there.
//
// No simplification. Marvel Super Heroes (#4455).
func init() {
	const (
		drainLabel = "Archnemesis — the enchanted player loses 2 life; you draw a card and gain 2 life"
		moveLabel  = "Archnemesis — you may attach this Aura to the attacking player"
	)
	drain := OncePerBatchPerPlayer(On(game.EventAttack, msc18YouAttackTheEnchantedPlayer, drainLabel,
		Do(Instruction(msc18TheAttackedPlayerLoses2), DrawCards{N: 1}, GainLife{Amount: 2})))
	move := OncePerBatchPerPlayer(Optional(On(game.EventAttack, msc18APlayerAttacksYou, moveLabel,
		Do(Instruction(msc18AttachToTheAttacker))),
		"Archnemesis — attach this Aura to the attacking player?"))
	Register(Spec{
		OracleID:     "c408717c-e07b-484f-9c81-5666595adf81",
		Name:         "Archnemesis",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("enchant opponent", Opponent()),
		Triggered:    []game.TriggeredAbility{drain, move},
	})
}

// msc18YouAttackTheEnchantedPlayer is "whenever you attack enchanted player".
func msc18YouAttackTheEnchantedPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Actor != source.Controller {
		return false
	}
	return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == b17DefendingPlayer(g, ev)
}

// msc18APlayerAttacksYou is "whenever a player attacks you".
func msc18APlayerAttacksYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventAttack || ev.Actor == source.Controller {
		return false
	}
	return b17DefendingPlayer(g, ev) == source.Controller
}

// msc18TheAttackedPlayerLoses2 is "that player loses 2 life": the player the
// attack named, read off the triggering event.
func msc18TheAttackedPlayerLoses2(ctx *Context) error {
	who := b17DefendingPlayer(ctx.Game, ctx.Trigger().Event)
	if p := ctx.Game.PlayerByIDForEffect(who); p == nil || p.Eliminated {
		return nil
	}
	return GainLife{Player: who, Amount: -2}.Apply(ctx)
}

// msc18AttachToTheAttacker moves the Aura to the player who attacked.
func msc18AttachToTheAttacker(ctx *Context) error {
	aura, ok := battlefieldCardForEffect(ctx.Game, ctx.Item.SourceCardID)
	if !ok {
		return nil
	}
	who := ctx.Trigger().Event.Actor
	p := ctx.Game.PlayerByIDForEffect(who)
	if p == nil || p.Eliminated || who == ctx.Controller() {
		return nil
	}
	if aura.AttachedTo.Kind == game.TargetPlayer && aura.AttachedTo.ID == who {
		return nil
	}
	return ctx.Game.AttachForEffect(ctx.Item.SourceCardID, game.TargetRef{Kind: game.TargetPlayer, ID: who})
}
