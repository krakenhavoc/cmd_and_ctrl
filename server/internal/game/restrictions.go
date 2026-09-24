package game

import "github.com/google/uuid"

// restrictions.go is the S24 restriction vocabulary: the engine's
// answer to "enchanted creature can't attack or block" (Pacifism),
// "equipped creature can't be blocked" (Whispersilk Cloak), "this
// creature can't block" (Carrion Feeder) and "its activated
// abilities can't be activated" (Arrest, Faith's Fetters).
//
// # Why a field of its own and not a keyword
//
// Every continuous effect the catalog could express before this
// file either changed a characteristic (P/T, types, colours, name)
// or granted a keyword. A restriction is neither. "Can't attack" is
// not an ability the creature has — it is an effect the AURA has,
// which is why CR 613 gives restrictions no layer at all, and why
// "enchanted creature loses all abilities" does not switch Pacifism
// off. Modelling it as a string in Characteristic.Abilities would
// have got that backwards, rendered a badge for a non-keyword, and
// put the restriction one layer-6 ability-strip away from vanishing.
//
// So restrictions ride in their own field on Characteristic, written
// by ordinary StaticAbility.Apply funcs (the layer pass is the only
// vehicle that knows which permanents an effect applies to) and
// OR-accumulated. Because the field is only ever OR'd into and no
// layer clears it, the layer a restriction is applied in does not
// matter and timestamp order does not matter — which is exactly
// CR 613's position on effects that are not applied in a layer.
//
// # The taxonomy, and why these five bits
//
// The shapes printed on real cards are not interchangeable and the
// difference is WHO is restricted, not what the text says:
//
//	"can't attack"            CantAttack       the creature
//	"can't block"             CantBlock        the creature
//	"can't be blocked"        CantBeBlocked    the DEFENDER's options
//	"activated abilities
//	 can't be activated"      CantActivate     non-mana activations
//	 (…unless they're
//	  mana abilities)         CantActivateMana mana activations
//
// CantBeBlocked lives on the attacker because that is the permanent
// the effect is attached to, but it is consumed on the defender's
// side of the table — Game.BlockPairRefusalLocked is the one
// predicate that sees both
// cards, so that is where it is read.
//
// The activation pair is two bits rather than one because Faith's
// Fetters prints the difference: Arrest stops mana abilities too,
// Fetters explicitly does not. Two bits also map one-to-one onto the
// engine's two activation entry points, so neither gate has to know
// about the other's card text.
//
// # What this deliberately does not express, and how it would
//
//   - "can't attack unless its controller pays {2}" (Propaganda,
//     Ghostly Prison, Norn's Annex) is a COST to attack, not a
//     prohibition, and it is SHIPPED — in attack_tax.go, not here
//     (ADR 0080, #1063). A bit cannot carry a cost, and the shape it
//     wanted was the one this note predicted: an attack-cost pipeline
//     keyed on the DEFENDING player rather than on the attacker,
//     consulted by the declaration verbs at CR 508.1a. Nothing about
//     the five bits changed to make room for it.
//   - "no more than two creatures can attack you each combat"
//     (Crawlspace) and "no more than one creature can attack each
//     combat and no more than one creature can block each combat"
//     (Silent Arbiter) are COUNT restrictions over the whole
//     declaration, not over one permanent, and they are SHIPPED
//     without disturbing anything in this file (#1507): the block
//     half is BlockRule.Limit, one more set check in
//     checkBlockDeclarationLocked, and the attack half is
//     game.AttackLimit (attack_limits.go), judged by both attack
//     declaration verbs and the enumerator. ADR 0045 Decisions 43-45.
//   - Goad is not a restriction at all: CR 701.15b makes it two
//     REQUIREMENTS ("attacks each combat if able and attacks a player
//     other than the goading player if able"), judged over the whole
//     declaration since #1571 (attack_requirements.go). Landwalk's "can't be blocked as long as
//     defending player controls an Island" is the same shape on the
//     block side, and since #705 it lives where ADR 0045's addendum
//     put it: a keyword read by Game.BlockPairRefusalLocked, which
//     sees the defending player's lands (landwalk.go).
//
// # When restrictions are checked
//
// At DECLARATION, and only there (CR 508.1c, 509.1b/c). A creature
// that becomes pacified after attackers are declared stays an
// attacking creature — CR 506.4 removes a permanent from combat when
// it leaves the battlefield, changes control or stops being a
// creature, and "acquired a restriction" is not on that list. So
// nothing in this file runs during the damage step, and nothing
// retroactively un-declares an attack.

// Restriction is a set of declaration-time and activation-time
// prohibitions on one permanent, as a bitmask. The zero value is
// "no restrictions", which is what every permanent the catalog says
// nothing about carries.
type Restriction uint8

