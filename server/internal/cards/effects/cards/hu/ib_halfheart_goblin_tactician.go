package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ib Halfheart, Goblin Tactician — Legendary Creature — Goblin Advisor
// {3}{R}, 3/2 (Jumpstart 2022, #5948):
//
//	"Whenever another Goblin you control becomes blocked, sacrifice it. If
//	 you do, it deals 4 damage to each creature blocking it.
//	 Sacrifice two Mountains: Create two 1/1 red Goblin creature tokens."
//
// "Becomes blocked" is the once-per-attacker EventBecomesBlocked (CR 506.4,
// Battle-Scarred Goblin's), narrowed to another Goblin its controller controls
// as it is blocked, read through the source's text. The blockers are read as
// the trigger resolves, before the Goblin goes (a sacrificed creature is out of
// combat), and "if you do" is the sacrifice really happening; the damage is
// then dealt by the Goblin as it last existed (CR 608.2h), as one instruction
// to every creature that was blocking it. A Goblin that has left, or that
// changed controllers, is not sacrificed.
//
// The second ability is a cost of two Mountains (read through the text, CR
// 612.2), paid as the ability is announced; it makes two tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6742c4b9-8fa6-439f-844d-04a09fa20c45",
		Name:         "Ib Halfheart, Goblin Tactician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Goblin"))
			}, "Ib Halfheart — sacrifice it; if you do, it deals 4 damage to each creature blocking it", ibHalfheartBlocked),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice two Mountains: Create two 1/1 red Goblin creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 2, CreatureTokens: 2},
			Cost:    effects.SacrificeN(2, "two Mountains", effects.Subtype("Mountain")),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Goblin"), N: 2}),
		}},
	})
}

// ibHalfheartBlocked sacrifices the blocked Goblin and, if it really went,
// has it deal 4 damage to each creature that was blocking it. The blockers are
// read before the sacrifice (a sacrificed creature is out of combat) and ride
// the sacrifice's tail, so a CR 614 window that keeps the Goblin on the
// battlefield cancels the damage.
func ibHalfheartBlocked(g *game.Game, item *game.StackItem) error {
	goblin := item.Trigger.Event.CardID
	c, ok := g.LookupCardForEffect(goblin)
	if !ok || !effects.OnBattlefield(g, goblin) || c.Controller != item.Controller {
		return nil
	}
	blockers := effects.BlockersOf(g, goblin)
	tail := game.KeyedTail{Ref: ibHalfheartSacrificed, Player: item.Controller, Source: item.SourceCardID,
		Carry: blockers, Asked: game.NewAskedItem(item)}
	return g.SacrificeAllKeyedForEffect(uuid.Nil, []uuid.UUID{goblin}, tail)
}

// ibHalfheartSacrificed is the "if you do": the sacrificed Goblin, as it last
// existed, deals 4 damage to each creature that was blocking it, as one
// instruction. An on-disk key: never renamed, never reused.
var ibHalfheartSacrificed = effects.CardTail("then/s82-gap-b156-ib-halfheart-sacrificed", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	return ctx.Game.DealDamageEachKeyedForEffect(o.IDs[0], o.Tail.Carry, 4, game.KeyedTail{})
})
