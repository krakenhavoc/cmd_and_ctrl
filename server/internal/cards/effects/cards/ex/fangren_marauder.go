package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fangren Marauder — Creature — Beast {5}{G}, 5/5 (Mirrodin Besieged, S82 OLD
// b466, #5576):
//
//	"Whenever an artifact is put into a graveyard from the battlefield, you
//	 may gain 5 life."
//
// Disciple of the Vault's watcher: one optional trigger per artifact that
// goes from the battlefield to a graveyard, anyone's, itself included when it
// is one (it isn't). A bounce or an exile does not count. "You may" is the
// yes/no prompt as the trigger is put on the stack (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e154e612-4041-4fb7-903f-917588dbe58e",
		Name:         "Fangren Marauder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Fangren Marauder: gain 5 life?"},
			Key:            "Fangren Marauder — you may gain 5 life",
			Purpose:        game.Purpose{LifeGain: 5},
			Effect:         effects.Do(effects.GainLife{Amount: 5}),
		}},
	})
}
