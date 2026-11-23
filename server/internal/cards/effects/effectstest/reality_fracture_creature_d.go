package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_d_test.go — the Reality Fracture creatures of
// slice fra-creature-d.

const (
	RfdMarwynClearcutterOracle = "aff4e43d-cafa-47de-8e39-011f9254a26a"
	RfdMarwynPreserverOracle   = "eb03f6c1-c94f-4f09-b2e9-a55ce5533653"
	RfdMasterOfBarbsOracle     = "e8c36337-135c-460d-a1da-cc01743dda71"
	RfdMemnarchOracle          = "f2630b02-fb48-4167-8a47-8284f7d4c484"
	RfdNivMizzetOracle         = "a5350327-e84c-44b6-8359-87882e521048"
	RfdObNixilisOracle         = "3a49b518-61c8-4dea-8008-c5edf8384581"
	RfdOmnathOracle            = "579225b5-50e2-4891-8871-0bf6dbc07e33"
	RfdPiaAetherOracle         = "6519dc16-4960-4220-aaa8-7e4279743714"
	RfdPiaRebuilderOracle      = "a49c0c64-ea87-44f4-9acf-e7181962aeb2"
	RfdWitchstalkerOracle      = "c21fafaf-4510-4752-a501-606ae68e6699"
	RfdProftConsultingOracle   = "45473bdb-b96f-4213-a46d-b2dbdf6b5e56"
	RfdProftSinisterOracle     = "a350856c-87f2-4605-999f-4e26aa362a55"
	RfdRampartHunterOracle     = "0867755a-9fb9-42ad-897e-6206fbfbc1f8"
	RfdRankRatOracle           = "64e7ef06-ba30-4b19-9d44-ca77944cc930"
	RfdRescueGirlOracle        = "df6fc0b4-333b-4f6c-804c-ba6c8e29c88e"
	RfdRuricBiomagusOracle     = "873fa095-5662-407a-bd24-9a1c4712a428"
	RfdRuricMagecrusherOracle  = "453b2b3c-7df5-4515-9565-82f603cf451b"
	RfdSaheeliConsulOracle     = "dfa95c0e-393d-4b22-9de3-45a7efc0ba14"
	RfdSaheeliJewelOracle      = "83705258-1f42-41db-a32e-de99ffd759eb"
)

// rfdHandCard seeds a card with real power/toughness into a hand.
func RfdHandCard(p *game.Player, name, typeLine, manaCost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// rfdCast casts a creature from the active seat's hand and lets it
// (and anything it triggers) settle.
func RfdCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := RfdHandCard(me, name, typeLine, "", oracle, power, toughness)
	ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

func RfdCountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

func RfdOnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := BattlefieldCard(g, id)
	return ok
}
