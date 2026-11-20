package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Creeperhulk — Creature — Plant Elemental {3}{G}{G}, 5/5 (Commander 2014,
// #5553):
//
//	"Trample
//	 {1}{G}: Until end of turn, target creature you control has base power
//	 and toughness 5/5 and gains trample."
//
// Eidolon of Astral Winds' shape as an activated ability: the base P/T is a
// layer 7b setting under counters and pumps (CR 613.4b), and the trample a
// keyword grant, both until end of turn on the one target still legal as it
// resolves. Creeperhulk may target itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c6423251-759a-48f1-a9ae-b49b98172c08",
		Name:            "Creeperhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}: Until end of turn, target creature you control has base power and toughness 5/5 and gains trample.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{G}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  creeperhulkBecomesFiveFive,
		}},
	})
}

// creeperhulkBecomesFiveFive makes the target a 5/5 with trample until end
// of turn.
func creeperhulkBecomesFiveFive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	const label = "Creeperhulk — base power and toughness 5/5 and trample until end of turn"
	if err := untilEndOfTurn(ctx, target, nil, label, game.SetBasePTMods(5, 5)...); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: target, Keywords: []string{"trample"}, Label: label}.Apply(ctx)
}
