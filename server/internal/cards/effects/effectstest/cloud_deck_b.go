package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cloud_deck_b_test.go — the second slice of the Cloud, Ex-SOLDIER deck
// request (#3656): Summon: Kujata, Ultimate Magic: Meteor, Lifestream's
// Blessing, Heidegger, Cid, Bronze Guardian, Unfinished Business,
// Aerith, Yuffie and Cait Sith.

const (
	SummonKujataOracle      = "9fe77f70-85cf-4a18-9451-2fad4c341f07"
	UltimateMeteorOracle    = "31cc5d9d-9384-4a9b-ba7f-6a4edbcb7a35"
	LifestreamsBlessingOrcl = "7cbdba72-ed33-40c6-a9c0-4d65c280a4d0"
	HeideggerOracle         = "215024ba-811e-4011-bbc8-8dd14a34f973"
	CidFreeflierOracle      = "d3bdfb25-dfb2-4df4-b502-ef0909a39f86"
	BronzeGuardianOracle    = "922a030c-857b-4c61-93df-b70ebb32abd3"
	UnfinishedBusinessOracl = "800cfd70-6a97-40eb-8dc0-40b762e4db76"
	AerithLastAncientOracle = "d6cba688-faff-4d9f-8148-37d389fa5fb3"
	YuffieMateriaOracle     = "acdcbb49-9d71-4c36-8ccb-650fb11c3a58"
	CaitSithOracle          = "31851cb6-0caf-4575-aa5e-41aeadccafc1"
)

func CloudDeckPickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, ids ...uuid.UUID) {
	t.Helper()
	pick := LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	var refs []game.TargetRef
	for _, id := range ids {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func CloudDeckKujata(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return TdCastFromHand(t, g, game.Card{Name: "Summon: Kujata", TypeLine: "Enchantment Creature — Saga Ox",
		OracleID: SummonKujataOracle, ManaCost: "{5}{R}", Power: 7, Toughness: 5}, game.CastSpellParams{})
}

func CloudDeckForetellAndCast(t *testing.T, g *game.Game, name, typeLine, oracle, foretellCost string, pay []game.ManaToken, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	AdvanceTo(t, g, game.StepPrecombatMain)
	card := HandCardForTest(active, name, typeLine, oracle)
	active.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(active.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell %s: %v", name, err)
	}
	g.WithWriteLock(func() { g.Turn.Seq++ })
	active.ManaPool.AddMana(pay...)
	if err := g.CastSpell(active.ID, card, game.CastSpellParams{
		Strict:          true,
		FromZone:        "exile",
		AlternativeCost: game.AltCostKeyForetell,
		Targets:         targets,
	}); err != nil {
		t.Fatalf("cast the foretold %s: %v", name, err)
	}
}

func CloudDeckMana(color string, n int) []game.ManaToken {
	out := make([]game.ManaToken, n)
	for i := range out {
		out[i] = game.ManaToken{Color: color}
	}
	return out
}

func LookupBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, ok = c, true
			}
		}
	})
	return out, ok
}
