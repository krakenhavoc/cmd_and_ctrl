package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const VisageOfDreadOracleID = "7cad43df-31c9-47b5-bc05-4a0f6544b396"

func init() {
	effects.Register(effects.Spec{
		OracleID:     VisageOfDreadOracleID,
		Name:         "Visage of Dread",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Visage of Dread — target opponent reveals their hand, you choose an artifact or creature card",
					effects.TargetRevealsYouChooseDiscardAbility(effects.Or(effects.Artifact(), effects.Creature()), "artifact or creature card")),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with two creatures {5}{B}", "{5}{B}", effects.CraftWithN(2, "creature")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:        VisageOfDreadOracleID + "#1",
		Name:            "Dread Osseosaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEntersOrAttacks("Dread Osseosaur — mill two cards", effects.Do(effects.MillCards{N: 2})),
				"Dread Osseosaur — mill two cards?"),
		},
	})
}
