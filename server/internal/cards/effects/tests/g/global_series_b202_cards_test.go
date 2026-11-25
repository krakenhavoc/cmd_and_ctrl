package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// global_series_b202_cards_test.go — Global Series Jiang Yanggu & Mu
// Yanling, S82 older-sets batch 202 (#5528).

const (
	gs1B202JiaoOracle      = "0056e07b-416b-487e-9e6c-3697db402dd4"
	gs1B202TurtleOracle    = "51a886f2-9b0a-4964-9d59-99dc1a68a97c"
	gs1B202ZhengOracle     = "a45e0854-28c7-41f5-a9fe-5b76a8070c5b"
	gs1B202DogOracle       = "a8e69725-cd57-4340-83b1-498bc0bd77b8"
	gs1B202FishOracle      = "e8ec9ba4-dc6b-437d-863f-5ec37f4ae9cd"
	gs1B202StormcloudOracl = "0dc8bdbb-996b-47d7-a5e5-7bdb6d9280fe"
	gs1B202SnakeOracle     = "77b69977-200e-44c2-bfde-228e6501b03a"
	gs1B202SparrowOracle   = "82657b8b-de2d-46ab-ab56-e2790e326d95"
	gs1B202QilinOracle     = "cbe1c80a-db82-460a-b6e9-d0e429628aa1"
	gs1B202PangolinOracle  = "1f384f19-7fb4-4120-8001-5b401b4debbd"
	gs1B202PhoenixOracle   = "2b75759a-c3cc-432d-9609-4f8fa2456870"
	gs1B202BlessingOracle  = "31585036-cdc3-47f5-9fc7-913212826f2a"
)

func TestGs1B202EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gs1B202JiaoOracle:      "Leopard-Spotted Jiao",
		gs1B202TurtleOracle:    "Armored Whirl Turtle",
		gs1B202ZhengOracle:     "Ferocious Zheng",
		gs1B202DogOracle:       "Moon-Eating Dog",
		gs1B202FishOracle:      "Vivid Flying Fish",
		gs1B202StormcloudOracl: "Stormcloud Spirit",
		gs1B202SnakeOracle:     "Feiyi Snake",
		gs1B202SparrowOracle:   "Colorful Feiyi Sparrow",
		gs1B202QilinOracle:     "Heavenly Qilin",
		gs1B202PangolinOracle:  "Reckless Pangolin",
		gs1B202PhoenixOracle:   "Screeching Phoenix",
		gs1B202BlessingOracle:  "Qilin's Blessing",
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

func TestGs1B202VanillaCreaturesHaveNoAbilities(t *testing.T) {
	for _, oracle := range []string{gs1B202JiaoOracle, gs1B202TurtleOracle, gs1B202ZhengOracle} {
		spec, _ := effects.Lookup(oracle)
		if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Triggered) != 0 ||
			len(spec.Static) != 0 || len(spec.BlockRules) != 0 {
			t.Errorf("%s is vanilla but its spec carries abilities", spec.Name)
		}
	}
}

func TestGs1B202PrintedKeywords(t *testing.T) {
	for _, tc := range []struct{ name, oracle, kw string }{
		{"Stormcloud Spirit", gs1B202StormcloudOracl, "flying"},
		{"Feiyi Snake", gs1B202SnakeOracle, "reach"},
		{"Colorful Feiyi Sparrow", gs1B202SparrowOracle, "flying"},
		{"Heavenly Qilin", gs1B202QilinOracle, "flying"},
		{"Screeching Phoenix", gs1B202PhoenixOracle, "flying"},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 2, 2)
		if !effectstest.HasKeywordOnBattlefield(t, g, id, tc.kw) {
			t.Errorf("%s does not have %s", tc.name, tc.kw)
		}
	}
}

func gs1B202Planeswalker(g *game.Game, owner uuid.UUID, subtype string) {
	id := effectstest.B12Permanent(g, owner, "Walker "+subtype, "Legendary Planeswalker — "+subtype)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterLoyalty, 3) })
}

func TestGs1B202MoonEatingDogFliesOnlyWithYourYanlingWalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dog := effectstest.B12Push(g, me.ID, "Moon-Eating Dog", "Creature — Dog", gs1B202DogOracle, 3, 3)
	if effectstest.HasKeywordOnBattlefield(t, g, dog, "flying") {
		t.Fatal("flying with no planeswalker")
	}
	gs1B202Planeswalker(g, opp.ID, "Yanling")
	gs1B202Planeswalker(g, me.ID, "Gideon")
	if effectstest.HasKeywordOnBattlefield(t, g, dog, "flying") {
		t.Fatal("flying from an opponent's Yanling or a non-Yanling walker")
	}
	gs1B202Planeswalker(g, me.ID, "Yanling")
	if !effectstest.HasKeywordOnBattlefield(t, g, dog, "flying") {
		t.Error("no flying with your own Yanling planeswalker")
	}
}

func TestGs1B202VividFlyingFishFliesOnlyWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fish := effectstest.B12Push(g, me.ID, "Vivid Flying Fish", "Creature — Fish Lizard", gs1B202FishOracle, 1, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, fish, "flying") {
		t.Fatal("flying while not attacking")
	}
	effectstest.DeclareAttack(t, g, opp.ID, fish)
	if !effectstest.HasKeywordOnBattlefield(t, g, fish, "flying") {
		t.Error("no flying while attacking")
	}
}

func TestGs1B202HeavenlyQilinGivesAnotherCreatureYouControlFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qilin := effectstest.B12Push(g, me.ID, "Heavenly Qilin", "Creature — Kirin", gs1B202QilinOracle, 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, qilin)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []game.TargetRef{{Kind: game.TargetCard, ID: qilin}, {Kind: game.TargetCard, ID: theirs}} {
		if err := g.ResolvePickTarget(p.ID, me.ID, bad); err == nil {
			t.Errorf("%s was accepted: the Qilin itself or an opponent's creature", bad.ID)
		}
	}
	effectstest.AvB34Pick(t, g, me.ID, home)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, home, "flying") {
		t.Error("the other creature did not gain flying")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "flying") {
		t.Error("the opponent's creature gained flying")
	}
}

func TestGs1B202RecklessPangolinGrowsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pangolin := effectstest.B12Push(g, me.ID, "Reckless Pangolin", "Creature — Pangolin", gs1B202PangolinOracle, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, pangolin)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, pangolin); p != 3 || tough != 3 {
		t.Errorf("the Pangolin is %d/%d after attacking, want 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, other); p != 2 || tough != 2 {
		t.Errorf("another creature is %d/%d, want 2/2", p, tough)
	}
}

func TestGs1B202ScreechingPhoenixPumpsYourTeam(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	phoenix := effectstest.B12Push(g, me.ID, "Screeching Phoenix", "Creature — Phoenix", gs1B202PhoenixOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "R", "C", "C")
	effectstest.ApaActivate(t, g, me, phoenix, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		want int
	}{{phoenix, 5}, {bear, 3}, {theirs, 2}} {
		if p, _ := effectstest.PtOf(t, g, tc.id); p != tc.want {
			t.Errorf("power %d, want %d", p, tc.want)
		}
	}
	if p, tough := effectstest.PtOf(t, g, bear); tough != 2 {
		t.Errorf("toughness changed: %d/%d", p, tough)
	}
}

func TestGs1B202QilinsBlessingGivesPlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Qilin's Blessing", "Instant", gs1B202BlessingOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Qilin's Blessing", "Instant", gs1B202BlessingOracle, nil); err == nil {
		t.Error("cast with no target")
	}
}
