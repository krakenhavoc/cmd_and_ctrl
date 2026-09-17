package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Confidant — Creature — Human Wizard {1}{B}, 2/1:
//
//	"At the beginning of your upkeep, reveal the top card of your
//	 library and put it into your hand. You lose life equal to its
//	 mana value."
//
// Bob. The two-mana card advantage engine that charges you in life
// for cards you did not choose, which is why a deck playing him keeps
// its curve low and why he is a Commander staple rather than a
// Commander auto-include.
//
// # Reveal, then move, then pay — and the order is the card
//
// REVEALING IS NOT DRAWING. Nothing here emits EventDrawCard, so a
// Psychosis Crawler or a Niv-Mizzet sitting next to Bob does not fire
// on the upkeep flip, and a replacement effect that watches draws
// does not see it either. That is exactly what the card prints, and
// it is the difference a draw-payoff deck notices first.
//
// The reveal happens FROM THE LIBRARY, before anything moves
// (CR 701.20b — a reveal is not a zone change). #549's frame is what
// makes that observable: without it the flip would be a card
// appearing in a hidden zone with no announcement, and the whole
// table is entitled to know what Bob just charged its controller for.
// The knowledge is sticky, so the revealed card stays readable to
// every seat while it sits in that hand — see game/reveal.go.
//
// MANA VALUE IS READ BEFORE THE MOVE, for the reason
// reanimateSingleTarget reads its card before moving it: that is the
// last moment the card is guaranteed findable where the effect put
// it. The number is the same in either zone (CR 202.3e — {X} is zero
// anywhere but the stack), so this is about not losing the card, not
// about picking a zone to read it in.
//
// A LAND FLIP COSTS NOTHING and an empty library reveals nothing and
// is not an error — revealing is not drawing, so bottoming out this
// way does not set up the CR 704.5b loss. Bob at an empty library
// simply stops doing anything, which is what he does in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2068185c-1b50-47d0-aa3f-bf505d199428",
		Name:         "Dark Confidant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			// "YOUR upkeep" — the drawer is the Confidant's
			// controller, and the step's Actor is the active
			// player.
			AtYourUpkeep("Dark Confidant — reveal the top card of your library and put it into your hand", func(g *game.Game, item *game.StackItem) error {
				return darkConfidantFlip(g, item)
			}),
		},
	})
}

// darkConfidantFlip is the whole trigger: reveal one off the top,
// read what it costs, put it in hand, then pay.
func darkConfidantFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var revealed []uuid.UUID
	if err := (RevealTopOfLibrary{
		Player:   item.Controller,
		N:        1,
		Reason:   "Dark Confidant — reveal the top card of your library",
		Revealed: &revealed,
	}).Apply(ctx); err != nil {
		return err
	}
	if len(revealed) == 0 {
		// An empty library. Nothing was revealed, so there is
		// nothing to put into hand and nothing to pay for.
		return nil
	}
	flipped := revealed[0]
	life := 0
	if c, ok := g.LookupCardForEffect(flipped); ok {
		life = c.ManaValue()
	}
	if err := (BounceToHand{Target: flipped}).Apply(ctx); err != nil {
		return err
	}
	if life == 0 {
		// A land, or a card with no mana cost. The clause still
		// happened; it just cost nothing.
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -life)
}
