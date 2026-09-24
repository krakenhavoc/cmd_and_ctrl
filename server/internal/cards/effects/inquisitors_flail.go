package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inquisitor's Flail — Artifact — Equipment {2} (EDHREC rank 4135):
//
//	"If equipped creature would deal combat damage, it deals double
//	 that damage instead.
//	 If another creature would deal combat damage to equipped
//	 creature, it deals double that damage to equipped creature
//	 instead.
//	 Equip {2}"
//
// Four mana all in for a damage doubler on one creature, and the
// symmetry is the card: the Flail does not protect, it escalates.
// On a trampler or a commander it halves the clock; on a blocker it
// is a liability, because the creature that blocks it hits back for
// double too. It goes in the deck that wants combat to end quickly
// in either direction.
//
// TWO REPLACEMENTS, NOT ONE, and they are deliberately asymmetric in
// scope:
//
//   - The first is OUTBOUND and unrestricted in target: equipped
//     creature's combat damage to anything — a player, a
//     planeswalker, a blocker, a battle — is doubled.
//   - The second is INBOUND and narrow: only damage from ANOTHER
//     CREATURE, and only damage dealt TO the equipped creature. A
//     Lightning Bolt is not doubled; neither is a Blasphemous Act;
//     neither is the equipped creature's own damage to itself. That
//     is why the second clause carries both an IsCombatDamage check
//     and a source-is-a-creature check that excludes the host.
//
// BOTH CLAUSES ARE COMBAT-DAMAGE ONLY. A Flailed creature that pings
// with an activated ability does not double it.
//
// AN UNBLOCKED CREATURE FIGHTING A FLAILED BLOCKER GETS BOTH: the
// attacker's damage to the blocker is doubled by the second clause
// and the blocker's damage back is doubled by the first. Applying
// both to one exchange is correct — they are different events, and
// each is replaced exactly once.
//
// The engine looks the damage source up wherever it now is: combat
// damage is dealt before state-based actions run, so a creature that
// traded is still on the battlefield when its damage is replaced.
// With a second doubler on the board the affected player orders them
// (CR 616.1) and ×4 is ×4 either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a89ae357-b5aa-4256-beb3-a2e5e7f43200",
		Name:         "Inquisitor's Flail",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return b39IsCombatDamage(ev) && src.IsAttachedTo(ev.DamageSource)
				},
				Replace:    b39DoubleTheDamage,
				Controller: b39SourceController,
				Label:      "Inquisitor's Flail: double the equipped creature's combat damage",
			},
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					if !b39IsCombatDamage(ev) || !src.IsAttachedTo(ev.DamageTarget) {
						return false
					}
					// "ANOTHER creature": the dealer must be a
					// creature, and must not be the equipped
					// creature itself.
					if ev.DamageSource == ev.DamageTarget {
						return false
					}
					// #1430, CR 608.2h: the dealer's type is read
					// from the event's last-known characteristics,
					// not a current-zone lookup — a creature that
					// dealt its damage and then left (or was
					// type-changed on the way out) is still judged
					// by what it was when it dealt the damage.
					ch, ok := damageSourceCharacteristics(ev, g)
					return ok && hasFold(ch.Types, "Creature")
				},
				Replace:    b39DoubleTheDamage,
				Controller: b39SourceController,
				Label:      "Inquisitor's Flail: double the combat damage dealt to the equipped creature",
			},
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// b39IsCombatDamage is the guard both of the Flail's clauses open
// with: a real combat-damage event with a positive amount and a
// source to attribute it to.
func b39IsCombatDamage(ev *game.ReplacementEvent) bool {
	return ev.Kind == game.RepEventDamage && ev.IsCombatDamage &&
		ev.DamageAmount > 0 && ev.DamageSource != uuid.Nil
}

// b39DoubleTheDamage is the "it deals double that damage instead"
// body (CR 614.1b), shared by both of the Flail's clauses.
func b39DoubleTheDamage(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
	ev.DamageAmount *= 2
	return nil
}

// b39SourceController names the replacement's controller for the
// CR 616.1 ordering prompt: whoever controls the Equipment.
func b39SourceController(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
	return src.Controller
}
