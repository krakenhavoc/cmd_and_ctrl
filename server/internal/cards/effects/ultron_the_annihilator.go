package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ultron the Annihilator — Legendary Artifact Creature — Robot Villain
// {3}{B}{B}, 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Ultron enters or attacks, create a 2/2 colorless Robot Villain
//	 artifact creature token.
//	 Whenever another artifact is put into your graveyard from the battlefield
//	 or an artifact card is put into your graveyard from anywhere other than
//	 the battlefield, each opponent loses 1 life."
//
// The token is Age of Ultron's Robot Villain template. The drain is
// per artifact, one trigger each (WhenAnotherArtifactIsPutIntoYourGraveyard,
// #4721), and a token artifact dying counts because it is put into the
// graveyard before it ceases to exist (CR 111.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33f2c6cb-f864-4bee-b365-aedf33fbedf5",
		Name:            "Ultron the Annihilator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Ultron the Annihilator — create a 2/2 Robot Villain artifact creature token",
				Do(CreateToken{Template: TokenCard("2/2 colorless Robot Villain artifact"), N: 1})),
			WhenAnotherArtifactIsPutIntoYourGraveyard("Ultron the Annihilator — each opponent loses 1 life",
				Do(EachOpponentLosesLife{N: 1})),
		},
	})
}
