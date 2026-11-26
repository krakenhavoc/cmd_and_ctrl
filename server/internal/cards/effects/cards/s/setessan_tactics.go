package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const setessanTacticsFightGrant = "setessan-tactics/tap-fight"

// Setessan Tactics — Instant {1}{G} (Journey into Nyx):
//
//	"Strive — This spell costs {G} more to cast for each target beyond the first.
//	 Until end of turn, any number of target creatures each get +1/+1 and gain
//	 "{T}: This creature fights another target creature.""
func init() {
	effects.Register(effects.Spec{
		OracleID:     "755e2f3e-bb32-4902-af8d-d80f9899c69f",
		Name:         "Setessan Tactics",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
		Targets:      effects.TargetCreature("any number of target creatures").WithCount(0, 0),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMorePerTargetBeyondFirst("{G}", "Strive — This spell costs {G} more to cast for each target beyond the first."),
		},
		Grants: []effects.AbilityGrant{{
			Key: setessanTacticsFightGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: This creature fights another target creature.",
				Cost:    game.AbilityCost{Tap: true},
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Targets: effects.Another(effects.TargetCreature("another target creature")),
				Effect:  effects.B30SourceFightsFirstLegalTarget,
			}},
			Text: "{T}: This creature fights another target creature.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.S82OldB366LegalTargetCards(ctx) {
				if err := (effects.GrantAbilitiesFor{
					Target: id,
					Keys:   []string{setessanTacticsFightGrant},
					Also:   []game.Mod{game.ModifyPTMod(1, 1)},
					Label:  "Setessan Tactics — +1/+1 and {T}: fights another target creature",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
