package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_or_graveyard_search.go — the "search your library and/or
// graveyard" family (#3217, ADR 0015's 2026-10-13 amendment).
//
// The War of the Spark cycle ("When this creature enters, you may search
// your library and/or graveyard for a card named <Planeswalker>, reveal
// it, and put it into your hand. If you search your library this way,
// shuffle.") is a dozen creatures with one shape, so the trigger is
// built here and each card names only its keywords and its planeswalker.
// The zone behaviour is the engine's: a named card already in the
// graveyard may be taken or left, a card taken from the graveyard skips
// the shuffle, and a card taken from the library (or nothing found)
// shuffles.

// mayFetchNamedFromLibraryOrGraveyard is the "you may search your
// library and/or graveyard for a card named X, reveal it, and put it
// into your hand" enters trigger. `self` is the creature's name, used
// only for the stack and prompt labels. Declining the "you may" neither
// searches nor shuffles.
func mayFetchNamedFromLibraryOrGraveyard(self, named string) game.TriggeredAbility {
	return On(game.EventETB, b06SelfETB, self+" — search for "+named,
		func(g *game.Game, item *game.StackItem) error {
			return MayChoice{
				Question: self + " — search your library and/or graveyard for a card named " + named + "?",
				Then:     maySearchNamedLibraryOrGraveyard,
				Word:     named,
			}.Apply(NewContext(g, item))
		})
}

func init() {
	for _, c := range []struct {
		oracleID, name, fetches string
		keywords                []string
	}{
		{"e32573e8-91a8-438b-966f-1c6d098d54ae", "Goldmane Griffin", "Ajani, Inspiring Leader", []string{"flying", "vigilance"}},
		{"53b17c51-a216-484f-80df-3904284df0ea", "Chandra's Firemaw", "Chandra, Flame's Catalyst", []string{"haste"}},
		{"b3f6e68e-9490-4719-8d2e-d1ec22a96132", "Teferi's Wavecaster", "Teferi, Timeless Voyager", []string{"flash"}},
		{"c4bf4397-22e5-458d-b9d1-3f3a0a27e9e5", "Yanling's Harbinger", "Mu Yanling, Celestial Wind", []string{"flying"}},
	} {
		Register(Spec{
			OracleID:        c.oracleID,
			Name:            c.name,
			Completeness:    CompletenessFull,
			PrintedKeywords: c.keywords,
			Triggered:       []game.TriggeredAbility{mayFetchNamedFromLibraryOrGraveyard(c.name, c.fetches)},
		})
	}
}
