package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spymaster's Vault — Land (Modern Horizons 3, #5948):
//
//	"This land enters tapped unless you control a Swamp.
//	 {T}: Add {B}.
//	 {B}, {T}: Target creature you control connives X, where X is the
//	 number of creatures that died this turn."
//
// Castle Locthwain's entry and tap; X is read as the ability resolves
// (the turn's creature-death tally, tokens included), so a creature that
// dies in response raises it. X = 0 connives nothing (ConniveX).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69ddca4b-5cc0-45f3-b2e6-a047c8d601be",
		Name:         "Spymaster's Vault",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B06EntersTappedUnlessLandType("swamp")},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: Target creature you control connives X, where X is the number of creatures that died this turn.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				return effects.ConniveX{Controller: item.Controller, Creature: t.ID, X: effects.B11CreaturesDiedThisTurn(g)}.Apply(ctx)
			},
		}},
	})
}
