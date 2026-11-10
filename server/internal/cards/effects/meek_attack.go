package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meek Attack — Enchantment {2}{R} (Lorwyn Eclipsed, #4458):
//
//	"{1}{R}: You may put a creature card with total power and toughness 5
//	 or less from your hand onto the battlefield. That creature gains haste.
//	 At the beginning of the next end step, sacrifice that creature."
//
// Quicksilver Amulet's put from hand, optional, over creature cards whose
// printed power plus toughness is five or less. The tail runs once the card
// has landed: it gains haste until end of turn and a delayed trigger is made
// for it (CR 603.7a), bound to that object. Haste has no printed duration, but
// the creature is sacrificed at the next end step, which is this turn's end
// step or, when the ability is activated during an end step, the next turn's,
// by which time haste changes nothing, so until end of turn is exact. The
// sacrifice happens only if you still control that same object (the ruling):
// one that left the battlefield, even if it came back, stays. Declining the
// put makes no delayed trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9fff42d-31cf-4edf-8353-4a3b996bb77b",
		Name:         "Meek Attack",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: You may put a creature card with total power and toughness 5 or less from your hand onto the battlefield. That creature gains haste. At the beginning of the next end step, sacrifice that creature.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    ManaCost("{1}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match: And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.Power+c.Toughness <= 5
					}),
					Optional: true,
					Label:    "Meek Attack — you may put a creature card with total power and toughness 5 or less from your hand onto the battlefield",
					ThenTail: meekAttackEntered,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// meekAttackEntered is "that creature gains haste. At the beginning of the
// next end step, sacrifice that creature". An on-disk key.
var meekAttackEntered = PutFromHandTail("meek-attack/entered", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: res.Entered, Keywords: []string{"haste"},
		Label: "Meek Attack — haste"}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label:  "Meek Attack — sacrifice that creature",
		Body:   stdB66MeekAttackSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})
