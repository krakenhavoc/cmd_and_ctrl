package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfall Invocation — Sorcery {3}{W}{W} (EDHREC rank 2210):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Destroy all creatures. If the gift was promised, return a
//	 creature card put into your graveyard this way to the battlefield
//	 under your control."
//
// Wrath of God that can keep its best creature for the price of a
// card to an opponent. The wipe is DestroyAllMatching (S23): every
// creature dies as one event, so the aristocrats payoffs see all of
// them.
//
// Sandbox simplification, declared — Long River's Pull's posture: the
// GIFT is not offered. Gift is a cast-time promise (CR 702.174) that
// needs a prompt in the cast flow and a per-cast flag the resolution
// reads, and neither exists. So the spell is always its base mode —
// destroy all creatures — which is the printed card with one option
// removed, weaker and never stronger. When a cast-time promise lands,
// the second clause is a draw for the promisee plus a pick among the
// controller's creature cards the wipe put into the graveyard.
//
// And the catalog-wide one every board wipe carries (#446): the mass
// destroy path does not consult indestructible, so an indestructible
// creature dies to this too. Declared until the engine fix lands.
func init() {
	Register(Spec{
		OracleID:     "7024532b-f99b-43a7-b0ed-5b3e7ec7592b",
		Name:         "Starfall Invocation",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The gift can't be promised, so the spell only destroys all creatures — it never returns one of yours to the battlefield.",
			"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Creature()}.Apply(ctx)
		},
	})
}
