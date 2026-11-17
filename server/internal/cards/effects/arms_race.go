package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arms Race — Enchantment {3}{R} (The Brothers' War, S82):
//
//	"{3}{R}: You may put an artifact card from your hand onto the
//	 battlefield. That artifact gains haste. Sacrifice it at the beginning
//	 of the next end step."
//
// Meek Attack's put from hand over artifact cards: optional, the tail runs
// once the card has landed and grants haste (until end of turn, which is
// exact: the artifact is sacrificed at the next end step) and schedules
// the sacrifice for that object (CR 603.7a), shared with Meek Attack. It
// is skipped if the artifact is no longer that object under your control
// (CR 603.7c). Declining the put makes no delayed trigger. Haste only
// matters if the artifact is a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e801c5bd-68e6-4a74-bed3-addd18db3b2b",
		Name:         "Arms Race",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}: You may put an artifact card from your hand onto the battlefield. That artifact gains haste. Sacrifice it at the beginning of the next end step.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{3}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Artifact(),
					Optional: true,
					Label:    "Arms Race — you may put an artifact card from your hand onto the battlefield",
					ThenTail: armsRaceEntered,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// armsRaceEntered is "that artifact gains haste. Sacrifice it at the
// beginning of the next end step", run once the card has landed. An on-disk
// key.
var armsRaceEntered = PutFromHandTail("arms-race/entered", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: res.Entered, Keywords: []string{"haste"},
		Label: "Arms Race — haste"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label:  "Arms Race — sacrifice that artifact",
		Body:   stdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
