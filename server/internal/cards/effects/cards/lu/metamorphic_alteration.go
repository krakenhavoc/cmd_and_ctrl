package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Metamorphic Alteration — Enchantment — Aura {1}{U} (Core Set 2019, #5527):
//
//	"Enchant creature
//	 As this Aura enters, choose a creature.
//	 Enchanted creature is a copy of the chosen creature."
//
// Dinosaur Headdress's copy for as long as the attachment stays on the
// creature (CR 707.2, 611.2b, game.WhileAttachedToDuration), with Scheming
// Fence's "as this enters, choose a permanent" stored on the Aura
// (Card.ChosenPermanent). The choice is not optional: the prompt takes
// exactly one of the creatures on the battlefield (any, including the one it
// enchants). It is asked as the Aura enters, so the copy begins when the
// answer arrives, or, when the Aura becomes attached after that, at the attach
// (AsAttached, which also starts a new copy on the creature a moved Aura now
// enchants). The copy holds the chosen creature's copiable values as they
// were when it began and keeps them if the chosen creature leaves; counters
// and other effects on the enchanted creature are not touched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2518a4f2-607f-41bb-9389-2009138d6b09",
		Name:         "Metamorphic Alteration",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			var ids []uuid.UUID
			var epochs []int
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsCreature() {
					ids = append(ids, c.InstanceID)
					epochs = append(epochs, c.ObjectEpoch)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:   card.Controller,
				Source:    card.InstanceID,
				Question:  "Metamorphic Alteration — choose a creature; the enchanted creature is a copy of it",
				Cards:     ids,
				Min:       1,
				Max:       1,
				LiveZone:  game.ZoneBattlefield,
				ThenKey:   metamorphicAlterationChosen,
				Carry:     append([]uuid.UUID(nil), ids...),
				CarryNums: epochs,
			})
			return nil
		},
		AsAttached: func(card *game.Card, ctx *effects.Context) error {
			chosen := ctx.Game.ChosenPermanentOf(card)
			if chosen == nil || card.AttachedTo.Kind != game.TargetCard {
				return nil
			}
			metamorphicAlterationCopy(ctx.Game, card.InstanceID, card.AttachedTo.ID, chosen.InstanceID)
			return nil
		},
	})
}

// metamorphicAlterationCopy makes host a copy of the chosen creature for as
// long as the Aura stays on it.
func metamorphicAlterationCopy(g *game.Game, aura, host, chosen uuid.UUID) {
	d, ok := g.WhileAttachedToDuration(aura, host)
	if !ok {
		return
	}
	v, ok := g.CopiableValuesForEffect(chosen)
	if !ok {
		return
	}
	g.BecomeCopyForEffect(aura, chosen, []uuid.UUID{host}, v, d, "Metamorphic Alteration")
}

// metamorphicAlterationChosen stores the answer as the Aura's chosen
// permanent (the object it was offered as) and starts the copy when the Aura
// is already attached. An on-disk key.
var metamorphicAlterationChosen = game.RegisterChooseCardsThen("choose-cards/metamorphic-alteration-chosen", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	picked := r.Picked[0]
	for i, id := range r.Carry {
		if id != picked || i >= len(r.Nums) {
			continue
		}
		c, ok := g.LookupCardForEffect(picked)
		if !ok || c.ObjectEpoch != r.Nums[i] {
			return nil
		}
		g.SetChosenPermanentForEffect(r.Source, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
		if aura, ok := g.LookupCardForEffect(r.Source); ok && aura.AttachedTo.Kind == game.TargetCard && aura.AttachedTo.ID != uuid.Nil {
			metamorphicAlterationCopy(g, r.Source, aura.AttachedTo.ID, picked)
		}
		return nil
	}
	return nil
})
