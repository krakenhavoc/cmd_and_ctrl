package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ragost, Deft Gastronaut — Legendary Creature — Lobster Citizen {R}{W},
// 2/2 (Edge of Eternities, #4461):
//
//	"Artifacts you control are Foods in addition to their other types and
//	 have "{2}, {T}, Sacrifice this artifact: You gain 3 life."
//	 {1}, {T}, Sacrifice a Food: Ragost deals 3 damage to each opponent.
//	 At the beginning of each end step, if you gained life this turn,
//	 untap Ragost."
//
// Gemcutter Buccaneer's shape for Foods: a layer-4 subtype add on every
// artifact the controller controls (the artifacts keep every type and
// ability they have) and a layer-6 grant of the Food ability as a catalog
// bundle, so it is each artifact's own ability, "this artifact" being the
// one that has it, and each Food can pay only one cost. The sacrifice cost
// of Ragost's own ability is Bosco's "Sacrifice a Food": any Food you
// control, a Food token or an artifact Ragost made one, read post-layer.
// The untap is Angelic Accord's each-end-step shape: the intervening "if"
// is checked as the step begins and again as the trigger resolves (CR 603.4),
// on anyone's end step, from life gained at any time that turn.
//
// No simplification.
const ragostFoodGrant = "ragost-deft-gastronaut/food-life"

func init() {
	Register(Spec{
		OracleID:     "76f5445a-5193-4794-8593-b44f961334b6",
		Name:         "Ragost, Deft Gastronaut",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: ragostFoodGrant,
			Activated: []ActivatedAbility{{
				Label:   "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), game.AbilityCost{Tap: true, SacrificeSelf: true}),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
				},
			}},
			Text: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
		}},
		Static: []game.StaticAbility{
			stdB3AreAlsoFood("Artifacts you control are Foods in addition to their other types.", stdB3ArtifactsYouControl),
			Labelled(GrantAbilities(stdB3ArtifactsYouControl, ragostFoodGrant),
				"Artifacts you control have \"{2}, {T}, Sacrifice this artifact: You gain 3 life.\""),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice a Food: Ragost deals 3 damage to each opponent.",
			Purpose: game.Purpose{Answers: game.AnswerValue, DamageEachOpponent: 3},
			Cost: Plus(ManaCost("{1}"), TapCost(),
				game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Effect: Do(DamageEachOpponent{N: 3}),
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 1)
			}, "Ragost — untap Ragost (you gained life this turn)", func(g *game.Game, item *game.StackItem) error {
				if b15LifeGainedThisTurn(g, item.Controller) < 1 {
					return nil
				}
				ctx := NewContext(g, item)
				if !onBattlefield(g, ctx.Source()) {
					return nil
				}
				return UntapTarget{Target: ctx.Source()}.Apply(ctx)
			}),
		},
	})
}
