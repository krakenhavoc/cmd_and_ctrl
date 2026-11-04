package effects

import (
	"fmt"

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

// EachOpponentGainsLifeRider is the mana-ability rider "Each opponent
// gains N life" (Grove of the Burnwillows): every other seat still in the
// game gains n life once the mana is in the pool (CR 605.3b). It is life
// GAIN through ChangePlayerLifeForEffect, so each opponent's replacements
// and "whenever you gain life" payoffs see it, which is the whole
// drawback of the card. Declare it with ManaAbility.OpponentsGainLife,
// which also tells the auto-tapper what it does (#4189).
func EachOpponentGainsLifeRider(n int) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == controller {
				continue
			}
			if err := g.ChangePlayerLifeForEffect(source, p.ID, n); err != nil {
				return err
			}
		}
		return nil
	}
}

// manaActivationTriggers grows the Triggered rows the mana abilities'
// PingOpponentsTrigger declarations stand for (#4928), in ability order.
// Nil when none declares one, which is every card but a handful.
func manaActivationTriggers(name string, abilities []ManaAbility) []game.TriggeredAbility {
	var out []game.TriggeredAbility
	for _, a := range abilities {
		if a.PingOpponentsTrigger > 0 {
			out = append(out, manaActivationPing(name, a.Label, a.PingOpponentsTrigger))
		}
	}
	return out
}

// manaActivationPing is "Whenever you tap this creature for mana, it
// deals N damage to each opponent" (Zhur-Taa Druid) and its reflexive
// twin "… Add {R}. When you do, this creature deals N damage to each
// opponent" (Rubble Rouser, CR 603.12): a triggered ability that fires
// when the row labelled `label` is activated, and goes on the stack.
//
// CR 605.1b is why it is not a rider: a triggered ability is a mana
// ability only if "it could add mana to a player's mana pool when it
// resolves", and damage cannot. The mana ability itself still resolves
// at once (CR 605.3b), so the mana is in the pool before the trigger is
// even put on the stack (CR 603.3, the next time a player would receive
// priority).
//
// The watch is game.EventManaAbilityActivated, which the hand-clicked
// activation and the auto-tapper both emit with the row's label — not
// EventTapCard, which also fires for an attack, a crew or an opponent's
// tap effect. "You" is the activator: the source's controller.
func manaActivationPing(name, label string, n int) game.TriggeredAbility {
	return On(game.EventManaAbilityActivated,
		func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return source != nil && ev.Source == source.InstanceID && ev.Label == label && ev.Actor == source.Controller
		},
		fmt.Sprintf("%s — %d damage to each opponent", name, n),
		Do(DamageEachOpponent{N: n}))
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
		Cost:      ManaAbilityCost{Tap: true},
		Produced:  "{" + a + "|" + b + "}",
		Label:     "Add {" + a + "} or {" + b + "}. " + self + " deals 1 damage to you.",
		PainToYou: 1,
	}
}

// painlessColorless is the FIRST ability on every painland and every
// Talisman: "{T}: Add {C}", with no rider at all.
//
// Keeping it at index 0 is deliberate: game.autoTapAbilityFor, the
// pick for a plan entry that names no ability, takes the first
// acceptable one. The planner offers both halves since #2392 (the
// painful one through its PainToYou declaration), and orders the
// coloured half in its pain tier, after every painless source.
func painlessColorless() ManaAbility {
	return ManaAbility{
		Cost:     ManaAbilityCost{Tap: true},
		Produced: "{C}",
		Label:    "Add {C}",
	}
}
