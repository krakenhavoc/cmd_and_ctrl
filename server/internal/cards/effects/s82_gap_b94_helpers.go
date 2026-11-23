package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b94_helpers.go — card-side helpers of set GAP, batch 94
// (Doctor Who enchantments and instants, S82, #5948).

// GapB94IsDoctorCard is "a Doctor card": a card with the Doctor subtype.
func GapB94IsDoctorCard(c game.Card) bool { return c.HasSubtypeWord("Doctor") }

// GapB94IsTimeLord is "a Time Lord": the engine reads a type line word by word,
// so the creature type Time Lord is the two words Time and Lord (no other
// type is spelled either).
func GapB94IsTimeLord(c game.Card) bool {
	return c.HasSubtypeWord("Time") && c.HasSubtypeWord("Lord")
}

// GapB94TimeLordCard is a CardPredicate over GapB94IsTimeLord.
func GapB94TimeLordCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return GapB94IsTimeLord(c) }
}

// GapB94TappedCreature is a tapped creature, for "tapped creatures".
func GapB94TappedCreature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() && c.Tapped }
}

// GapB94CounterKinds is the kinds of counter on a permanent or card, sorted:
// the stable list a "choose a counter on it" prompt offers and re-derives.
func GapB94CounterKinds(c game.Card) []string {
	var kinds []string
	for kind, n := range c.Counters {
		if n > 0 {
			kinds = append(kinds, kind)
		}
	}
	sort.Strings(kinds)
	return kinds
}

// --- The Girl in the Fireplace's tokens ----------------------------------

// GapHumanNobleToken is the 1/1 white Human Noble creature token with
// vanishing 3 and "Prevent all damage that would be dealt to this token."
// It is made with three time counters (TokenSpec.WithCounters), which is how
// vanishing 3 enters (CR 702.63a); the upkeep removal and the last-counter
// sacrifice are the template's two triggers.
func GapHumanNobleToken() game.Card { return TokenFromCatalog(PrintedGapHumanNobleToken) }

// PrintedGapHumanNobleToken is that Human Noble as PRINTED. Its slug is an
// on-disk identity and is never renamed.
func PrintedGapHumanNobleToken() TokenTemplate {
	return TokenTemplate{
		Slug: "human-noble-vanishing-prevent-damage",
		Card: game.Card{
			Name:      "Human Noble",
			TypeLine:  "Token Creature — Human Noble",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		Triggered: Cmdr2B17VanishingTriggers(),
		Replacements: []game.ReplacementEffect{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Prevention: true, // CR 615.1a — "prevent"
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDamage && ev.DamageTarget == src.InstanceID && ev.DamageAmount > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Prevent all damage that would be dealt to this creature.",
		}},
		Text: "Vanishing 3\nPrevent all damage that would be dealt to this creature.",
	}
}

// GapHorseToken is the 2/2 white Horse creature token with "Doctors you
// control have horsemanship."
func GapHorseToken() game.Card { return TokenFromCatalog(PrintedGapHorseToken) }

// PrintedGapHorseToken is that Horse as PRINTED. Its slug is an on-disk
// identity and is never renamed.
func PrintedGapHorseToken() TokenTemplate {
	return TokenTemplate{
		Slug: "horse-doctors-have-horsemanship",
		Card: game.Card{
			Name:      "Horse",
			TypeLine:  "Token Creature — Horse",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"W"},
		},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Doctor"}, YoursOnly: true}, "horsemanship"),
		},
		Text: "Doctors you control have horsemanship. (They can't be blocked except by creatures with horsemanship.)",
	}
}
