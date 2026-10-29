package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gandalf, Wandering Wizard — Legendary Creature — Avatar Wizard
// {4}{U}, 4/5:
//
//	"Ward {3}
//	 {6}: Gandalf's owner shuffles him into their library and draws
//	 three cards."
//
// The ability is Chaos Warp's tuck-then-shuffle with a draw after the
// shuffle: the OWNER (not the activator, who may have stolen him) shuffles
// the card in and then draws, so the three cards can include Gandalf
// himself. The shuffle and the draw are one keyed tail (thenShuffleThenDraw)
// so a commander-zone offer or any other pause in the move still finishes
// in printed order. If Gandalf has already left the battlefield when the
// ability resolves there is nothing to shuffle in, and the owner draws
// three (CR 608.2c, do as much as possible).
//
// No simplification. The Hobbit (#4454).
func init() {
	Register(Spec{
		OracleID:     "c996bc53-184d-4dc3-8244-e8d9528516a0",
		Name:         "Gandalf, Wandering Wizard",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Gandalf, Wandering Wizard — ward {3}"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}: Gandalf's owner shuffles him into their library and draws three cards.",
			Cost:    ManaCost("{6}"),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				owner := item.Controller
				if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
					owner = c.Owner
				}
				if !onBattlefield(g, item.SourceCardID) {
					return g.DrawNForEffect(owner, 3)
				}
				return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{item.SourceCardID}, game.TuckOptions{},
					game.KeyedTail{Ref: thenShuffleThenDraw, Player: owner, N: 3})
			},
		}},
	})
}

// thenShuffleThenDraw is "Tail.Player shuffles their library, then draws
// Tail.N cards". An on-disk key.
var thenShuffleThenDraw = game.RegisterTail("then/shuffle-then-draw-n", func(g *game.Game, o game.TailOutcome) error {
	if err := g.ShuffleLibraryForEffect(o.Tail.Player); err != nil {
		return err
	}
	return g.DrawNForEffect(o.Tail.Player, o.Tail.N)
})
