package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dauntless Bodyguard — Creature — Human Knight {W}, 2/1 (Dominaria, S82 OLD
// batch 214, #5530):
//
//	"As this creature enters, choose another creature you control.
//	 Sacrifice this creature: The chosen creature gains indestructible until
//	 end of turn."
//
// The choice is made as it enters (CR 614.12), mandatory when there is another
// creature, and is stored as the Bodyguard's linked permanent with its CR 400.7
// epoch (Tyrannical Pitlord's shape): a chosen creature that leaves and returns
// is a new object nobody chose. The sacrifice is the cost, so the chosen
// creature is read from the Bodyguard as it last existed (CR 608.2h); a chosen
// creature that has left, or changed controllers, gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70973a2d-66e6-4662-ab10-99678775dda1",
		Name:         "Dauntless Bodyguard",
		Completeness: effects.CompletenessFull,
		AsEnters:     dauntlessBodyguardChoose,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: The chosen creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSacOutlet},
			Cost:    effects.SacrificeThis(),
			Effect:  dauntlessBodyguardProtect,
		}},
	})
}

// dauntlessBodyguardChoose is "As this creature enters, choose another creature
// you control."
func dauntlessBodyguardChoose(card *game.Card, ctx *effects.Context) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID != card.InstanceID && c.Controller == card.Controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	ctx.Game.QueueChooseLinkedPermanentAsEntersForEffect(card.Controller, card.InstanceID,
		"Dauntless Bodyguard — choose another creature you control", ids)
	return nil
}

func dauntlessBodyguardProtect(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	chosen := g.LinkedPermanentForEffect(ref)
	if chosen == nil {
		return nil
	}
	if live, ok := g.LookupCardForEffect(chosen.InstanceID); !ok || !g.Battlefield.Contains(live.InstanceID) {
		return nil
	}
	return effects.UntilEndOfTurn(ctx, chosen.InstanceID, nil, "Dauntless Bodyguard — indestructible until end of turn", game.AddKeywordsMod("indestructible"))
}
