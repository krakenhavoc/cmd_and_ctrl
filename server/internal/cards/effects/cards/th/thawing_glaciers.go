package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thawing Glaciers — Land (Alliances, S82 OLD batch 820):
//
//	"This land enters tapped.
//	 {1}, {T}: Search your library for a basic land card, put that card onto
//	 the battlefield tapped, then shuffle. Return this land to its owner's
//	 hand at the beginning of the next cleanup step."
//
// The return is a delayed trigger (CR 603.7) scheduled as the ability
// resolves, whether or not the search found a land.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6792d9f-8b74-43c4-814f-ba4adab2fdea",
		Name:         "Thawing Glaciers",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Search your library for a basic land card, put that card onto the battlefield tapped, then shuffle. Return this land to its owner's hand at the beginning of the next cleanup step.",
			Purpose: game.Purpose{Lands: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.SearchLibrary{
					Player:        item.Controller,
					Predicate:     effects.IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Thawing Glaciers — a basic land card, onto the battlefield tapped",
				}).Apply(ctx); err != nil {
					return err
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ScheduleDelayedTrigger{
					At:     game.StepCleanup,
					Label:  "Thawing Glaciers — return it to its owner's hand",
					Cards:  []uuid.UUID{item.SourceCardID},
					Body:   effects.S82GapB221ReturnListedToHandBody,
					Params: game.EffectParams{Object: game.ObjectRef{ID: item.SourceCardID, Epoch: c.ObjectEpoch}},
				}.Apply(ctx)
			},
		}},
	})
}
