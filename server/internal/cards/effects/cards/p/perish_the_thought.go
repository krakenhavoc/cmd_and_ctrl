package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Perish the Thought — Sorcery {2}{B} (Rise of the Eldrazi, S82 OLD b485):
//
//	"Target opponent reveals their hand. You choose a card from it. That
//	 player shuffles that card into their library."
//
// The revealed-hand pick (ADR 0116) with no filter, so a land may be
// chosen. The card stays in the hand for the pick (PickKeep) and is then
// put into its owner's library through the tuck route, with the owner's
// library shuffled once it has arrived (a commander's CR 903.9 prompt is
// answered first). An empty hand chooses nothing, and the opponent still
// shuffles nothing: no card, no shuffle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8955996a-76c9-496a-9238-5fd89bcdaae4",
		Name:         "Perish the Thought",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Label:  "card",
				Keep:   true,
				Then:   perishTheThoughtShuffle,
			}.Apply(ctx)
		},
	})
}

// perishTheThoughtShuffle puts the chosen card into its owner's library and
// shuffles that library. An on-disk key: never renamed, never reused.
var perishTheThoughtShuffle = effects.RevealedPickThen("revealed-pick/perish-the-thought-shuffle",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		if len(pick.Chosen) == 0 {
			return nil
		}
		then := game.KeyedTail{Ref: perishTheThoughtShuffled, Player: pick.Chooser, Source: pick.Source,
			Carry: []uuid.UUID{pick.FromPlayer}}
		return effects.PutIntoLibrary{Card: pick.Chosen[0].InstanceID, ThenTail: then}.Apply(ctx)
	})

// perishTheThoughtShuffled is the shuffle, run once the card has reached the
// library (or been sent elsewhere): the player who revealed the hand shuffles.
var perishTheThoughtShuffled = effects.CardTail("then/perish-the-thought-shuffled",
	func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return nil
		}
		return ctx.Game.ShuffleLibraryForEffect(o.Tail.Carry[0])
	})
