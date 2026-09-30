package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mantle of the Ancients — Enchantment — Aura {3}{W}{W}:
//
//	"Enchant creature you control
//	 When this Aura enters, return any number of target Aura and/or
//	 Equipment cards from your graveyard to the battlefield attached
//	 to enchanted creature.
//	 Enchanted creature gets +1/+1 for each Aura and Equipment
//	 attached to it."
//
// The trigger's targets are cards in the controller's graveyard, any
// number of them. Each returns under its owner's control and is
// attached to the enchanted creature by the same resolution, so the
// CR 704.5m / 704.5n checks never see it unattached.
//
// An Aura that could not enchant that creature is not returned
// (CR 303.4j: it stays where it is) — the same enchant predicate the
// Aura's own cast used decides, read without targeting, so a
// hexproof creature is still a legal host. Equipment needs only a
// creature. A target that left the graveyard in response is skipped
// (CR 608.2b).
//
// The bonus counts the Auras and Equipment on the host, this one
// included, whoever controls them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7424560f-557f-4bc9-a3e7-eb890c73aaa3",
		Name:         "Mantle of the Ancients",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, aurasAndEquipmentOnTheHost),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Mantle of the Ancients — return Auras and Equipment attached to enchanted creature",
				mantleOfTheAncientsReturn),
				TargetCardInGraveyard("any number of target Aura and/or Equipment cards in your graveyard",
					YouOwn(), Or(HasSubtype("Aura"), HasSubtype("Equipment"))).WithCount(0, 0)),
		},
	})
}

// mantleOfTheAncientsReturn returns each still-legal target to the
// battlefield and attaches it to the enchanted creature.
func mantleOfTheAncientsReturn(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	hostRef := game.TargetRef{Kind: game.TargetCard, ID: host.InstanceID}
	for _, t := range NewContext(g, item).LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		card, ok := g.LookupCardForEffect(t.ID)
		if !ok || !canBeAttachedTo(g, item.Controller, card, host.InstanceID) {
			continue
		}
		back, err := g.ReturnToBattlefieldForEffect(t.ID, uuid.Nil, false)
		if errors.Is(err, game.ErrCardNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if back == uuid.Nil {
			back = t.ID
		}
		if err := g.AttachForEffect(back, hostRef); err != nil {
			return err
		}
	}
	return nil
}

// canBeAttachedTo reports whether `attachment` (a card about to enter)
// may be attached to the battlefield permanent `host` when an effect
// puts it there (CR 303.4j, CR 301.5): an Aura needs the host to meet
// its enchant clause, an Equipment needs a creature. An Aura with no
// catalogued enchant clause is a manual object in this sandbox and is
// allowed.
//
// Read without the protection-style targeting gate: CR 303.4j asks
// whether the Aura could enchant the object, not whether it could
// target it. Caller must hold g.mu.
func canBeAttachedTo(g *game.Game, controller uuid.UUID, attachment game.Card, host uuid.UUID) bool {
	hostCard, ok := g.LookupCardForEffect(host)
	if !ok {
		return false
	}
	if attachment.HasSubtype("Aura") {
		spec := game.TargetSpecFor(game.CatalogKey(attachment))
		if spec == nil {
			return true
		}
		for _, id := range g.SpecCandidatesForEffect(controller, spec).Cards {
			if id == host {
				return true
			}
		}
		return false
	}
	return hostCard.IsCreature()
}
