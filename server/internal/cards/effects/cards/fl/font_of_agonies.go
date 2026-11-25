package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Font of Agonies — Enchantment {B} (Ravnica Allegiance, #5521):
//
//	"Whenever you pay life, put that many blood counters on this
//	 enchantment.
//	 {1}{B}, Remove four blood counters from this enchantment: Destroy target
//	 creature."
//
// "Pay life" is life paid as a cost, by any means (CR 119.4): a spell's or
// ability's cost, Phyrexian mana, ward; damage and an effect's life loss are
// not. "That many" is the life that actually moved.
//
// No simplification.
func init() {
	const name = "Font of Agonies"
	effects.Register(effects.Spec{
		OracleID:     "ab44c6fe-6a66-42f0-9e75-8b25641e1f13",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventChangeLife, effects.YouPayLife, name+" — put that many blood counters on this enchantment",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount >= 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.AddCounter{Target: item.SourceCardID, Kind: "blood", N: -item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Remove four blood counters from this enchantment: Destroy target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.RemoveCountersFromThis("blood", 4)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
