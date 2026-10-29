package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const gollumSilentSlinkerOracleID = "0296d57e-e5b9-456f-bb21-eb584adefb4c"

func init() {
	Register(Spec{
		OracleID:        gollumSilentSlinkerOracleID,
		Name:            "Gollum, Silent Slinker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
	Register(Spec{
		OracleID:     gollumSilentSlinkerOracleID + "#1",
		Name:         "Meager Meal",
		Completeness: CompletenessFull,
		Targets: Clauses(
			UpToOneTargetCreature("up to one target creature"),
			TargetPlayer("target player"),
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetPlayer {
					return GainLife{Player: t.ID, Amount: 2}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}
