package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vazi, Keen Negotiator — Legendary Creature — Human Advisor
// {2}{B}{R}{G}, 3/3 (Streets of New Capenna Commander):
//
//	"Haste
//	 {T}: Target opponent creates X Treasure tokens, where X is the number
//	 of Treasure tokens you created this turn.
//	 Whenever an opponent casts a spell or activates an ability, if mana
//	 from a Treasure was spent to cast it or activate it, put a +1/+1
//	 counter on target creature, then draw a card."
//
// X is counted as the ability resolves: the Treasure permanents that
// entered under your control this turn, as they entered (the turn tally,
// so a Treasure you have since sacrificed still counts, CR 111.7 aside).
// The trigger is one row over both events (CR 603.2): an opponent's cast
// and an opponent's activation, each judged by the payment record of the
// spell or ability on the stack (StackItem.Paid), whose mana kinds are a
// snapshot, so a Treasure that sacrificed itself to pay still counts.
// Mana abilities use no stack and are never activations of this kind.
// Vazi declares ReadsPoolManaOnly: Treasure mana can only come out of the
// pool, so no waived payment is ever asked about it. The condition is an
// intervening if (CR 603.4): it is the payment as the spell or ability was
// put on the stack, which cannot change, so it is judged as the trigger
// fires. If the creature target is gone on resolution, nothing happens,
// the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7446c013-63b0-402b-b8b9-dd25ccc553df",
		Name:            "Vazi, Keen Negotiator",
		ReadsManaSpent:  game.ReadsPoolManaOnly,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.OnAny([]game.EventKind{game.EventCast, game.EventActivateAbility}, s82SncB11OpponentPaidWithTreasureMana,
					"Vazi, Keen Negotiator — put a +1/+1 counter on target creature, then draw a card",
					effects.S82SncB11VaziCounterThenDraw),
				effects.TargetCreature("target creature"),
			),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target opponent creates X Treasure tokens, where X is the number of Treasure tokens you created this turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				opp := effects.TargetedPlayer(ctx)
				n := g.EnteredWithSubtypeThisTurn(item.Controller, "Treasure")
				if opp == uuid.Nil || n <= 0 {
					return nil
				}
				return effects.CreateToken{Controller: opp, Template: effects.TreasureToken(), N: n}.Apply(ctx)
			},
		}},
	})
}

// s82SncB11OpponentPaidWithTreasureMana is Vazi's condition: an opponent
// of the source's controller cast a spell or activated an ability (not a
// mana ability), and mana from a Treasure was spent to do it, read off the
// payment record of the item on the stack.
func s82SncB11OpponentPaidWithTreasureMana(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	switch ev.Kind {
	case game.EventCast:
		return g.StackItemPaidForEffect(ev.CardID).Spent().FromTreasure()
	case game.EventActivateAbility:
		return ev.StackItemID != uuid.Nil && g.StackItemPaidForEffect(ev.StackItemID).Spent().FromTreasure()
	}
	return false
}
