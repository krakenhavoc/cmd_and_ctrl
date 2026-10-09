package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Visage of Dread // Dread Osseosaur — a transforming artifact (#2124,
// ADR 0137):
//
//	Visage of Dread — Artifact {1}{B}
//	  "When this artifact enters, target opponent reveals their hand.
//	   You choose an artifact or creature card from it. That player
//	   discards that card.
//	   Craft with two creatures {5}{B}"
//	Dread Osseosaur — Creature — Dinosaur Skeleton Horror, 5/4
//	  "Menace
//	   Whenever this creature enters or attacks, you may mill two cards."
//
// The enters trigger is Grief's revealed-hand pick (ADR 0116) with an
// artifact-or-creature filter. Craft with two creatures takes any two
// from among creatures you control and creature cards in your
// graveyard, mixed (CR 702.167b). The Osseosaur's trigger is one ability
// watching both events, asked as a "you may".
//
// No simplification.
const visageOfDreadOracleID = "7cad43df-31c9-47b5-bc05-4a0f6544b396"

func init() {
	Register(Spec{
		OracleID:     visageOfDreadOracleID,
		Name:         "Visage of Dread",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Visage of Dread — target opponent reveals their hand, you choose an artifact or creature card",
					TargetRevealsYouChooseDiscardAbility(Or(Artifact(), Creature()), "artifact or creature card")),
				TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with two creatures {5}{B}", "{5}{B}", CraftWithN(2, "creature")),
		},
	})

	Register(Spec{
		OracleID:        visageOfDreadOracleID + "#1",
		Name:            "Dread Osseosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEntersOrAttacks("Dread Osseosaur — mill two cards", Do(MillCards{N: 2})),
				"Dread Osseosaur — mill two cards?"),
		},
	})
}
