package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_planeswalker_b_test.go — slice fra-planeswalker-b of
// the Reality Fracture set (#2795).

const (
	RfpbGarrukCurseBreakerOracle  = "184d4672-05a9-4182-8b3f-3561fb33ed62"
	RfpbGarrukVeiledButcherOracle = "4a6972fe-348a-4a19-a1d0-bdf8f1f9c792"
	RfpbGideonOathlessOracle      = "7afa530f-ea51-4bee-af3b-59b83ea5cb25"
	RfpbGideonsMemorialOracle     = "c8f08704-7330-4fa9-b456-2e96aa252a2c"
	RfpbIdentityEchoOracle        = "4dbf8c43-d7c5-4037-875e-6accd87aaf6f"
	RfpbInnovativeCommonsOracle   = "6812e4d3-0034-4fc8-9ee9-520f76a0ea94"
	RfpbJaceMultiverseOracle      = "3321c134-bf5b-4f63-8035-fca0bbdfa86b"
	RfpbKonstrariAnnexOracle      = "31ea4f3f-392a-4319-a95b-eb28790a80e6"
	RfpbLichsRelicOracle          = "0540eaca-0e03-4831-955a-192ae2b87d35"
	RfpbLilianaFaultlessOracle    = "22c2e66d-a2a1-46db-a70a-dff5ad5d6c14"
	RfpbLilianaRepentantOracle    = "5eb4403f-f199-4f75-a7c6-e76783f9b07d"
	RfpbLivingLibraryOracle       = "81f022a7-fd19-4340-b8a0-6b11ab3e9d91"
	RfpbLoyalTutorOracle          = "7e52f151-d1f6-4fcb-9b21-1baecae27da6"
	RfpbMabelOracle               = "36924990-8e3a-434e-bcd1-f03603cb350d"
	RfpbMassacreGirlOracle        = "81f35914-73bf-429f-ba20-32e23ac30159"
	RfpbMeticulousCommonsOracle   = "7fdd471f-ec95-4aa7-8f8a-ad8946de5ab4"
)

// rfpbWalker seats a planeswalker with the given type line and loyalty.
func RfpbWalker(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, loyalty int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: owner, Controller: owner,
		Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

// rfpbCastCreature casts a creature with the given power from the active
// seat's hand (free) and settles it.
func RfpbCastCreature(t *testing.T, g *game.Game, name string, power int) uuid.UUID {
	t.Helper()
	ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Test", Power: power, Toughness: power,
		Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

func RfpbCardRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

func RfpbAddMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// rfpbCastRelic casts Lich's Relic for {B} with extra mana in the pool and
// settles the spell, leaving the "you may pay {2}" prompt open.
func RfpbCastRelic(t *testing.T, g *game.Game, me *game.Player, extra int) {
	t.Helper()
	ToMainForCost(t, g)
	relic := HandCardFull(me, "Lich's Relic", "Artifact — Equipment", "{B}", RfpbLichsRelicOracle, []string{"B"})
	RfpbAddMana(me, "B")
	for i := 0; i < extra; i++ {
		RfpbAddMana(me, "B")
	}
	if err := g.CastSpell(me.ID, relic, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Lich's Relic: %v", err)
	}
	PassPriorityAroundTable(t, g)
}
