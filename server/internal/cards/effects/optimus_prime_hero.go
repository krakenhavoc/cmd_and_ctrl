package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Optimus Prime, Hero // Optimus Prime, Autobot Leader (#3277, #3195,
// ADR 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {3}{U}{R}{W}, 4/8:
//
//	"More Than Meets the Eye {2}{U}{R}{W}
//	 At the beginning of each end step, bolster 1.
//	 When Optimus Prime dies, return it to the battlefield converted
//	 under its owner's control."
//
// Back face, Legendary Artifact — Vehicle, 6/8:
//
//	"Living metal
//	 Trample
//	 Whenever you attack, bolster 2. The chosen creature gains trample
//	 until end of turn. When that creature deals combat damage to a
//	 player this turn, convert Optimus Prime."
//
// Bolster is effects.Bolster (CR 701.39). "The chosen creature" is its
// Then: the trample grant and the delayed trigger are bound to that one
// creature and its object (CR 603.7, 400.7), as Captain Howler binds its
// draw. "When that creature deals combat damage to a player this turn" is
// a one-shot delayed trigger ("when", so it fires once, and it ends at
// cleanup); its convert records Optimus Prime's turn-over count when the
// trigger is created, so a Prime that turned over in between ignores it
// (CR 701.28e). "Whenever you attack" is one trigger per declaration,
// however many creatures attack (CR 603.2c).
//
// "Return it to the battlefield converted" is a new object entering back
// face up (CR 712.14a), not a convert, and it enters untapped.
//
// No simplification.
const optimusPrimeOracleID = "199c4a1a-2082-417c-a174-f479d7334d15"

func init() {
	Register(Spec{
		OracleID:         optimusPrimeOracleID,
		Name:             "Optimus Prime, Hero",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{2}{U}{R}{W}")},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Optimus Prime, Hero — bolster 1",
				func(g *game.Game, item *game.StackItem) error {
					return Bolster{N: 1}.Apply(NewContext(g, item))
				}),
			WhenThisDies("Optimus Prime, Hero — return it to the battlefield converted",
				func(g *game.Game, item *game.StackItem) error {
					return ReturnFromGraveyard{
						Target:      item.SourceCardID,
						Dest:        game.ZoneBattlefield,
						Transformed: true, // CR 712.14a: back face up, under its owner's control
					}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:        optimusPrimeOracleID + "#1",
		Name:            "Optimus Prime, Autobot Leader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Optimus Prime, Autobot Leader — bolster 2; the chosen creature gains trample; convert when it deals combat damage to a player",
				optimusPrimeBolster)),
		},
	})
}

func optimusPrimeBolster(g *game.Game, item *game.StackItem) error {
	return Bolster{N: 2, Then: func(ctx *Context, chosen uuid.UUID) error {
		if err := (GrantKeywordUntilEOT{
			Target:   chosen,
			Keywords: []string{"trample"},
			Label:    "Optimus Prime, Autobot Leader — the chosen creature gains trample until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
		c, ok := ctx.Game.LookupCardForEffect(chosen)
		if !ok {
			return nil
		}
		return DelayedOnEvent{
			Label:      "Optimus Prime, Autobot Leader — convert Optimus Prime",
			On:         []game.EventKind{game.EventDealDamage},
			Condition:  thatCreatureHitAPlayerCondition,
			CondParams: game.EffectParams{Object: game.ObjectRef{ID: chosen, Epoch: c.ObjectEpoch}},
			Body:       transformersConvertThisBody,
		}.Apply(ctx)
	}}.Apply(NewContext(g, item))
}
