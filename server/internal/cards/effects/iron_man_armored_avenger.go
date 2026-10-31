package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iron Man, Armored Avenger — Legendary Artifact Creature — Human Hero
// {3}{U}, 2/2 (Marvel Super Heroes Commander):
//
//	"Flying
//	 Whenever you draw a card, put a +1/+1 counter on target creature.
//	 Whenever Iron Man attacks, other attacking modified creatures gain flying
//	 until end of turn. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// The draw trigger is a targeted counter on any creature (the target is picked
// as the trigger goes on the stack, per card drawn). "Modified" is Kodama of
// the West Tree's CR 700.9 reading (b07IsModified). The set of attackers that
// gain flying is locked as the attack trigger resolves (CR 611.2c); Iron Man
// himself is "other"-excluded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c81195d4-19c6-4cac-b312-7f5c33dff962",
		Name:            "Iron Man, Armored Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouDraw("Iron Man — put a +1/+1 counter on target creature",
				Do(Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }))),
				TargetCreature("target creature")),
			WheneverThisAttacks("Iron Man — other attacking modified creatures gain flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					self := item.SourceCardID
					return GrantKeywordUntilEOT{
						Match: func(g *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.InstanceID != self && c.IsCreature() && c.AttackingTarget != uuid.Nil && b07IsModified(g, c)
						},
						Keywords: []string{"flying"},
						Label:    "Iron Man — flying until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
