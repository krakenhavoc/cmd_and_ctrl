package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b20_test.go — Ravnica: Clue Edition, BNS batch 20 (#5929). Each
// test plays the real catalog entry and asserts what its text says.

const (
	bnsB20SnowDay     = "057729d0-f7ed-465b-91aa-fb2c6d6612b2"
	bnsB20StreetSpasm = "95385d84-550c-4d6c-a889-62bdbc1d518d"
	bnsB20Tandem      = "e42cd4da-f049-4304-a16f-911769985e6c"
	bnsB20Frog        = "74c4c1e2-c50d-4c8f-889a-0d5674dc6d67"
	bnsB20Mist        = "631f04d6-5693-4414-a378-fd839c3b6f6d"
	bnsB20Vines       = "5372445b-5f67-435e-a736-61c6cb4f9c6e"
	bnsB20Wildsize    = "e1f439c7-217f-4b20-a77d-b9014c5de8c0"
	bnsB20Depose      = "c547c9ab-a303-48fb-9579-37f9de9a558b"
	bnsB20Integrity   = "12f252b1-d3da-43af-ab31-0c8a6807c3c4"
	bnsB20Status      = "df7d4964-35a2-42f7-a4f5-05122f78cbba"
	bnsB20Ballroom    = "cb91e842-9f06-4863-a328-2cabe1bcfe27"
	bnsB20Billiard    = "a4fc174e-7fa6-41a8-ae03-255f226840f9"
)

func TestBnsB20EveryCardIsFullWithNoCaveat(t *testing.T) {
	for name, oracle := range map[string]string{
		"Snow Day": bnsB20SnowDay, "Street Spasm": bnsB20StreetSpasm, "Tandem Tactics": bnsB20Tandem,
		"Turn to Frog": bnsB20Frog, "Turn to Mist": bnsB20Mist, "Vines of the Recluse": bnsB20Vines,
		"Wildsize": bnsB20Wildsize, "Depose": bnsB20Depose, "Deploy": bnsB20Depose + "#1",
		"Integrity": bnsB20Integrity, "Intervention": bnsB20Integrity + "#1",
		"Status": bnsB20Status, "Statue": bnsB20Status + "#1",
		"Ballroom": bnsB20Ballroom, "Billiard Room": bnsB20Billiard,
	} {
		spec := requireFullNoCaveats(t, oracle, name)
		if spec.Name != name {
			t.Errorf("%s registered as %q", name, spec.Name)
		}
	}
}

func bnsB20Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

func bnsB20Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card, mana string, params game.CastSpellParams) error {
	t.Helper()
	id := bnsB20Hand(me, c)
	advanceToMain(t, g)
	floatForTest(g, me, mana)
	params.Strict = true
	if err := g.CastSpell(me.ID, id, params); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

func bnsB20Split(me *game.Player, oracle string, faces ...game.Face) uuid.UUID {
	card := game.Card{InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID, Faces: faces}
	card.SettleImported()
	me.Hand.PushTop(card)
	return card.InstanceID
}

func bnsB20CastFace(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, face int, mana string, targets []game.TargetRef) error {
	t.Helper()
	advanceToMain(t, g)
	floatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Face: face, Targets: targets}); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

func TestBnsB20SnowDayTapsFreezesAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	seedLibrary(me, "A", "B", "C")
	handBefore := me.Hand.Size()
	err := bnsB20Cast(t, g, me, game.Card{Name: "Snow Day", TypeLine: "Instant", OracleID: bnsB20SnowDay, ManaCost: "{4}{U}{U}"},
		"CCCCUU", game.CastSpellParams{Targets: cardRefs(a, b)})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, id := range []uuid.UUID{a, b} {
		if c := findBattlefieldCardByID(g, id); c == nil || !c.Tapped {
			t.Error("a target was not tapped")
		}
	}
	discardFromHand(t, g, me.ID)
	// The spell left the hand, two were drawn and one discarded.
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand %d, want %d (draw two, discard one)", me.Hand.Size(), handBefore+1)
	}
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepUpkeep)
	for _, id := range []uuid.UUID{a, b} {
		if c := findBattlefieldCardByID(g, id); c == nil || !c.Tapped {
			t.Error("a target untapped during its controller's next untap step")
		}
	}
}

func TestBnsB20SnowDayRefusesThreeTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Snow Day", TypeLine: "Instant", OracleID: bnsB20SnowDay, ManaCost: "{4}{U}{U}"},
		"CCCCUU", game.CastSpellParams{Targets: cardRefs(a, b, c)}); err == nil {
		t.Fatal("cast with three targets")
	}
}

func bnsB20Spasm() game.Card {
	return game.Card{Name: "Street Spasm", TypeLine: "Instant", OracleID: bnsB20StreetSpasm, ManaCost: "{X}{R}", Colors: []string{"R"}}
}

func TestBnsB20StreetSpasmDealsXToAGroundCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	if err := bnsB20Cast(t, g, me, bnsB20Spasm(), "CCR", game.CastSpellParams{XValue: 2, Targets: cardRefs(bird)}); err == nil {
		t.Fatal("targeted a creature with flying")
	}
	g = newCatalogGame(t)
	me = g.Seats[0]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := bnsB20Cast(t, g, me, bnsB20Spasm(), "CCR", game.CastSpellParams{XValue: 2, Targets: cardRefs(mine)}); err == nil {
		t.Fatal("targeted a creature the caster controls")
	}
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := bnsB20Cast(t, g, me, bnsB20Spasm(), "CCR", game.CastSpellParams{XValue: 2, Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if onBattlefieldNow(g, bear) {
		t.Error("X=2 did not kill the 2/2")
	}
}

func TestBnsB20StreetSpasmOverloadHitsEachOpposingGroundCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bear2 := b12Creature(g, g.Seats[2].ID, "Bear2", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	// X=2: {2}{2}{R}{R} = six mana.
	if err := bnsB20Cast(t, g, me, bnsB20Spasm(), "CCCCRR", game.CastSpellParams{XValue: 2, AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	if onBattlefieldNow(g, bear) || onBattlefieldNow(g, bear2) {
		t.Error("an opposing ground creature survived the overload")
	}
	if !onBattlefieldNow(g, bird) {
		t.Error("a creature with flying was hit")
	}
	if !onBattlefieldNow(g, mine) {
		t.Error("the caster's own creature was hit")
	}
}

func TestBnsB20StreetSpasmOverloadCostsXTwice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if err := bnsB20Cast(t, g, me, bnsB20Spasm(), "CCRR", game.CastSpellParams{XValue: 2, AlternativeCost: "overload"}); err == nil {
		t.Fatal("overload with X=2 paid for with four mana")
	}
}

func TestBnsB20TandemTacticsPumpsTwoAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	life := me.Life
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Tandem Tactics", TypeLine: "Instant", OracleID: bnsB20Tandem, ManaCost: "{1}{W}"},
		"CW", game.CastSpellParams{Targets: cardRefs(a, b)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectivePower(t, g, a) != 3 || effectiveToughness(t, g, a) != 4 || effectivePower(t, g, b) != 4 || effectiveToughness(t, g, b) != 5 {
		t.Error("each target should get +1/+2")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestBnsB20TandemTacticsWithNoTargetStillGains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Tandem Tactics", TypeLine: "Instant", OracleID: bnsB20Tandem, ManaCost: "{1}{W}"},
		"CW", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestBnsB20TurnToFrogMakesABlueFrogUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dragon", TypeLine: "Creature — Dragon",
		Power: 5, Toughness: 5, Colors: []string{"R"}, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Turn to Frog", TypeLine: "Instant", OracleID: bnsB20Frog, ManaCost: "{1}{U}"},
		"CU", game.CastSpellParams{Targets: cardRefs(dragon)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	eff := layeredCard(t, g, dragon).Effective()
	if eff.Power != 1 || eff.Toughness != 1 {
		t.Errorf("%d/%d, want 1/1", eff.Power, eff.Toughness)
	}
	if hasEffectiveKeyword(t, g, dragon, "flying") {
		t.Error("it kept flying")
	}
	if len(eff.Colors) != 1 || eff.Colors[0] != "U" {
		t.Errorf("colours %v, want [U]", eff.Colors)
	}
	if !containsString(eff.Subtypes, "Frog") || containsString(eff.Subtypes, "Dragon") {
		t.Errorf("subtypes %v, want only Frog", eff.Subtypes)
	}
	advanceToEndStep(t, g)
	advanceToNextSeatsTurn(t, g)
	if p := effectivePower(t, g, dragon); p != 5 {
		t.Errorf("power %d after the turn ended, want 5", p)
	}
}

func TestBnsB20TurnToMistReturnsTheCreatureAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Turn to Mist", TypeLine: "Instant", OracleID: bnsB20Mist, ManaCost: "{1}{W/U}"},
		"CU", game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if onBattlefieldNow(g, bear) {
		t.Fatal("the creature was not exiled")
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == opp.ID {
			back = true
		}
	}
	if !back {
		t.Error("the creature did not return under its owner's control at the end step")
	}
}

func TestBnsB20VinesOfTheReclusePumpsGrantsReachAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Tapped = true })
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Vines of the Recluse", TypeLine: "Instant", OracleID: bnsB20Vines, ManaCost: "{G}"},
		"G", game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectivePower(t, g, bear) != 3 || effectiveToughness(t, g, bear) != 4 || !hasEffectiveKeyword(t, g, bear, "reach") {
		t.Error("expected a 3/4 with reach")
	}
	if c := findBattlefieldCardByID(g, bear); c == nil || c.Tapped {
		t.Error("the creature was not untapped")
	}
}

