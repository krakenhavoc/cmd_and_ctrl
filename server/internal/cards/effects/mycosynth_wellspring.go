package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mycosynth Wellspring — Artifact {2} (EDHREC rank 2419):
//
//	"When this artifact enters or is put into a graveyard from the
//	 battlefield, you may search your library for a basic land card,
//	 reveal it, put it into your hand, then shuffle."
//
// The artifact-sacrifice deck's land-smoother — two searches from
// one two-drop. One printed ability with two trigger conditions is
// ONE TriggeredAbility watching two kinds (the Sun Titan posture):
// its own EventETB, and its own EventLTB bound for a graveyard
// (cardDied — a bounce or an exile is not "put into a graveyard").
// "You may" is the search prompt's decline; the pick is revealed and
// goes to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ec43ec6-b625-4be8-8f79-3679e6657dbc",
		Name:         "Mycosynth Wellspring",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventLTB}, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b22SelfEnteredOrDied(ev, source)
			}, "Mycosynth Wellspring — search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Mycosynth Wellspring — a basic land card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
