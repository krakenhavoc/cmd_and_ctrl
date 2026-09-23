package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clever Concealment — Instant {2}{W}{W}:
//
//	"Convoke (Your creatures can help cast this spell. Each creature
//	 you tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Any number of target nonland permanents you control phase out.
//	 (Treat them and anything attached to them as though they don't
//	 exist until your next turn.)"
//
// The modern Teferi's Protection for the board half only, and the
// card that proves the MASS shape of CR 702.26: a whole side of the
// table leaves at once, simultaneously, dragging every Aura and
// Equipment with it (CR 702.26g), and comes back at the start of your
// next turn with its counters, its damage and its attachments intact
// (CR 702.26d).
//
// Convoke is the S22 tap cost and pairs with the effect the way the
// designers meant: the creatures you tap to cast it are the creatures
// you are saving, and they phase out tapped and phase in tapped —
// then untap in the same untap step, because CR 502.1's phase-in runs
// before CR 502.3's untap. Nothing in this file arranges that; it is
// the order of the step's turn-based actions.
//
// "ANY NUMBER OF TARGET" is Min 0, Max unbounded — the Deepglow Skate
// shape — so the caster picks, and a target that stopped being legal
// in response is dropped per slot (CR 608.2b) rather than fizzling
// the spell. Zero targets is a legal announcement and a spell that
// does nothing, as printed.
//
// NONLAND, so a Clever Concealment does not hide your mana base: your
// lands stay, which is the card's one real cost and the reason it is
// not simply better than a Fog.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "42bb7ea9-f6e4-4551-8d93-3b1eae84b865",
		Name:         "Clever Concealment",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets: TargetPermanent("any number of target nonland permanents you control",
			Nonland(), YouControl()).WithCount(0, 0),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// ONE call with the whole list, not one call per target:
			// CR 702.26a phases them out simultaneously, and an
			// Equipment named beside the creature it is attached to
			// must not be dragged out twice.
			return PhaseOut{Targets: legalTargetCards(item, ctx.Game)}.Apply(ctx)
		},
	})
}
