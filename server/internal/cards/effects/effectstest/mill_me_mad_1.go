package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S58 PR 10, "Mill Me Mad (1)": the Lhurgoyfs and the sacrifice
// outlets.

const (
	MortivoreOracle         = "2f554218-2072-4d9b-b4d4-34f161f1f2a9"
	NecrogoyfOracle         = "e034a02d-3482-44eb-90dc-c38be104197f"
	BeastOfBurdenOracle     = "67389ccc-dacc-4ce3-a9a2-f551a0d6e9d3"
	UrborgLhurgoyfOracle    = "f84d8217-51a4-49ba-9df6-d080ed38f37d"
	AltarOfDementiaOracle   = "d64e9152-ef24-4394-aeb0-9c3befc56549"
	BushmeatPoacherOracle   = "0287d541-3f73-4c8b-9a77-87f99898758d"
	DiscipleGriselbrandOra  = "2d92a035-dd7a-4426-a8c0-f04e0b836dad"
	SplinterfrightOracle    = "e54c6fa9-59ca-47dc-9354-91681228371e"
	ApocalypseDemonOracle   = "5b8084b2-0c3c-4a54-86dc-41f4fe513e99"
	CruelSomnophageTestOrac = "997bdec5-f67b-4822-a3ba-c636e2685e8a"
)

func PutCreatureCardsInGraveyard(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear",
			Owner: p.ID, Controller: p.ID,
		})
	}
}
