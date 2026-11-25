package game

import "github.com/google/uuid"

// aura_returns_as_enchantment.go — #6235: Takklemaggot.
//
//	"When enchanted creature dies, that creature's controller chooses a
//	 creature that this card could enchant. If the player does, return
//	 this card to the battlefield under your control attached to that
//	 creature. If they don't, return this card to the battlefield under
//	 your control as a non-Aura enchantment. It loses "enchant creature"
//	 and gains "At the beginning of that player's upkeep, this
//	 enchantment deals 1 damage to that player.""
//
// Two engine pieces, both data:
//
//   - CreaturesAuraCardCouldEnchantForEffect is "a creature that this
//     card could enchant": the battlefield creatures the Aura card, put
//     onto the battlefield under `controller`, could legally be attached
//     to (CR 303.4f's legal choice: the enchant clause, protection, "can't
//     be enchanted"). Choosing one targets nothing.
//   - EntersAs.LoseSubtypes / LoseKeywords / ChosenPlayer (entry_overwrite.go)
//     put the card onto the battlefield as a non-Aura enchantment that
//     has lost enchant and remembers "that player" (CR 205.1a set,
//     CR 613.1d and 613.1f; the effect is the returning ability's, pinned
//     to the new object, CR 400.7). An Aura entering that way enchants
//     nothing, so CR 303.4f/g never asks for a host or leaves it where it
//     was.

// CreaturesAuraCardCouldEnchantForEffect is every creature on the
// battlefield the Aura card `auraID` could enchant if it entered the
// battlefield under `controller`'s control now, in battlefield order.
// Nil for a card that is not an Aura with an enchant clause.
//
// Caller must hold g.mu (write: the layer cache may be refreshed).
func (g *Game) CreaturesAuraCardCouldEnchantForEffect(auraID, controller uuid.UUID) []uuid.UUID {
	z := g.findCardZoneLocked(auraID)
	if z == nil || z.Kind == ZoneBattlefield {
		return nil
	}
	aura, ok := g.cardInZoneLocked(z, auraID)
	if !ok || !aura.IsAura() {
		return nil
	}
	spec := TargetSpecFor(catalogKeyOf(&aura))
	if spec == nil {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	aura.Controller = controller
	ev := &ReplacementEvent{Kind: RepEventMove, Actor: controller, CardID: auraID, OldZone: z.Kind, NewZone: ZoneBattlefield,
		auraHostCreaturesOnly: true}
	var out []uuid.UUID
	for _, id := range g.auraHostCandidatesLocked(ev, aura, spec).Cards {
		if c := findBattlefieldCard(g, id); c != nil && c.IsCreature() {
			out = append(out, id)
		}
	}
	return out
}
