package legal

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// interacts.go — #2853, owner answer 2 (2026-10-09). An activated
// ability with no target can still be an answer to something on the
// stack: a sacrifice outlet in response to removal, a regeneration
// shield, Selfless Spirit's indestructible, Spore Frog's fog, a pump in
// response to a burn spell. Move.Interacts says so, and smart autopass
// stops for it as it stops for a targeted ability. Pure value (draw,
// mana, ramp, a fetch, tokens, scry, a Clue) stays unmarked.
//
// The owner's goal is both halves at once: no pointless stops and no
// missed windows. So the rules below are as narrow as the printed text
// allows, and lean towards "interacts" only where the text cannot say.
// The answer is read from the ability's shape, in this order:
//
//  1. its cost: sacrificing, exiling or returning to hand a creature
//     you control (or a permanent the cost does not name as plainly
//     non-creature) is an outlet or a save, whatever the ability buys.
//     Sacrificing or returning a land, a Food, a Treasure, a Clue or a
//     plain artifact is a value cost and does not count;
//  2. its declared purpose (ADR 0126 §6): a pump, a damage shield, a
//     creature it damages, a sweep;
//  3. its printed effect text, the ability row's Label after the cost:
//     regenerate, protection, indestructible, hexproof, shroud, phasing,
//     prevention, a power/toughness change or +1/+1 and -1/-1 counters,
//     an exile that returns (a blink), damage to each creature, a
//     destroy that is not "destroy this".
//
// The effect itself is a closure, so the text is the only description
// of it the enumerator has. interacts_internal_test.go pins both lists.

// abilityInteracts reports whether an activated ability with no target
// can matter in response to something on the stack.
func abilityInteracts(ab game.ActivatedAbilityShape) bool {
	if sacrificeInteracts(ab.Cost.SacrificeOther) || exileCostInteracts(ab.Cost.ExilePermanents) {
		return true
	}
	if rc := ab.Cost.ReturnToHand; !rc.Empty() && sacrificeInteracts(rc.Filter) {
		return true
	}
	if purposeInteracts(ab.Purpose) {
		return true
	}
	return effectTextInteracts(ab.Label)
}

// manaAbilityInteracts is the same question for a mana ability: only a
// creature sacrifice outlet (Ashnod's Altar, Phyrexian Altar) answers
// anything. The mana it makes never does.
func manaAbilityInteracts(sacrificeOther *game.TargetSpec) bool {
	return sacrificeInteracts(sacrificeOther)
}

// valueSacrifices are the printed sacrifice objects that are a price,
// not an answer: lands and the food, treasure, clue and blood tokens.
var valueSacrifices = []string{
	"land", "swamp", "forest", "island", "mountain", "plains", "desert",
	"food", "treasure", "clue", "blood", "artifact", "room", "token",
}

// sacrificeInteracts reads a sacrifice cost's printed object ("a
// creature", "another creature", "a Food"). A creature, a permanent,
// or a creature type the list does not know (a Goblin, an Eldrazi) is
// an outlet; a land, a token kind or a plain artifact is not.
func sacrificeInteracts(spec *game.TargetSpec) bool {
	if spec == nil {
		return false
	}
	label := strings.ToLower(strings.ReplaceAll(spec.Label, "noncreature", ""))
	if strings.Contains(label, "creature") || strings.Contains(label, "permanent") {
		return true
	}
	for _, v := range valueSacrifices {
		if strings.Contains(label, v) {
			return false
		}
	}
	return true
}

// exileCostInteracts: exiling a creature (or any permanent) you control
// as a cost saves it from what is on the stack; exiling a land does not.
func exileCostInteracts(c *game.ExilePermanentsCost) bool {
	if c.Empty() {
		return false
	}
	switch strings.ToLower(c.CardType) {
	case "", "creature", "planeswalker":
		return true
	}
	return false
}

func purposeInteracts(p game.Purpose) bool {
	return p.Pump != nil || p.PreventCombatDamageToSelf || p.DamageToCreature > 0 ||
		p.Sweep != (game.Sweep{})
}

// interactingPhrases are effect-text fragments of an answer, lower case.
var interactingPhrases = []string{
	"regenerate",
	"protection from",
	"indestructible",
	"hexproof",
	"shroud",
	"phase out",
	"phases out",
	"prevent",
	// Redirection and damage shields: Opal-Eye, Beacon of Destiny,
	// Aegis of Honor, Personal Incarnation.
	"the next time a source",
	"the next time an instant",
	"would be dealt",
	// Pumps the text does not spell as +N/+N.
	"monstrosity",
	"adapt ",
	"base power",
	// Saves from destruction and combat tricks.
	"persist",
	"undying",
	"first strike",
	"double strike",
	"deathtouch",
	// Ranger-Captain of Eos: an answer to what comes next.
	"can't cast",
}

var (
	// ptChange matches a printed power/toughness change or counter:
	// "+1/+1", "-2/-0", "+X/+0".
	ptChange = regexp.MustCompile(`[+-]([0-9]+|x)/[+-]([0-9]+|x)`)
	// eachCreatureDamage is a pinging sweep: "damage to each creature",
	// "to each other creature", "to each creature with flying". Damage
	// to each opponent or player alone is not an answer.
	eachCreatureDamage = regexp.MustCompile(`damage to each (other )?creature`)
	// blinkReturn is the second half of a blink: "return it", "return
	// that card", "return them".
	// selfBounce is an effect that returns something of its own to its
	// owner's hand: "Return this creature to its owner's hand."
	selfBounce  = regexp.MustCompile(`^\s*return [^.]*to its owner's hand`)
	blinkReturn = regexp.MustCompile(`return (it|that|them|those)\b`)
)

// effectTextInteracts reads the printed effect, the part of an ability
// row's label after its cost ("{1}{B}: Regenerate this creature."). A
// label with no colon is read whole.
func effectTextInteracts(label string) bool {
	text := strings.ToLower(label)
	if i := strings.Index(text, ":"); i >= 0 {
		text = text[i+1:]
	}
	for _, p := range interactingPhrases {
		if strings.Contains(text, p) {
			return true
		}
	}
	if ptChange.MatchString(text) || eachCreatureDamage.MatchString(text) {
		return true
	}
	// "Destroy all artifacts, creatures, and enchantments" answers;
	// "Destroy this enchantment" is the card leaving on its own terms.
	if strings.Contains(strings.ReplaceAll(text, "destroy this", ""), "destroy") {
		return true
	}
	// Returning itself to its owner's hand: Aethertide Whale, Arcanis,
	// Batterskull. A save from removal.
	if selfBounce.MatchString(text) {
		return true
	}
	// A blink: exile one of your permanents, then return it. "Return
	// all cards exiled with this" is not one.
	return strings.Contains(text, "exile") && blinkReturn.MatchString(text)
}
