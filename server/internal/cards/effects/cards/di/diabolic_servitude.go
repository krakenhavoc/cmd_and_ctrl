package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diabolic Servitude — Enchantment {3}{B} (Commander 2015, S82 OLD #5546):
//
//	"When this enchantment enters, return target creature card from your
//	 graveyard to the battlefield.
//	 When the creature put onto the battlefield with this enchantment dies,
//	 exile it and return this enchantment to its owner's hand.
//	 When this enchantment leaves the battlefield, exile the creature put
//	 onto the battlefield with this enchantment."
//
// Three linked abilities (CR 607.2e). The enters trigger returns the target
// and records it as the Servitude's linked permanent, as the object it
// became (Havengul Mystery's link). The second watches that object die: it
// exiles the card from the graveyard (if it is still there) and bounces the
// Servitude. The third reads the link as the Servitude leaves
// (game.LinkedPermanentForEffect, last-known information once it has gone)
// and exiles the creature only while it is still that same object on the
// battlefield, so a creature that was already exiled by the second ability, or
// that has left and returned as a new object, is left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e275cbfb-3d25-4189-84f0-3162eadc7d65",
		Name:         "Diabolic Servitude",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Diabolic Servitude — return target creature card from your graveyard to the battlefield", diabolicServitudeReturn),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
			effects.On(game.EventLTB, diabolicServitudeCreatureDied,
				"Diabolic Servitude — the creature put onto the battlefield with it died: exile it and return the Servitude to its owner's hand",
				diabolicServitudeFallBack),
			effects.On(game.EventLTB, effects.Self,
				"Diabolic Servitude — it left the battlefield: exile the creature put onto the battlefield with it",
				diabolicServitudeExileCreature),
		},
	})
}

// diabolicServitudeReturn returns the target and records it as the linked
// permanent, as the new object it became.
func diabolicServitudeReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self, hasSelf := ctx.SourceRef()
	if hasSelf {
		g.SetLinkedPermanentForEffect(self.ID, game.ObjectRef{})
	}
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	back := effects.ReturnFromGraveyard{Target: target, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}
	if err := back.Apply(ctx); err != nil {
		return err
	}
	if c, found := g.LookupCardForEffect(target); found && hasSelf && effects.OnBattlefield(g, target) {
		g.SetLinkedPermanentForEffect(self.ID, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
	}
	return nil
}

// diabolicServitudeCreatureDied is "the creature put onto the battlefield with
// this enchantment dies": the linked permanent went to a graveyard.
func diabolicServitudeCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	l := source.LinkedPermanent
	return l.ID != uuid.Nil && ev.CardID == l.ID && ev.NewZone == game.ZoneGraveyard
}

// diabolicServitudeFallBack returns the Servitude to its owner's hand and
// forgets the link; the hand move is a keyed bounce, so exiling the dead
// creature is its tail, told which card it was.
func diabolicServitudeFallBack(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	card, ok := g.LookupCardForEffect(self.ID)
	if !ok {
		return nil
	}
	exileDead := effects.WithCarry(ctx.Tail(diabolicServitudeExileDead), card.LinkedPermanent.ID)
	g.SetLinkedPermanentForEffect(self.ID, game.ObjectRef{})
	if !effects.OnBattlefield(g, self.ID) {
		return exileDead.Run(g, nil)
	}
	return g.BounceCardsToHandKeyedForEffect([]uuid.UUID{self.ID}, exileDead)
}

// diabolicServitudeExileDead is "exile it": the creature that died, if it is
// still in the graveyard. An on-disk key: never renamed, never reused.
var diabolicServitudeExileDead = effects.CardTail("tail/diabolic-servitude-exile-dead", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, dead := range o.Tail.Carry {
		if z := ctx.Game.FindCardZoneForEffect(dead); dead != uuid.Nil && z != nil && z.Kind == game.ZoneGraveyard {
			return effects.ExileTarget{Target: dead}.Apply(ctx)
		}
	}
	return nil
})

// diabolicServitudeExileCreature exiles the creature this Servitude put onto
// the battlefield, if it is still that object there.
func diabolicServitudeExileCreature(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	self, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	linked := g.LinkedPermanentForEffect(self)
	if linked == nil {
		return nil
	}
	return effects.ExileTarget{Target: linked.InstanceID}.Apply(ctx)
}
