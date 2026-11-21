package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Status // Statue — split card, oracle df7d4964-35a2-42f7-a4f5-05122f78cbba
// (Ravnica: Clue Edition):
//
//	Status {B/G} Instant: "Target creature gets +1/+1 and gains deathtouch
//	 until end of turn."
//	Statue {2}{B}{G} Instant: "Destroy target artifact, creature, or
//	 enchantment."
//
// Two specs, Statue under "<oracle_id>#1". Status is Massive Might's pump
// and grant; Statue is a plain destroy on a three-type target clause.
//
// No simplification.
func init() {
	const oracle = "df7d4964-35a2-42f7-a4f5-05122f78cbba"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Status",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "deathtouch")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(1, 1, "Status — +1/+1 and deathtouch", "deathtouch")(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Statue",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact, creature, or enchantment", effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id := effects.FirstLegalBattlefieldTarget(ctx); id != [16]byte{} {
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			}
			return nil
		},
	})
}
