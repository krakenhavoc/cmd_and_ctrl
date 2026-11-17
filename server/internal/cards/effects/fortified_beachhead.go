package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fortified Beachhead — Land (The Brothers' War, #5482):
//
//	"As this land enters, you may reveal a Soldier card from your hand. This
//	 land enters tapped unless you revealed a Soldier card this way or you
//	 control a Soldier.
//	 {T}: Add {W} or {U}.
//	 {5}, {T}: Soldiers you control get +1/+1 until end of turn."
//
// Temple of the Dragon Queen's entry (the reveal-land replacement that does
// not apply, and so asks nothing, while a Soldier is controlled), here over
// Soldiers. The pump is a one-shot set locked at resolution (CR 611.2c).
//
// No simplification.
func init() {
	const name = "Fortified Beachhead"
	Register(Spec{
		OracleID:      "387fe395-e4a0-4fb2-8d6c-88a1a21d2ed8",
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{broB05EntersTappedUnlessRevealOrControl(name, "Soldier")},
		ManaAbilities: []ManaAbility{dualManaAbility("W", "U")},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Soldiers you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Effect: Do(BoostUntilEOT{
				Match: And(Creature(), YouControl(), OfCreatureType("Soldier")), Power: 1, Toughness: 1,
				Label: "Fortified Beachhead — Soldiers you control get +1/+1",
			}),
		}},
	})
}
