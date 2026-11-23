package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Motherlode, Excavator — Legendary Artifact Creature — Robot
// {3}{R}{R}, 5/5 (Fallout, #5948):
//
//	"When The Motherlode enters, choose target opponent. You get an amount
//	 of {E} (energy counters) equal to the number of nonbasic lands that
//	 player controls.
//	 Whenever The Motherlode attacks, you may pay {E}{E}{E}{E}. When you do,
//	 destroy target nonbasic land defending player controls, and creatures
//	 that player controls without flying can't block this turn."
//
// The enters trigger counts the chosen opponent's nonbasic lands as it
// resolves (CR 608.2b: an opponent who is no longer a legal target gives
// nothing). The attack trigger pins the defending player as the attack is
// declared; the energy is paid as it resolves (MayPayEnergy), and paying
// raises the CR 603.12 reflexive trigger, which chooses its land as it goes
// on the stack. Its two halves are one ability: the land is destroyed, and
// every creature the defending player controls then that has no flying can't
// block this turn, a set fixed as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	label := "The Motherlode — you may pay {E}{E}{E}{E}; when you do, destroy target nonbasic land defending player controls, and their creatures without flying can't block this turn"
	attack := game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: effects.ThisAttacked,
		Key:       label,
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Player = effects.B17DefendingPlayer(g, ev)
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return effects.MayPayEnergy{
				N:        4,
				Question: "The Motherlode — pay {E}{E}{E}{E} to destroy a nonbasic land and stop their ground creatures blocking?",
				Then:     motherlodePaid,
				Carry:    []uuid.UUID{item.Params.Player},
			}.Apply(effects.NewContext(g, item))
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "5ae50388-7e36-4a83-a618-0148944c5099",
		Name:         "The Motherlode, Excavator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("The Motherlode — you get {E} equal to the number of nonbasic lands target opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						n := len(effects.MatchingBattlefield(ctx, effects.And(effects.Land(), effects.B03Nonbasic(), effects.ControlledBy(t.ID))))
						return effects.GetEnergy{N: n}.Apply(ctx)
					}
					return nil
				}), effects.TargetPlayer("target opponent", effects.Opponent())),
			attack,
		},
	})
}

// motherlodePaid raises the reflexive trigger once the energy is paid, with
// the defending player on its params.
var motherlodePaid = effects.PayUnlessBranches("pay-unless/the-motherlode-paid",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 || r.Carry[0] == uuid.Nil {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "The Motherlode — destroy target nonbasic land defending player controls; their creatures without flying can't block",
			Body:   motherlodeBody,
			Params: game.EffectParams{Player: r.Carry[0]},
		}.Apply(ctx)
	}, nil)

var motherlodeBody = game.ReflexiveBody("the-motherlode/destroy-and-no-blocks", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if p.Player == uuid.Nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), effects.ControlledBy(p.Player))) {
		if !game.HasKeyword(&c, "flying") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "The Motherlode — can't block this turn"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}, func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
	return effects.TargetPermanent("target nonbasic land defending player controls", effects.Land(), effects.B03Nonbasic(), effects.ControlledBy(p.Player))
})
