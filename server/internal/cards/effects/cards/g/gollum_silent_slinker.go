package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gollum, Silent Slinker // Meager Meal — an Adventure card (CR 715,
// The Hobbit, #4454), oracle 0296d57e…:
//
//	Gollum, Silent Slinker — Legendary Creature — Halfling Horror {3}{B}, 4/3
//	  "Menace"
//	Meager Meal — Sorcery — Adventure {B}
//	  "Put a +1/+1 counter on up to one target creature. Target player
//	   gains 2 life. (Then exile this card. You may cast the creature
//	   later from exile.)"
//
// Two target clauses, announced together (CR 601.2c): slot 0 is the
// optional creature, slot 1 the player. They are independent: with no
// creature picked the player still gains the life, and a pick that left
// in response gets no counter while the other still resolves
// (CR 608.2b).
//
// No simplification.

// gollumSilentSlinkerOracleID is shared by both faces of the card.
const GollumSilentSlinkerOracleID = "0296d57e-e5b9-456f-bb21-eb584adefb4c"

func init() {
	effects.Register(effects.Spec{
		OracleID:        GollumSilentSlinkerOracleID,
		Name:            "Gollum, Silent Slinker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
	effects.Register(effects.Spec{
		OracleID:     GollumSilentSlinkerOracleID + "#1",
		Name:         "Meager Meal",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.UpToOneTargetCreature("up to one target creature"),
			effects.TargetPlayer("target player"),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetPlayer {
					return effects.GainLife{Player: t.ID, Amount: 2}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
