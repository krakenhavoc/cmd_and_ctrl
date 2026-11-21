package testsbr

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var protoRows = []effectstest.ProtoRow{
	{Oracle: "19f344ee-031d-4346-9553-c5d32b02df43", Name: "Blitz Automaton", TypeLine: "Artifact Creature — Construct", Cost: "{7}", Power: 6, Toughness: 4,
		Text: "Prototype {2}{R} — 3/2" + effectstest.ProtoReminder + "\nHaste", InsetCost: "{2}{R}", InsetColor: "R", InsetPower: 3, InsetToughness: 2, Keywords: []string{"haste"}},
	{Oracle: "02457c7f-eced-4bf3-9b61-309b4ceab2d5", Name: "Goring Warplow", TypeLine: "Artifact Creature — Construct", Cost: "{6}", Power: 5, Toughness: 4,
		Text: "Prototype {1}{B} — 1/1" + effectstest.ProtoReminder + "\nDeathtouch", InsetCost: "{1}{B}", InsetColor: "B", InsetPower: 1, InsetToughness: 1, Keywords: []string{"deathtouch"}},
	{Oracle: "61df991d-32e7-4b6b-b8e6-700ba915c3f2", Name: "Rust Goliath", TypeLine: "Artifact Creature — Construct", Cost: "{10}", Power: 10, Toughness: 10,
		Text: "Prototype {3}{G}{G} — 3/5" + effectstest.ProtoReminder + "\nReach, trample", InsetCost: "{3}{G}{G}", InsetColor: "G", InsetPower: 3, InsetToughness: 5, Keywords: []string{"reach", "trample"}},
	{Oracle: "82caf3e2-da42-4851-b4db-546e2403c4bf", Name: "Boulderbranch Golem", TypeLine: "Artifact Creature — Golem", Cost: "{7}", Power: 6, Toughness: 5,
		Text: "Prototype {3}{G} — 3/3" + effectstest.ProtoReminder + "\nWhen this creature enters, you gain life equal to its power.", InsetCost: "{3}{G}", InsetColor: "G", InsetPower: 3, InsetToughness: 3, Keywords: nil},
	{Oracle: "0c460bdd-460b-4312-9ed2-32f3809a7d36", Name: "Combat Thresher", TypeLine: "Artifact Creature — Construct", Cost: "{7}", Power: 3, Toughness: 3,
		Text: "Prototype {2}{W} — 1/1" + effectstest.ProtoReminder + "\nDouble strike\nWhen this creature enters, draw a card.", InsetCost: "{2}{W}", InsetColor: "W", InsetPower: 1, InsetToughness: 1, Keywords: []string{"double strike"}},
	{Oracle: "225f53ef-da42-47a0-b773-d2aad3338c72", Name: "Cradle Clearcutter", TypeLine: "Artifact Creature — Golem", Cost: "{6}", Power: 3, Toughness: 6,
		Text: "Prototype {2}{G} — 1/3" + effectstest.ProtoReminder + "\n{T}: Add an amount of {G} equal to this creature's power.", InsetCost: "{2}{G}", InsetColor: "G", InsetPower: 1, InsetToughness: 3, Keywords: nil},
	{Oracle: "f376f4c6-4526-4ae3-a84a-e8ea50fe78b6", Name: "Depth Charge Colossus", TypeLine: "Artifact Creature — Dreadnought", Cost: "{9}", Power: 9, Toughness: 9,
		Text: "Prototype {4}{U}{U} — 6/6" + effectstest.ProtoReminder + "\nThis creature doesn't untap during your untap step.\n{3}: Untap this creature.", InsetCost: "{4}{U}{U}", InsetColor: "U", InsetPower: 6, InsetToughness: 6, Keywords: nil},
	{Oracle: "dc8cca38-df2a-4cef-973c-a482505be66f", Name: "Fallaji Dragon Engine", TypeLine: "Artifact Creature — Dragon", Cost: "{8}", Power: 5, Toughness: 5,
		Text: "Prototype {2}{R} — 1/3" + effectstest.ProtoReminder + "\nFlying\n{2}: This creature gets +1/+0 until end of turn.", InsetCost: "{2}{R}", InsetColor: "R", InsetPower: 1, InsetToughness: 3, Keywords: []string{"flying"}},
	{Oracle: "ccf13531-60bd-4ebd-9a5b-68e2d05ce620", Name: "Spotter Thopter", TypeLine: "Artifact Creature — Thopter", Cost: "{8}", Power: 4, Toughness: 5,
		Text: "Prototype {3}{U} — 2/3" + effectstest.ProtoReminder + "\nFlying\nWhen this creature enters, scry X, where X is its power.", InsetCost: "{3}{U}", InsetColor: "U", InsetPower: 2, InsetToughness: 3, Keywords: []string{"flying"}},
	{Oracle: "b04f45de-cb35-45a4-b039-0bf008c3ca56", Name: "Iron-Craw Crusher", TypeLine: "Artifact Creature — Wurm", Cost: "{7}", Power: 4, Toughness: 6,
		Text: "Prototype {2}{G}{G} — 2/5" + effectstest.ProtoReminder + "\nWhenever this creature attacks, target attacking creature gets +X/+0 until end of turn, where X is this creature's power.", InsetCost: "{2}{G}{G}", InsetColor: "G", InsetPower: 2, InsetToughness: 5, Keywords: nil},
	{Oracle: "3afdc98e-aeeb-43e9-bdcd-0c51da44d2f8", Name: "Autonomous Assembler", TypeLine: "Artifact Creature — Assembly-Worker", Cost: "{5}", Power: 4, Toughness: 5,
		Text: "Prototype {1}{W} — 2/2" + effectstest.ProtoReminder + "\nVigilance\n{1}, {T}: Put a +1/+1 counter on target Assembly-Worker you control.", InsetCost: "{1}{W}", InsetColor: "W", InsetPower: 2, InsetToughness: 2, Keywords: []string{"vigilance"}},
}

