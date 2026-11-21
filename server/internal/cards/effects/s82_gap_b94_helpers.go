package effects

import (
	"slices"
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b94_helpers.go — card-side helpers of set GAP, batch 94
// (Doctor Who enchantments and instants, S82, #5948).

// gapB94IsDoctorCard is "a Doctor card": a card with the Doctor subtype.
func gapB94IsDoctorCard(c game.Card) bool { return c.HasSubtypeWord("Doctor") }

// gapB94IsTimeLord is "a Time Lord": the engine reads a type line word by word,
// so the creature type Time Lord is the two words Time and Lord (no other
// type is spelled either).
func gapB94IsTimeLord(c game.Card) bool {
	return c.HasSubtypeWord("Time") && c.HasSubtypeWord("Lord")
}

// gapB94TimeLordCard is a CardPredicate over gapB94IsTimeLord.
func gapB94TimeLordCard() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return gapB94IsTimeLord(c) }
}

// gapB94TappedCreature is a tapped creature, for "tapped creatures".
func gapB94TappedCreature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() && c.Tapped }
}

// gapB94CounterKinds is the kinds of counter on a permanent or card, sorted:
// the stable list a "choose a counter on it" prompt offers and re-derives.
func gapB94CounterKinds(c game.Card) []string {
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

// gapHumanNobleToken is the 1/1 white Human Noble creature token with
// vanishing 3 and "Prevent all damage that would be dealt to this token."
// It is made with three time counters (TokenSpec.WithCounters), which is how
// vanishing 3 enters (CR 702.63a); the upkeep removal and the last-counter
// sacrifice are the template's two triggers.
func gapHumanNobleToken() game.Card { return tokenFromCatalog(printedGapHumanNobleToken) }

// printedGapHumanNobleToken is that Human Noble as PRINTED. Its slug is an
// on-disk identity and is never renamed.
func printedGapHumanNobleToken() tokenTemplate {
	return tokenTemplate{
		Slug: "human-noble-vanishing-prevent-damage",
		Card: game.Card{
			Name:      "Human Noble",
			TypeLine:  "Token Creature — Human Noble",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"W"},
		},
		Triggered: cmdr2B17VanishingTriggers(),
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

// gapHorseToken is the 2/2 white Horse creature token with "Doctors you
// control have horsemanship."
func gapHorseToken() game.Card { return tokenFromCatalog(printedGapHorseToken) }

// printedGapHorseToken is that Horse as PRINTED. Its slug is an on-disk
// identity and is never renamed.
func printedGapHorseToken() tokenTemplate {
	return tokenTemplate{
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

// --- "choose a counter on it" prompts -------------------------------------

// gapB94AddAnotherCounterOptions are the options of "choose a counter on
// it. You may put an additional counter of that kind on it" for a
// permanent: one "put another <kind>" per kind it has, then a decline.
func gapB94AddAnotherCounterOptions(c game.Card) []game.ChoiceOption {
	var out []game.ChoiceOption
	for _, kind := range gapB94CounterKinds(c) {
		out = append(out, game.ChoiceOption{Label: "Put another " + kind + " counter on " + c.Name})
	}
	return append(out, game.ChoiceOption{Label: "Don't put another counter on " + c.Name})
}

// gapB94AskAdditionalCounter asks the controller about the first permanent
// of `ids` that still has a counter (CR 608.2b: one that left or lost its
// counters is skipped), carrying the rest.
func gapB94AskAdditionalCounter(ctx *Context, ids []uuid.UUID) error {
	for len(ids) > 0 {
		c, ok := ctx.Game.LookupCardForEffect(ids[0])
		if !ok || len(gapB94CounterKinds(c)) == 0 {
			ids = ids[1:]
			continue
		}
		return PickOption{
			Player:   ctx.Controller(),
			Question: "The Caves of Androzani — choose a counter on " + c.Name + " and you may put another of that kind on it",
			Options:  gapB94AddAnotherCounterOptions(c),
			ThenKey:  gapB94AdditionalCounterAnswered,
			Carry:    slices.Clone(ids),
			WithItem: true,
		}.Apply(ctx)
	}
	return nil
}

// gapB94AdditionalCounterAnswered applies one answer and asks about the next
// permanent. An on-disk key: never renamed, never reused. Registered in
// init, since the answer asks the next question, which names this key.
var gapB94AdditionalCounterAnswered game.OptionPickThen

func init() {
	gapB94AdditionalCounterAnswered = OptionPickThen("option-pick/gap-caves-of-androzani-counter", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) == 0 || r.Option == nil {
			return nil
		}
		id := r.Carry[0]
		if c, ok := ctx.Game.LookupCardForEffect(id); ok {
			kinds := gapB94CounterKinds(c)
			if r.Index >= 0 && r.Index < len(kinds) {
				if err := ctx.Game.AddCounterByForEffect(r.Chooser, id, kinds[r.Index], 1); err != nil {
					return err
				}
			}
		}
		return gapB94AskAdditionalCounter(ctx, r.Carry[1:])
	})
}
