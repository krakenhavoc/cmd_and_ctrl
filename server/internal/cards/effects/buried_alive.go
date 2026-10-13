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
// "Up to three" is a real choice at any library size: a search that names
// a quality may find nothing and an "up to N" one may find fewer (CR
// 701.23b, #3053), so the searcher is asked even when the library holds
// three or fewer creature cards, and the shuffle stays mandatory. The
// caveat this card carried (the library's last creatures were taken
// without asking) described the engine before that change; a test now
// pins the prompt for a library of two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8203c621-a1a0-4865-8c9a-0d4064c86107",
		Name:         "Buried Alive",
		Purpose:      game.Purpose{SelfMillTutor: 3},
		Completeness: CompletenessFull,
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
