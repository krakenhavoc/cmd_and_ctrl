package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_f_test.go — slice fra-creature-f (tracker
// #2795): Reality Fracture creatures, one or more tests each.

const (
	OracleTraxosAcademy    = "7b9a574f-6d56-421b-8fbd-94da1e6e3ce1"
	OracleTraxosScourge    = "acf4abde-4034-4fcf-b531-78c935aecd8c"
	OracleUndulatingWit    = "77b684b8-227b-40de-80b4-15d0e70914f4"
	OracleHortimancer      = "7cd0bd74-2463-4e9d-9802-69ac4ef6d543"
	OracleVenserFervent    = "9b1ae6d8-0cf3-4a67-9241-37f4bc5ddbdd"
	OracleVerdantKraken    = "83c61083-b357-4775-a855-5a9013cc2bdf"
	OracleVinelasher       = "b5918fb0-c33f-4e33-ac6d-8f94723d4296"
	OracleVraskaSoul       = "cad63b31-667b-4518-bc49-be4ffb82bc63"
	OracleVraskaGlare      = "580de511-863c-4b7a-9d2e-fcdf6e12d54b"
	OracleWinterTeam       = "254f133a-eee5-48b9-b19f-110555e49300"
	OracleWreckingGecko    = "4a9e0294-c967-4e90-b776-6d3195034a23"
	OracleYargle           = "014d0bcd-b80f-4c43-8f00-f4c31e8f3370"
	OracleYoshimaruBelov   = "c00a0be5-d9a9-4263-a87e-7bf2d128ab3a"
	OracleYoshimaruStray   = "c87da757-3daa-4874-a1e4-be0aa2adcb28"
	OracleYurikoBlade      = "f89b0143-4b19-4ab1-a96b-356b21e435ef"
	RfCrFNoncreatureOracle = "00000000-0000-0000-0000-00000000f001"
)

// rfCrFHand puts a card in the active seat's hand.
func RfCrFHand(g *game.Game, name, typeLine, oracle, cost string, power, toughness int) uuid.UUID {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
	})
	return id
}

// rfCrFCast puts a free creature in hand, goes to the main phase and
// casts it.
func RfCrFCast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := RfCrFHand(g, name, typeLine, oracle, "", power, toughness)
	ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
