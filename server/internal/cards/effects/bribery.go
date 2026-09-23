package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bribery — Sorcery {3}{U}{U}:
//
//	"Search target opponent's library for a creature card and put
//	 that card onto the battlefield under your control. Then that
//	 player shuffles."
//
// The "Search / library-interaction primitives" row's non-owner-
// searcher half (docs/engine-seams.md): the search is conducted by
// the CASTER, not the searched library's owner, and the found
// creature enters under the CASTER's control rather than its own
// owner's — "the searcher is always the library's owner; no
// control-changing entry" was this card's declared skip on batch 28
// (#390).
//
// SearchLibrarySpec.LibraryOwner (#1230) is the one new field this
// needed. Every OTHER field on the spec already meant what Bribery
// wants once Player is read as "the caster" throughout: Player is
// who CHOOSES, who the fetched permanent's Controller is stamped as
// (searchEnterBattlefieldLocked), and who the move's Actor is — all
// three already correct for "under your control" with no change.
// LibraryOwner only redirects which pile is scanned, offered and
// shuffled, which is the opponent's.
//
// Aven Mindcensor is the row's other named card and remains open: its
// "search the top four cards instead" is a static REPLACEMENT on
// every search an affected player makes, board-wide, across every
// card in the catalog — not a parameter this or any other single
// search call can carry. Left declared on the tracking issue rather
// than invented here.
func init() {
	Register(Spec{
		OracleID:     "6d194882-ca37-49bb-ac9f-a751c53850a8",
		Name:         "Bribery",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			victim := item.Targets[0].ID
			return SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: victim,
				Predicate:    func(c game.Card) bool { return c.IsCreature() },
				Dest:         game.ZoneBattlefield,
				Limit:        1,
				Reveal:       true,
				Shuffle:      true,
				Reason:       "Bribery — search target opponent's library for a creature card",
			}.Apply(ctx)
		},
	})
}
