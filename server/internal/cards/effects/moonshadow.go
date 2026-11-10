package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonshadow — Creature — Elemental {B}, 7/7 (Lorwyn Eclipsed, #4458):
//
//	"Menace
//	 This creature enters with six -1/-1 counters on it.
//	 Whenever one or more permanent cards are put into your graveyard from
//	 anywhere while this creature has a -1/-1 counter on it, remove a
//	 -1/-1 counter from this creature."
//
// The counters are an entry replacement (Brambleback Brute). The trigger is
// a "one or more" batch (OncePerBatch, CR 603.2c): every card a death,
// discard, mill, surveil or resolving spell puts in your graveyard is read
// by graveyardArrivals, a permanent card meaning a creature, artifact,
// enchantment, land, planeswalker or battle card as it sits in the
// graveyard (CR 603.6c), and never a token. The "while" is read as the
// ability would trigger, so it does not trigger with no -1/-1 counter, and
// the counter is removed only if the Moonshadow is still the same permanent
// with one as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3635b3f-839c-4fe5-83a5-7c70e8630184",
		Name:            "Moonshadow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterMinusOne, 6, "Moonshadow: enters with six -1/-1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(graveyardArrivalKinds, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return source.Counters[game.CounterMinusOne] > 0 && stdB64PermanentCardsToYourGraveyard.When()(ev, source, lki, g)
			}, "Moonshadow — remove a -1/-1 counter from this creature",
				Do(Instruction(stdB64RemoveAMinusCounterFromThis)))),
		},
	})
}
