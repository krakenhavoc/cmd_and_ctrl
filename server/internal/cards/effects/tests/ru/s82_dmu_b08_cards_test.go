package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b08_cards_test.go — Dominaria United, batch 8 (S82, #5487):
// nine kicker spells, each played through the real catalog entry, kicked
// and unkicked.

const (
	dmuB08SabotageOracle  = "a47ff11e-dbcf-45b1-a13d-fff677ea42c9"
	dmuB08GrowthOracle    = "0611beb7-a38d-4fbf-af9b-ee2a5b9a2896"
	dmuB08ChargeOracle    = "862511a3-672f-44fe-8121-a9afe1d1404f"
	dmuB08HymnOracle      = "2f541bf9-1e5b-4d16-8558-a5cfce7e93ad"
	dmuB08EspionageOracle = "32dce35e-182d-41be-9fb2-b0738d0781b8"
	dmuB08ShotOracle      = "aa597788-4149-48b8-9f43-8c0de063eb68"
	dmuB08ScoutOracle     = "422aca0b-8e8f-4774-9a20-6f7a1cae967e"
	dmuB08StrengthOracle  = "f0b24d59-b3f4-47e6-befd-40a25ad976e2"
	dmuB08GeyserOracle    = "d2240ce8-0bd6-4c4f-a73e-fb0a2af5bfac"
)

var dmuB08Both = []struct {
	name     string
	optional []int
	kicked   bool
}{{"unkicked", nil, false}, {"kicked", []int{0}, true}}

func TestDmuB08EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dmuB08SabotageOracle:  "Aggressive Sabotage",
		dmuB08GrowthOracle:    "Colossal Growth",
		dmuB08ChargeOracle:    "Heroic Charge",
		dmuB08HymnOracle:      "Hurloon Battle Hymn",
		dmuB08EspionageOracle: "Phyrexian Espionage",
		dmuB08ShotOracle:      "Runic Shot",
		dmuB08ScoutOracle:     "Scout the Wilderness",
		dmuB08StrengthOracle:  "Strength of the Coalition",
		dmuB08GeyserOracle:    "Tolarian Geyser",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestDmuB08AggressiveSabotageDiscardsTwoAndBurnsWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			opp.Hand.Cards = nil
			for i := 0; i < 3; i++ {
				opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
			}
			life := opp.Life
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Aggressive Sabotage", "Sorcery", dmuB08SabotageOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			_ = me
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.DiscardOwed(g, opp.ID); got != 2 {
				t.Fatalf("owes %d discards, want 2", got)
			}
			effectstest.DiscardFromHand(t, g, opp.ID)
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Hand.Size() != 1 {
				t.Errorf("hand is %d cards, want 1", opp.Hand.Size())
			}
			want := life
			if tc.kicked {
				want -= 3
			}
			if opp.Life != want {
				t.Errorf("life is %d, want %d", opp.Life, want)
			}
		})
	}
}

func TestDmuB08ColossalGrowthIsInsteadWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			bear := effectstest.SeedBear(g, g.Seats[1].ID)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Colossal Growth", "Instant", dmuB08GrowthOracle, effectstest.CardTarget(bear), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := 5
			if tc.kicked {
				want = 6
			}
			if p, tough := effectstest.PtOf(t, g, bear); p != want || tough != want {
				t.Errorf("target is %d/%d, want %d/%d", p, tough, want, want)
			}
			for _, kw := range []string{"trample", "haste"} {
				if got := effectstest.HasKeywordOnBattlefield(t, g, bear, kw); got != tc.kicked {
					t.Errorf("%s = %v, want %v", kw, got, tc.kicked)
				}
			}
		})
	}
}

func TestDmuB08HeroicChargePumpsYourCreaturesOnly(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			mine, theirs := effectstest.SeedBear(g, g.Seats[0].ID), effectstest.SeedBear(g, g.Seats[1].ID)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Heroic Charge", "Instant", dmuB08ChargeOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 3 {
				t.Errorf("mine is %d/%d, want 4/3", p, tough)
			}
			if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
				t.Errorf("theirs is %d/%d, want 2/2", p, tough)
			}
			if got := effectstest.HasKeywordOnBattlefield(t, g, mine, "trample"); got != tc.kicked {
				t.Errorf("trample = %v, want %v", got, tc.kicked)
			}
			if effectstest.HasKeywordOnBattlefield(t, g, theirs, "trample") {
				t.Error("an opposing creature gained trample")
			}
		})
	}
}

func TestDmuB08HurloonBattleHymnDamagesAndGainsLifeWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			bear := effectstest.SeedBear(g, opp.ID)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Hurloon Battle Hymn", "Instant", dmuB08HymnOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, tc.optional, nil); err == nil {
				t.Fatal("a player was accepted as the target")
			}
			life := me.Life
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Hurloon Battle Hymn", "Instant", dmuB08HymnOracle, effectstest.CardTarget(bear), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(bear) {
				t.Error("the creature survived 4 damage")
			}
			want := life
			if tc.kicked {
				want += 4
			}
			if me.Life != want {
				t.Errorf("life is %d, want %d", me.Life, want)
			}
		})
	}
}

