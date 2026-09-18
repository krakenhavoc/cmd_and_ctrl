package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_rider_helpers.go — the shared vocabulary for the S22
// mana-ability-rider batch: the painland / Talisman / Ancient Tomb
// shape of "{T}: Add X. This permanent deals N damage to you."
//
// Before this batch a painland in a deck was worse than absent. The
// engine's synthetic land ability (ManaAbilitiesForCard →
// basicLandColor) only fires for lands with the BASIC supertype, and
// "Land" alone has no types to derive from — so an unregistered
// Shivan Reef sat on the battlefield producing nothing at all. Every
// card in this batch is registered here for that reason, not for
// completeness.

// PainRider is the post-production half of a painland-style mana
// ability: "This land deals N damage to you."
//
// Three things about it are load-bearing and easy to get wrong:
//
//   - It is DAMAGE, not life loss. It routes through
//     DealDamageToPlayerForEffect, which means the CR 614 replacement
//     pipeline sees it — a damage-prevention shield or a doubler
//     applies, and Platinum Emperion-style "your life total can't
//     change" would stop it. Life loss would dodge all of that.
//   - The source of the damage is the permanent itself ("THIS LAND
//     deals 1 damage"), not the controller, which is what a
//     source-reading effect will want.
//   - It is not a cost. The ability is activatable at 1 life, the
//     damage happens, and the player loses on the next SBA pass.
//     ActivateManaAbility runs that pass on the way out whenever a
//     rider fired, so the loss is immediate rather than deferred to
//     some later boundary.
func PainRider(n int) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		return g.DealDamageToPlayerForEffect(source, controller, n)
	}
}

// painDual is the second ability shared by all eight painlands and
// all six Talismans: "{T}: Add {A} or {B}. This permanent deals 1
// damage to you."
//
// `self` is the printed self-reference ("This land" / "This
// artifact") so the menu copy reads like the card.
//
// Pipe syntax for the colour pair, with NarrowToCommanderIdentity
// off: the printed text names two specific colours and says nothing
// about the command zone, so a Shivan Reef in a mono-blue deck still
// offers {R} (after {U}). (The narrowing is only for Arcane Signet
// and Command Tower and their kin, which do print the clause.)
func painDual(a, b, self string) ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{" + a + "|" + b + "}",
		Label:    "Add {" + a + "} or {" + b + "}. " + self + " deals 1 damage to you.",
		Rider:    PainRider(1),
	}
}

// painlessColorless is the FIRST ability on every painland and every
// Talisman: "{T}: Add {C}", with no rider at all.
//
// Keeping it at index 0 is deliberate and the auto-tapper depends on
// it: game.autoTapAbilityFor takes the first tap ability with no
// sacrifice cost, no life cost and no rider, so the planner reaches
// for the painless half and never spends the player's life without
// being asked. The colored half stays a deliberate click in the
// ability menu.
func painlessColorless() ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{C}",
		Label:    "Add {C}",
	}
}
