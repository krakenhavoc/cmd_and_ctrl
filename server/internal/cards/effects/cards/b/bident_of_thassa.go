package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bident of Thassa — Legendary Enchantment Artifact for {2}{U}{U}:
//
//	"Whenever a creature you control deals combat damage to a player,
//	you may draw a card.
//	{1}{U}, {T}: Creatures your opponents control attack this turn if
//	able."
//
// S19 sub-PR 7: the first combat-damage trigger. EventDealDamage
// carries Combat=true for damage dealt in the combat damage step;
// combatDamageToPlayerBy checks the source is a creature the
// Bident's controller controls and the target is a player. One
// event per attacker per player, so three unblocked attackers queue
// three "you may draw" prompts — as in paper.
//
// #1571: the activated ability is a CR 508.1d attack requirement on
// "creatures your opponents control", until end of turn. It is a
// requirement, not a characteristic change, so CR 611.2c does not lock
// the affected set: a creature an opponent casts after the ability
// resolved has to attack too, which is why it is a live-scope record
// (OpponentsCreaturesAttackIfAble) rather than a snapshot. The engine
// judges it with every other requirement — a creature that is tapped,
// summoning sick, or could attack only by paying a Propaganda-style tax
// is not made to (CR 508.1d).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1afaef7-9fa3-4662-a95f-adfb0da9fd11",
		Name:         "Bident of Thassa",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OptionalDraw(effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			}, "Bident of Thassa — draw a card", effects.Do(effects.DrawCards{N: 1})), "Bident of Thassa — draw a card?", 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Creatures your opponents control attack this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.OpponentsCreaturesAttackIfAble{
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Bident of Thassa — creatures your opponents control attack this turn if able",
				}.Apply(ctx)
			},
		}},
	})
}
