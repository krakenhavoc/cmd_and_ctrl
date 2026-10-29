package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eagle's Rescue — Enchantment — Aura {2}{W/U}{W/U} (The Hobbit):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying.
//	 {2}{W/U}{W/U}: Return this card from your graveyard to the battlefield
//	 attached to target creature you control with power 1 or less.
//	 Activate only as a sorcery."
//
// The two statics are Iona's Blessing's shape (PumpAttached, GrantToAttached
// over AttachedToSource). The recursion is a graveyard ability with a target
// of its own, returning itself with Mantle of the Ancients' return-attached
// (ReturnFromGraveyardTogether with AttachedTo). Both halves of CR 608.2b are
// honoured: the card must still be in the graveyard, and the creature must
// still be a legal target (power is read at resolution) and still able to be
// enchanted by this Aura (canBeAttachedTo), otherwise the card stays put.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "95e1c368-b58a-4e69-9583-587560ad558d",
		Name:         "Eagle's Rescue",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("flying"),
		},
		Activated: []ActivatedAbility{{
			Label:        "{2}{W/U}{W/U}: Return this card from your graveyard to the battlefield attached to target creature you control with power 1 or less. Activate only as a sorcery.",
			Cost:         ManaCost("{2}{W/U}{W/U}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature you control with power 1 or less", YouControl(), PowerLE(1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id := item.SourceCardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				host, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				card, ok := g.LookupCardForEffect(id)
				if !ok || !canBeAttachedTo(g, item.Controller, card, host) {
					return nil
				}
				return ReturnFromGraveyardTogether{
					Targets:    []uuid.UUID{id},
					Controller: item.Controller,
					AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
				}.Apply(ctx)
			},
		}},
	})
}
