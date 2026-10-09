package legal

import (
	"regexp"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_interacts.go — #2871. Some untargeted activated abilities are
// not an answer to a spell, so abilityInteracts leaves them out, but
// they change a fight: a creature that gains flying or menace before
// blocks, a manland or a crewed Vehicle that becomes a blocker, a
// creature that can block an additional creature. Counted on every
// opponent spell, a board of them would stop you every time, so they
// get a bit of their own, Move.CombatInteracts, and the client counts
// it only in a combat window (beginning of combat, declare attackers,
// declare blockers, or an attack or block trigger on the stack). ADR
// 0009, amendment #2871.
//
// A creature-token maker is narrower still (owner answer, 2026-10-09):
// its token is a blocker, so it matters only to a player who is being
// attacked. It also sets Move.CombatDefenderOnly, and the client counts
// it only while the viewer defends against an attacker.
//
// It is read from the ability's shape, like abilityInteracts:
//
//  1. a crew cost (CR 702.122), the Vehicle becoming an artifact
//     creature until end of turn, or a crew row whose cost is printed
//     another way ("Crew — remove a loyalty counter");
//  2. its printed effect text, the label after the cost:
//     - becoming a creature (a manland, an animated artifact, Lurking
//       Evil);
//     - gaining or losing one of the evasion and combat keywords below;
//     - "can block an additional creature" or "can block any number";
//     - "can't be blocked", "must be blocked", "attack this turn if
//       able";
//     - tapping or untapping the enchanted creature;
//     - double damage;
//     - tokens that enter "tapped and attacking";
//  3. defender only: making a creature token that enters untapped,
//     populate or amass. A row that also grants a combat keyword
//     ("They gain haste") is read by rule 2 instead.
//
// abilityInteracts is asked first: a row it already marks is never
// also marked here.

// combatKind is how an untargeted ability matters in combat.
type combatKind int

const (
	combatNone     combatKind = iota
	combatAny                 // in any combat window
	combatDefender            // only to a player being attacked
)

// combatPhrases are effect-text fragments of a combat ability, lower
// case. The keyword ones are matched as whole words (combatKeyword).
var combatPhrases = []string{
	"can block an additional",
	"can block any number",
	"can't be blocked",
	"must be blocked",
	"attack this turn if able",
	"tap enchanted creature",
	"deals double",
	"tapped and attacking",
}

// defenderPhrases make or grow a creature token without saying so in
// a power and toughness.
var defenderPhrases = []string{
	"populate",
	"amass",
}

// combatKeyword is a keyword that decides attacks or blocks when an
// untargeted ability grants it.
var combatKeyword = regexp.MustCompile(`\b(flying|haste|trample|menace|lifelink|vigilance|reach)\b`)

var (
	// becomesCreature is an animation: "becomes a 3/3 Elemental
	// creature", "becomes an artifact creature", "become 2/2 creatures".
	becomesCreature = regexp.MustCompile(`\bbecomes? [^.]*\bcreatures?\b`)
	// createsCreature is a creature token, read from its power and
	// toughness or the word: "create a 1/1 white Soldier creature
	// token", "create X 1/1 Goblins", "create an X/X black Aetherborn
	// creature token". The first group catches "create a tapped 3/3",
	// which cannot block.
	createsCreature = regexp.MustCompile(`\bcreate (?:a |an |one |two |three |four |five |x |that many )?(tapped )?[^.]*?(?:\b(?:[0-9]+|x)/(?:[0-9]+|x)\b|\bcreature tokens?\b)`)
)

// abilityCombatKind reports how an untargeted activated ability
// changes attacks or blocks.
func abilityCombatKind(ab game.ActivatedAbilityShape) combatKind {
	if ab.Cost.Crew > 0 || strings.HasPrefix(strings.ToLower(ab.Label), "crew") {
		return combatAny
	}
	return combatTextKind(ab.Label)
}

// combatTextKind reads the printed effect after the cost. A label with
// no colon is read whole.
func combatTextKind(label string) combatKind {
	text := strings.ToLower(label)
	if i := strings.Index(text, ":"); i >= 0 {
		text = text[i+1:]
	}
	for _, p := range combatPhrases {
		if strings.Contains(text, p) {
			return combatAny
		}
	}
	if becomesCreature.MatchString(text) || grantsCombatKeyword(text) {
		return combatAny
	}
	for _, p := range defenderPhrases {
		if strings.Contains(text, p) {
			return combatDefender
		}
	}
	if m := createsCreature.FindStringSubmatch(text); m != nil && m[1] == "" {
		return combatDefender
	}
	return combatNone
}

// grantsCombatKeyword: the keyword is something the ability gives or
// takes away ("gains flying", "have trample", "loses flying"), not a
// word in a "with flying" filter or a token's description, which the
// token rule judges.
var grantsKeyword = regexp.MustCompile(`\b(gains?|has|have|loses?)\b[^.]*`)

func grantsCombatKeyword(text string) bool {
	for _, m := range grantsKeyword.FindAllString(text, -1) {
		if combatKeyword.MatchString(m) {
			return true
		}
	}
	return false
}
