package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fetchlands_test.go — the sacrifice-to-fetch land family, the two
// one-drop mana dorks and Lotus Petal, from the frequency-ordered
// Commander staples pass.

const (
	EvolvingWildsOracle       = "a75445d3-1303-4bb5-89ad-26ea93fecd48"
	TerramorphicExpanseOracle = "1bd3e453-aa21-4ee6-95c2-d6d920ee8e7a"
	PollutedDeltaOracle       = "ef86989d-ce80-4e55-aece-7d11710eeffa"
	LlanowarElvesOracle       = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	ElvishMysticOracle        = "3f3b2c10-21f8-4e13-be83-4ef3fa36e123"
	LotusPetalOracle          = "32e5339e-9e4f-46f8-b305-f9d6d3ba8bb5"
	BojukaBogOracle           = "04b7362d-0490-4cb0-b5d7-2a7732f659ce"
)

// stapleLibraryCard seeds a named land into a player's library.
func StapleLibraryCard(p *game.Player, name, typeLine string) uuid.UUID {
	return PushLibraryCardForTest(p, game.Card{Name: name, TypeLine: typeLine})
}

func StapleCountLands(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsLand() {
			n++
		}
	}
	return n
}
