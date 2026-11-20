package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dominion Assault Force — Creature — Jem'Hadar Warrior {3}{G}, 4/4
// (Star Trek, #4973):
//
//	"Vigilance
//	 When this creature enters, choose one —
//	 • Exile target card from a graveyard.
//	 • Surveil 1."
//
// Federation Probe's modal enters trigger: the mode and its target are
// chosen as the trigger goes on the stack, and a graveyard card that left
// in response is not exiled (CR 608.2b). Surveil is CR 701.25a.
//
// No simplification.
func init() {
	etb := effects.WhenThisEnters("Dominion Assault Force — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = effects.ChooseOne(
		effects.ModeDoing("Exile target card from a graveyard.",
			effects.TargetCardInGraveyard("target card from a graveyard"),
			effects.ExileTheModesTarget),
		effects.ModeDoing("Surveil 1.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Surveil{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:        "b093916c-cc0a-4b3b-9de6-8a2db53cdfe0",
		Name:            "Dominion Assault Force",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
