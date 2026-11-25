package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b100_helpers.go — Jumpstart Auras (S82 older sets, batch 100).

// S82OldB100AnimateAttachedLand is the Zendikon animation with a colour,
// "Enchanted land is a P/T <colour> <subtype> creature. It's still a land."
// The land gains the Creature type and the subtype in layer 4, the colour in
// layer 5 and the base power and toughness in layer 7b, all kept while the
// Aura is on it (CR 613.6). The colour and the subtype are words of the
// Aura's text (CR 612.2).
func S82OldB100AnimateAttachedLand(color, subtype string, power, toughness int) []game.StaticAbility {
	typed := game.StaticAbility{
		Label:                 "Enchanted land is a " + subtype + " creature. It's still a land.",
		Layer:                 game.Layer4Type,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if !TypeListHasEffects(c.Types, "Creature") {
				c.Types = append(c.Types, "Creature")
			}
			for _, s := range source.TextSubtypes(subtype) {
				if !TypeListHasEffects(c.Subtypes, s) {
					c.Subtypes = append(c.Subtypes, s)
				}
			}
		},
	}
	return []game.StaticAbility{typed, SetAttachedColors(color), SetAttachedBasePT(power, toughness)}
}

// S82OldB100ReturnTheDiedCardToHand is "return that card to its owner's hand"
// for a dies trigger on the enchanted permanent: the card the dies event
// named comes back from the graveyard, and a card that has left the
// graveyard by then stays where it is (CR 400.7).
func S82OldB100ReturnTheDiedCardToHand(g *game.Game, item *game.StackItem) error {
	return ReturnFromGraveyard{Target: item.Trigger.Event.CardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}

// S82OldB100AnotherAuraAttachedTo is "another Aura is attached to enchanted
// creature": an Aura other than the source is attached to the host, read as
// the board stands in this layer pass (CR 303.4).
func S82OldB100AnotherAuraAttachedTo(host *game.Card, g *game.Game, source *game.Card) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.InstanceID != source.InstanceID && a.IsAura() && a.IsAttachedTo(host.InstanceID) {
			return true
		}
	}
	return false
}

// S82OldB100CreatureCardsInControllersGraveyard counts the creature cards in
// the graveyard of the player who controls `host`.
func S82OldB100CreatureCardsInControllersGraveyard(g *game.Game, host *game.Card) int {
	p := g.PlayerByIDForEffect(host.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			n++
		}
	}
	return n
}

// S82OldB100EnchantedUntapped is "enchanted creature is untapped" for an
// Aura's trigger condition: the Aura is attached to a permanent that is not
// tapped.
func S82OldB100EnchantedUntapped(g *game.Game, aura *game.Card) bool {
	if aura == nil || aura.AttachedTo.Kind != game.TargetCard {
		return false
	}
	host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	return ok && !host.Tapped
}

// S82OldB100TapEnchanted is the effect body of "tap enchanted creature".
func S82OldB100TapEnchanted(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := EnchantedCreatureOf(ctx)
	if host == uuid.Nil {
		return nil
	}
	return TapTarget{Target: host}.Apply(ctx)
}

// S82OldB100OpponentsCreatureDied is "whenever a creature an opponent
// controls dies" for an ability the source's controller owns: a creature
// that another player controlled as it died (CR 603.10a).
func S82OldB100OpponentsCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := DiedCreature(ev, g)
	if !ok {
		return false
	}
	owner := LeftUnderControlOf(ev, dead)
	return owner != uuid.Nil && owner != source.Controller
}
