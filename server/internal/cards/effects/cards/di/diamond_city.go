package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diamond City — Land (Fallout):
//
//	"This land enters with a shield counter on it. (If it would be dealt
//	 damage or destroyed, remove a shield counter from it instead.)
//	 {T}: Add {C}.
//	 {T}: Move a shield counter from this land onto target creature.
//	 Activate only if two or more creatures entered the battlefield under
//	 your control this turn."
//
// The entry counter is Captain America's replacement. The condition is
// Park Heights Pegasus's creature-entry tally for the turn, checked as the
// ability is activated. The move is Simic Fluxmage's: the shield counter is
// put on the target first and, if it landed, taken off the land, so a target
// that is gone by resolution, or a land that no longer has a counter, does
// nothing (CR 608.2b, 701.34). The shield counter's own protection is the
// engine's (CR 122.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88e29d50-1680-495d-be84-b92b4c9e636f",
		Name:         "Diamond City",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterShield, 1, "This land enters with a shield counter on it."),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Move a shield counter from this land onto target creature. Activate only if two or more creatures entered the battlefield under your control this turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive},
			Cost:    effects.TapCost(),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return g.EnteredWithCardTypeThisTurn(controller, "creature") >= 2
			},
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || effects.SourceIsNewObject(g, item) {
					return nil
				}
				self, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) || self.Counters[game.CounterShield] <= 0 {
					return nil
				}
				return g.AddCounterByKeyedForEffect(item.Controller, t.ID, game.CounterShield, 1, ctx.Tail(diamondCityTakeOne))
			},
		}},
	})
}

// diamondCityTakeOne removes the shield counter from the land, if the one
// onto the creature landed. An on-disk key: never renamed, never reused.
var diamondCityTakeOne = game.RegisterTail("diamond-city/take-one", func(g *game.Game, o game.TailOutcome) error {
	if o.Amount <= 0 {
		return nil
	}
	return g.AddCounterForEffect(o.Tail.Source, game.CounterShield, -1)
})
