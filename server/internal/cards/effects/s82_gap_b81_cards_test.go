package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec)
		}
	}
}

func TestGapB81DemolishDestroysAnArtifactOrALandAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := gptB1Seats(g)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Their Land", "Land")
	bear := b16Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castAt(t, g, "Demolish", "Sorcery", gapB81Demolish, gptB2CardRef(rock))
	castAt(t, g, "Demolish", "Sorcery", gapB81Demolish, gptB2CardRef(land))
	if _, ok := battlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	if _, ok := battlefieldCard(g, land); ok {
		t.Error("the land survived")
	}
	active := g.Seats[g.Turn.ActiveSeat]
	id := avB24Hand(g, active, game.Card{Name: "Demolish", TypeLine: "Sorcery", OracleID: gapB81Demolish})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: []game.TargetRef{gptB2CardRef(bear)}, Strict: true}); err == nil {
		t.Error("a plain creature was a legal target")
	}
}

func TestGapB81MuggingDamagesAndStopsABlock(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := gptB1Seats(g)
	big := b16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	castAt(t, g, "Mugging", "Sorcery", gapB81Mugging, gptB2CardRef(big))
	c, ok := battlefieldCard(g, big)
	if !ok || c.DamageMarked != 2 {
		t.Fatalf("damage %d on a surviving target, want 2 (present %v)", c.DamageMarked, ok)
	}
	if lc, _ := g.LookupCardForEffect(big); lc.Effective().Restrictions&game.CantBlock == 0 {
		t.Error("the damaged creature can still block")
	}
	small := b16Creature(g, opp.ID, "Mouse", "Creature — Mouse", 1, 2)
	castAt(t, g, "Mugging", "Sorcery", gapB81Mugging, gptB2CardRef(small))
	if _, ok := battlefieldCard(g, small); ok {
		t.Error("a 1/2 survived two damage")
	}
}

func TestGapB81SiegeOfTowersAnimatesAMountainForGood(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := gptB1Seats(g)
	mtn := b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	forest := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := avB24Hand(g, me, game.Card{Name: "Siege of Towers", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: gapB81SiegeOfTowers})
	floatForTest(g, me, "RC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{gptB2CardRef(forest)}, Strict: true}); err == nil {
		t.Fatal("a Forest was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{gptB2CardRef(mtn)}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	v := layeredCard(t, g, mtn)
	if v.CurrentPower() != 3 || v.CurrentToughness() != 1 || !v.IsLand() || !v.IsCreature() {
		t.Fatalf("Mountain is %d/%d land=%v creature=%v, want a 3/1 land creature", v.CurrentPower(), v.CurrentToughness(), v.IsLand(), v.IsCreature())
	}
	endTurn(t, g)
	if !layeredCard(t, g, mtn).IsCreature() {
		t.Error("the animation ended with the turn")
	}
}

func TestGapB81SiegeOfTowersReplicatesOntoASecondMountain(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := gptB1Seats(g)
	a := b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	b := b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := avB24Hand(g, me, game.Card{Name: "Siege of Towers", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: gapB81SiegeOfTowers})
	floatForTest(g, me, "RRCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{gptB2CardRef(a)}, OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with one replicate: %v", err)
	}
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) && latestPickTarget(g, me.ID) == nil {
			break
		}
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, gptB2CardRef(b)); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	for name, id := range map[string]uuid.UUID{"original": a, "copy's": b} {
		if v := layeredCard(t, g, id); !v.IsCreature() || v.CurrentPower() != 3 || v.CurrentToughness() != 1 {
			t.Errorf("the %s target is not a 3/1 creature", name)
		}
	}
}

func TestGapB81SproutingRenewalModes(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := gptB1Seats(g)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castModal(t, g, "Sprouting Renewal", "Sorcery", gapB81SproutingRenewal, []int{1}, []game.TargetRef{gptB2CardRef(rock)})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, rock); ok {
		t.Error("the artifact survived")
	}
	castModal(t, g, "Sprouting Renewal", "Sorcery", gapB81SproutingRenewal, []int{0}, nil)
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elf Knight" && c.Controller == me.ID && c.Power == 2 && c.Toughness == 2 {
			found = true
		}
	}
	if !found {
		t.Error("no 2/2 Elf Knight token")
	}
	spec, _ := Lookup(gapB81SproutingRenewal)
	if spec.TapCost == nil {
		t.Error("no convoke")
	}
}

func TestGapB81StitchInTimeGrantsAnExtraTurnExactlyWhenTheFlipIsWon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	wins, losses := 0, 0
	for i := 0; i < 40 && (wins < 2 || losses < 2); i++ {
		castCatalogSpell(t, g, "Stitch in Time", "Sorcery", gapB81StitchInTime, nil)
		passPriorityAroundTable(t, g)
		c := pr7bPrompt(t, g, game.PendingChoiceCoinCall)
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
		endTurn(t, g)
		assertTurnOf(t, g, me, true, "an extra turn owed for a won flip")
	}
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "no extra turn beyond the won flips")
}
