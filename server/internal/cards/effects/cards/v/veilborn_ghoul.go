package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veilborn Ghoul — Creature — Zombie {4}{B}, 4/1 (Magic 2013, S82 OLD, #5568):
//
//	"This creature can't block.
//	 Whenever a Swamp you control enters, you may return this card from your
//	 graveyard to your hand."
//
// Bloodghast's can't-block static and graveyard-only trigger, over a Swamp
// entering (any land with the Swamp subtype, basic or not) under its owner's
// control, returning to hand. The "you may" is the CR 603.5 prompt; a card that
// has left the graveyard before it resolves does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8381b04a-666b-4e33-b277-6db39fdfe905",
		Name:         "Veilborn Ghoul",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.InGraveyard(effects.On(game.EventETB, effects.S82OldB430SwampEnteredUnderYourControl,
				"Veilborn Ghoul — return it from your graveyard to your hand",
				effects.ReturnThisCardFromYourGraveyardToHand)), "Veilborn Ghoul — return it to your hand?"),
		},
	})
}
