package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Namor, Atlantean King — Legendary Creature — Mutant Merfolk Noble
// {1}{W}{U}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever you cast a noncreature spell, create a 1/1 blue Merfolk
//	 creature token.
//	 Whenever Namor attacks a player who has more life than you, other
//	 creatures you control attacking that player get +2/+0 until end of
//	 turn."
//
// The attack condition is judged when Namor is declared as an attacker (it is
// part of the trigger event, not an intervening "if"): the player he attacks
// directly, not a planeswalker or battle of theirs, has more life than his
// controller at that moment. As the trigger resolves, each OTHER creature his
// controller controls that is attacking that same player gets +2/+0; a
// creature attacking someone else, or one declared after, does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "171a0a09-4aee-466c-aebd-3b0d4c1f51b4",
		Name:            "Namor, Atlantean King",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Namor — create a 1/1 blue Merfolk creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 blue Merfolk"), N: 1})),
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclared(ev, source) {
					return false
				}
				defender := effects.B17DefendingPlayer(g, ev)
				if defender == uuid.Nil || defender != ev.Target {
					return false
				}
				them, me := g.PlayerByIDForEffect(defender), g.PlayerByIDForEffect(source.Controller)
				return them != nil && me != nil && them.Life > me.Life
			}, "Namor — other creatures you control attacking that player get +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := item.Trigger.Event.Target
					ctx := effects.NewContext(g, item)
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || c.InstanceID == item.SourceCardID ||
							!c.IsCreature() || c.AttackingTarget != defender {
							continue
						}
						if err := (effects.BoostUntilEOT{Target: c.InstanceID, Power: 2, Label: "Namor — +2/+0"}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