func TestBnsB20WildsizePumpsTramplesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B")
	before := me.Hand.Size()
	if err := bnsB20Cast(t, g, me, game.Card{Name: "Wildsize", TypeLine: "Instant", OracleID: bnsB20Wildsize, ManaCost: "{2}{G}"},
		"CCG", game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 4 || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("expected a 4/4 with trample")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d: the spell leaves and a card is drawn", me.Hand.Size(), before+1)
	}
}

func bnsB20DeposeDeploy(me *game.Player) uuid.UUID {
	return bnsB20Split(me, bnsB20Depose,
		game.Face{Name: "Depose", TypeLine: "Instant", ManaCost: "{1}{W/U}", OracleText: "Tap target creature.\nDraw a card."},
		game.Face{Name: "Deploy", TypeLine: "Instant", ManaCost: "{2}{W}{U}", OracleText: "Create two 1/1 colorless Thopter artifact creature tokens with flying, then you gain 1 life for each creature you control."})
}

func TestBnsB20DeposeTapsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibrary(me, "A", "B")
	id := bnsB20DeposeDeploy(me)
	before := me.Hand.Size()
	if err := bnsB20CastFace(t, g, me, id, 0, "CW", cardRefs(bear)); err != nil {
		t.Fatalf("cast Depose: %v", err)
	}
	if c := findBattlefieldCardByID(g, bear); c == nil || !c.Tapped {
		t.Error("the creature was not tapped")
	}
	if me.Hand.Size() != before { // the split card left, one card drawn
		t.Errorf("hand %d, want %d", me.Hand.Size(), before)
	}
}

func TestBnsB20DeployMakesTwoThoptersThenGainsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	id := bnsB20DeposeDeploy(me)
	if err := bnsB20CastFace(t, g, me, id, 1, "CCWU", nil); err != nil {
		t.Fatalf("cast Deploy: %v", err)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Thopter" {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("%d Thopters, want 2", n)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (Bear and two Thopters)", me.Life, life+3)
	}
}

func bnsB20IntegrityIntervention(me *game.Player) uuid.UUID {
	return bnsB20Split(me, bnsB20Integrity,
		game.Face{Name: "Integrity", TypeLine: "Instant", ManaCost: "{R/W}", OracleText: "Target creature gets +2/+2 until end of turn."},
		game.Face{Name: "Intervention", TypeLine: "Instant", ManaCost: "{2}{R}{W}", OracleText: "Intervention deals 3 damage to any target and you gain 3 life."})
}

