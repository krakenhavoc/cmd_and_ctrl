package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fuss // Bother — split card (Murders at Karlov Manor, #4469):
//
//	Fuss — Instant {2}{R/W}: "Put a +1/+1 counter on each attacking
//	       creature you control."
//	Bother — Sorcery {4}{W/U}{W/U}: "Create three 1/1 colorless Thopter
//	         artifact creature tokens with flying. Surveil 2."
//
// Each half is an ADR 0034 face (ADR 0103), as Cease // Desist's are:
// Fuss under the bare oracle ID, Bother under "#1". Only one half can
// be cast (CR 709.3; the card has no fuse).
//
// Fuss is Drana, Liberator of Malakir's counters: the creatures you
// control that are attacking as it resolves, each getting one. Outside
// combat there are none and it does nothing. Bother makes the Thopters,
// then surveils (CR 701.25a); the surveil is the last instruction, so
// nothing waits on its answer.
//
// No simplification.
const fussBotherOracleID = "88e9006a-cfdf-40f3-80d5-23465f3dceca"

func init() {
	effects.Register(effects.Spec{
		OracleID:     fussBotherOracleID,
		Name:         "Fuss",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B13PutCounterOnEach(ctx, effects.B13AttackingCreaturesYouControl(ctx.Game, ctx.Controller()))
		},
	})
	effects.Register(effects.Spec{
		OracleID:     fussBotherOracleID + "#1",
		Name:         "Bother",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 3}).Apply(ctx); err != nil {
				return err
			}
			return effects.Surveil{N: 2}.Apply(ctx)
		},
	})
}
