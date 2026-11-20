package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wick's Patrol — Creature — Rat Warlock {4}{B}{B}, 5/3 (Bloomburrow,
// #4467):
//
//	"When this creature enters, mill three cards. When you do, target
//	 creature an opponent controls gets -X/-X until end of turn, where X
//	 is the greatest mana value among cards in your graveyard."
//
// Primal Witchstalker's mill-then-reflexive shape, with the ruling's
// threshold: a library of two or fewer cards mills what it has and the
// reflexive trigger does not happen. That is read before the mill and
// filed for the second step, which waits for any prompt the mill asks
// (CR 608.2c). The CR 603.12 trigger picks its creature as it goes on
// the stack, after the mill. X is read once, as it resolves (the
// ruling), over every card in your graveyard; with X = 0 nothing
// changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29423a9a-31a6-4605-8b49-a75895038a3a",
		Name:         "Wick's Patrol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wick's Patrol — mill three cards", effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
					if p == nil || p.Library == nil {
						return nil
					}
					if p.Library.Size() >= 3 {
						ctx.RecordStep(wicksPatrolMilledThree, nil, 1)
					}
					return effects.MillCards{N: 3}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if ctx.StepRecord(wicksPatrolMilledThree).N() == 0 {
						return nil
					}
					return effects.WhenYouDo("Wick's Patrol — target creature an opponent controls gets -X/-X until end of turn (X = greatest mana value in your graveyard)",
						wicksPatrolShrinkBody).Apply(ctx)
				}),
			)),
		},
	})
}

// wicksPatrolMilledThree files that the library held three cards to mill.
const wicksPatrolMilledThree = "wicks-patrol/milled-three"

// wicksPatrolShrinkBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var wicksPatrolShrinkBody = game.ReflexiveBody("wicks-patrol/shrink",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		x := effects.RfCreatureCGreatestManaValueInGraveyard(g, item.Controller)
		if x <= 0 {
			return nil
		}
		return effects.BoostUntilEOT{Target: target, Power: -x, Toughness: -x, Label: "Wick's Patrol — -X/-X"}.Apply(ctx)
	},
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())
	}))