func TestBnsB20IntegrityPumpsTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := bnsB20IntegrityIntervention(me)
	if err := bnsB20CastFace(t, g, me, id, 0, "R", cardRefs(bear)); err != nil {
		t.Fatalf("cast Integrity: %v", err)
	}
	if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 4 {
		t.Error("expected 4/4")
	}
}

func TestBnsB20InterventionDamagesAnyTargetAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := bnsB20IntegrityIntervention(me)
	life, oppLife := me.Life, opp.Life
	if err := bnsB20CastFace(t, g, me, id, 1, "CCRW", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("cast Intervention: %v", err)
	}
	if opp.Life != oppLife-3 || me.Life != life+3 {
		t.Errorf("opp %d (want %d), me %d (want %d)", opp.Life, oppLife-3, me.Life, life+3)
	}
}

func bnsB20StatusStatue(me *game.Player) uuid.UUID {
	return bnsB20Split(me, bnsB20Status,
		game.Face{Name: "Status", TypeLine: "Instant", ManaCost: "{B/G}", OracleText: "Target creature gets +1/+1 and gains deathtouch until end of turn."},
		game.Face{Name: "Statue", TypeLine: "Instant", ManaCost: "{2}{B}{G}", OracleText: "Destroy target artifact, creature, or enchantment."})
}

func TestBnsB20StatusPumpsAndGrantsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := bnsB20StatusStatue(me)
	if err := bnsB20CastFace(t, g, me, id, 0, "G", cardRefs(bear)); err != nil {
		t.Fatalf("cast Status: %v", err)
	}
	if effectivePower(t, g, bear) != 3 || effectiveToughness(t, g, bear) != 3 || !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("expected a 3/3 with deathtouch")
	}
}

func TestBnsB20StatueDestroysAnArtifactCreatureOrEnchantmentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	id := bnsB20StatusStatue(me)
	if err := bnsB20CastFace(t, g, me, id, 1, "CCBG", cardRefs(land)); err == nil {
		t.Fatal("Statue targeted a land")
	}
	for _, tc := range []struct{ name, typeLine string }{
		{"Bear", "Creature — Bear"}, {"Rod", "Artifact"}, {"Ward", "Enchantment"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := b12Permanent(g, opp.ID, tc.name, tc.typeLine)
		if tc.typeLine == "Creature — Bear" {
			victim = b12Creature(g, opp.ID, tc.name, tc.typeLine, 2, 2)
		}
		id := bnsB20StatusStatue(me)
		if err := bnsB20CastFace(t, g, me, id, 1, "CCBG", cardRefs(victim)); err != nil {
			t.Fatalf("Statue on %s: %v", tc.name, err)
		}
		if onBattlefieldNow(g, victim) {
			t.Errorf("Statue left the %s", tc.name)
		}
	}
}

func bnsB20Land(t *testing.T, name, oracle string, colors []string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, name, oracle)
	passPriorityAroundTable(t, g)
	top100AssertEnteredTapped(t, g, id, name)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Tapped = false })
	spm5TapForEach(t, g, me, id, 2, colors)

	// {4}, {T}: Investigate.
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("investigate activated with three mana")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, id) {
		t.Error("the {T} cost was not paid")
	}
	clues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.Name == "Clue" {
			clues++
		}
	}
	if clues != 1 {
		t.Errorf("%d Clues, want 1", clues)
	}
}

func TestBnsB20BallroomEntersTappedMakesWOrBAndInvestigates(t *testing.T) {
	bnsB20Land(t, "Ballroom", bnsB20Ballroom, []string{"W", "B"})
}

func TestBnsB20BilliardRoomEntersTappedMakesBOrRAndInvestigates(t *testing.T) {
	bnsB20Land(t, "Billiard Room", bnsB20Billiard, []string{"B", "R"})
}
