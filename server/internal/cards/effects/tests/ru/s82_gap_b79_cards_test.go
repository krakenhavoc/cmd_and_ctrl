package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b79_cards_test.go — S82 GAP batch 79 (Ravnica Remastered
// instants, #5948). Every test plays the real catalog entry through the
// engine.

const (
	gapB79Assure    = "087018b6-6355-4534-9a35-9336b4b7fccc"
	gapB79Warrant   = "19053c47-97ce-4a25-a92c-1f4061e85bf3"
	gapB79Downsize  = "46537bac-b1eb-4c02-ba43-6c6b16c7729c"
	gapB79Zeal      = "6b70af46-ff27-4eec-9692-248b06248de6"
	gapB79Homing    = "6e7f8e25-49a9-49a0-9bf4-129439344550"
	gapB79Turnabout = "715d0777-4d36-463b-9102-6bb6bb00b7e9"
	gapB79Response  = "ae93854b-c808-44cb-b0be-7d1035c597d7"
	gapB79Lost      = "e0462a2c-fb88-495f-813b-6476ee3e62bb"
)

func TestGapB79EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapB79Assure:          "Assure",
		gapB79Assure + "#1":   "Assemble",
		gapB79Warrant:         "Warrant",
		gapB79Warrant + "#1":  "Warden",
		gapB79Downsize:        "Downsize",
		gapB79Zeal:            "Arrester's Zeal",
		gapB79Homing:          "Homing Lightning",
		gapB79Turnabout:       "Turnabout",
		gapB79Response:        "Response",
		gapB79Response + "#1": "Resurgence",
		gapB79Lost:            "Totally Lost",
	}
	for oracle, name := range want {
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

// gapB79CastSplit casts one half of a split card for the active seat.
func gapB79CastSplit(t *testing.T, g *game.Game, oracle, left, right string, face int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: left, TypeLine: "Instant", ManaCost: "{0}"},
			{Name: right, TypeLine: "Instant", ManaCost: "{0}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", left, err)
	}
}

// gapB79CastNow casts a card for the active seat in whatever step it is.
func gapB79CastNow(g *game.Game, name, oracle string, targets []game.TargetRef, strict bool) error {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Strict: strict})
}

func TestGapB79AssureCountersAndProtectsAndAssembleMakesThreeKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	gapB79CastSplit(t, g, gapB79Assure, "Assure", "Assemble", 0, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the creature is not indestructible")
	}
	gapB79CastSplit(t, g, gapB79Assure, "Assure", "Assemble", 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.NamedOnBattlefield(g, "Elf Knight", me.ID); n != 3 {
		t.Fatalf("%d Elf Knights, want 3", n)
	}
	for _, c := range effectstest.TokensNamed(g, "Elf Knight") {
		if c.Power != 2 || c.Toughness != 2 {
			t.Errorf("Elf Knight is %d/%d, want 2/2", c.Power, c.Toughness)
		}
		if !game.HasKeyword(&c, "vigilance") {
			t.Error("an Elf Knight lacks vigilance")
		}
	}
}

func TestGapB79WarrantTucksAnAttackerOnlyAndWardenMakesASphinx(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if err := gapB79CastNow(g, "Warrant", gapB79Warrant, effectstest.CardTarget(idle), true); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := gapB79CastNow(g, "Warrant", gapB79Warrant, effectstest.CardTarget(bear), false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, bear) {
		t.Error("the attacker stayed")
	}
	if cards := me.Library.Cards; len(cards) == 0 || cards[len(cards)-1].InstanceID != bear {
		t.Error("the attacker is not on top of its owner's library")
	}
}

func TestGapB79WardenMakesASphinx(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gapB79CastSplit(t, g, gapB79Warrant, "Warrant", "Warden", 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.TokensNamed(g, "Sphinx")
	if len(toks) != 1 || toks[0].Controller != me.ID || toks[0].Power != 4 || toks[0].Toughness != 4 {
		t.Fatalf("Sphinx tokens: %+v", toks)
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !game.HasKeyword(&toks[0], kw) {
			t.Errorf("the Sphinx lacks %s", kw)
		}
	}
}

func TestGapB79TotallyLostTucksANonlandPermanentOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Totally Lost", "Instant", gapB79Lost, effectstest.CardTarget(land)); err == nil {
		t.Fatal("targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Totally Lost", "Instant", gapB79Lost, effectstest.CardTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, rock) {
		t.Error("the artifact stayed")
	}
	if cards := opp.Library.Cards; len(cards) == 0 || cards[len(cards)-1].InstanceID != rock {
		t.Error("the artifact is not on top of its owner's library")
	}
}

