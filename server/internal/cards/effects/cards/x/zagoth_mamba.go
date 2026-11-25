package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zagoth Mamba — Creature — Nightmare Snake {B}, 1/1 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Whenever this creature mutates, target creature an opponent controls
//	 gets -2/-2 until end of turn."
//
// The mamba has no mutate cost of its own: it triggers when a creature
// spell is merged with it as the target of another card's mutate. Mutate
// is the engine's (ADR 0223) and the trigger is Sawtusk Demolisher's,
// firing on the merge whichever card ends on top (CR 702.140d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aad05e4e-28cf-4dfe-bf5b-277a50d183ba",
		Name:         "Zagoth Mamba",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Zagoth Mamba — target creature an opponent controls gets -2/-2 until end of turn",
				effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Zagoth Mamba — -2/-2"}.Apply(ctx)
				}))),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
