package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Booby Trap — Artifact {6} (Tempest, S82 OLD batch 766, #5628):
//
//	"As this artifact enters, choose an opponent and a card name other than a
//	 basic land card name.
//	 The chosen player reveals each card they draw.
//	 When the chosen player draws a card with the chosen name, sacrifice this
//	 artifact. If you do, this artifact deals 10 damage to that player."
//
// The as-enters choices are the controller's, stored on the permanent (the
// opponent as its chosen player, the name as its chosen name). The reveal is the
// RevealsChosenPlayersDraws static (#6238) and the trigger reads the drawer and
// the drawn card's name through ChosenPlayerDrawsChosenName. The drawer is
// stamped on the trigger as it goes on the stack, so "that player" is still
// known if the artifact is gone; the damage is dealt by the artifact as it last
// existed, only if it was sacrificed (CR 608.2h).
//
// No simplification.
func init() {
	const label = "Booby Trap — sacrifice this artifact; if you do, it deals 10 damage to that player"
	effects.Register(effects.Spec{
		OracleID:                  "3e46381a-7850-4e50-9213-e1a180299977",
		Name:                      "Booby Trap",
		Completeness:              effects.CompletenessFull,
		AsEnters:                  effects.BoobyTrapEnters("Booby Trap"),
		RevealsChosenPlayersDraws: true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDrawCard},
			AppliesTo: effects.ChosenPlayerDrawsChosenName(),
			Key:       label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = ev.Actor
				return item
			},
			Effect: effects.BoobyTrapStrikes,
		}},
	})
}
