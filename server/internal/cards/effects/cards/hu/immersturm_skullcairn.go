package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immersturm Skullcairn — Land (Kaldheim, #5505):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {1}{B}{R}{R}, {T}, Sacrifice this land: It deals 3 damage to target
//	 player. That player discards a card. Activate only as a sorcery."
//
// The damage is the sacrificed land's (last known information, CR 608.2h);
// the discard is the player's own pick and follows whether or not the
// damage was prevented.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "354a7376-fb4b-424d-8964-93727302dccb",
		Name:         "Immersturm Skullcairn",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}{R}{R}, {T}, Sacrifice this land: It deals 3 damage to target player. That player discards a card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B}{R}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets:      effects.TargetPlayer("target player"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.DealDamageToTheTarget(3)(g, item); err != nil {
					return err
				}
				return effects.TargetPlayerDiscards(1)(g, item)
			},
		}},
	})
}
