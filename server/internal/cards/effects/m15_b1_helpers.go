package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// m15_b1_helpers.go — S82 Magic 2015 batch 1 (#5556): the Paragon cycle.
// Append-only.

// m15B1Paragon is a whole M15 Paragon: "Other <color> creatures you
// control get +1/+1." (Glen Elendra Liege's anthem, read through the
// layered colours, the Paragon itself excluded) and "<cost>, {T}: Another
// target <color> creature you control gains <keyword> until end of turn."
// (Alabaster Mage's grant, with the target re-checked at resolution).
func m15B1Paragon(oracleID, name, cost, color, colorWord, keyword string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16Anthem(g2OtherCreaturesYouControlOfColor(color), 1, 1)},
		Activated: []ActivatedAbility{{
			Label:   cost + ", {T}: Another target " + colorWord + " creature you control gains " + keyword + " until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, keyword)),
			Cost:    Plus(ManaCost(cost), TapCost()),
			Targets: Another(TargetCreature("another target "+colorWord+" creature you control", OfColor(color), YouControl())),
			Effect:  jmpB7FirstLegalTargetGains(name+" — "+keyword+" until end of turn", keyword),
		}},
	}
}
