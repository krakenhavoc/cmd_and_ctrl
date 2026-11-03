package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ascendant Dustspeaker — Creature — Orc Cleric {4}{W}, 3/4 (Secrets of
// Strixhaven):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on another target
//	 creature you control.
//	 At the beginning of combat on your turn, exile up to one target card
//	 from a graveyard."
//
// Both targets are chosen as the trigger goes on the stack and re-checked
// at resolution (CR 608.2b): "another" excludes the Dustspeaker itself (a
// second creature of the same name stays legal), and "up to one" lets the
// combat trigger go on the stack with no card and do nothing.
func init() {
	enters := WhenThisEnters("Ascendant Dustspeaker — a +1/+1 counter on another target creature you control",
		PlusOneCounterOnTheTarget)
	enters.Targets = Another(TargetCreature("another target creature you control", YouControl()))
	Register(Spec{
		OracleID:        "295f8b8d-102a-47af-93c2-f8182c5f11ca",
		Name:            "Ascendant Dustspeaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			enters,
			Targeting(
				AtBeginningOfYourCombat("Ascendant Dustspeaker — exile up to one target card from a graveyard",
					b27ExileChosenTarget),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
	})
}
