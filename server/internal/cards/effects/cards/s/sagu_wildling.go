package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sagu Wildling // Roost Seek — Tarkir: Dragonstorm (#4463), an omen
// card (CR 720, #5175):
//
//	Sagu Wildling — Creature — Dragon {4}{G}, 3/3
//	  "Flying
//	   When this creature enters, you gain 3 life."
//	Roost Seek — Sorcery — Omen {G}
//	  "Search your library for a basic land card, reveal it, put it into
//	   your hand, then shuffle. (Also shuffle this card.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle of the card itself is the engine's (CR 720.3d, game/omen.go).
//
// No simplification.

const SaguWildlingOracleID = "7cee6d43-5dec-4989-85ec-c635ec783ec7"

func init() {
	effects.Register(effects.Spec{
		OracleID:        SaguWildlingOracleID,
		Name:            "Sagu Wildling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sagu Wildling — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     SaguWildlingOracleID + "#1",
		Name:         "Roost Seek",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Roost Seek — a basic land card",
			}.Apply(ctx)
		},
	})
}
