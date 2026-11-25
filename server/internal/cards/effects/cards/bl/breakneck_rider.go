package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Breakneck Rider // Neck Breaker — {1}{R}{R} Creature — Human Scout
// Werewolf 3/3 // Creature — Werewolf 4/3 (Shadows over Innistrad, S82 OLD
// b293):
//
//	Front: "At the beginning of each upkeep, if no spells were cast last
//	        turn, transform this creature."
//	Back:  "Attacking creatures you control get +1/+0 and have trample.
//	        At the beginning of each upkeep, if a player cast two or more
//	        spells last turn, transform this creature."
//
// The original-Innistrad werewolf triggers (innistrad_werewolves.go) and
// Instigator Gang's attacking-creatures boost, with trample granted to the
// same set. Each half is its own static: one ability line, two effects, as
// printed.
//
// No simplification.
func init() {
	const oracle = "51685f85-06ad-4de2-a7a6-06e6f59a1cbb"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Breakneck Rider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfNoSpellsLastTurn("Breakneck Rider"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Neck Breaker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AttackingCreaturesYouControlGetPower(1),
			effects.AttackingCreaturesHave("Attacking creatures you control have trample.", nil, "trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.TransformAtUpkeepIfAPlayerCastTwoSpellsLastTurn("Neck Breaker"),
		},
	})
}
