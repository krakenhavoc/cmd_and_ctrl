package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prepare // Fight — Instant // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Prepare {1}{W}: "Untap target creature. It gets +2/+2 and gains lifelink
//	 until end of turn."
//	Fight {3}{G}: "Aftermath. Target creature you control fights target creature
//	 an opponent controls."
//
// Two registered faces. Prepare untaps, then pumps and grants lifelink to the one
// target, all until end of turn. Fight is the two-clause fight (CR 701.14): each
// creature deals damage equal to its power to the other, and nothing happens if
// either is gone; it is cast only from a graveyard and exiled as it resolves
// (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fed7a6ad-68d8-4923-9ef0-619f73bfd6c2",
		Name:         "Prepare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "lifelink")), Answers: game.AnswerPump},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Prepare — +2/+2 until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Prepare — lifelink"}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "fed7a6ad-68d8-4923-9ef0-619f73bfd6c2#1",
		Name:         "Fight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.FightTargetClauses)),
	})
}
