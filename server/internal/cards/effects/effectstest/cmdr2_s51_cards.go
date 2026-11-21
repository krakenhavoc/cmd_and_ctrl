package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s51_cards_test.go — batch 51's planeswalkers: Nissa, Wrenn and
// Seven, Tezzeret, Tamiyo and Liliana, through the real loyalty path.

const (
	NissaShakesOracle    = "5aa3abf1-d56b-4f42-8c84-7e5a2c15ee0f"
	WrennAndSevenOracle  = "63a509ea-cedc-40e9-b4e8-0ff4fc356485"
	TezzeretBetrayOracle = "5dedd5a9-bc5d-47bb-9ae1-7d683843d041"
	TamiyoFieldOracle    = "9cf99129-f51a-4418-a8cc-8ca2992b17fa"
	LilianaDeathOracle   = "e1b6cc44-41b6-43d0-9fe9-722cebb1c47b"

	S51ArtifactProbeOracle = "test-s51-artifact-probe"
	S51CreatureProbeOracle = "test-s51-creature-probe"
)

func init() {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	effects.Register(effects.Spec{
		OracleID: S51ArtifactProbeOracle, Name: "S51 Artifact Probe",
		Activated: []effects.ActivatedAbility{{Label: "{2}: Do nothing.", Cost: effects.ManaCost("{2}"), Purpose: game.Purpose{Answers: game.AnswerValue}, Effect: noop}},
	})
	effects.Register(effects.Spec{
		OracleID: S51CreatureProbeOracle, Name: "S51 Creature Probe",
		Activated: []effects.ActivatedAbility{{Label: "{2}: Do nothing.", Cost: effects.ManaCost("{2}"), Purpose: game.Purpose{Answers: game.AnswerValue}, Effect: noop}},
	})
}

func S51ActivePlayer(t *testing.T) (*game.Game, int, *game.Player, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	return g, seat, g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}
