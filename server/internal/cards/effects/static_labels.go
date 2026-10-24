package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// static_labels.go — #2410: the player-facing wording the shared static
// and replacement constructors put on their rows. An art tile's ◆ chip
// lists a card's static abilities by StaticAbility.Label; a constructor
// that sets it gives every card that calls it wording at once, built
// from the constructor's own arguments.
//
// The tone is docs/adding-cards.md's Caveats rule: a sentence a player
// reads, no engine vocabulary (no layer, no "static", no "replacement").
// What a label cannot know — the closure a card passes for "who
// qualifies" — it does not claim: a row whose subject is a predicate
// says what it gives and leaves the subject to the card.

// signedPT renders a power/toughness change the way the card prints it:
// "+1/+1", "-2/-2", "+3/-1".
func signedPT(power, toughness int) string {
	return fmt.Sprintf("%+d/%+d", power, toughness)
}

// joinWords joins "a", "a and b", "a, b and c".
func joinWords(words []string) string {
	switch len(words) {
	case 0:
		return ""
	case 1:
		return words[0]
	}
	return strings.Join(words[:len(words)-1], ", ") + " and " + words[len(words)-1]
}

// irregularPlurals are the creature types whose plural is not "+s".
var irregularPlurals = map[string]string{
	"Elf": "Elves", "Wolf": "Wolves", "Werewolf": "Werewolves",
	"Dwarf": "Dwarves", "Thief": "Thieves", "Fungus": "Fungi",
	"Sphinx": "Sphinxes", "Fish": "Fish", "Sheep": "Sheep",
	"Wolverine": "Wolverines", "Cyclops": "Cyclopes", "Octopus": "Octopuses",
	"Ouphe": "Ouphes", "Djinn": "Djinn", "Homunculus": "Homunculi",
	"Mouse": "Mice", "Merfolk": "Merfolk", "Samurai": "Samurai",
	"Kithkin": "Kithkin", "Moonfolk": "Moonfolk", "Orc": "Orcs",
}

func pluralType(t string) string {
	if p, ok := irregularPlurals[t]; ok {
		return p
	}
	return t + "s"
}

// subject is the filter as a noun phrase: "Other Goblins you control",
// "Elves", "Creatures you control of the chosen type".
func (f TribeFilter) subject() string {
	var who string
	switch {
	case f.Chosen && f.YoursOnly:
		who = "creatures you control of the chosen type"
	case f.Chosen:
		who = "creatures of the chosen type"
	case len(f.Tribes) == 0:
		who = "creatures"
	default:
		plurals := make([]string, len(f.Tribes))
		for i, t := range f.Tribes {
			plurals[i] = pluralType(t)
		}
		who = strings.Join(plurals, " and ")
	}
	if f.YoursOnly && !f.Chosen {
		who += " you control"
	}
	if f.Others {
		who = "other " + who
	}
	return strings.ToUpper(who[:1]) + who[1:]
}

// restrictionClause names the bits of a Restriction a player reads on a
// card: "can't attack", "can't attack or block". Empty when r holds a
// bit this wording does not cover, so the row falls back to its kind
// rather than saying half of it.
func restrictionClause(r game.Restriction) string {
	var parts []string
	rest := r
	for _, p := range []struct {
		bit  game.Restriction
		text string
	}{
		{game.CantAttack, "attack"},
		{game.CantBlock, "block"},
	} {
		if rest&p.bit != 0 {
			parts = append(parts, p.text)
			rest &^= p.bit
		}
	}
	var clauses []string
	if len(parts) > 0 {
		clauses = append(clauses, "can't "+strings.Join(parts, " or "))
	}
	if rest&game.CantBeBlocked != 0 {
		clauses = append(clauses, "can't be blocked")
		rest &^= game.CantBeBlocked
	}
	if rest != 0 || len(clauses) == 0 {
		return ""
	}
	return strings.Join(clauses, " and ")
}

// queryNoun renders a PermanentQuery as the thing a player sees on the
// board: "Island", "artifact", "Island or Swamp". Empty for a query it
// cannot say plainly.
func queryNoun(q game.PermanentQuery) string {
	words := append(append([]string{}, q.Subtypes...), q.Types...)
	if len(words) == 0 {
		return ""
	}
	return strings.Join(words, " or ")
}

