package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Identity Thief — Creature — Shapeshifter {2}{U}{U}, 0/3:
//
//	"Whenever this creature attacks, you may exile another target nontoken
//	 creature. If you do, this creature becomes a copy of that creature
//	 until end of turn. Return the exiled card to the battlefield under its
//	 owner's control at the beginning of the next end step."
//
// The target is chosen as the trigger goes on the stack and the "may" is
// asked as it resolves. The copiable values are read off the creature
// before it is exiled (CR 707.2) and the copy begins only if the exile
// really happened ("if you do"): a creature that left, or that the owner
// put in the command zone instead, copies nothing. The return is the
// shared delayed trigger of the flicker family, under its owner's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a200daec-983c-4ea6-ae98-1c963b50fbb1",
		Name:         "Identity Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			func() game.TriggeredAbility {
				t := Optional(On(game.EventAttack, ThisAttacked,
					"Identity Thief — you may exile another target nontoken creature; if you do, this creature becomes a copy of it until end of turn and the card returns at the beginning of the next end step",
					identityThiefSteal),
					"Identity Thief — exile the target creature and become a copy of it until end of turn?")
				t.TargetsFrom = AnotherTarget(func(other CardPredicate) *game.TargetSpec {
					return TargetCreature("another target nontoken creature", other, Not(IsTokenPredicate()))
				})
				return t
			}(),
		},
	})
}

func identityThiefSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	values, ok := g.CopiableValuesForEffect(id)
	if !ok {
		return nil
	}
	self := ctx.Source()
	if err := exileTargetsThenScheduleReturn(ctx, "Identity Thief — return the exiled card"); err != nil {
		return err
	}
	if !identityThiefInExile(g, id) {
		return nil
	}
	return BecomeCopy{
		Targets:  []uuid.UUID{self},
		Of:       id,
		Values:   &values,
		Duration: CopyUntilEndOfTurn,
		Label:    "Identity Thief — becomes a copy until end of turn",
	}.Apply(ctx)
}

func identityThiefInExile(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneExile
}
