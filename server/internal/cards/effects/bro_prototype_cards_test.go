package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bro_prototype_cards_test.go — the eleven Brothers' War prototype cards
// (#5656), each cast both ways through the engine, and each clause that
// reads "this creature's power" checked against both sizes.

const protoReminder = " (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)"

// protoRow is one card as printed, with what its inset says.
type protoRow struct {
	oracle, name, typeLine, cost string
	power, toughness             int
	text                         string

	insetCost, insetColor      string
	insetPower, insetToughness int
	keywords                   []string
}

var protoRows = []protoRow{
	{"19f344ee-031d-4346-9553-c5d32b02df43", "Blitz Automaton", "Artifact Creature — Construct", "{7}", 6, 4,
		"Prototype {2}{R} — 3/2" + protoReminder + "\nHaste", "{2}{R}", "R", 3, 2, []string{"haste"}},
	{"02457c7f-eced-4bf3-9b61-309b4ceab2d5", "Goring Warplow", "Artifact Creature — Construct", "{6}", 5, 4,
		"Prototype {1}{B} — 1/1" + protoReminder + "\nDeathtouch", "{1}{B}", "B", 1, 1, []string{"deathtouch"}},
	{"61df991d-32e7-4b6b-b8e6-700ba915c3f2", "Rust Goliath", "Artifact Creature — Construct", "{10}", 10, 10,
		"Prototype {3}{G}{G} — 3/5" + protoReminder + "\nReach, trample", "{3}{G}{G}", "G", 3, 5, []string{"reach", "trample"}},
	{"82caf3e2-da42-4851-b4db-546e2403c4bf", "Boulderbranch Golem", "Artifact Creature — Golem", "{7}", 6, 5,
		"Prototype {3}{G} — 3/3" + protoReminder + "\nWhen this creature enters, you gain life equal to its power.", "{3}{G}", "G", 3, 3, nil},
	{"0c460bdd-460b-4312-9ed2-32f3809a7d36", "Combat Thresher", "Artifact Creature — Construct", "{7}", 3, 3,
		"Prototype {2}{W} — 1/1" + protoReminder + "\nDouble strike\nWhen this creature enters, draw a card.", "{2}{W}", "W", 1, 1, []string{"double strike"}},
	{"225f53ef-da42-47a0-b773-d2aad3338c72", "Cradle Clearcutter", "Artifact Creature — Golem", "{6}", 3, 6,
		"Prototype {2}{G} — 1/3" + protoReminder + "\n{T}: Add an amount of {G} equal to this creature's power.", "{2}{G}", "G", 1, 3, nil},
	{"f376f4c6-4526-4ae3-a84a-e8ea50fe78b6", "Depth Charge Colossus", "Artifact Creature — Dreadnought", "{9}", 9, 9,
		"Prototype {4}{U}{U} — 6/6" + protoReminder + "\nThis creature doesn't untap during your untap step.\n{3}: Untap this creature.", "{4}{U}{U}", "U", 6, 6, nil},
	{"dc8cca38-df2a-4cef-973c-a482505be66f", "Fallaji Dragon Engine", "Artifact Creature — Dragon", "{8}", 5, 5,
		"Prototype {2}{R} — 1/3" + protoReminder + "\nFlying\n{2}: This creature gets +1/+0 until end of turn.", "{2}{R}", "R", 1, 3, []string{"flying"}},
	{"ccf13531-60bd-4ebd-9a5b-68e2d05ce620", "Spotter Thopter", "Artifact Creature — Thopter", "{8}", 4, 5,
		"Prototype {3}{U} — 2/3" + protoReminder + "\nFlying\nWhen this creature enters, scry X, where X is its power.", "{3}{U}", "U", 2, 3, []string{"flying"}},
	{"b04f45de-cb35-45a4-b039-0bf008c3ca56", "Iron-Craw Crusher", "Artifact Creature — Wurm", "{7}", 4, 6,
		"Prototype {2}{G}{G} — 2/5" + protoReminder + "\nWhenever this creature attacks, target attacking creature gets +X/+0 until end of turn, where X is this creature's power.", "{2}{G}{G}", "G", 2, 5, nil},
	{"3afdc98e-aeeb-43e9-bdcd-0c51da44d2f8", "Autonomous Assembler", "Artifact Creature — Assembly-Worker", "{5}", 4, 5,
		"Prototype {1}{W} — 2/2" + protoReminder + "\nVigilance\n{1}, {T}: Put a +1/+1 counter on target Assembly-Worker you control.", "{1}{W}", "W", 2, 2, []string{"vigilance"}},
}

