package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Echoing Assault — Enchantment {4}{R} (Bloomburrow Commander):
//
//	"Creature tokens you control have menace.
//	 Whenever you attack a player, choose target nontoken creature that's
//	 attacking that player. Create a token that's a copy of that creature,
//	 except it's 1/1. The token enters tapped and attacking that player.
//	 Sacrifice it at the beginning of the next end step."
//
// Menace is the layer 6 grant to creature tokens (Brudiclad's scope). The
// attack trigger is once per player attacked in a declaration (Long-Range
// Sensor's shape, CR 508.3e: an attack at a planeswalker is no attack at a
// player). The target is a nontoken creature attacking THAT player
// (TargetSpec.AttackingThatPlayer), re-checked as the trigger resolves (CR
// 608.2b). The token takes the creature's copiable values (CR 707.2) with
// its power and toughness set to 1 (Offspring's clause), enters tapped and
// attacking that player without being declared (CR 508.4), and a delayed
// trigger sacrifices it at the beginning of the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d15dba3a-3b85-43fb-b5f3-c4d8d1e23b1d",
		Name:         "Echoing Assault",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(creatureTokensYouControl, "menace"), "Creature tokens you control have menace."),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatchPerPlayer(Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16YouAttackedAPlayer(ev, source, g)
			}, "Echoing Assault — create a 1/1 token copy of target nontoken attacker, tapped and attacking that player", echoingAssaultCopy),
				AttackingThatPlayer(TargetCreature("target nontoken creature that's attacking that player", Not(IsTokenPredicate()))))),
		},
	})
}

// echoingAssaultCopy makes the 1/1 copy of the chosen attacker.
func echoingAssaultCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, t.ID)
	if !ok {
		return nil
	}
	offspringOneOne(&tmpl)
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: item.Controller, Source: ctx.Source(),
		Groups: []game.TokenGroup{{Template: copiedToken(tmpl), Count: 1, Entry: game.TokenEntryOptions{Tapped: true}, Attacking: triggerAttackTarget(item)}},
	}, ctx.ItemTail(echoingAssaultSacrifice))
}

// echoingAssaultSacrifice is "Sacrifice it at the beginning of the next end
// step" over the tokens that landed. An on-disk key.
var echoingAssaultSacrifice = CardTail("echoing-assault/sacrifice-at-next-end-step", func(ctx *Context, o game.TailOutcome) error {
	var ids []uuid.UUID
	ids = append(ids, o.IDs...)
	if len(ids) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Echoing Assault — sacrifice the token",
		Cards: ids,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
})