func TestDmuB08PhyrexianEspionageDrawsTwoAndOpponentsDiscardWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.SeedLibrary(me, "A", "B", "C")
			for _, o := range g.Seats[1:] {
				o.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: o.ID, Controller: o.ID})
			}
			hand := me.Hand.Size()
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Phyrexian Espionage", "Sorcery", dmuB08EspionageOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 2 {
				t.Errorf("drew %d cards, want 2 (the spell left the hand)", got+1)
			}
			for _, o := range g.Seats[1:] {
				owed := effectstest.DiscardOwed(g, o.ID)
				if tc.kicked && owed != 1 {
					t.Errorf("%s owes %d discards, want 1", o.ID, owed)
				}
				if !tc.kicked && owed != 0 {
					t.Errorf("%s owes %d discards, want 0", o.ID, owed)
				}
			}
		})
	}
}

func TestDmuB08RunicShotDestroysATappedCreatureAndScriesWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			untapped := effectstest.SeedBear(g, opp.ID)
			tapped := effectstest.SeedBear(g, opp.ID)
			effectstest.TapForTest(t, g, tapped)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Runic Shot", "Sorcery", dmuB08ShotOracle, effectstest.CardTarget(untapped), tc.optional, nil); err == nil {
				t.Fatal("an untapped creature was accepted as the target")
			}
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Runic Shot", "Sorcery", dmuB08ShotOracle, effectstest.CardTarget(tapped), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(tapped) {
				t.Error("the tapped creature survived")
			}
			if !g.Battlefield.Contains(untapped) {
				t.Error("the untapped creature died")
			}
			c := effectstest.ScryChoiceFor(g, me.ID)
			if tc.kicked && (c == nil || len(c.ScryCards) != 2) {
				t.Errorf("kicked cast: scry prompt %v, want one looking at 2", c)
			}
			if !tc.kicked && c != nil {
				t.Error("unkicked cast asked for a scry")
			}
		})
	}
}

func TestDmuB08ScoutTheWildernessFetchesALandAndMakesSoldiersWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			forest := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Scout the Wilderness", "Sorcery", dmuB08ScoutOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerSearchByID(t, g, me.ID, forest)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(forest) {
				t.Fatal("the basic land did not reach the battlefield")
			}
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == forest && !c.Tapped {
					t.Error("the land entered untapped")
				}
			}
			want := 0
			if tc.kicked {
				want = 2
			}
			if got := effectstest.CountOnBattlefield(g, "Soldier", me.ID); got != want {
				t.Errorf("%d Soldier tokens, want %d", got, want)
			}
		})
	}
}

func TestDmuB08StrengthOfTheCoalitionPumpsOneAndCountersAllWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			a, b := effectstest.SeedBear(g, g.Seats[0].ID), effectstest.SeedBear(g, g.Seats[0].ID)
			theirs := effectstest.SeedBear(g, g.Seats[1].ID)
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Strength of the Coalition", "Instant", dmuB08StrengthOracle, effectstest.CardTarget(theirs), tc.optional, nil); err == nil {
				t.Fatal("an opposing creature was accepted as the target")
			}
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Strength of the Coalition", "Instant", dmuB08StrengthOracle, effectstest.CardTarget(a), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			wantA, wantB, counters := 4, 2, 0
			if tc.kicked {
				wantA, wantB, counters = 5, 3, 1
			}
			if p, _ := effectstest.PtOf(t, g, a); p != wantA {
				t.Errorf("target power %d, want %d", p, wantA)
			}
			if p, _ := effectstest.PtOf(t, g, b); p != wantB {
				t.Errorf("other creature power %d, want %d", p, wantB)
			}
			if got := effectstest.CounterCount(g, b, game.CounterPlusOne); got != counters {
				t.Errorf("other creature has %d counters, want %d", got, counters)
			}
			if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
				t.Errorf("an opposing creature got %d counters", got)
			}
		})
	}
}

func TestDmuB08TolarianGeyserBouncesDrawsAndGainsLifeWhenKicked(t *testing.T) {
	for _, tc := range dmuB08Both {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedLibrary(me, "A", "B")
			bear := effectstest.SeedBear(g, opp.ID)
			life, hand := me.Life, me.Hand.Size()
			if _, err := effectstest.CastWithOptionalCosts(t, g, "Tolarian Geyser", "Sorcery", dmuB08GeyserOracle, effectstest.CardTarget(bear), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(bear) {
				t.Error("the creature was not returned")
			}
			if opp.Hand.Size() == 0 {
				t.Error("the creature did not reach its owner's hand")
			}
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("hand grew by %d, want 1 (the spell left it, one card drawn)", got+1)
			}
			want := life
			if tc.kicked {
				want += 3
			}
			if me.Life != want {
				t.Errorf("life is %d, want %d", me.Life, want)
			}
		})
	}
}
