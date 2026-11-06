package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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

const saguWildlingOracleID = "7cee6d43-5dec-4989-85ec-c635ec783ec7"

func init() {
	Register(Spec{
		OracleID:        saguWildlingOracleID,
		Name:            "Sagu Wildling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sagu Wildling — you gain 3 life", Do(GainLife{Amount: 3})),
		},
	})
	Register(Spec{
		OracleID:     saguWildlingOracleID + "#1",
		Name:         "Roost Seek",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Roost Seek — a basic land card",
			}.Apply(ctx)
		},
	})
}