const (
	// CantAttack is "~ can't attack" (CR 508.1c). Checked when an
	// attacker is declared; a creature that acquires it while
	// already attacking keeps attacking.
	CantAttack Restriction = 1 << iota

	// CantBlock is "~ can't block" (CR 509.1b). Checked when a
	// blocker is declared.
	CantBlock

	// CantBeBlocked is "~ can't be blocked" (CR 509.1b) — a
	// restriction on the DEFENDING player's legal blocks, carried on
	// the attacker because that is the permanent the effect is
	// attached to.
	CantBeBlocked

	// CantActivate is "its activated abilities can't be activated"
	// for everything that is not a mana ability (CR 602.5),
	// including loyalty abilities, which are activated abilities
	// (CR 606.1).
	CantActivate

	// CantActivateMana is the same prohibition extended to mana
	// abilities. Split from CantActivate because Faith's Fetters
	// prints the difference — "unless they're mana abilities" — and
	// because the engine activates the two through different entry
	// points.
	CantActivateMana
)

// CantAttackOrBlock is the pair every "can't attack or block" card
// prints together. Spelled as a named constant because writing the
// two bits out on each card invites writing one of them.
const CantAttackOrBlock = CantAttack | CantBlock

// Has reports whether every bit in `want` is set. A zero `want` is
// vacuously true, which no caller should be asking.
func (r Restriction) Has(want Restriction) bool {
	return r&want == want
}

// Names renders the set as stable snake_case tokens for the wire, in
// declaration order. Returns nil for the empty set so the JSON field
// can be omitempty.
//
// These strings are protocol, not debug output: the client reads
// them to disable an "attack with all" entry and to grey out an
// ability row with an honest reason. Renaming one is a wire change.
func (r Restriction) Names() []string {
	var out []string
	for _, e := range []struct {
		bit  Restriction
		name string
	}{
		{CantAttack, "cant_attack"},
		{CantBlock, "cant_block"},
		{CantBeBlocked, "cant_be_blocked"},
		{CantActivate, "cant_activate"},
		{CantActivateMana, "cant_activate_mana"},
	} {
		if r&e.bit != 0 {
			out = append(out, e.name)
		}
	}
	return out
}

// RestrictionsOn returns the restriction set currently applying to a
// permanent, post-layers. nil is the empty set so callers can skip
// defensive nil checks.
//
// Reads the EFFECTIVE characteristic, so the caller must have
// recomputed layers (every declaration path already does —
// BlockPairRefusalLocked
// has the same requirement for keywords).
func RestrictionsOn(c *Card) Restriction {
	if c == nil {
		return 0
	}
	return c.Effective().Restrictions
}

// Restricted reports whether `want` is among the restrictions on this
// permanent. The one read every gate in the engine goes through.
func Restricted(c *Card, want Restriction) bool {
	return RestrictionsOn(c).Has(want)
}

// CanAttack reports whether this permanent may be declared as an
// attacker under the RESTRICTION vocabulary alone (CR 508.1c).
//
// Deliberately narrow: tapped state, summoning sickness (CR 302.6)
// and defender (CR 702.3) are separate gates with their own errors,
// and the single-card DeclareAttacker verb is intentionally lax about
// some of them for sandbox hand-forcing. AttackerEligible is the
// predicate that folds all of them together for callers that want the
// strict answer.
func CanAttack(c *Card) bool {
	return !Restricted(c, CantAttack)
}

// CanActivateAbilities reports whether this permanent's non-mana
// activated abilities may be activated (Arrest, Faith's Fetters).
func CanActivateAbilities(c *Card) bool {
	return !Restricted(c, CantActivate)
}

// CanActivateManaAbilities reports whether this permanent's mana
// abilities may be activated. Separate from CanActivateAbilities
// because Faith's Fetters stops one and not the other.
func CanActivateManaAbilities(c *Card) bool {
	return !Restricted(c, CantActivateMana)
}

// AttackerEligible reports whether card `a` could be declared as an
// attacker by `seat` right now, ignoring which defender it would be
// pointed at: a creature that seat controls, untapped (CR 508.1a),
// not already attacking, without defender (CR 702.3), not summoning
// sick (CR 302.6), and under no "can't attack" restriction
// (CR 508.1c).
//
// The mirror of BlockerEligible, and it exists for the same reason
// that one does (#328) — and, more sharply, for the reason #544
// happened: this is the ONE copy of the rule that the engine's bulk
// DeclareAttackers and internal/legal's move enumerator both run. Two
// copies of a declaration rule drift, and when the enumerator's copy
// is the laxer one a bot seat is offered a move the engine refuses,
// which is not a bad move but a hung table.
//
// The target-side half of attack legality (CR 506.2 — not yourself,
// not a planeswalker you control, not a battle you protect) is NOT
// here: it needs the game and a target, and it already has a shared
// predicate of its own in canAttackTargetLocked / AttackTargetsForEffect.
//
// Deliberately NOT a check that `seat` is the active player, and not
// an authorization check: the bulk verb takes the card's own
// controller as `seat` and authorization lives one layer up in
// actions.Dispatch. The enumerator adds the active-seat test itself.
//
// nil is ineligible so callers can skip defensive nil checks. Layers
// must be fresh — HasKeyword and the restriction read both go through
// the effective characteristic.
func AttackerEligible(a *Card, seat uuid.UUID) bool {
	if a == nil {
		return false
	}
	if a.Controller != seat || !a.IsCreature() {
		return false
	}
	if a.Tapped {
		return false
	}
	if a.AttackingTarget != uuid.Nil {
		return false
	}
	if HasKeyword(a, "defender") || HasSummoningSickness(a) {
		return false
	}
	return CanAttack(a)
}