func protoRowNamed(t *testing.T, name string) protoRow {
	t.Helper()
	for _, r := range protoRows {
		if r.name == name {
			return r
		}
	}
	t.Fatalf("no prototype row for %s", name)
	return protoRow{}
}

func (r protoRow) card(t *testing.T, owner uuid.UUID) game.Card {
	return prototypeCard(t, owner, r.oracle, r.name, r.typeLine, r.cost, r.power, r.toughness, r.text)
}

// protoManaFor is exactly the mana a cost asks for, generic paid with {C}.
func protoManaFor(t *testing.T, cost string) string {
	t.Helper()
	parsed, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%s): %v", cost, err)
	}
	rest := cost
	if i := strings.Index(cost, "}"); parsed.Generic > 0 && i >= 0 {
		rest = cost[i+1:]
	}
	return strings.Repeat("{C}", parsed.Generic) + rest
}

// castProtoRow casts the card from the active player's hand at their
// main phase, prototyped or not, paying exactly its cost, and resolves
// it. It leaves any prompt the entry opened unanswered.
func castProtoRow(t *testing.T, g *game.Game, r protoRow, prototyped bool) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c := r.card(t, me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	params := game.CastSpellParams{Strict: true}
	cost := r.cost
	if prototyped {
		params.Face = game.PrototypeFace
		cost = r.insetCost
	}
	me.ManaPool = nil
	floatMana(t, g, me, protoManaFor(t, cost))
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s (prototyped=%v): %v", r.name, prototyped, err)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Fatalf("%s (prototyped=%v) left %d mana floating, want its whole cost spent", r.name, prototyped, left)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatalf("%s did not resolve onto the battlefield", r.name)
	}
	return c.InstanceID
}

// pushProtoRow puts the card on the battlefield as the permanent a cast
// would have made, untapped and able to attack and tap.
func pushProtoRow(t *testing.T, g *game.Game, owner uuid.UUID, r protoRow, prototyped bool) uuid.UUID {
	t.Helper()
	c := r.card(t, owner)
	if prototyped {
		c.SetFace(game.PrototypeFace)
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestBroPrototypeCardsAreRegisteredFull(t *testing.T) {
	for _, r := range protoRows {
		spec, ok := Lookup(r.oracle)
		if !ok {
			t.Errorf("%s is not registered", r.name)
			continue
		}
		if spec.Name != r.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", r.name, spec.Name, spec.Completeness, spec.Caveats)
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
			g := newCatalogGame(t)
			id := castProtoRow(t, g, r, prototyped)
			perm, _ := g.LookupCardForEffect(id)
			eff := perm.Effective()
			wantCost, wantP, wantT, wantColors := r.cost, r.power, r.toughness, ""
			if prototyped {
				wantCost, wantP, wantT, wantColors = r.insetCost, r.insetPower, r.insetToughness, r.insetColor
			}
			if perm.Prototyped() != prototyped || perm.ManaCost != wantCost || eff.Power != wantP || eff.Toughness != wantT ||
				strings.Join(eff.Colors, "") != wantColors {
				t.Errorf("%s (prototyped=%v) = %s %d/%d %v, want %s %d/%d %q",
					r.name, prototyped, perm.ManaCost, eff.Power, eff.Toughness, eff.Colors, wantCost, wantP, wantT, wantColors)
			}
			if perm.Name != r.name || !perm.IsArtifact() || !perm.IsCreature() {
				t.Errorf("%s (prototyped=%v) lost its name or types: %q %v", r.name, prototyped, perm.Name, eff.Types)
			}
			for _, kw := range r.keywords {
				if !game.HasKeyword(&perm, kw) {
					t.Errorf("%s (prototyped=%v) has no %s", r.name, prototyped, kw)
				}
			}
		}

		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		c := r.card(t, me.ID)
		me.Hand.PushTop(c)
		advanceTo(t, g, game.StepPrecombatMain)
		floatMana(t, g, me, protoManaFor(t, r.insetCost))
		if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("%s was cast normally for its prototype cost", r.name)
		}
	}
}

