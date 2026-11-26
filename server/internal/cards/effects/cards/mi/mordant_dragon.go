package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mordant Dragon — Creature — Dragon {3}{R}{R}{R}, 5/5 (Commander Starter
// Decks):
//
//	"Flying
//	 {1}{R}: This creature gets +1/+0 until end of turn.
//	 Whenever this creature deals combat damage to a player, you may have it
//	 deal that much damage to target creature that player controls."
//
// Bellows Lizard's firebreathing. The damage trigger targets (a clause that
// is a fact about what happened, TargetsFrom, so the creature must be one the
// damaged player controls), asks the "may" as it resolves (CR 603.5) and, on
// a yes, deals the combat damage's amount again as the Dragon's own damage
// (the damage dealt is read from the triggering event, so a Dragon that has
// since left still deals it, CR 608.2h). It is re-judged as it resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Mordant Dragon"
	effects.Register(effects.Spec{
		OracleID:        "c02ed1b7-1f8d-4196-a9b6-65e1b5a763d7",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, name+" — +1/+0 until end of turn"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Optional(game.TriggeredAbility{
				Watches:     []game.EventKind{game.EventDealDamage},
				AppliesTo:   effects.ThisDealtCombatDamageToAPlayer,
				TargetsFrom: effects.S82OldB458TargetCreatureThatPlayerControls,
				Key:         name + " — have it deal that much damage to target creature that player controls",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					amount := item.Trigger.Event.Amount
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
					}
					return nil
				},
			}, name+" — have it deal that much damage to target creature that player controls?"),
		},
	})
}
