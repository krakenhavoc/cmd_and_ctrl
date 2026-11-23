package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Susur Secundi, Void Altar — Land — Planet:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Planet. Station only as a sorcery.)
//	 12+ | {1}{B}, {T}, Pay 2 life, Sacrifice a creature: Draw cards
//	 equal to the sacrificed creature's power. Activate only as a
//	 sorcery."
//
// Adagia's shape (a station Planet land with a charge-gated ability)
// with Greater Good's last-known-power read for the draw.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50d6cadc-07e4-479e-90f4-e3a20f769bab",
		Name:         "Susur Secundi, Void Altar",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{
			effects.Station(),
			{
				Label:        "{1}{B}, {T}, Pay 2 life, Sacrifice a creature: Draw cards equal to the sacrificed creature's power. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.PayLife(2), effects.SacrificeACreature()),
				SorcerySpeed: true,
				ActiveWhen:   effects.AtChargeCounters(12),
				Effect: func(g *game.Game, item *game.StackItem) error {
					fed, ok := effects.B17PermanentSacrificedToPay(g, item)
					if !ok {
						return nil
					}
					return effects.DrawCards{Player: item.Controller, N: effects.DepartedCreaturePower(g, fed)}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
