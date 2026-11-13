package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acquire — Sorcery {3}{U}{U} (Fifth Dawn, #5604):
//
//	"Search target opponent's library for an artifact card and put that
//	 card onto the battlefield under your control. Then that player
//	 shuffles."
//
// Bribery's search for an artifact card: the caster searches the
// opponent's library (SearchLibrary.LibraryOwner), the found card enters
// under the caster's control, and the opponent's library is shuffled.
// Finding nothing is allowed; the library is still shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1c3d749-5cdb-4af2-967f-6413d6cb04fe",
		Name:         "Acquire",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
				return nil
			}
			return SearchLibrary{
				Player:       ctx.Controller(),
				LibraryOwner: item.Targets[0].ID,
				Predicate:    MatchArtifact,
				Dest:         game.ZoneBattlefield,
				Limit:        1,
				Shuffle:      true,
				Reason:       "Acquire — search target opponent's library for an artifact card",
			}.Apply(ctx)
		},
	})
}
