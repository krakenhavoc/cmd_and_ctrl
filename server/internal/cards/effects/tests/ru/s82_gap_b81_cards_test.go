package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB81SproutingRenewal = "674202f0-6832-4392-a118-d513162bcd03"
	gapB81Demolish         = "08c6d2fd-d9ae-4f95-bb9d-7e16b1039814"
	gapB81Mugging          = "2e4f08c2-1bad-4a85-8aad-9bb34a5579b1"
	gapB81SiegeOfTowers    = "d8eb0897-0d2f-4f2d-a09b-87d57d5a0fa9"
	gapB81StitchInTime     = "59b2a90e-542f-4fb0-b290-ac79dc2892a4"
)

func TestGapB81EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB81SproutingRenewal: "Sprouting Renewal",
		gapB81Demolish:         "Demolish",
		gapB81Mugging:          "Mugging",
		gapB81SiegeOfTowers:    "Siege of Towers",
		gapB81StitchInTime:     "Stitch in Time",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec)
		}
	}
}

func TestGapB81DemolishDestroysAnArtifactOrALandAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	bear := effectstest.B16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastAt(t, g, "Demolish", "Sorcery", gapB81Demolish, r35GptB2CardRef(rock))
	effectstest.CastAt(t, g, "Demolish", "Sorcery", gapB81Demolish, r35GptB2CardRef(land))
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, land); ok {
		t.Error("the land survived")
	}
	active := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.AvB24Hand(g, active, game.Card{Name: "Demolish", TypeLine: "Sorcery", OracleID: gapB81Demolish})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: []game.TargetRef{r35GptB2CardRef(bear)}, Strict: true}); err == nil {
		t.Error("a plain creature was a legal target")
	}
}

func TestGapB81MuggingDamagesAndStopsABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	big := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.CastAt(t, g, "Mugging", "Sorcery", gapB81Mugging, r35GptB2CardRef(big))
	c, ok := effectstest.BattlefieldCard(g, big)
	if !ok || c.DamageMarked != 2 {
		t.Fatalf("damage %d on a surviving target, want 2 (present %v)", c.DamageMarked, ok)
	}
	if lc, _ := g.LookupCardForEffect(big); lc.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the damaged creature can still block")
	}
	small := effectstest.B16Creature(g, opp.ID, "Mouse", "Creature — Mouse", 1, 2)
	effectstest.CastAt(t, g, "Mugging", "Sorcery", gapB81Mugging, r35GptB2CardRef(small))
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("a 1/2 survived two damage")
	}
}

func TestGapB81SiegeOfTowersAnimatesAMountainForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	mtn := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.AvB24Hand(g, me, game.Card{Name: "Siege of Towers", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: gapB81SiegeOfTowers})
	effectstest.FloatForTest(g, me, "RC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{r35GptB2CardRef(forest)}, Strict: true}); err == nil {
		t.Fatal("a Forest was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{r35GptB2CardRef(mtn)}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	v := effectstest.LayeredCard(t, g, mtn)
	if v.CurrentPower() != 3 || v.CurrentToughness() != 1 || !v.IsLand() || !v.IsCreature() {
		t.Fatalf("Mountain is %d/%d land=%v creature=%v, want a 3/1 land creature", v.CurrentPower(), v.CurrentToughness(), v.IsLand(), v.IsCreature())
	}
	effectstest.EndTurn(t, g)
	if !effectstest.LayeredCard(t, g, mtn).IsCreature() {
		t.Error("the animation ended with the turn")
	}
}

func TestGapB81SiegeOfTowersReplicatesOntoASecondMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	a := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	b := effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.AvB24Hand(g, me, game.Card{Name: "Siege of Towers", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: gapB81SiegeOfTowers})
	effectstest.FloatForTest(g, me, "RRCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{r35GptB2CardRef(a)}, OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, r35GptB2CardRef(b)); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	for name, id := range map[string]uuid.UUID{"original": a, "copy's": b} {
		if v := effectstest.LayeredCard(t, g, id); !v.IsCreature() || v.CurrentPower() != 3 || v.CurrentToughness() != 1 {
			t.Errorf("the %s target is not a 3/1 creature", name)
		}
	}
}

func TestGapB81SproutingRenewalModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.GptB1Seats(g)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastModal(t, g, "Sprouting Renewal", "Sorcery", gapB81SproutingRenewal, []int{1}, []game.TargetRef{r35GptB2CardRef(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	effectstest.CastModal(t, g, "Sprouting Renewal", "Sorcery", gapB81SproutingRenewal, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elf Knight" && c.Controller == me.ID && c.Power == 2 && c.Toughness == 2 {
			found = true
		}
	}
	if !found {
		t.Error("no 2/2 Elf Knight token")
	}
	spec, _ := effects.Lookup(gapB81SproutingRenewal)
	if spec.TapCost == nil {
		t.Error("no convoke")
	}
}

func TestGapB81StitchInTimeGrantsAnExtraTurnExactlyWhenTheFlipIsWon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	wins, losses := 0, 0
	for i := 0; i < 40 && (wins < 2 || losses < 2); i++ {
		effectstest.CastCatalogSpell(t, g, "Stitch in Time", "Sorcery", gapB81StitchInTime, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceCoinCall)
		if err := g.ResolveCoinCall(c.ID, g.Seats[me].ID, "heads"); err != nil {
			t.Fatal(err)
		}
		var won bool
		for _, ev := range g.Events {
			if ev.Kind == game.EventFlipCoin {
				won = ev.Won
			}
		}
		if won {
			wins++
		} else {
			losses++
		}
	}
	if wins == 0 || losses == 0 {
		t.Fatalf("flips did not cover both outcomes (%d won, %d lost)", wins, losses)
	}
	// Each won flip queued one extra turn for the caster; each lost flip none.
	for i := 0; i < wins; i++ {
		effectstest.EndTurn(t, g)
		effectstest.AssertTurnOf(t, g, me, true, "an extra turn owed for a won flip")
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "no extra turn beyond the won flips")
}
