package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Assure // Assemble — split card {G/W}{G/W} // {4}{G}{W} (Ravnica
// Remastered):
//
//	Assure (Instant): "Put a +1/+1 counter on target creature. That creature
//	 gains indestructible until end of turn."
//	Assemble (Instant): "Create three 2/2 green and white Elf Knight creature
//	 tokens with vigilance."
//
// The counter and the grant are two instructions on the same target; each
// does nothing if the creature has left (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "087018b6-6355-4534-9a35-9336b4b7fccc",
		Name:         "Assure",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.PutPlusOneCounterOnTheTarget),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"},
					Label: "Assure — indestructible until end of turn"}.Apply(ctx)
			}),
		),
	})
	effects.Register(effects.Spec{
		OracleID:     "087018b6-6355-4534-9a35-9336b4b7fccc#1",
		Name:         "Assemble",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("2/2 green and white Elf Knight with vigilance"), N: 3}),
	})
}
