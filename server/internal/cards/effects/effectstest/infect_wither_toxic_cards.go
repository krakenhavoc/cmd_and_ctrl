package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// infect_wither_toxic_cards_test.go plays ADR 0056's first card wave
// (#748) through real combat and real casts: the damage tail turning
// infect, wither and toxic damage into counters is exercised through
// the cards that print or grant the keywords.

const (
	PlagueMyrOracle           = "2f328e05-5edf-4b21-9c2a-50dcf1e7b3ec"
	BlightedAgentOracle       = "e48ea9ea-64bc-4f1c-a424-592d48569244"
	IchorRatsOracle           = "f8148664-49c0-421f-93a5-cd59e4e9ea36"
	PunctureBlastOracle       = "e296581d-01ac-43bf-898c-2edb4c81bcbe"
	BloatedContaminatorOracle = "090018e0-4dcb-4b3c-b4e0-7ba62de0484d"
	TaintedStrikeOracle       = "95a53dd7-76dc-46f1-8833-fafd02ba49c4"
	TriumphOfTheHordesOracle  = "3ded0c0c-40ce-4d14-a9a6-b023bc19ee0e"
	KarumonixOracle           = "c017f54c-e4c0-411e-b8c1-eb20b1b86c56"
)

func PoisonOn(p *game.Player) int { return p.Counters[game.CounterPoison] }

func MinusOneOn(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterMinusOne]
			}
		}
	})
	return n
}
