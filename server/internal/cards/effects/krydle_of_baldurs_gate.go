package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krydle of Baldur's Gate — Legendary Creature — Human Elf Rogue {U}{B}, 1/3
// (Adventures in the Forgotten Realms):
//
//	"Whenever Krydle deals combat damage to a player, that player loses 1
//	 life and mills a card, then you gain 1 life and scry 1.
//	 Whenever you attack, you may pay {2}. If you do, target creature can't
//	 be blocked this turn."
//
// The damage trigger reads the damaged player off its event and does the four
// instructions in order: the player loses 1 life, mills one card, then
// Krydle's controller gains 1 and scrys 1 (the scry is last, so nothing
// follows it). "Whenever you attack" is one trigger per declaration
// (OncePerBatch, the Kiora shape). Its target is chosen as it goes on the
// stack (CR 603.3d), before the payment is offered; the "if you do" half is
// the registered pay branch, which reads that target back and skips it if it
// is no longer legal (CR 608.2b). No simplification.
func init() {
	Register(Spec{
		OracleID:     "77212081-0b49-4d92-9c5a-1f21e6907fce",
		Name:         "Krydle of Baldur's Gate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Krydle of Baldur's Gate — that player loses 1 life and mills a card, then you gain 1 life and scry 1",
				krydleDrain),
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Krydle of Baldur's Gate — you may pay {2}: target creature can't be blocked this turn", krydleMayPay),
				TargetCreature("target creature"))),
		},
	})
}

func krydleDrain(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Target == uuid.Nil {
		return nil
	}
	ctx := NewContext(g, item)
	victim := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(victim) != nil {
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), victim, -1); err != nil {
			return err
		}
		if err := (MillCards{Player: victim, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if err := (GainLife{Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	return Scry{N: 1}.Apply(ctx)
}

func krydleMayPay(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}",
		Question: "Krydle of Baldur's Gate — pay {2}: target creature can't be blocked this turn?",
		Then:     krydlePaid,
	}.Apply(NewContext(g, item))
}

var krydlePaid = PayUnlessBranches("pay-unless/krydle-of-baldurs-gate-unblockable",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return RestrictUntilEOT{
				Target:       t.ID,
				Restrictions: game.CantBeBlocked,
				Label:        "Krydle of Baldur's Gate — can't be blocked",
			}.Apply(ctx)
		}
		return nil
	}, nil)
