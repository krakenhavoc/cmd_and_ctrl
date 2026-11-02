package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Man, Inferno Incarnate — Legendary Creature — Elemental Villain
// {2}{R}, 0/0 (Marvel's Spider-Man, #4460):
//
//	"When Molten Man enters, search your library for a basic Mountain
//	 card, put it onto the battlefield tapped, then shuffle.
//	 Molten Man gets +1/+1 for each Mountain you control.
//	 When Molten Man leaves the battlefield, sacrifice a land."
//
// The fetch is the basic-land search (Aang's IsBasicLandWithSubtype, read
// through the searching object's text, CR 612.2): the land arrives tapped
// and the library is shuffled even when no Mountain is found. The pump is a
// Layer 7c self-static counting effective Mountains the controller has
// (Desert Were-Worm's count), so a nonbasic Mountain and a land made a
// Mountain count, and Molten Man itself does not unless it is one. It is a
// 0/0 until a Mountain is out: with none under your control when it enters,
// it dies to state-based actions (CR 704.5f) before the trigger resolves,
// as printed. The leaves trigger asks the controller to sacrifice a land of
// their choice (SacrificeChoice); with no land there is nothing to
// sacrifice (CR 701.21a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6fcf3c1-0626-4d53-92c0-caa22fa61e9a",
		Name:         "Molten Man, Inferno Incarnate",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Molten Man gets +1/+1 for each Mountain you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				mountain := source.TextSubtype("Mountain")
				n := permanentsYouControl(g, source, func(o *game.Card) bool { return o.HasSubtype(mountain) })
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Molten Man — search for a basic Mountain card, put it onto the battlefield tapped, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     IsBasicLandWithSubtype("Mountain"),
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Molten Man — a basic Mountain card, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
			WhenThisLeaves("Molten Man — sacrifice a land",
				func(g *game.Game, item *game.StackItem) error {
					return SacrificeChoice{
						Player:     item.Controller,
						Candidates: permanentsControlledByMatching(g, item.Controller, Land()),
						Question:   "Molten Man — sacrifice a land",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
