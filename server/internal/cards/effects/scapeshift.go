package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scapeshift — Sorcery {2}{G}{G}:
//
//	"Sacrifice any number of lands. Search your library for up to that
//	 many land cards, put them onto the battlefield tapped, then
//	 shuffle."
//
// The card on the "resolution-time choose N of your own permanents"
// seam row (#1214), and the one that shows why the existing sacrifice
// prompt could not carry it. PlayerSacrificesNForEffect asks for a
// FIXED count, one permanent at a time; "any number" is one question
// over the whole board with a floor of zero, and the sentence after it
// READS THE COUNT — so the clause has to be a continuation rather than
// anything hung off the queue's return value (#1019).
//
// It does not target. A land with hexproof is sacrificed like any
// other, because nothing here is a target, and "up to that many" is
// read off what actually left the battlefield rather than off how many
// the player clicked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c235015e-a7a9-4f8d-bf4a-cf68b3847f83",
		Name:         "Scapeshift",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question:   "Scapeshift — sacrifice any number of lands",
				Candidates: scapeshiftLands,
				Sacrifice:  true,
				Then:       scapeshiftFetch,
			}.Apply(ctx)
		},
	})
}

// scapeshiftLands is every land the chooser controls, with a floor of
// zero and no ceiling but the board — "any number".
//
// Caller holds g.mu.
func scapeshiftLands(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == of && c.IsLand() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 0, len(out)
}

// scapeshiftFetch is the second sentence, run once the lands have
// actually gone: "search your library for up to that many land cards,
// put them onto the battlefield tapped, then shuffle."
//
// `picked.Count()` is what really left the battlefield, not what was
// clicked — the sacrifice went through the one sacrifice path, so a
// sacrificed land whose move was replaced is counted the way the rules
// count it.
//
// Caller holds g.mu.
func scapeshiftFetch(ctx *Context, picked game.PromptedPicks) error {
	n := picked.Count()
	if n == 0 {
		// "Search for up to zero cards, then shuffle." The shuffle is
		// printed unconditionally and still happens.
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return ctx.Game.SearchLibraryThenForEffect(game.SearchLibrarySpec{
		Player:        ctx.Controller(),
		Source:        ctx.Source(),
		Pred:          b03IsLandCard,
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Scapeshift — land cards onto the battlefield tapped",
	})
}
