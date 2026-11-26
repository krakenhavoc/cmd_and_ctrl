package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vance's Blasting Cannons // Spitfire Bastion — Legendary Enchantment {3}{R}
// (Ixalan, S82 older sets):
//
//	"At the beginning of your upkeep, exile the top card of your library. If
//	 it's a nonland card, you may cast that card this turn.
//	 Whenever you cast your third spell in a turn, you may transform Vance's
//	 Blasting Cannons."
//
//	Spitfire Bastion — Legendary Land:
//	"{T}: Add {R}.
//	 {2}{R}, {T}: Spitfire Bastion deals 3 damage to any target."
//
// The upkeep is an impulse exile that may only be CAST (ExileTopWithPermission
// CastOnly: a land stays stranded in exile, which is "if it's a nonland
// card"), until end of turn. The third spell is Captain Ripley Vance's
// R37YouCastYourThirdSpellEachTurn, with the "may" asked on resolution and the
// turn-over refused if the permanent has turned since (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "5e7eca9c-a7b8-4b7b-a0a0-e8937530145a"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Vance's Blasting Cannons",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Vance's Blasting Cannons — exile the top card of your library; if it's a nonland card, you may cast it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1, CastOnly: true}.Apply(effects.NewContext(g, item))
				}),
			effects.Optional(effects.On(game.EventCast, effects.R37YouCastYourThirdSpellEachTurn,
				"Vance's Blasting Cannons — transform it", effects.Do(effects.TransformThis{})),
				"Vance's Blasting Cannons — transform it?"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Spitfire Bastion",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, {T}: Spitfire Bastion deals 3 damage to any target.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.DealDamageToTheTarget(3),
		}},
	})
}
