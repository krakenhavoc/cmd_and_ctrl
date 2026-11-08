package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thought-Stalker Warlock — Creature — Lizard Warlock {2}{B}, 2/2
// (Bloomburrow, #4467):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 When this creature enters, choose target opponent. If they lost life
//	 this turn, they reveal their hand, you choose a nonland card from it,
//	 and they discard that card. Otherwise, they discard a card."
//
// "If they lost life this turn" is a clause of the effect, not an
// intervening if, so it is read as the trigger resolves, off the per-turn
// life-lost tally (damage, paid life or a "loses life" effect). The first
// branch is Thoughtseize's ChooseFromRevealedHand over a nonland card
// (CR 701.9b: the Warlock's controller chooses); a hand with no nonland
// card is still revealed and nothing is discarded. The second is the
// plain discard, the opponent's own choice. Menace rides PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25a63171-a1a0-4c61-ac61-49bb59fe83c1",
		Name:            "Thought-Stalker Warlock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Thought-Stalker Warlock — target opponent discards (you choose a nonland card if they lost life this turn)",
				thoughtStalkerWarlockDiscard),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// thoughtStalkerWarlockDiscard is the entry trigger's effect.
func thoughtStalkerWarlockDiscard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := TargetedPlayer(ctx)
	if victim == uuid.Nil {
		return nil
	}
	if b18LifeLostThisTurn(g, victim) > 0 {
		return ChooseFromRevealedHand{Player: victim, Filter: Nonland(), Label: "nonland card"}.Apply(ctx)
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: victim, Source: item.SourceCardID, N: 1})
	return nil
}
