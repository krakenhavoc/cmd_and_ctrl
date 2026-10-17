package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Werewolf Pack Leader — Creature — Human Werewolf {G}{G}, 3/3:
//
//	"Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, draw a card.
//	 {3}{G}: Until end of turn, this creature has base power and
//	 toughness 5/3, gains trample, and isn't a Human."
//
// Pack tactics is an ability word. Its trigger is an intervening "if"
// (CR 603.4): the total power of the creatures declared as attackers is
// read when the attack is declared (the attacking creatures then) and
// again as the trigger resolves (the creatures in the declaration's own
// event batch, so a creature put onto the battlefield attacking is not
// "attacked with"). A creature that has left counts at the power it last
// had (CR 608.2h).
//
// The activated ability is one record until end of turn (CR 611.2c, 613.1):
// "isn't a Human" is a layer-4 subtype loss (game.LoseSubtypesMod, CR
// 613.1d, #4072), the base 5/3 layer 7b and trample layer 6. While it
// lasts a Human lord (Mayor of Avabruck) no longer pumps the creature,
// and a non-Human lord (Mikaeus, the Unhallowed) does.
//
// No simplification.
func init() {
	const label = "Werewolf Pack Leader — pack tactics: draw a card"
	Register(Spec{
		OracleID:     "086a911f-9ed4-4260-ae46-58c9333b95f5",
		Name:         "Werewolf Pack Leader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
					packTacticsAttackingPower(g, source.Controller) >= 6
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if packTacticsTotalPower(g, item.Controller, ctx.Trigger().Event.Batch) < 6 {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}: Until end of turn, this creature has base power and toughness 5/3, gains trample, and isn't a Human.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{3}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return untilEndOfTurn(ctx, ctx.Source(), nil,
					"Werewolf Pack Leader — base 5/3, trample and not a Human until end of turn",
					game.SetBasePowerMod(5), game.SetBaseToughnessMod(3), game.AddKeywordsMod("trample"),
					game.LoseSubtypesMod(ctx.TextSubtype("Human")))
			},
		}},
	})
}

// packTacticsAttackingPower is the check as the attack is declared: the
// whole declaration is staged before its first EventAttack is announced
// (CR 508.1), but the later attackers' events are not logged yet, so the
// attackers are read off the battlefield. Only the creatures declared so
// far are attacking at that moment, so a creature put onto the
// battlefield attacking afterwards is not counted.
func packTacticsAttackingPower(g *game.Game, player uuid.UUID) int {
	total := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.AttackingTarget != uuid.Nil {
			if p := c.PowerForComparison(); p > 0 {
				total += p
			}
		}
	}
	return total
}

// packTacticsTotalPower is the total power of the creatures `player`
// declared as attackers in the declaration `batch`: a creature still on
// the battlefield at its current power, one that has left at the power it
// last had.
func packTacticsTotalPower(g *game.Game, player uuid.UUID, batch uint64) int {
	total := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventAttack || ev.Batch != batch || ev.Actor != player {
			continue
		}
		if c, ok := battlefieldCardForEffect(g, ev.CardID); ok {
			if p := c.PowerForComparison(); p > 0 {
				total += p
			}
			continue
		}
		if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok && info.Power > 0 {
			total += info.Power
		}
	}
	return total
}
