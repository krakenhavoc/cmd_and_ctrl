package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siphon Insight — Instant {U}{B} (Outlaws of Thunder Junction
// Commander, #5368):
//
//	"Look at the top two cards of target opponent's library. Exile one
//	 of them face down and put the other on the bottom of that library.
//	 You may play the exiled card for as long as it remains exiled, and
//	 you may spend mana as though it were mana of any color to cast that
//	 spell.
//	 Flashback {1}{U}{B}"
//
// Gonti, Lord of Luxury's look-and-exile over two cards
// (cmdr2S53SiphonInsightLook): the caster alone sees the cards, picks
// the one to exile face down, and the other goes to the bottom of that
// library. "Play", not "cast": an exiled land may be played as the
// caster's land drop. Mana of any COLOR, not any type: the grant is
// AnyColor without AnyType, so a {C} in the card's cost still wants
// colorless mana. The permission lasts for as long as the card stays in
// exile and belongs to the player who cast Siphon Insight. A library of
// one card exiles it with no choice to make.
//
// Flashback is the shared alternative cost from the graveyard
// (CR 702.34a); the flashed-back Siphon Insight is exiled as it leaves
// the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "aee82122-d13e-4175-b499-9dde718e4da9",
		Name:             "Siphon Insight",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target opponent", Opponent()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{1}{U}{B}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return cmdr2S53SiphonInsightLook(ctx.Game, item)
		},
	})
}