func queriesNoun(qs []game.PermanentQuery) string {
	var nouns []string
	for _, q := range qs {
		if n := queryNoun(q); n != "" {
			nouns = append(nouns, n)
		}
	}
	return strings.Join(nouns, " or ")
}

// attackTargetLabel words an AttackTargetRestriction.
func attackTargetLabel(r game.AttackTargetRestriction) string {
	switch {
	case r.NotOwner && r.NotOwnersPlaneswalkers:
		return "Can't attack its owner or its owner's planeswalkers."
	case r.NotOwner:
		return "Can't attack its owner."
	case r.NotPlayer != uuid.Nil:
		return "Can't attack a certain player or their planeswalkers."
	case len(r.DefenderMustControl) > 0:
		if n := queriesNoun(r.DefenderMustControl); n != "" {
			return "Can't attack unless defending player controls " + articleFor(strings.ToLower(n)) + " " + n + "."
		}
		return "Can't attack unless defending player controls a certain permanent."
	case r.DefenderMustBePoisoned:
		return "Can't attack unless defending player is poisoned."
	case r.DefenderMustBeMonarch:
		return "Can't attack unless defending player is the monarch."
	case r.DefenderGraveyardAtLeast > 0:
		return fmt.Sprintf("Can't attack unless defending player has %d or more cards in their graveyard.", r.DefenderGraveyardAtLeast)
	case len(r.ControllerMustControlMore) > 0:
		if n := queriesNoun(r.ControllerMustControlMore); n != "" {
			return "Can't attack unless you control more " + n + "s than defending player."
		}
		return "Can't attack unless you control more of a certain permanent than defending player."
	case r.ControllerMustHaveCitysBlessing:
		return "Can't attack unless you have the city's blessing."
	case r.NotAlreadyAttackedThisTurn:
		return "Can't attack a player it already attacked this turn."
	case r.PlayerMustHaveActedLastTurn:
		return "Can't attack a player who didn't cast a spell or put a nontoken permanent onto the battlefield during their last turn."
	}
	return ""
}

// counterCountPhrase is "a +1/+1 counter" / "three charge counters" in
// the wording of an enters-with-counters row.
func counterCountPhrase(n int, kind string) string {
	if n == 1 {
		return "a " + kind + " counter"
	}
	return fmt.Sprintf("%d %s counters", n, kind)
}

// rowLabelFromHeader turns a card's CR 616 prompt header ("Pariah —
// damage to you is dealt to enchanted creature instead") into the
// chip's sentence by dropping the card's name: "Damage to you is dealt
// to enchanted creature instead." Used by the families whose cards each
// already carry a header that reads as the printed ability.
func rowLabelFromHeader(header string) string {
	if _, after, ok := strings.Cut(header, " — "); ok {
		header = after
	}
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	header = strings.ToUpper(header[:1]) + header[1:]
	if !strings.HasSuffix(header, ".") {
		header += "."
	}
	return header
}

// Labelled is s with the wording its ◆ chip shows (#2410), for a card
// whose static comes from a shared constructor that cannot know the
// card's subject.
func Labelled(s game.StaticAbility, label string) game.StaticAbility {
	s.Label = label
	return s
}

// attachedTypesLabel words SetAttachedTypes: "Attached permanent becomes
// an Elk creature."
func attachedTypesLabel(types, subtypes []string) string {
	var words []string
	words = append(words, subtypes...)
	for _, t := range types {
		words = append(words, strings.ToLower(t))
	}
	if len(words) == 0 {
		return "Attached permanent's type is changed."
	}
	phrase := strings.Join(words, " ")
	return "Attached permanent becomes " + articleFor(strings.ToLower(phrase)) + " " + phrase + "."
}

// colorWords turns wire colour letters into "white", "blue and black".
func colorWords(letters []string) string {
	names := map[string]string{"W": "white", "U": "blue", "B": "black", "R": "red", "G": "green"}
	words := make([]string, 0, len(letters))
	for _, l := range letters {
		if n, ok := names[l]; ok {
			words = append(words, n)
		} else {
			words = append(words, l)
		}
	}
	return joinWords(words)
}
