package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// returned_aura.go — a creature that "returns to the battlefield. It's an
// Aura enchantment with enchant <clause> and '<ability>'. It loses all
// other abilities" (#2809, ADR 0154's 2026-10-23 amendment). Harold and
// Bob, First Numens is the first card; Bronzehide Lion prints the same
// shape with another clause and grant.
//
// A card of this family declares three things and nothing else is
// per-card:
//
//   - Spec.ReturnedAura, the enchant clause of its Aura form
//     (EnchantLand(Subtype("Forest"), YouControl())). CR 303.4f asks it
//     of the controller as the card enters; CR 704.5m holds the Aura to
//     it afterwards.
//   - ReturnedAuraForm(grant) in Spec.Static, its Aura form's three
//     statics, and the AbilityGrant bundle `grant` names in Spec.Grants.
//   - WhenThisDiesReturnItAsAnAura as its dies trigger.
//
// The engine side is game/returned_aura.go: Card.ReturnedAura is set as
// the entry lands, before the host is attached and before EventETB, and
// cleared by every zone change.

// ReturnedAuraForm is the Aura form's statics, in the order they have to
// be declared in.
//
//  1. The grant "Enchanted <thing> has '<ability>'" (CR 113.10), which
//     the Aura itself has. Layer 6, to the permanent the Aura is attached
//     to. Declared BEFORE the removal: both are the one effect's
//     layer-6 half and share the object's timestamp (CR 613.7a), the
//     bucket's stable sort keeps declaration order, and a grant the
//     removal has not yet silenced is applied (CR 613.6,
//     ability_removal.go). ContinuesAfterRemoval makes it the
//     later-layer half of the effect that started in layer 4, so it is
//     nobody's victim.
//  2. "It's an Aura enchantment" (CR 205.1a, CR 613.1d): the card types
//     become Enchantment and the subtypes Aura, so the creature type and
//     every creature subtype are gone. Supertypes are not card types and
//     stay (CR 205.4b): a legendary creature returns a legendary Aura.
//  3. "Loses all other abilities" (CR 613.1f), applied to the card
//     itself: Reach, vigilance and the dies trigger. Also
//     ContinuesAfterRemoval, as LoseAllAbilities is for every printed
//     "loses all abilities and is a …".
//
// The type change and the removal apply only to the permanent whose
// Card.ReturnedAura is set, so the creature the card prints is untouched;
// the grant applies to whatever the Aura is attached to, which a creature
// never is.
func ReturnedAuraForm(grantKey string) []game.StaticAbility {
	self := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && target.InstanceID == source.InstanceID && source.ReturnedAura
	}
	grant := GrantAbilitiesToAttached(grantKey)
	grant.ContinuesAfterRemoval = true
	return []game.StaticAbility{
		grant,
		{
			Label:     "It's an Aura enchantment.",
			Layer:     game.Layer4Type,
			AppliesTo: self,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Types = []string{"Enchantment"}
				c.SetSubtypes([]string{"Aura"})
			},
		},
		{
			Label:                 "It loses all other abilities.",
			Layer:                 game.Layer6Ability,
			RemovesAbilities:      true,
			ContinuesAfterRemoval: true,
			AppliesTo:             self,
			Apply:                 func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		},
	}
}

// WhenThisDiesReturnItAsAnAura is the printed trigger: "When <this>
// dies, if it was a creature, return it to the battlefield. It's an Aura
// enchantment …". "If it was a creature" is CR 603.4's intervening-if,
// checked against the CR 603.10a last-known information, so the Aura
// that dies (its Forest was destroyed) finds it false and stays in the
// graveyard, which is also why the form loses the trigger.
//
// The return is ReturnFromGraveyard{AsAura}: the controller chooses what
// it enchants as it enters (CR 303.4f) and with no legal host the card
// stays in the graveyard (CR 303.4g). Nothing happens if something moved
// the card out of the graveyard in response (CR 608.2b).
func WhenThisDiesReturnItAsAnAura(name string) game.TriggeredAbility {
	return On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
		return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature")
	}, name+" — return it to the battlefield as an Aura",
		func(g *game.Game, item *game.StackItem) error {
			id := item.SourceCardID
			if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
				return nil
			}
			return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, AsAura: true}.Apply(NewContext(g, item))
		})
}
