package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kemuri-Onna — Creature — Spirit {4}{B}, 3/3 (Saviors of Kamigawa, #5601):
//
//	"When this creature enters, target player discards a card.
//	 Whenever you cast a Spirit or Arcane spell, you may return this
//	 creature to its owner's hand."
//
// The enters trigger uses the stack. The target player picks the card from their hand (a real discard prompt,
// not a random one); a target that left the game is skipped (CR 608.2b). The return is the cycle's
// shared sokB1OnnaReturn: optional, asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "54db986e-cf8a-4d2c-8ce9-13da11b11464",
		Name:         "Kemuri-Onna",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Kemuri-Onna — target player discards a card", KemuriOnnaDiscard),
				TargetPlayer("target player")),
			SokB1OnnaReturn("Kemuri-Onna"),
		},
	})
}

// kemuriOnnaDiscard asks the still-legal target player to discard a card
// of their choice.
func KemuriOnnaDiscard(g *game.Game, item *game.StackItem) error {
	player, ok := FirstLegalPlayerTarget(NewContext(g, item))
	if !ok {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: item.SourceCardID, N: 1})
	return nil
}
