package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loyal Inventor — Creature — Human Artificer {2}{U}, 2/3 (Assassin's Creed,
// #5948):
//
//	"Vigilance
//	 When this creature enters, you may search your library for an artifact
//	 card, reveal it, then shuffle. Put that card into your hand if you
//	 control an Assassin. Otherwise, put that card on top of your library."
//
// Campus Guide's optional, revealed search, with the destination read as the
// search is made: into your hand if you control an Assassin, otherwise on top
// of your library after the shuffle (CR 701.24). The Assassin is looked for
// as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eff22b28-88b8-4251-937f-4c06b6ca7bd2",
		Name:            "Loyal Inventor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Loyal Inventor — search for an artifact card; put it into your hand if you control an Assassin, otherwise on top of your library",
				func(g *game.Game, item *game.StackItem) error {
					assassin := HasSubtype("Assassin")
					controlsAssassin := false
					for _, c := range g.Battlefield.Cards {
						if c.Controller == item.Controller && assassin(g, item.Controller, c) {
							controlsAssassin = true
							break
						}
					}
					search := SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsArtifact() },
						Dest:      game.ZoneHand,
						Limit:     1,
						Optional:  true,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Loyal Inventor — an artifact card",
					}
					if !controlsAssassin {
						search.Dest, search.ToTop = game.ZoneLibrary, true
					}
					return search.Apply(NewContext(g, item))
				}),
		},
	})
}
