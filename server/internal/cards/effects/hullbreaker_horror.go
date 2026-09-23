package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// hullbreakerHorrorLabel is the stack label for the "whenever you
// cast a spell" trigger's item.
const hullbreakerHorrorLabel = "Hullbreaker Horror — choose up to one"

// Hullbreaker Horror — Creature — Kraken Horror {5}{U}{U}, 7/8
// (EDHREC rank ~360):
//
//	"Flash
//	 This spell can't be countered.
//	 Whenever you cast a spell, choose up to one —
//	 • Return target spell you don't control to its owner's hand.
//	 • Return target nonland permanent to its owner's hand."
//
// # Modal trigger, real mode prompt (#764)
//
// Roadmap batch 02 (#295) filed this under "a modal and multi-target
// clause", which #764 closed: TriggeredAbility.Modes is the same
// game.ModeSpec a modal spell declares, asked as a mode_pick prompt at
// CR 603.3c — after the harvest, before targets, a full priority round
// before resolution. "Choose up to one" is ChooseN's (0, 1) bound:
// declining is a legal answer, exactly as Ertai Resurrected's own
// "up to one" bullet is, except this one is a REAL mode choice rather
// than that older card's "the mode is the target" workaround — #764
// postdates Ertai's file and gave a trigger the same picker a modal
// spell has always had.
//
// Two bullets, two different target zones, so they cannot share one
// clause the way Kolaghan's Command's same-zone pair can (Aang, Swift
// Savior's two-zone shape is the nearer relative): the first is
// ReturnSpellToHand over the stack (bounce, not counter — the card
// never says "counter", so a spell printed "can't be countered" is
// still a legal target and nothing watching "whenever a spell is
// countered" fires, Reprieve's own point); the second is
// BounceTheModesTarget, the shared body Mystic Confluence and Sublime
// Epiphany already call, over the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4a84e78-d9b9-4c67-8a4b-4329e65f0f15",
		Name:            "Hullbreaker Horror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CantBeCountered: true,
		Triggered:       []game.TriggeredAbility{hullbreakerHorrorTrigger()},
	})
}

// hullbreakerHorrorTrigger declares the modal "whenever you cast a
// spell" ability. Each bullet supplies its own body (ModeDoing), so
// the base Effect the constructor takes never runs anything itself —
// the engine dispatches whichever bullet was chosen, in announce
// order, exactly as Gala Greeters' alliance trigger does.
func hullbreakerHorrorTrigger() game.TriggeredAbility {
	t := WheneverYouCast(nil, hullbreakerHorrorLabel,
		func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = ChooseN("Choose up to one", 0, 1,
		ModeDoing("Return target spell you don't control to its owner's hand.",
			TargetSpell("target spell you don't control", OpponentControls()),
			hullbreakerHorrorBounceSpell),
		ModeDoing("Return target nonland permanent to its owner's hand.",
			TargetPermanent("target nonland permanent", Nonland()),
			BounceTheModesTarget),
	)
	return t
}

// hullbreakerHorrorBounceSpell is the first bullet's body: a bounce,
// not a counter (ReturnSpellToHand, not CounterTarget) — the printed
// text never says "counter".
func hullbreakerHorrorBounceSpell(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ReturnSpellToHand{StackID: t.ID}.Apply(ctx)
}
