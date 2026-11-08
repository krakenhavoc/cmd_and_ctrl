package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shrouded Shepherd // Cleave Shadows — Creature — Spirit Warrior {1}{W},
// 2/2 // Sorcery — Adventure {1}{B} (Wilds of Eldraine, #4471):
//
//	Shrouded Shepherd — "When this creature enters, target creature you
//	 control gets +2/+2 until end of turn."
//	Cleave Shadows — "Creatures your opponents control get -1/-1 until end
//	 of turn. (Then exile this card. You may cast the creature later from
//	 exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The entry pump is Fang Guardian's targeted +2/+2, and the
// Shepherd may target itself. Cleave Shadows is Seeker's Folly's -1/-1
// over the creatures opponents control as it resolves (CR 611.2c), so it
// declares the same opponents-only sweep.
//
// No simplification.
const shroudedShepherdOracleID = "de52b262-418e-418c-b2aa-8d5a776e3a75"

func init() {
	Register(Spec{
		OracleID:     shroudedShepherdOracleID,
		Name:         "Shrouded Shepherd",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Shrouded Shepherd — target creature you control gets +2/+2 until end of turn",
				boostTargetUntilEOT("Shrouded Shepherd — +2/+2", 2, 2)),
				TargetCreature("target creature you control", YouControl())),
				ForTargets(BenefitUntilEOT(0, 2, 2))),
		},
	})
	Register(Spec{
		OracleID:     shroudedShepherdOracleID + "#1",
		Name:         "Cleave Shadows",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return BoostUntilEOT{
				Match:     And(Creature(), OpponentControls()),
				Power:     -1,
				Toughness: -1,
				Label:     "Cleave Shadows — -1/-1 until end of turn",
			}.Apply(ctx)
		},
	})
}
