package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pawpatch Recruit — Creature — Rabbit Warrior {G}, 2/1 (Bloomburrow,
// #4467):
//
//	"Offspring {2}
//	 Trample
//	 Whenever a creature you control becomes the target of a spell or
//	 ability an opponent controls, put a +1/+1 counter on target
//	 creature you control other than that creature."
//
// Offspring is the shared keyword (offspring.go). The trigger reads
// MACO Strike Team's condition: a creature you control became the target
// of a spell or ability whose controller is an opponent. Its own target
// is "other than that creature", which is a fact about the event, so
// the clause is built from the trigger context (TargetsFrom), excluding
// the targeted object by identity. With no other creature you control
// the trigger has no legal target and is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "50ab0194-c7dc-4f11-abc0-1a82393cc42e",
		Name:            "Pawpatch Recruit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		OptionalCosts:   []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Pawpatch Recruit"),
			{
				Watches:     []game.EventKind{game.EventBecomesTarget},
				AppliesTo:   macoYourCreatureTargetedByAnOpponent,
				TargetsFrom: pawpatchRecruitOtherCreature,
				Key:         "Pawpatch Recruit — put a +1/+1 counter on target creature you control other than that creature",
				Effect:      PlusOneCounterOnTheTarget,
			},
		},
	})
}

// pawpatchRecruitOtherCreature is "target creature you control other
// than that creature", where that creature is the one just targeted.
func pawpatchRecruitOtherCreature(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	return TargetCreature("target creature you control other than that creature",
		YouControl(), NotSelf(tc.Event.CardID))
}