func protoRowNamed(t *testing.T, name string) effectstest.ProtoRow {
	t.Helper()
	for _, r := range protoRows {
		if r.Name == name {
			return r
		}
	}
	t.Fatalf("no prototype row for %s", name)
	return effectstest.ProtoRow{}
}

func TestBroPrototypeCardsAreRegisteredFull(t *testing.T) {
	for _, r := range protoRows {
		spec, ok := effects.Lookup(r.Oracle)
		if !ok {
			t.Errorf("%s is not registered", r.Name)
			continue
		}
		if spec.Name != r.Name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", r.Name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// Every card, cast both ways: the prototyped permanent has the inset's
// mana cost, colour, power and toughness (CR 718.3b), the normal one
// its printed ones, and both keep the name, types and keywords (CR
// 718.5). The inset cost does not pay for the normal cast.
func TestBroPrototypeCardsCastBothWays(t *testing.T) {
	for _, r := range protoRows {
		for _, prototyped := range []bool{true, false} {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastProtoRow(t, g, r, prototyped)
			perm, _ := g.LookupCardForEffect(id)
			eff := perm.Effective()
			wantCost, wantP, wantT, wantColors := r.Cost, r.Power, r.Toughness, ""
			if prototyped {
				wantCost, wantP, wantT, wantColors = r.InsetCost, r.InsetPower, r.InsetToughness, r.InsetColor
			}
			if perm.Prototyped() != prototyped || perm.ManaCost != wantCost || eff.Power != wantP || eff.Toughness != wantT ||
				strings.Join(eff.Colors, "") != wantColors {
				t.Errorf("%s (prototyped=%v) = %s %d/%d %v, want %s %d/%d %q",
					r.Name, prototyped, perm.ManaCost, eff.Power, eff.Toughness, eff.Colors, wantCost, wantP, wantT, wantColors)
			}
			if perm.Name != r.Name || !perm.IsArtifact() || !perm.IsCreature() {
				t.Errorf("%s (prototyped=%v) lost its name or types: %q %v", r.Name, prototyped, perm.Name, eff.Types)
			}
			for _, kw := range r.Keywords {
				if !game.HasKeyword(&perm, kw) {
					t.Errorf("%s (prototyped=%v) has no %s", r.Name, prototyped, kw)
				}
			}
		}

		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		c := r.Card(t, me.ID)
		me.Hand.PushTop(c)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		effectstest.FloatMana(t, g, me, effectstest.ProtoManaFor(t, r.InsetCost))
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("%s was cast normally for its prototype cost", r.Name)
		}
	}
}

// Boulderbranch Golem: "its power" is 3 prototyped and 6 otherwise.
func TestBoulderbranchGolemGainsLifeEqualToItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 3}, {false, 6}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		effectstest.CastProtoRow(t, g, protoRowNamed(t, "Boulderbranch Golem"), tc.prototyped)
		if got := me.Life - before; got != tc.want {
			t.Errorf("prototyped=%v: gained %d life, want %d", tc.prototyped, got, tc.want)
		}
	}
}

func TestCombatThresherDrawsACardWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := me.Hand.Size()
	effectstest.CastProtoRow(t, g, protoRowNamed(t, "Combat Thresher"), true)
	// The Thresher went into the hand and out again; the trigger drew one.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want one card drawn", before, got)
	}
}

