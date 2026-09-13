package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Buried Alive — Sorcery for {2}{B}:
//
//	"Search your library for up to three creature cards, put them
//	 into your graveyard, then shuffle."
//
// Entomb three times over at sorcery speed. The three creatures are
// picked as a set, so it is one search, one prompt and one shuffle
// rather than three of each.
//
// The predicate is Card.IsCreature(), which reads the printed type
// line. That is right for a library search: the layer engine
// maintains effective characteristics for battlefield permanents
// only, so a card in a library has nothing but its printed line to
// be judged on.
//
// DECLARED SIMPLIFICATION — "UP TO THREE" IS NOT ALWAYS A CHOICE.
// The search chooser queues a prompt whenever the library holds more
// matches than Limit, and there the searcher really does pick any
// number up to three. In the degenerate case where the library holds
// three or FEWER creature cards, SearchLibraryThenForEffect takes
// them all synchronously with no prompt, so the controller cannot
// choose to bury fewer than everything. Setting Optional would force
// the prompt, but Optional also lets the searcher decline the
// SHUFFLE — and Buried Alive's shuffle is mandatory, so that trade
// makes the card wrong in a way that matters more (a declined
// shuffle leaves the library in a known order). Taking the whole
// bottom of an almost-empty library is what a Buried Alive caster
// would have chosen anyway.
//
// Nothing here is stronger than printed: the forced case takes
// cards the controller asked for, out of their own library, into
// their own graveyard.
func init() {
	Register(Spec{
		OracleID:     "8203c621-a1a0-4865-8c9a-0d4064c86107",
		Name:         "Buried Alive",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If your library holds three or fewer creatures, all of them are put into your graveyard — you cannot choose to bury fewer."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsCreature() },
				Dest:      game.ZoneGraveyard,
				Limit:     3,
				Shuffle:   true,
				Reason:    "Buried Alive — put up to three creature cards into your graveyard",
			}.Apply(ctx)
		},
	})
}
