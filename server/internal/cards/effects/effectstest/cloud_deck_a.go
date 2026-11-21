package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloud_deck_a_test.go — the first slice of the Cloud, Ex-SOLDIER deck
// request (#3656): Cloud, Armory Automaton, Wrecking Ball Arm, Slayers'
// Stronghold, both Barrets, SOLDIER Military Program, Red XIII, Elena,
// Zack Fair and Conformer Shuriken.

const (
	CloudExSoldierOracle   = "653f0452-5802-4639-86c2-1500c3964fcb"
	ArmoryAutomatonOracle  = "fa3335be-80b1-49d7-9809-5bef13be79f4"
	WreckingBallArmOracle  = "9f48cb32-46be-4c47-b558-2ea359efdd9b"
	SlayersStrongholdOracl = "2de7367b-a5a4-43f4-8f8f-b931ea28150d"
	BarretWallaceOracle    = "913672bc-cc88-4964-a331-7dbce40f2bda"
	BarretAvalancheOracle  = "83f776d9-f3b6-40c2-8008-1ace4d110825"
	SoldierProgramOracle   = "3e2034ee-adc5-4a24-9605-c9722bf813c1"
	RedXIIIOracle          = "9d302e2a-d852-4878-ac9d-d51d9924144b"
	ElenaTurkRecruitOracle = "f5b6835b-7ebe-4855-8b3e-bffb74fd50ee"
	ZackFairOracle         = "e84d146a-1c67-49df-b062-622794e45c42"
	ConformerShurikenOracl = "8e7681ba-6412-4165-b374-5636feec216d"
)

func CloudDeckCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters["+1/+1"]
			}
		}
	})
	return n
}

func CloudDeckAttach(g *game.Game, attachment uuid.UUID, host game.TargetRef) error {
	var err error
	g.WithWriteLock(func() { err = g.AttachForEffect(attachment, host) })
	return err
}
