package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx Summoner — Artifact Creature — Sphinx {3}{U}{B}, 3/3 (Conflux,
// #5587):
//
//	"Flying
//	 When this creature enters, you may search your library for an
//	 artifact creature card, reveal it, put it into your hand, then
//	 shuffle."
//
// Avarax's optional tutor (CR 603.5) for an artifact creature card. A
// search for a card with a quality may fail to find one (CR 701.23b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77ae7caa-e180-4efb-9348-ab389364b186",
		Name:            "Sphinx Summoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Sphinx Summoner — search your library for an artifact creature card",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() && c.IsCreature() },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Sphinx Summoner — an artifact creature card",
					}.Apply(NewContext(g, item))
				}),
				"Search your library for an artifact creature card?"),
		},
	})
}
