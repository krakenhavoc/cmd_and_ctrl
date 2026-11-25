package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghired, Conclave Exile — Legendary Creature — Human Shaman {2}{R}{G}{W}, 2/5
// (Commander 2019):
//
//	"When Ghired enters, create a 4/4 green Rhino creature token with trample.
//	 Whenever Ghired attacks, populate. The token enters tapped and attacking.
//	 (To populate, create a token that's a copy of a creature token you
//	 control.)"
//
// The enters trigger is the printed Rhino. The attack trigger is populate
// (CR 701.36), the copy chosen among your creature tokens as it resolves, whose
// continuation is told the token that landed: it is tapped and put into combat
// attacking (CR 506.3c: never declared as an attacker, so nothing triggers on
// it), first against the player or planeswalker Ghired is attacking, and then
// its controller is offered the choice of what it attacks (CR 508.4), "keep
// attacking" being the first answer. Populate with nothing to copy creates
// nothing, and the tail runs for no token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce08fbfb-5d51-4dfc-80ae-87d5d9c7726f",
		Name:         "Ghired, Conclave Exile",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ghired, Conclave Exile — create a 4/4 green Rhino creature token with trample",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 green Rhino with trample"), N: 1})),
			effects.WheneverThisAttacks("Ghired, Conclave Exile — populate; the token enters tapped and attacking",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					tail := ctx.ItemTail(ghiredPopulated)
					if item.Trigger != nil && item.Trigger.Event.Target != uuid.Nil {
						tail.Carry = []uuid.UUID{item.Trigger.Event.Target}
					}
					return effects.Populate{ThenTail: tail}.Apply(ctx)
				}),
		},
	})
}

// ghiredPopulated taps the populated token and puts it into combat. An on-disk
// key: never renamed, never reused.
var ghiredPopulated = effects.CardTail("then/ghired-populated", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		defender := uuid.Nil
		if len(o.Tail.Carry) > 0 {
			defender = o.Tail.Carry[0]
		}
		if !ctx.Game.PutOnBattlefieldAttackingForEffect(id, defender) {
			refs := ctx.Game.AttackTargetsForEffect(ctx.Controller())
			if len(refs) == 0 || !ctx.Game.PutOnBattlefieldAttackingForEffect(id, refs[0].ID) {
				continue
			}
		}
		ctx.Game.QueueReselectAttackForEffect(game.ReselectAttackPrompt{
			Chooser:  ctx.Controller(),
			Attacker: id,
			Source:   ctx.Source(),
			Question: ctx.Game.ReselectAttackQuestion(id),
		})
	}
	return nil
})
