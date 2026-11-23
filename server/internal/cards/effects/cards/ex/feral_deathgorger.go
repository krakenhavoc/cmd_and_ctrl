package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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

const FeralDeathgorgerOracleID = "1722f814-2d4e-4d59-9b74-10ef8e89def9"

func init() {
	effects.Register(effects.Spec{
		OracleID:        FeralDeathgorgerOracleID,
		Name:            "Feral Deathgorger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersExileFromASingleGraveyard("Feral Deathgorger", 2, effects.ExileTargetCards),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     FeralDeathgorgerOracleID + "#1",
		Name:         "Dusk Sight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve:    effects.EffectThenDrawACard(effects.PlusOneCounterOnTheTarget),
	})
}
