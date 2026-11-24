package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master Skald — Creature — Dwarf Warrior {4}{W}, 4/4 (Kaldheim):
//
//	"When this creature enters, you may exile a creature card from your
//	 graveyard. If you do, return target artifact or enchantment card from
//	 your graveyard to your hand."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d), so the
// trigger is removed when the graveyard holds no artifact or enchantment card
// to name. As it resolves the controller picks zero or one creature card of
// their graveyard to exile; only if a card really reached exile does the
// target come back to hand (and a target that was itself the exiled card is
// gone, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34bccfc0-453d-4d1a-8046-70f9c3065e88",
		Name:         "Master Skald",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Master Skald — you may exile a creature card from your graveyard; if you do, return target artifact or enchantment card from your graveyard to your hand",
					masterSkaldOffer),
				effects.TargetCardInGraveyard("target artifact or enchantment card from your graveyard", effects.YouOwn(), effects.Or(effects.Artifact(), effects.Enchantment()))),
		},
	})
}

// masterSkaldOffer asks for up to one creature card of the controller's
// graveyard.
func masterSkaldOffer(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	me := ctx.Controller()
	var cards []uuid.UUID
	if p := g.PlayerByIDForEffect(me); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.HasCardType("creature") {
				cards = append(cards, c.InstanceID)
			}
		}
	}
	if len(cards) == 0 {
		return nil
	}
	_, err := g.ChooseCardsRunThenKeyedForEffect([]game.ChooseCardsPrompt{{
		Chooser:    me,
		FromPlayer: me,
		Source:     ctx.Source(),
		Question:   "Master Skald — you may exile a creature card from your graveyard to return an artifact or enchantment card",
		Cards:      cards,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneGraveyard,
	}}, masterSkaldExiled, ctx.Source(), ctx.RunData([]uuid.UUID{me}))
	return err
}

// masterSkaldExiled exiles the card picked, if any. An on-disk key.
var masterSkaldExiled = effects.RunThen("prompt-run/s82-old-b062-master-skald-exile", func(ctx *effects.Context, r game.RunDone) error {
	var chosen []uuid.UUID
	for _, player := range r.Carry {
		chosen = append(chosen, r.Picks().By(player)...)
	}
	if len(chosen) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(chosen, ctx.ItemTail(masterSkaldAfterExile))
})

// masterSkaldAfterExile returns the target once a card really reached exile.
// An on-disk key.
var masterSkaldAfterExile = effects.CardTail("then/s82-old-b062-master-skald-return", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		}
	}
	return nil
})
