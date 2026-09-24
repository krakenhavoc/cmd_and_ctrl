package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice296c_helpers.go — shared bodies behind roadmap slice 296-c
// (the Equipment slice: Helm of the Host, Commander's Plate, The
// Reaver Cleaver, Brotherhood Regalia, Hammer of Nazahn, Buster
// Sword, Tarrian's Soulcleaver, Conqueror's Flail, Embercleave,
// Caduceus Staff of Hermes, Bloodforged Battle-Axe, Sting the
// Glinting Dagger, Illusionist's Bracers). Own file per the #231
// convention.

// attachedHostFor looks up the permanent an attachment (equipment or
// aura) with instance ID `sourceID` is currently attached to. Nil
// when the source can't be found or isn't attached to a card — a
// quiet no-op rather than an error, the same posture every other
// attachment read in this package takes.
func attachedHostFor(g *game.Game, sourceID uuid.UUID) *game.Card {
	src, ok := g.LookupCardForEffect(sourceID)
	if !ok {
		return nil
	}
	return g.AttachedHostOf(&src)
}

// dealtCombatDamageToPlayerOrPlaneswalkerByAttached is The Reaver
// Cleaver's trigger condition: "whenever THIS CREATURE [the equipped
// creature] deals combat damage to a player or planeswalker". The
// Swords' attachedCreatureDealtCombatDamageToPlayer widened to also
// match a planeswalker target, which none of the existing Swords
// need.
func dealtCombatDamageToPlayerOrPlaneswalkerByAttached(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if !source.IsAttachedTo(ev.Source) {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) != nil {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsPlaneswalker()
}

// anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield is
// Tarrian's Soulcleaver's trigger condition. "Another" excludes only
// the Soulcleaver's own instance, per the printed "another" — the
// equipped creature dying counts, which is the whole reason the card
// still grows when its own carrier trades in combat.
func anotherArtifactOrCreaturePutIntoGraveyardFromBattlefield(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return c.IsArtifact() || c.IsCreature()
}

// colorsAmongPermanentsControlledBy counts the distinct colors among
// the permanents `source`'s controller controls — Conqueror's Flail's
// "for each color among permanents you control". Colorless permanents
// contribute nothing, which is correct: a colorless-only board makes
// the Flail a plain +0/+0.
func colorsAmongPermanentsControlledBy(g *game.Game, source *game.Card) int {
	if g == nil || source == nil {
		return 0
	}
	seen := map[string]bool{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != source.Controller {
			continue
		}
		for _, color := range []string{"W", "U", "B", "R", "G"} {
			if !seen[color] && c.HasColor(color) {
				seen[color] = true
			}
		}
	}
	return len(seen)
}

// removeLegendaryFromTypeLine strips a leading "Legendary" supertype
// from a token template's type line — Helm of the Host's "except the
// token isn't legendary". A type line that never carried it is
// returned unchanged.
func removeLegendaryFromTypeLine(tl string) string {
	const prefix = "Legendary "
	if len(tl) > len(prefix) && tl[:len(prefix)] == prefix {
		return tl[len(prefix):]
	}
	return tl
}
