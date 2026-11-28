package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Psychic Theft — Sorcery {1}{U} (Prophecy, S82 OLD batch 673, #5617):
//
//	"Target player reveals their hand. You choose an instant or sorcery card
//	 from it and exile that card. You may cast that card for as long as it
//	 remains exiled. At the beginning of the next end step, if you haven't cast
//	 the card, return it to its owner's hand."
//
// Psychic Intrusion's revealed-hand pick, narrowed to instant and sorcery
// cards in the hand only, then a cast-only permission for the chooser for as
// long as the card remains exiled (mana as normal, the card's own cost). The
// delayed trigger is Mnemonic Betrayal's: scheduled for the next end step over
// the exiled card, it returns the card to its owner's hand if it is still in
// exile. A card that was cast has left exile and is not returned. A hand with
// no instant or sorcery card is still revealed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66953dfd-d908-41d3-abb4-309c6a165dc3",
		Name:         "Psychic Theft",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Filter: effects.Or(effects.Instant(), effects.Sorcery()),
				Label:  "instant or sorcery card",
				Exile:  true,
				Then:   psychicTheftMayCast,
			}.Apply(ctx)
		},
	})
}

// psychicTheftMayCast is the "you may cast that card" continuation. An on-disk
// key: never renamed, never reused.
var psychicTheftMayCast = effects.S82OldB673PsychicTheftPick("revealed-pick/psychic-theft-may-cast", "Psychic Theft")
