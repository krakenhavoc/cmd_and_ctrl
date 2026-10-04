package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avatar Destiny — Enchantment — Aura {2}{G}{G}:
//
//	"Enchant creature you control
//	 Enchanted creature gets +1/+1 for each creature card in your
//	 graveyard and is an Avatar in addition to its other types.
//	 When enchanted creature dies, mill cards equal to its power.
//	 Return this card to its owner's hand and up to one creature card
//	 milled this way to the battlefield under your control."
//
// The bonus is Blanchwood Armor's PumpAttachedPer over the Aura
// controller's creature cards in graveyard, recounted on every layer
// pass; the Avatar subtype is an additive layer-4 change (CR 205.1b,
// not "is an Avatar", so nothing else is lost).
//
// The dies trigger reads the creature's power as it last existed on the
// battlefield (CR 608.2h), which includes this Aura's own bonus — it
// was still attached when the creature died. A negative power mills 0
// (CR 107.1b). "Milled this way" is MillToZone.Then's list: only cards
// that actually reached the graveyard. The creature card is picked as
// the ability resolves (not targeted); "up to one" lets the controller
// pick none.
//
// "This card" is the Aura in its owner's graveyard, where the state-based
// action put it once its creature died (CR 704.5m). It returns only if
// it is still that object (CR 400.7): one exiled or returned in response
// is left where it is, while the creature still comes back. The Aura
// goes to hand before the creature enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecb81888-589b-4486-a444-e0c26384854e",
		Name:         "Avatar Destiny",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return b11CreatureCardsInGraveyard(g, source.Controller)
			}),
			{
				Layer:     game.Layer4Type,
				AppliesTo: AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasType(c.Subtypes, "Avatar") {
						c.Subtypes = append(c.Subtypes, "Avatar")
					}
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Avatar Destiny — mill cards equal to its power, then return this card to hand and up to one milled creature card to the battlefield", avatarDestinyMillAndReturn),
		},
	})
}

func avatarDestinyMillAndReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	if info, ok := ctx.TriggeringPermanent(); ok && info.Power > 0 {
		n = info.Power
	}
	return MillToZone{N: n, Then: avatarDestinyChooseMilledCreature}.Apply(ctx)
}

// avatarDestinyChooseMilledCreature asks for up to one creature card
// from the ones this mill put into the graveyard, then returns both.
func avatarDestinyChooseMilledCreature(ctx *Context, milled []uuid.UUID) error {
	var creatures []uuid.UUID
	for _, id := range milled {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	item := ctx.Item
	if len(creatures) == 0 {
		return avatarDestinyReturn(ctx, uuid.Nil)
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Avatar Destiny — return up to one creature card milled this way to the battlefield",
		Cards:    creatures,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneGraveyard,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			pick := uuid.Nil
			if len(picked) > 0 {
				pick = picked[0]
			}
			return avatarDestinyReturn(NewContext(g, item), pick)
		},
	})
	return nil
}

// avatarDestinyReturn puts this card into its owner's hand (if it is
// still the Aura that fell off) and then `creature`, if any, onto the
// battlefield under the ability's controller.
func avatarDestinyReturn(ctx *Context, creature uuid.UUID) error {
	if sourceCardInGraveyardAfterLeaving(ctx) {
		if err := (ReturnFromGraveyard{Target: ctx.Item.SourceCardID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	if creature == uuid.Nil {
		return nil
	}
	return ReturnFromGraveyard{Target: creature, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}