func TestGapB79DownsizeShrinksOneOrEachCreatureYouDontControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 5, 5)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 5, 5)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Downsize", "Instant", gapB79Downsize, effectstest.CardTarget(mine)); err == nil {
		t.Fatal("targeted my own creature")
	}
	effectstest.CastCatalogSpell(t, g, "Downsize", "Instant", gapB79Downsize, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if pa, _ := effectstest.EffectivePTOf(t, g, a); pa != 1 {
		t.Errorf("the target has power %d, want 1", pa)
	}
	if pb, _ := effectstest.EffectivePTOf(t, g, b); pb != 5 {
		t.Errorf("the other creature has power %d, want 5", pb)
	}
	effectstest.CastWithAltCost(t, g, "Downsize", "Instant", gapB79Downsize, "overload")
	effectstest.PassPriorityAroundTable(t, g)
	if pa, _ := effectstest.EffectivePTOf(t, g, a); pa != -3 {
		t.Errorf("A has power %d, want -3", pa)
	}
	if pb, _ := effectstest.EffectivePTOf(t, g, b); pb != 1 {
		t.Errorf("B has power %d, want 1", pb)
	}
	if pm, _ := effectstest.EffectivePTOf(t, g, mine); pm != 5 {
		t.Errorf("my creature has power %d, want 5", pm)
	}
}

func TestGapB79ArrestersZealHasFlyingOnlyInYourMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Arrester's Zeal", "Instant", gapB79Zeal, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.EffectivePTOf(t, g, bear); p != 4 || to != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("no flying when cast in my main phase")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	bear2 := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	id := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: id, Name: "Arrester's Zeal", TypeLine: "Instant", OracleID: gapB79Zeal, Owner: me2.ID, Controller: me2.ID})
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(bear2)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if p, _ := effectstest.EffectivePTOf(t, g2, bear2); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if effectstest.HasEffectiveKeyword(t, g2, bear2, "flying") {
		t.Error("flying when cast outside a main phase")
	}
}

func TestGapB79HomingLightningHitsEveryCreatureWithTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 6)
	b := effectstest.B12Creature(g, me.ID, "Goblin", "Creature — Goblin", 1, 6)
	c := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 6)
	effectstest.CastCatalogSpell(t, g, "Homing Lightning", "Instant", gapB79Homing, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{a: 4, b: 4, c: 0} {
		if got := effectstest.FindCardAnywhere(t, g, id).DamageMarked; got != want {
			t.Errorf("%s took %d, want %d", effectstest.FindCardAnywhere(t, g, id).Name, got, want)
		}
	}
}

func TestGapB79ResponseHitsAnAttackerAndResurgenceAddsACombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 6)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 6)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if err := gapB79CastNow(g, "Response", gapB79Response, effectstest.CardTarget(idle), true); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := gapB79CastNow(g, "Response", gapB79Response, effectstest.CardTarget(attacker), false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, attacker).DamageMarked; got != 5 {
		t.Errorf("the attacker took %d, want 5", got)
	}
}

func TestGapB79ResurgenceGrantsFirstStrikeAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	gapB79CastSplit(t, g, gapB79Response, "Response", "Resurgence", 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"first strike", "vigilance"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("their creature gained %s", kw)
		}
	}
	var sawCombat bool
	for range 40 {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
		if g.Turn.Step == game.StepBeginCombat {
			sawCombat = true
			break
		}
	}
	if !sawCombat {
		t.Fatal("no combat followed the main phase")
	}
	var combats int
	for range 40 {
		if g.Turn.Step == game.StepBeginCombat {
			combats++
		}
		if g.Turn.Step == game.StepCleanup || g.Turn.ActiveSeat != 0 {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	if combats < 2 {
		t.Errorf("%d combat phases after the main phase, want 2 (one extra)", combats)
	}
}

func TestGapB79TurnaboutTapsOrUntapsTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Turnabout", "Instant", gapB79Turnabout, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if pick == nil {
		t.Fatalf("no option prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, 2); err != nil { // tap creatures
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	tapped := func(id uuid.UUID) bool { return effectstest.FindBattlefieldCardForTest(g, id).Tapped }
	if !tapped(bear) {
		t.Error("their creature was not tapped")
	}
	if tapped(rock) {
		t.Error("their artifact was tapped (not the chosen type)")
	}
	if tapped(mine) {
		t.Error("my creature was tapped (not the target player's)")
	}
	effectstest.CastCatalogSpell(t, g, "Turnabout", "Instant", gapB79Turnabout, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	pick = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if pick == nil {
		t.Fatal("no second option prompt")
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, 3); err != nil { // untap creatures
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if tapped(bear) {
		t.Error("their creature stayed tapped")
	}
}
