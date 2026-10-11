package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// High-Society Hunter — 5/3 Creature — Vampire Noble for {3}{B}{B}
// (EDHREC rank 4427):
//
//	"Flying
//	 Whenever this creature attacks, you may sacrifice another
//	 creature. If you do, put a +1/+1 counter on this creature.
//	 Whenever another nontoken creature dies, draw a card."
//
// A five-mana evasive body that turns the whole table's attrition into
// cards. The second trigger is the reason it plays in Commander: it
// watches ANY nontoken creature die, anyone's, so every board wipe and
// every combat the table has without you is a fistful of cards.
// Roadmap batch 42 (#449), "no new machinery".
//
// "Another NONTOKEN creature" excludes tokens, which is the printed
// card's own brake — the Hunter is not a sacrifice-deck payoff, it is
// a grindy-table payoff, and a Bitterblossom board does nothing for
// it. It also excludes itself, so the Hunter's own death draws
// nothing.
//
// The attack trigger's counter follows ONLY from the sacrifice: "if
// you do" is a real condition, and declining the prompt leaves the
// Hunter a 5/3.
//
// The sacrifice is chosen as the trigger resolves, not as it goes on
// the stack (CR 608.2c, 701.21): the trigger targets nothing, asks
// "you may sacrifice another creature" through MayChoice, and then a
// SacrificeChoice over the creatures you control at that moment. So an
// opponent's response cannot take the choice away, and a hexproof or
// shroud creature is a legal pick. With no other creature there is no
// question and no counter; a "No" leaves the Hunter a 5/3. The counter
// is the sacrifice's continuation, so it follows only a sacrifice that
// happened, and AddCounter does nothing if the Hunter has left the
// battlefield and come back as a new object (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0dc158ba-7ec8-4558-91f5-0a87ef8380d4",
		Name:            "High-Society Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("High-Society Hunter — you may sacrifice another creature; if you do, put a +1/+1 counter on this creature",
				hunterSacrifice),
			On(game.EventLTB,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					if !AnotherCreatureDied(ev, source, lki, g) {
						return false
					}
					dead, ok := diedCreature(ev, g)
					return ok && !IsToken(dead)
				},
				"High-Society Hunter — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}

// hunterSacrifice is the attack trigger's resolution: the optional
// sacrifice, and the counter only once a creature has gone.
func hunterSacrifice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(hunterFodder(g, item)) == 0 {
		return nil
	}
	return MayChoice{
		Question: "High-Society Hunter — sacrifice another creature to put a +1/+1 counter on it?",
		OnYes: func(ctx *Context) error {
			// Recomputed after the answer: the board can change while
			// the question is open.
			fodder := hunterFodder(ctx.Game, item)
			if len(fodder) == 0 {
				return nil
			}
			return SacrificeChoice{
				Player:     item.Controller,
				Candidates: fodder,
				Question:   "High-Society Hunter — sacrifice another creature",
				Then: func(ctx *Context) error {
					return AddCounter{
						Target: item.SourceCardID,
						Kind:   game.CounterPlusOne,
						N:      1,
					}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	}.Apply(ctx)
}

// hunterFodder is every creature the trigger's controller controls
// except the Hunter itself.
func hunterFodder(g *game.Game, item *game.StackItem) []uuid.UUID {
	var out []uuid.UUID
	for _, id := range permanentsControlledByMatching(g, item.Controller, Creature()) {
		if id != item.SourceCardID {
			out = append(out, id)
		}
	}
	return out
}
