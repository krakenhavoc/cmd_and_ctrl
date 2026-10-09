package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enlightened Confidant — Creature — Kor Cleric {1}{W}, 2/1:
//
//	"Lifelink
//	 At the beginning of your end step, if you gained life this turn,
//	 surveil 1. If you put a card with mana value less than or equal to
//	 the amount of life you gained this turn into your graveyard this
//	 way, put that card into your hand."
//
// An intervening-if end-step trigger (CR 603.4): asked as the end step
// begins and again as the trigger resolves. The life total is the turn's
// tally (b15LifeGainedThisTurn), the card is the one that was on top when
// the surveil began, and it comes back only if it really reached the
// graveyard — a replacement that exiled it instead leaves it alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a139eb7f-3853-490c-a839-35d0e5aed7b5",
		Name:            "Enlightened Confidant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b15LifeGainedThisTurn(g, source.Controller) > 0
			}, "Enlightened Confidant — surveil 1; a cheap enough card goes to hand", rfMiscAConfidantEndStep),
		},
	})
}
