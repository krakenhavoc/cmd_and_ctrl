package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chief of the Wilds — Legendary Creature — Wolf {2}{B}{G}, 4/4
// (The Hobbit, #4454):
//
//	"Menace
//	 Whenever another Wolf you control enters, put two +1/+1 counters on
//	 Chief of the Wilds.
//	 If a triggered ability of another Wolf or battle you control
//	 triggers, that ability triggers an additional time."
//
// Menace rides PrintedKeywords. The counters are an enters watcher that
// excludes the Chief itself. The doubling is CR 603.2d's shape
// (DoublesAbilitiesOf, as Bifur, Melodic Rider has it) over a source that is
// a Wolf other than this Chief, or a battle, under the Chief's controller:
// "Wolf" is read through the Chief's own text (CR 612.2), and the source is
// read as it is or last was, so a Wolf's dies trigger is doubled too. The
// Chief's own abilities are not doubled, "another" excludes it, and a second
// Chief doubles the first.
//
// No simplification.
func init() {
	doubler := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{SourceMatch: chiefOfTheWildsDoubles})
	doubler.Label = "Chief of the Wilds"
	doubler.RowLabel = "Triggered abilities of your other Wolves and battles trigger an additional time."
	Register(Spec{
		OracleID:        "bf06d873-08be-4a45-b8fb-fe1dfadc3a39",
		Name:            "Chief of the Wilds",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, chiefOfTheWildsWolfEntered, "Chief of the Wilds — two +1/+1 counters",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 2})),
		},
		TriggerDoublers: []game.TriggerDoubler{doubler},
	})
}

// chiefOfTheWildsWolfEntered is "another Wolf you control enters".
func chiefOfTheWildsWolfEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Wolf"))
}

// chiefOfTheWildsDoubles is "a triggered ability of another Wolf or battle
// you control": the source, as it is or last was, is a battle or a Wolf, and
// is not the Chief itself. DoublesAbilitiesOf has already checked that the
// source is the Chief's controller's.
func chiefOfTheWildsDoubles(_ *game.Game, q game.TriggerDoublingQuery) bool {
	if q.Source.InstanceID == q.Doubler.InstanceID {
		return false
	}
	src := characteristicCard(q.Source.InstanceID, q.Source, q.SourceLKI)
	return src.IsBattle() || src.HasSubtype(q.Doubler.TextSubtype("Wolf"))
}
