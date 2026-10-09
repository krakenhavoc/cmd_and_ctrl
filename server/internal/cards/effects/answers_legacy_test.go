package effects

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answers_legacy_test.go — ADR 0142 delivery row 2. The #2853
// interacts read as it stood before ADR 0142 (internal/legal/interacts.go
// at 3407bee98), frozen, so TestAnswersOfKeepsEveryInteractsVerdict can
// hold answersOf to every verdict it gave. Never edited; deleted with
// that test in S4.

// legacyAbilityInteracts reports whether an activated ability with no target
// can matter in response to something on the stack.
func legacyAbilityInteracts(ab game.ActivatedAbilityShape) bool {
	if legacySacrificeInteracts(ab.Cost.SacrificeOther) || legacyExileCostInteracts(ab.Cost.ExilePermanents) {
		return true
	}
	if rc := ab.Cost.ReturnToHand; !rc.Empty() && legacySacrificeInteracts(rc.Filter) {
		return true
	}
	if legacyPurposeInteracts(ab.Purpose) {
		return true
	}
	return legacyEffectTextInteracts(ab.Label)
}

// legacyValueSacrifices are the printed sacrifice objects that are a price,
// not an answer: lands and the food, treasure, clue and blood tokens.
var legacyValueSacrifices = []string{
	"land", "swamp", "forest", "island", "mountain", "plains", "desert",
	"food", "treasure", "clue", "blood", "artifact", "room", "token",
}

// legacySacrificeInteracts reads a sacrifice cost's printed object ("a
// creature", "another creature", "a Food"). A creature, a permanent,
// or a creature type the list does not know (a Goblin, an Eldrazi) is
// an outlet; a land, a token kind or a plain artifact is not.
func legacySacrificeInteracts(spec *game.TargetSpec) bool {
	if spec == nil {
		return false
	}
	label := strings.ToLower(strings.ReplaceAll(spec.Label, "noncreature", ""))
	if strings.Contains(label, "creature") || strings.Contains(label, "permanent") {
		return true
	}
	for _, v := range legacyValueSacrifices {
		if strings.Contains(label, v) {
			return false
		}
	}
	return true
}

// legacyExileCostInteracts: exiling a creature (or any permanent) you control
// as a cost saves it from what is on the stack; exiling a land does not.
func legacyExileCostInteracts(c *game.ExilePermanentsCost) bool {
	if c.Empty() {
		return false
	}
	switch strings.ToLower(c.CardType) {
	case "", "creature", "planeswalker":
		return true
	}
	return false
}

func legacyPurposeInteracts(p game.Purpose) bool {
	return p.Pump != nil || p.PreventCombatDamageToSelf || p.DamageToCreature > 0 ||
		p.Sweep != (game.Sweep{})
}

// legacyInteractingPhrases are effect-text fragments of an answer, lower case.
var legacyInteractingPhrases = []string{
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
	// legacyPtChange matches a printed power/toughness change or counter:
	// "+1/+1", "-2/-0", "+X/+0".
	legacyPtChange = regexp.MustCompile(`[+-]([0-9]+|x)/[+-]([0-9]+|x)`)
	// legacyEachCreatureDamage is a pinging sweep: "damage to each creature",
	// "to each other creature", "to each creature with flying". Damage
	// to each opponent or player alone is not an answer.
	legacyEachCreatureDamage = regexp.MustCompile(`damage to each (other )?creature`)
	// legacyBlinkReturn is the second half of a blink: "return it", "return
	// that card", "return them".
	// legacySelfBounce is an effect that returns something of its own to its
	// owner's hand: "Return this creature to its owner's hand."
	legacySelfBounce  = regexp.MustCompile(`^\s*return [^.]*to its owner's hand`)
	legacyBlinkReturn = regexp.MustCompile(`return (it|that|them|those)\b`)
)

// legacyEffectTextInteracts reads the printed effect, the part of an ability
// row's label after its cost ("{1}{B}: Regenerate this creature."). A
// label with no colon is read whole.
func legacyEffectTextInteracts(label string) bool {
	text := strings.ToLower(label)
	if i := strings.Index(text, ":"); i >= 0 {
		text = text[i+1:]
	}
	for _, p := range legacyInteractingPhrases {
		if strings.Contains(text, p) {
			return true
		}
	}
	if legacyPtChange.MatchString(text) || legacyEachCreatureDamage.MatchString(text) {
		return true
	}
	// "Destroy all artifacts, creatures, and enchantments" answers;
	// "Destroy this enchantment" is the card leaving on its own terms.
	if strings.Contains(strings.ReplaceAll(text, "destroy this", ""), "destroy") {
		return true
	}
	// Returning itself to its owner's hand: Aethertide Whale, Arcanis,
	// Batterskull. A save from removal.
	if legacySelfBounce.MatchString(text) {
		return true
	}
	// A blink: exile one of your permanents, then return it. "Return
	// all cards exiled with this" is not one.
	return strings.Contains(text, "exile") && legacyBlinkReturn.MatchString(text)
}