// Boulderbranch Golem: "its power" is 3 prototyped and 6 otherwise.
func TestBoulderbranchGolemGainsLifeEqualToItsPower(t *testing.T) {
	for _, tc := range []struct {
		prototyped bool
		want       int
	}{{true, 3}, {false, 6}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		castProtoRow(t, g, protoRowNamed(t, "Boulderbranch Golem"), tc.prototyped)
		if got := me.Life - before; got != tc.want {
			t.Errorf("prototyped=%v: gained %d life, want %d", tc.prototyped, got, tc.want)
		}
	}
}

func TestCombatThresherDrawsACardWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepPrecombatMain)
	before := me.Hand.Size()
	castProtoRow(t, g, protoRowNamed(t, "Combat Thresher"), true)
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
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		id := pushProtoRow(t, g, me.ID, protoRowNamed(t, "Cradle Clearcutter"), tc.prototyped)
		advanceTo(t, g, game.StepPrecombatMain)
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := pushProtoRow(t, g, me.ID, protoRowNamed(t, "Depth Charge Colossus"), true)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	advanceTo(t, g, game.StepEnd)
	advanceToStepOf(t, g, seat, game.StepPrecombatMain)
	if !b16Tapped(t, g, id) {
		t.Fatal("the Colossus untapped during its controller's untap step")
	}
	me.ManaPool = nil
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the untap ability was activated with no mana")
	}
	floatMana(t, g, me, "{C}{C}{C}")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, id) {
		t.Error("{3} did not untap the Colossus")
	}
}

func TestFallajiDragonEngineFirebreathes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushProtoRow(t, g, me.ID, protoRowNamed(t, "Fallaji Dragon Engine"), true)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	floatMana(t, g, me, "{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 3 || tough != 3 {
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
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		castProtoRow(t, g, protoRowNamed(t, "Spotter Thopter"), tc.prototyped)
		scry := latestChoiceOfKind(g, game.PendingChoiceScry)
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
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		crusher := pushProtoRow(t, g, me.ID, protoRowNamed(t, "Iron-Craw Crusher"), tc.prototyped)
		bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		home := blbB1Push(g, me.ID, "Homebody", "Creature — Bear", "", 2, 2)
		declareAttack(t, g, opp.ID, crusher, bear)
		ch := pickTargetPrompt(t, g)
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: home}); err == nil {
			t.Fatal("a creature that is not attacking was accepted")
		}
		answerPickTarget(t, g, bear)
		passPriorityAroundTable(t, g)
		if p, tough := ptOf(t, g, bear); p != tc.wantPower || tough != 2 {
			t.Errorf("prototyped=%v: the bear is %d/%d, want %d/2", tc.prototyped, p, tough, tc.wantPower)
		}
	}
}

func TestIronCrawCrusherMayTargetItself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	crusher := pushProtoRow(t, g, me.ID, protoRowNamed(t, "Iron-Craw Crusher"), true)
	declareAttack(t, g, opp.ID, crusher)
	answerPickTarget(t, g, crusher)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, crusher); p != 4 || tough != 5 {
		t.Errorf("the prototyped Crusher is %d/%d, want 4/5", p, tough)
	}
}

// Autonomous Assembler: a counter on an Assembly-Worker you control,
// itself included, and on nothing else.
func TestAutonomousAssemblerPutsACounterOnAnAssemblyWorker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	id := pushProtoRow(t, g, me.ID, protoRowNamed(t, "Autonomous Assembler"), true)
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := pushProtoRow(t, g, opp.ID, protoRowNamed(t, "Autonomous Assembler"), false)
	advanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool = nil
	floatMana(t, g, me, "{C}")
	for what, target := range map[string]uuid.UUID{"a Bear": bear, "an opponent's Assembly-Worker": theirs} {
		err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
			Strict:  true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
		if err == nil {
			t.Fatalf("%s was accepted as the target", what)
		}
	}
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
	})
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("the prototyped Assembler is %d/%d, want 3/3", p, tough)
	}
	if !b16Tapped(t, g, id) {
		t.Error("the Assembler did not tap for the cost")
	}
}
