package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zask, Skittering Swarmlord — {3}{G}{G} Legendary Creature — Insect 5/5 (Jumpstart 2022, #5948):
//
//	"You may play lands and cast Insect spells from your graveyard.
//	 Whenever another Insect you control dies, put it on the bottom of its
//	 owner's library, then mill two cards.
//	 {1}{B/G}: Target Insect gets +1/+0 and gains deathtouch until end of turn."
//
// Two standing graveyard permissions (Haakon's Knight spells and Ramunap
// Excavator's lands). The dies trigger moves the card only while it is
// still the graveyard object that died (CR 400.7), and the mill happens
// whether or not there was a card to move (a token ceased to exist,
// CR 111.7). The activated ability's hybrid mana is paid with either
// colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c4966b9-26f9-46ff-8d41-7fd01c70986d",
		Name:         "Zask, Skittering Swarmlord",
		Completeness: effects.CompletenessFull,
		CastPermissions: []game.CastPermission{
			{
				Zone:   game.ZoneGraveyard,
				Filter: game.PermissionFilter{LandsOnly: true},
				Label:  "Play a land from your graveyard (Zask, Skittering Swarmlord)",
			},
			{
				Zone:   game.ZoneGraveyard,
				Filter: game.PermissionFilter{Subtype: "Insect"},
				Label:  "Cast Insect spells from your graveyard (Zask, Skittering Swarmlord)",
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				if !ok || ev.CardID == source.InstanceID || effects.LeftUnderControlOf(ev, dead) != source.Controller {
					return false
				}
				return effects.LeftAsSubtype(ev, dead, source.TextSubtype("Insect"))
			}, "Zask, Skittering Swarmlord — put it on the bottom of its owner's library, then mill two cards",
				zaskTuckAndMill),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B/G}: Target Insect gets +1/+0 and gains deathtouch until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "deathtouch")),
			Cost:    effects.ManaCost("{1}{B/G}"),
			Targets: effects.TargetCreature("target Insect", effects.Subtype("Insect")),
			Effect:  zaskPump,
		}},
	})
}

func zaskTuckAndMill(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if item.Trigger != nil {
		id := item.Trigger.Event.CardID
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard && zaskSameGraveyardObject(z, ctx, id) {
			if err := (effects.PutIntoLibrary{Card: id, ToBottom: true}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return effects.MillCards{Player: item.Controller, N: 2}.Apply(ctx)
}

// zaskSameGraveyardObject is sameGraveyardObject for the creature the
// trigger names rather than its source: the card in the graveyard is the
// object that died when its epoch is one past the battlefield epoch the
// trigger recorded for it.
func zaskSameGraveyardObject(z *game.Zone, ctx *effects.Context, id uuid.UUID) bool {
	obj := ctx.Trigger().Object
	if obj == nil || obj.Ref().ID != id {
		return true
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			return z.Cards[i].ObjectEpoch == obj.Ref().Epoch+1
		}
	}
	return false
}

func zaskPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: id, Power: 1, Label: "Zask, Skittering Swarmlord — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"deathtouch"}, Label: "Zask, Skittering Swarmlord — deathtouch"}.Apply(ctx)
}
