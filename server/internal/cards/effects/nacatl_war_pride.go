package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nacatl War-Pride — Creature — Cat Warrior, {3}{G}{G}{G}, 3/3:
//
//	"This creature must be blocked by exactly one creature if able.
//	 Whenever this creature attacks, create X tokens that are copies of
//	 it and that are tapped and attacking, where X is the number of
//	 creatures defending player controls. Exile the tokens at the
//	 beginning of the next end step."
//
// #1684: the "exactly one" requirement is #1597's exactlyOne kind, as
// a static. The attack trigger is Saheeli's token copy (TokenCopyTemplate,
// so each token carries the War-Pride's oracle ID and with it the same
// requirement) made tapped and attacking the War-Pride's own attack
// target, which is Adeline's shape: the tokens were never DECLARED as
// attackers, so they do not trigger again (CR 508.4). X is counted when
// the trigger resolves, off the defending player of that attack. The
// tokens are exiled by the shared end-step delayed trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "692fdca7-0eff-4d2d-b744-213b6f6e18fb",
		Name:         "Nacatl War-Pride",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{BlockRequirementWhere(game.BlockRequirementExactlyOne, selfOnly)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Nacatl War-Pride — X tapped and attacking token copies, where X is the number of creatures defending player controls",
				nacatlWarPrideCopies),
		},
	})
}

func nacatlWarPrideCopies(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	target := item.Trigger.Event.Target
	defender := g.DefendingPlayerForAttackForEffect(target)
	if defender == uuid.Nil {
		return nil
	}
	x := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsCreature() && c.Controller == defender {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(g, item.SourceCardID)
	if !ok {
		return nil
	}
	tmpl.Tapped = true
	tmpl.AttackingTarget = target
	made, err := g.CreateTokensForEffect(item.Controller, tmpl, x, game.TokenEntryOptions{Tapped: true})
	if err != nil {
		return err
	}
	if len(made) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: item.Controller,
		Label:      "Nacatl War-Pride — exile the tokens",
		Cards:      made,
		Body:       exileListedCardsBody,
	}.Apply(NewContext(g, item))
}
