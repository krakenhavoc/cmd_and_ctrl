package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Technodrome — {2} Artifact Creature — Construct 3/3 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Reach, trample
//	 This creature can't attack or block unless its power is 6 or
//	 greater.
//	 {T}, Sacrifice another artifact: Draw a card. Put a +1/+1 counter
//	 on this creature."
//
// The restriction is #4678's power floor: the static records 6 and the
// engine adds CantAttack | CantBlock after every layer has run while
// the finished power is below it (CR 613.11, game/power_floor.go), so
// counters and anthems count and a -N/-N takes it back out. Judged at
// declaration (CR 508.1c, 509.1b), like every restriction. The
// activated ability is Arcbound Ravager's counter after a draw, with
// Hungering Puppetbeast's "another artifact" sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51c0b46c-c379-420a-a4a4-d18724c759aa",
		Name:            "Technodrome",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Static:          []game.StaticAbility{CantAttackOrBlockUnlessPowerAtLeast(6)},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice another artifact: Draw a card. Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(TapCost(), SacrificeAnotherN(1, "another artifact", Artifact())),
			Effect:  Do(DrawCards{N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
