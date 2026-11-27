package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b650_cards_test.go — OLD batch 650 (Invasion, #5616): seven kicker
// cards. Each test plays the real catalog entry through the engine.

const (
	oldB650Offensive = "b5944ca8-d54e-4830-9f66-be8fe97d101d"
	oldB650Lancer    = "b5adbd37-8d98-4c65-b42d-cd9db94e7222"
	oldB650Charge    = "b8ce3f84-2053-4d16-ab3d-e552bd8d0e8c"
	oldB650Touch     = "cc22c210-efd0-494c-8560-448b038b3c5f"
	oldB650Titan     = "df166d0e-f944-404d-83b9-8cf53e220774"
	oldB650Urborg    = "f30943f1-473f-4fa4-b6a7-3dfbd1995ba9"
	oldB650Spider    = "fbe25c32-db20-464c-a0ed-d90e75b111a9"
)

func TestOldB650EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB650Offensive: "Savage Offensive", oldB650Lancer: "Benalish Lancer", oldB650Charge: "Vigorous Charge",
		oldB650Touch: "Orim's Touch", oldB650Titan: "Kavu Titan", oldB650Urborg: "Urborg Emissary",
		oldB650Spider: "Pincer Spider",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB650Cast casts a card from the active player's hand, kicked or not.
func oldB650Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, kicked bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	params := game.CastSpellParams{Targets: targets}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s (kicked=%v): %v", name, kicked, err)
	}
	return id
}

func TestOldB650KickedCreaturesEnterWithCountersAndAbilities(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		counters               int
		keyword                string
	}{
		{"Benalish Lancer", "Creature — Human Knight", oldB650Lancer, 2, 2, 2, "first strike"},
		{"Kavu Titan", "Creature — Kavu", oldB650Titan, 2, 2, 3, "trample"},
		{"Pincer Spider", "Creature — Spider", oldB650Spider, 2, 3, 1, ""},
	} {
		t.Run(tc.name+" kicked", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := oldB650Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, true, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.counters {
				t.Errorf("%d counters, want %d", got, tc.counters)
			}
			if p, _ := effectstest.PtOf(t, g, id); p != tc.power+tc.counters {
				t.Errorf("power %d, want %d", p, tc.power+tc.counters)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("a kicked %s lacks %s", tc.name, tc.keyword)
			}
		})
		t.Run(tc.name+" unkicked", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := oldB650Cast(t, g, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness, false, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 0 {
				t.Errorf("%d counters unkicked, want 0", got)
			}
			if tc.keyword != "" && effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
				t.Errorf("an unkicked %s has %s", tc.name, tc.keyword)
			}
		})
	}
}

func TestOldB650SavageOffensive(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
		oldB650Cast(t, g, "Savage Offensive", "Sorcery", oldB650Offensive, 0, 0, kicked, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, mine, "first strike") {
			t.Errorf("kicked=%v: your creature lacks first strike", kicked)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, "first strike") {
			t.Errorf("kicked=%v: their creature gained first strike", kicked)
		}
		want := 2
		if kicked {
			want = 3
		}
		if p, tough := effectstest.PtOf(t, g, mine); p != want || tough != want {
			t.Errorf("kicked=%v: %d/%d, want %d/%d", kicked, p, tough, want, want)
		}
		if p, _ := effectstest.PtOf(t, g, theirs); p != 3 {
			t.Errorf("kicked=%v: their creature is %d power, want 3", kicked, p)
		}
	}
}

func TestOldB650OrimsTouchPrevents(t *testing.T) {
	for _, tc := range []struct {
		kicked bool
		shield int
	}{{false, 2}, {true, 4}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		oldB650Cast(t, g, "Orim's Touch", "Instant", oldB650Touch, 0, 0, tc.kicked,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		for i := 0; i < 2; i++ {
			effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
		}
		if got := life - me.Life; got != 6-tc.shield {
			t.Errorf("kicked=%v: took %d from 6, want %d", tc.kicked, got, 6-tc.shield)
		}
	}
}

func TestOldB650UrborgEmissaryBouncesOnlyWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	oldB650Cast(t, g, "Urborg Emissary", "Creature — Human Wizard", oldB650Urborg, 3, 1, false, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(relic) {
		t.Fatal("an unkicked cast bounced a permanent")
	}
	oldB650Cast(t, g, "Urborg Emissary", "Creature — Human Wizard", oldB650Urborg, 3, 1, true, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the kicked Emissary did not return the permanent")
	}
	if z := effectstest.B12ZoneOf(g, relic); z != game.ZoneHand {
		t.Errorf("the permanent is in %s, want its owner's hand", z)
	}
}

func TestOldB650VigorousChargeGainsLifeOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		attacker := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
		oldB650Cast(t, g, "Vigorous Charge", "Instant", oldB650Charge, 0, 0, kicked,
			[]game.TargetRef{{Kind: game.TargetCard, ID: attacker}})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, attacker, "trample") {
			t.Fatalf("kicked=%v: the creature has no trample", kicked)
		}
		life, oppLife := me.Life, opp.Life
		effectstest.TrekB25ToCombatDamage(t, g, opp.ID, attacker)
		if opp.Life != oppLife-3 {
			t.Fatalf("kicked=%v: the defender is at %d, want %d", kicked, opp.Life, oppLife-3)
		}
		want := life
		if kicked {
			want = life + 3
		}
		if me.Life != want {
			t.Errorf("kicked=%v: life %d, want %d", kicked, me.Life, want)
		}
	}
}
