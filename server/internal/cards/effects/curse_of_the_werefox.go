package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curse of the Werefox — Sorcery {2}{G}:
//
//	"Create a Monster Role token attached to target creature you
//	 control. When you do, that creature fights up to one target
//	 creature you don't control. (If you control another Role on it, put
//	 that one into the graveyard. Enchanted creature gets +1/+1 and has
//	 trample. Creatures that fight each deal damage equal to their power
//	 to the other.)"
//
// "When you do" is a CR 603.12 reflexive trigger: it goes on the stack
// above the spell once the Role exists, and its target ("up to one
// creature you don't control") is chosen as it goes there, not when the
// spell was cast. The creature that fights rides the payload. The
// trigger is only created if the Role could be attached — a creature
// that left in response creates nothing and fights nothing.
//
// No simplifications.
var curseOfTheWerefoxFightBody = game.ReflexiveBody("curse-of-the-werefox/fight",
	simpleBody(curseOfTheWerefoxFight),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("up to one target creature you don't control", Not(YouControl())).WithCount(0, 1)
	}))

func curseOfTheWerefoxFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	mine := ctx.PayloadCards()
	if len(mine) == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return b10Fight(ctx, mine[0], t.ID)
		}
	}
	return nil
}

func init() {
	Register(Spec{
		OracleID:     "64b41b54-ca22-41df-8bd3-dc4f7832fbb7",
		Name:         "Curse of the Werefox",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (CreateRoleToken{Role: RoleMonster, Host: id}).Apply(ctx); err != nil {
				return err
			}
			return ReflexiveTrigger{
				Label: "Curse of the Werefox — that creature fights up to one target creature you don't control",
				Cards: []uuid.UUID{id},
				Body:  curseOfTheWerefoxFightBody,
			}.Apply(ctx)
		},
	})
}
