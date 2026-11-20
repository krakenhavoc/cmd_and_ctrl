package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gap2_test.go — card-level coverage for the play-rate slice cut from
// tracker #293's "Gap list 2/2" (removal, charms, a few creatures).

const (
	G2SlashTheRanksOracle     = "ba7f1b09-5727-484d-a502-3dcf4d618c56"
	G2OblivionRingOracle      = "bd9b9772-f5f9-4c6b-913e-7193bea5d0a7"
	G2UnstableObeliskOracle   = "060ae1bb-956a-4264-b405-a33151f31493"
	G2ReclaimOracle           = "797155cd-faf4-4321-8629-c8c352392748"
	G2HelpingHandOracle       = "170a5dbc-6a43-4715-9501-178b1eb7b08c"
	G2BantCharmOracle         = "324889d6-c857-41ed-bb60-408809fc9964"
	G2SultaiCharmOracle       = "46ed38d1-e642-4cea-99ed-a9c17fd982b1"
	G2GazeOfGraniteOracle     = "103d9ad0-d655-4bd5-a899-e9f8869e333d"
	G2PileOnOracle            = "361b0d7f-1e43-45c5-92c2-92baacaf326f"
	G2FecundityOracle         = "ffa64ac6-fe55-48b8-b015-849982cd7ad8"
	G2GlenElendraLiegeOracle  = "946bba74-0951-408c-b06f-167739b10934"
	G2CliffhavenVampireOracle = "1915a311-209b-4628-8122-af3055a78fed"
	G2WaywardServantOracle    = "2916f041-f080-450f-ae79-dc09076d971a"
	G2BeastcallerSavantOracle = "e28227eb-b3b5-42fa-b597-1fefd2a70186"
	G2InkwellLeviathanOracle  = "dbdd0963-b395-4cb8-ad9a-ed85a3e9f2e5"
	G2ValleyMightcallerOracle = "16e9c452-6288-4da8-813d-2eb6b7a538c3"
)

// g2CastFrom casts a card from `seat`'s hand with the given params.
func G2CastFrom(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: seat.ID, Controller: seat.ID})
	if err := g.CastSpell(seat.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}