// Cradle Clearcutter: one {G} from the prototyped 1/3, three from the 3/6.
func TestCradleClearcutterAddsGreenEqualToItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 1}, {false, 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.PushProtoRow(t, g, me.ID, protoRowNamed(t, "Cradle Clearcutter"), tc.prototyped)
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		me.ManaPool = nil
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("ActivateManaAbility: %v", err)
		}
		green := 0
		for _, m := range me.ManaPool {
			if m.Color == "G" {
				green += max(m.Count, 1)
			}
		}
		if green != tc.want || me.ManaPool.Total() != tc.want {
			t.Errorf("prototyped=%v: pool %+v, want %d {G}", tc.prototyped, me.ManaPool, tc.want)
		}
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
			t.Error("the tapped Clearcutter tapped for mana again")
		}
	}
}

// Depth Charge Colossus stays tapped through its controller's untap
// step, and {3} untaps it.
func TestDepthChargeColossusDoesntUntapAndUntapsForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.PushProtoRow(t, g, me.ID, protoRowNamed(t, "Depth Charge Colossus"), true)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepPrecombatMain)
	if !effectstest.B16Tapped(t, g, id) {
		t.Fatal("the Colossus untapped during its controller's untap step")
	}
	me.ManaPool = nil
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the untap ability was activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, id) {
		t.Error("{3} did not untap the Colossus")
	}
}

func TestFallajiDragonEngineFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushProtoRow(t, g, me.ID, protoRowNamed(t, "Fallaji Dragon Engine"), true)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("the prototyped Engine is %d/%d after two activations, want 3/3", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a third activation was paid from an empty pool")
	}
}

// Spotter Thopter: scry 2 prototyped, scry 4 otherwise.
func TestSpotterThopterScriesItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 2}, {false, 4}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.CastProtoRow(t, g, protoRowNamed(t, "Spotter Thopter"), tc.prototyped)
		scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
		if scry == nil {
			t.Fatalf("prototyped=%v: no scry prompt", tc.prototyped)
		}
		if scry.Chooser != me.ID || len(scry.ScryCards) != tc.want {
			t.Errorf("prototyped=%v: scry %d, want %d", tc.prototyped, len(scry.ScryCards), tc.want)
		}
		if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
}

// Iron-Craw Crusher: X is its power as the trigger resolves, the target
// must be attacking, and it may be the Crusher.
func TestIronCrawCrusherPumpsAnAttackerByItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		wantPower  int
	}{{true, 4}, {false, 6}} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		crusher := effectstest.PushProtoRow(t, g, me.ID, protoRowNamed(t, "Iron-Craw Crusher"), tc.prototyped)
		bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		home := effectstest.BlbB1Push(g, me.ID, "Homebody", "Creature — Bear", "", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, crusher, bear)
		ch := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: home}); err == nil {
			t.Fatal("a creature that is not attacking was accepted")
		}
		effectstest.AnswerPickTarget(t, g, bear)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, bear); p != tc.wantPower || tough != 2 {
			t.Errorf("prototyped=%v: the bear is %d/%d, want %d/2", tc.prototyped, p, tough, tc.wantPower)
		}
	}
}

func TestIronCrawCrusherMayTargetItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	crusher := effectstest.PushProtoRow(t, g, me.ID, protoRowNamed(t, "Iron-Craw Crusher"), true)
	effectstest.DeclareAttack(t, g, opp.ID, crusher)
	effectstest.AnswerPickTarget(t, g, crusher)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, crusher); p != 4 || tough != 5 {
		t.Errorf("the prototyped Crusher is %d/%d, want 4/5", p, tough)
	}
}

// Autonomous Assembler: a counter on an Assembly-Worker you control,
// itself included, and on nothing else.
func TestAutonomousAssemblerPutsACounterOnAnAssemblyWorker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	id := effectstest.PushProtoRow(t, g, me.ID, protoRowNamed(t, "Autonomous Assembler"), true)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.PushProtoRow(t, g, opp.ID, protoRowNamed(t, "Autonomous Assembler"), false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	effectstest.FloatMana(t, g, me, "{C}")
	for what, target := range map[string]uuid.UUID{"a Bear": bear, "an opponent's Assembly-Worker": theirs} {
		err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
			Strict:  true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
		if err == nil {
			t.Fatalf("%s was accepted as the target", what)
		}
	}
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("the prototyped Assembler is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("the Assembler did not tap for the cost")
	}
}
