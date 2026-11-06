package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feral Deathgorger // Dusk Sight — Tarkir: Dragonstorm (#4463), an omen
// card (CR 720, #5175):
//
//	Feral Deathgorger — Creature — Dragon {5}{B}, 3/5
//	  "Flying, deathtouch
//	   When this creature enters, exile up to two target cards from a
//	   single graveyard."
//	Dusk Sight — Sorcery — Omen {1}{B}
//	  "Put a +1/+1 counter on up to one target creature. Draw a card.
//	   (Then shuffle this card into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). The enters trigger
// is Arashin Sunshield's "from a single graveyard" (ADR 0106 §5).
//
// No simplification.

const feralDeathgorgerOracleID = "1722f814-2d4e-4d59-9b74-10ef8e89def9"

func init() {
	Register(Spec{
		OracleID:        feralDeathgorgerOracleID,
		Name:            "Feral Deathgorger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersExileFromASingleGraveyard("Feral Deathgorger", 2, ExileTargetCards),
		},
	})
	Register(Spec{
		OracleID:     feralDeathgorgerOracleID + "#1",
		Name:         "Dusk Sight",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve:    effectThenDrawACard(PlusOneCounterOnTheTarget),
	})
}
