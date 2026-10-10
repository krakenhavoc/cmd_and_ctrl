package legal

import (
	"regexp"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// interacts.go — #2853, owner answer 2 (2026-10-09), and ADR 0142. An
// activated ability with no target can still be an answer to something
// on the stack: a sacrifice outlet in response to removal, a
// regeneration shield, Selfless Spirit's indestructible, Spore Frog's
// fog, a pump in response to a burn spell. Move.Interacts says so, and
// smart autopass stops for it as it stops for a targeted ability. Pure
// value (draw, mana, ramp, a fetch, tokens, scry, a Clue) stays
// unmarked. Move.CombatInteracts is the combat tier: an ability that
// changes a fight (a granted first strike, crew) and counts only in a
// combat window (#2871).
//
// ADR 0142: what an untargeted activation answers is DECLARED on the
// catalog row (game.Purpose.Answers), and answersOf reads that first. A
// declaration wins outright: a declared row is never read as text.
//
// A row that declares nothing falls back to fallbackAnswers, which is
// the #2853 read mapped onto the declared vocabulary. The owner's goal
// is both halves at once: no pointless stops and no missed windows. So
// the rules below are as narrow as the printed text allows, and lean
// towards "interacts" only where the text cannot say. The answer is
// read from the ability's shape, in this order:
//
//  1. its cost: sacrificing, exiling or returning to hand a creature
//     you control (or a permanent the cost does not name as plainly
//     non-creature) is an outlet or a save, whatever the ability buys
//     (sac_outlet). Sacrificing or returning a land, a Food, a
//     Treasure, a Clue or a plain artifact is a value cost and does
//     not count. A crew cost is a Vehicle becoming a creature
//     (animate);
//  2. its declared purpose (ADR 0126 §6): a pump (pump), a damage
//     shield (prevent), a creature it damages, a sweep (remove);
//  3. its printed effect text, the ability row's Label after the cost:
//     regenerate, protection, indestructible, hexproof, shroud, phasing,
//     persist, undying, a blink or a self-bounce (protect); prevention
//     and redirection (prevent); a power/toughness change or +1/+1 and
//     -1/-1 counters, monstrosity, adapt, a base power (pump); damage
//     to each creature, a destroy that is not "destroy this" (remove);
//     "can't cast" (restrict); first strike, double strike and
//     deathtouch (combat_grant, ADR 0142 owner answer 2: they change
//     nothing outside combat).
//
// The effect itself is a closure, so the text is the only description
// of it the enumerator has for an undeclared row.
// interacts_internal_test.go pins both lists. The fallback only
// shrinks: internal/cards/effects' TestAnswersFallbackOnlyShrinks lists
// every catalog row still read this way.
//
// One more input is read from the game, not the row (ADR 0142 owner
// answer 4): a "Sacrifice this creature: …" move interacts while its
// creature is a target of an opponent's item on the stack, and
// combat-interacts while it is attacking or blocking
// (selfSacrificeThreatened).

// answersOf is what an untargeted activation can do in response, and
// whether that is the row's declaration (true) or the fallback read.
func answersOf(ab game.ActivatedAbilityShape) (game.Answers, bool) {
	if a := ab.Purpose.Answers; a != 0 {
		return a, true
	}
	return fallbackAnswers(ab), false
}

// fallbackAnswers is the #2853 read of an undeclared row, mapped onto
// ADR 0142's vocabulary.
func fallbackAnswers(ab game.ActivatedAbilityShape) game.Answers {
	return CostAnswers(ab.Cost) | purposeAnswers(ab.Purpose) | effectTextAnswers(ab.Label) | combatAnswers(ab)
}

// AnswersOf is answersOf for the catalog tests in internal/cards/effects
// (the ratchet and the disagreement record, ADR 0142 §4).
func AnswersOf(ab game.ActivatedAbilityShape) (game.Answers, bool) { return answersOf(ab) }

// FallbackAnswers is fallbackAnswers for the same tests: what the
// printed-text read would say of a row, declared or not.
func FallbackAnswers(ab game.ActivatedAbilityShape) game.Answers { return fallbackAnswers(ab) }

// CostAnswers is what an activated ability's cost alone answers: a
// creature sacrifice, exile or return outlet (sac_outlet), or crew
// (animate). effects.Register reads it too: a row whose cost is an
// outlet must declare sac_outlet or protect (ADR 0142 §3), so the cost
// rule stays true for every declared row.
func CostAnswers(c game.AbilityCost) game.Answers {
	var a game.Answers
	if sacrificeInteracts(c.SacrificeOther) || exileCostInteracts(c.ExilePermanents) {
		a |= game.AnswerSacOutlet
	}
	if rc := c.ReturnToHand; !rc.Empty() && sacrificeInteracts(rc.Filter) {
		a |= game.AnswerSacOutlet
	}
	if c.Crew > 0 {
		a |= game.AnswerAnimate
	}
	return a
}

// ManaCostAnswers is the same question for a mana ability's sacrifice
// cost: only a creature sacrifice outlet (Ashnod's Altar, Phyrexian
// Altar) answers anything. The mana it makes never does.
func ManaCostAnswers(sacrificeOther *game.TargetSpec) game.Answers {
	if sacrificeInteracts(sacrificeOther) {
		return game.AnswerSacOutlet
	}
	return 0
}

// manaAnswersOf is answersOf for a mana ability: its declaration
// (sac_outlet or value), else its sacrifice cost.
func manaAnswersOf(m game.ManaAbilityShape) (game.Answers, bool) {
	if m.Answers != 0 {
		return m.Answers, true
	}
	return ManaCostAnswers(m.SacrificeOther), false
}

// ManaAnswersOf is manaAnswersOf for the catalog tests.
func ManaAnswersOf(m game.ManaAbilityShape) (game.Answers, bool) { return manaAnswersOf(m) }

// selfSacrificeThreatened is ADR 0142 owner answer 4: "Sacrifice this
// creature: …" is an answer when the creature is about to be lost. It
// is read from the game: `stack` when the creature is a target of an
// item on the stack that its controller does not control, `combat` when
// it is attacking or blocking. A source that is not a creature on the
// battlefield, or a row that does not sacrifice its source, is neither.
func selfSacrificeThreatened(g *game.Game, source *game.Card, zone game.ZoneKind, sacrificesSelf bool) (stack, combat bool) {
	if !sacrificesSelf || source == nil || zone != game.ZoneBattlefield || !source.IsCreature() {
		return false, false
	}
	combat = source.AttackingTarget != uuid.Nil || source.BlockingTarget != uuid.Nil
	if g.Stack != nil {
		for i := range g.Stack.Cards {
			item := g.StackMeta[g.Stack.Cards[i].InstanceID]
			if item == nil || item.Controller == source.Controller {
				continue
			}
			for _, r := range item.Targets {
				if r.Kind == game.TargetCard && r.ID == source.InstanceID {
					return true, combat
				}
			}
		}
	}
	return false, combat
}

// combatFlags is an activate move's combat_interacts and
// combat_defender_only.
type combatFlags struct {
	any, defenderOnly bool
}

// untargetedFlags is an activate move's interacts and combat flags: the
// row's answers by tier, and owner answer 4's threatened self-sacrifice.
// A move with a target sets none (has_targets already stops you).
func untargetedFlags(g *game.Game, source *game.Card, zone game.ZoneKind, ab game.ActivatedAbilityShape, targets []game.TargetRef) (bool, combatFlags) {
	if hasTargets(targets) {
		return false, combatFlags{}
	}
	answers, _ := answersOf(ab)
	threatened, fighting := selfSacrificeThreatened(g, source, zone, ab.Cost.SacrificeSelf)
	return answerFlags(answers, threatened, fighting)
}

// answerFlags maps a row's answers and its threatened self-sacrifice
// onto the move's flags. combat_interacts is never set beside interacts,
// which already counts in every window (#2871), and an ability whose
// only combat answer is a token blocker (makes_blocker) matters only to
// a defending player (combat_defender_only).
func answerFlags(answers game.Answers, threatened, fighting bool) (bool, combatFlags) {
	if answers.HasTier(game.TierStack) || threatened {
		return true, combatFlags{}
	}
	if !answers.HasTier(game.TierCombat) && !fighting {
		return false, combatFlags{}
	}
	blockerOnly := answers&(game.AnswerCombatGrant|game.AnswerAnimate|game.AnswerMakesBlocker) == game.AnswerMakesBlocker
	return false, combatFlags{any: true, defenderOnly: blockerOnly && !fighting}
}

// AnswerFlags is answerFlags for the catalog test in internal/legal's
// external tests: interacts, combat_interacts, combat_defender_only.
func AnswerFlags(answers game.Answers) (interacts, combat, defenderOnly bool) {
	i, c := answerFlags(answers, false, false)
	return i, c.any, c.defenderOnly
}

// manaMoveInteracts is a mana move's interacts: a declared or read
// sacrifice outlet, or (owner answer 4) a creature that sacrifices
// itself while an opponent's item targets it. A mana move never sets
// combat_interacts.
func manaMoveInteracts(g *game.Game, source *game.Card, zone game.ZoneKind, ab game.ManaAbilityShape) bool {
	answers, _ := manaAnswersOf(ab)
	if answers.HasTier(game.TierStack) {
		return true
	}
	threatened, _ := selfSacrificeThreatened(g, source, zone, ab.SacrificeCost)
	return threatened
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

// purposeAnswers reads the amounts a row declares (ADR 0126 §6): a pump,
// a damage shield, damage to a creature, a sweep.
func purposeAnswers(p game.Purpose) game.Answers {
	var a game.Answers
	if p.Pump != nil {
		a |= game.AnswerPump
	}
	if p.PreventCombatDamageToSelf {
		a |= game.AnswerPrevent
	}
	if p.DamageToCreature > 0 || p.Sweep != (game.Sweep{}) {
		a |= game.AnswerRemove
	}
	return a
}

// interactingPhrases are effect-text fragments of an answer, lower
// case, and the answer each one reads as.
var interactingPhrases = []struct {
	phrase string
	answer game.Answers
}{
	{"regenerate", game.AnswerProtect},
	{"protection from", game.AnswerProtect},
	{"indestructible", game.AnswerProtect},
	{"hexproof", game.AnswerProtect},
	{"shroud", game.AnswerProtect},
	{"phase out", game.AnswerProtect},
	{"phases out", game.AnswerProtect},
	{"prevent", game.AnswerPrevent},
	// Redirection and damage shields: Opal-Eye, Beacon of Destiny,
	// Aegis of Honor, Personal Incarnation.
	{"the next time a source", game.AnswerPrevent},
	{"the next time an instant", game.AnswerPrevent},
	{"would be dealt", game.AnswerPrevent},
	// Pumps the text does not spell as +N/+N.
	{"monstrosity", game.AnswerPump},
	{"adapt ", game.AnswerPump},
	{"base power", game.AnswerPump},
	// Saves from destruction.
	{"persist", game.AnswerProtect},
	{"undying", game.AnswerProtect},
	// Combat tricks: they change a fight and nothing else, so they
	// count only in a combat window (ADR 0142 owner answer 2).
	{"first strike", game.AnswerCombatGrant},
	{"double strike", game.AnswerCombatGrant},
	{"deathtouch", game.AnswerCombatGrant},
	// Ranger-Captain of Eos: an answer to what comes next.
	{"can't cast", game.AnswerRestrict},
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

// effectTextInteracts reports whether the printed effect reads as an
// answer of either tier. effectTextAnswers says which.
func effectTextInteracts(label string) bool {
	return effectTextAnswers(label) != 0
}

// effectTextAnswers reads the printed effect, the part of an ability
// row's label after its cost ("{1}{B}: Regenerate this creature."). A
// label with no colon is read whole.
func effectTextAnswers(label string) game.Answers {
	text := strings.ToLower(label)
	if i := strings.Index(text, ":"); i >= 0 {
		text = text[i+1:]
	}
	var a game.Answers
	for _, p := range interactingPhrases {
		if strings.Contains(text, p.phrase) {
			a |= p.answer
		}
	}
	if ptChange.MatchString(text) {
		a |= game.AnswerPump
	}
	if eachCreatureDamage.MatchString(text) {
		a |= game.AnswerRemove
	}
	// "Destroy all artifacts, creatures, and enchantments" answers;
	// "Destroy this enchantment" is the card leaving on its own terms.
	if strings.Contains(strings.ReplaceAll(text, "destroy this", ""), "destroy") {
		a |= game.AnswerRemove
	}
	// Returning itself to its owner's hand: Aethertide Whale, Arcanis,
	// Batterskull. A save from removal.
	if selfBounce.MatchString(text) {
		a |= game.AnswerProtect
	}
	// A blink: exile one of your permanents, then return it. "Return
	// all cards exiled with this" is not one.
	if strings.Contains(text, "exile") && blinkReturn.MatchString(text) {
		a |= game.AnswerProtect
	}
	return a
}
