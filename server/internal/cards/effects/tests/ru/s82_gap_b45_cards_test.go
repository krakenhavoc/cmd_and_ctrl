package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b45_cards_test.go — GAP batch 45 (#5948, S82): Modern Horizons 3
// Commander creatures and enchantments. Each test plays the real catalog
// entry.

const (
	gapB45Electromancer   = "8e576de0-0846-4665-a07c-83fe483ac196"
	gapB45Cayth           = "d18a014e-16a0-4641-9546-ff3729437eed"
	gapB45Satya           = "ce8af7dd-49df-4419-9776-4157efc60b0f"
	gapB45AuroraShifter   = "d499e473-aed8-4c8b-8cf0-a43bcc5f40cf"
	gapB45BenthicAnomaly  = "c12e2563-4ad4-4637-b9ae-72f23cca31d8"
	gapB45Skulker         = "a09a4556-8913-495c-8768-e05413c48f0b"
	gapB45DrownerOfHope   = "2b3f2e85-5ea4-44d6-86e8-e8ce6d4f98c0"
	gapB45WonderscapeSage = "99994c36-21ec-4a28-b850-d3f2ba8b7f19"
	gapB45Aberration      = "60d0c941-39f6-4144-abef-02b0a84d602e"
	gapB45Lecturer        = "ac230b68-f945-4a1a-bd89-33e6d62e3c36"
	gapB45DesertWarfare   = "de301ef8-8ba0-423d-9117-a8492a58d01b"
	gapB45TarmogoyfNest   = "29fede1b-e379-46d4-93e9-a589c896add6"
)

func TestGapB45CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB45Electromancer:   "Overclocked Electromancer",
		gapB45Cayth:           "Cayth, Famed Mechanist",
		gapB45Satya:           "Satya, Aetherflux Genius",
		gapB45AuroraShifter:   "Aurora Shifter",
		gapB45BenthicAnomaly:  "Benthic Anomaly",
		gapB45Skulker:         "Deepfathom Skulker",
		gapB45DrownerOfHope:   "Drowner of Hope",
		gapB45WonderscapeSage: "Wonderscape Sage",
		gapB45Aberration:      "Angelic Aberration",
		gapB45Lecturer:        "Silverquill Lecturer",
		gapB45DesertWarfare:   "Desert Warfare",
		gapB45TarmogoyfNest:   "Tarmogoyf Nest",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB45DrownerOfHopeMakesScionsAndSacrificesOneToTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.M3cb1Opponent(g)
	drowner := effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Drowner of Hope", TypeLine: "Creature — Eldrazi",
		OracleID: gapB45DrownerOfHope, ManaCost: "{5}{U}", Power: 5, Toughness: 5})
	effectstest.PassPriorityAroundTable(t, g)
	scions := effectstest.S46Tokens(g, me.ID, "Eldrazi Scion")
	if len(scions) != 2 {
		t.Fatalf("%d Eldrazi Scions, want 2", len(scions))
	}
	victim := effectstest.M3cb1Push(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, drowner, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), SacrificeIDs: []uuid.UUID{scions[0].InstanceID}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, victim).Tapped {
		t.Error("the target was not tapped")
	}
	if got := len(effectstest.S46Tokens(g, me.ID, "Eldrazi Scion")); got != 1 {
		t.Errorf("%d Scions left, want 1", got)
	}
	if effectstest.LayeredCard(t, g, drowner).HasColor("U") {
		t.Error("devoid: the Drowner is blue")
	}
}

func TestGapB45SkulkerUnblockableNeedsColorlessMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	skulker := effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Deepfathom Skulker", TypeLine: "Creature — Eldrazi",
		OracleID: gapB45Skulker, ManaCost: "{5}{U}", Power: 4, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{U}{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, skulker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(skulker), Strict: true}); err == nil {
		t.Fatal("paid {3}{C} with coloured mana")
	}
	me.ManaPool = nil
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, skulker, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(skulker), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB45AngelicAberrationSacrificesSmallCreaturesForAngels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	small1 := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Mite", TypeLine: "Creature — Insect", Power: 1, Toughness: 1})
	small2 := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 5})
	big := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3})
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Angelic Aberration", TypeLine: "Creature — Eldrazi Angel",
		OracleID: gapB45Aberration, ManaCost: "{5}{W}", Power: 4, Toughness: 4})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if effectstest.HasID(pick.ChooseCards, big) {
		t.Error("a creature with base power and toughness above 1 is offered")
	}
	if !effectstest.HasID(pick.ChooseCards, small1) || !effectstest.HasID(pick.ChooseCards, small2) {
		t.Error("a creature with base power or toughness 1 or less is not offered")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{small1, small2}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(small1) || g.Battlefield.Contains(small2) || !g.Battlefield.Contains(big) {
		t.Error("exactly the chosen creatures are sacrificed")
	}
	angels := effectstest.S46Tokens(g, me.ID, "Eldrazi Angel")
	if len(angels) != 2 {
		t.Fatalf("%d Eldrazi Angels, want 2", len(angels))
	}
	if p, tg := effectstest.PtOf(t, g, angels[0].InstanceID); p != 4 || tg != 4 {
		t.Errorf("the Angel is %d/%d, want 4/4", p, tg)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, angels[0].InstanceID, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, angels[0].InstanceID, "vigilance") {
		t.Error("the Angel lacks flying or vigilance")
	}
	if !effectstest.LayeredCard(t, g, angels[0].InstanceID).IsColorless() {
		t.Error("the Angel is not colorless")
	}
}

func TestGapB45OverclockedElectromancerDoublesPowerAndGetsExcessEnergy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.M3cb1Opponent(g)
	mancer := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Overclocked Electromancer", TypeLine: "Creature — Lizard Wizard",
		OracleID: gapB45Electromancer, ManaCost: "{2}{R}", Power: 2, Toughness: 2})
	blocker := effectstest.M3cb1Push(g, opp.ID, game.Card{Name: "Mite", TypeLine: "Creature — Insect", Power: 0, Toughness: 1})
	gapB45AdvanceToAttackers(t, g, me, false)
	effectstest.DeclareAttack(t, g, opp.ID, mancer)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mancer); p != 4 {
		t.Fatalf("power after the attack trigger %d, want 4", p)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, mancer); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters["energy"]; got != 3 {
		t.Errorf("energy %d, want 3 (4 damage to a 0/1 is 3 excess)", got)
	}
}

func TestGapB45CaythGrantsFabricateToOtherNontokenCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Cayth, Famed Mechanist", TypeLine: "Legendary Creature — Dwarf Artificer",
		OracleID: gapB45Cayth, ManaCost: "{1}{U}{R}{W}", Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false) // Cayth's own fabricate: a Servo
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Servo")); got != 1 {
		t.Fatalf("%d Servos after Cayth's fabricate, want 1", got)
	}
	bear := effectstest.TrekB31Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true) // a counter on the Bear
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("the Bear has power %d, want 3 with a +1/+1 counter", p)
	}
	// A token entering is not given fabricate.
	g.WithWriteLock(func() {
		_ = effects.CreateToken{Controller: me.ID, Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a token was offered fabricate")
	}
}

func TestGapB45TarmogoyfNestGrantsATarmogoyfMaker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	nest := effectstest.CastCatalogSpell(t, g, "Tarmogoyf Nest", "Kindred Enchantment — Lhurgoyf Aura", gapB45TarmogoyfNest, effectstest.CardRefs(land))
	_ = nest
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Bolt", "Instant")
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	goyfs := effectstest.S46Tokens(g, me.ID, "Tarmogoyf")
	if len(goyfs) != 1 {
		t.Fatalf("%d Tarmogoyf tokens, want 1", len(goyfs))
	}
	if p, tg := effectstest.PtOf(t, g, goyfs[0].InstanceID); p != 1 || tg != 2 {
		t.Errorf("the Tarmogoyf is %d/%d, want 1/2 with an instant in a graveyard", p, tg)
	}
}

func TestGapB45WonderscapeSageDiscardsUnlessTheLandHadANonbasicType(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		discard  bool
	}{{"basic", "Basic Land — Forest", true}, {"desert", "Land — Desert", false}, {"no types", "Land", true}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		sage := effectstest.PushCatalogPermanent(g, me.ID, "Wonderscape Sage", "Creature — Moonfolk Wizard", gapB45WonderscapeSage, false)
		land := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Test Land", TypeLine: tc.typeLine})
		g.WithWriteLock(func() {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Drawn", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		})
		hand := me.Hand.Size()
		if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{land}, Strict: true}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tc.discard {
			if pc := effectstest.DiscardPromptFor(g, me); pc == nil {
				t.Errorf("%s: no discard prompt", tc.name)
			}
		} else if got := me.Hand.Size(); got != hand+2 { // the returned land and the drawn card
			t.Errorf("%s: hand %d -> %d, want +2 and no discard", tc.name, hand, got)
		}
	}
}

// gapB45AdvanceToAttackers walks to the declare attackers step, answering
// the Electromancer's energy prompt at the beginning of combat with pay.
func gapB45AdvanceToAttackers(t *testing.T, g *game.Game, me *game.Player, pay bool) {
	t.Helper()
	for i := 0; i < 40 && g.Turn.Step != game.StepDeclareAttackers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			c := effectstest.EnergyPrompt(g, me.ID)
			if c == nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
			if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
				t.Fatalf("answer: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
	}
}

func TestGapB45OverclockedElectromancerPaysThreeEnergyForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mancer := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Overclocked Electromancer", TypeLine: "Creature — Lizard Wizard",
		OracleID: gapB45Electromancer, ManaCost: "{2}{R}", Power: 2, Toughness: 2})
	g.WithWriteLock(func() {
		_ = effects.GetEnergy{N: 3}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	gapB45AdvanceToAttackers(t, g, me, true)
	if p, _ := effectstest.PtOf(t, g, mancer); p != 3 {
		t.Errorf("power %d, want 3 with a +1/+1 counter", p)
	}
	if got := me.Counters["energy"]; got != 0 {
		t.Errorf("energy %d, want 0 after paying three", got)
	}
}

func TestGapB45DesertWarfareReturnsASacrificedDesertAtYourNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Desert Warfare", "Enchantment", gapB45DesertWarfare, false)
	desert := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Test Desert", TypeLine: "Land — Desert"})
	g.WithWriteLock(func() {
		if err := (effects.SacrificePermanent{Target: desert}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(desert) {
		t.Fatal("the Desert came back at once")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(desert) {
		t.Error("the sacrificed Desert did not return at the end step")
	}
}

func TestGapB45DesertWarfareMakesSandWarriorsWithFiveDeserts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Desert Warfare", "Enchantment", gapB45DesertWarfare, false)
	for i := 0; i < 4; i++ {
		effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Test Desert", TypeLine: "Land — Desert"})
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Sand Warrior")); got != 0 {
		t.Fatalf("%d Sand Warriors with four Deserts, want 0", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Desert Warfare", "Enchantment", gapB45DesertWarfare, false)
	for i := 0; i < 5; i++ {
		effectstest.M3cb1Push(g2, me2.ID, game.Card{Name: "Test Desert", TypeLine: "Land — Desert"})
	}
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g2)
	warriors := effectstest.S46Tokens(g2, me2.ID, "Sand Warrior")
	if len(warriors) != 5 {
		t.Fatalf("%d Sand Warriors with five Deserts, want 5", len(warriors))
	}
	if !effectstest.EffectiveAbilitiesContain(t, g2, warriors[0].InstanceID, "haste") {
		t.Error("the Sand Warrior lacks haste")
	}
}

func TestGapB45BenthicAnomalyCopiesOneChosenCreatureAtTheTotals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	a := effectstest.M3cb1Push(g, opps[0].ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 3, Toughness: 2})
	b := effectstest.M3cb1Push(g, opps[1].ID, game.Card{Name: "Elk", TypeLine: "Creature — Elk", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 5})
	effectstest.CastCatalogSpell(t, g, "Benthic Anomaly", "Creature — Eldrazi Serpent", gapB45BenthicAnomaly, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, pick := range []uuid.UUID{a, b} {
		var leg *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && pc.Kind == game.PendingChoiceTheirPermanents && pc.Chooser == me.ID && effectstest.HasID(pc.ChooseCards, pick) {
				leg = pc
			}
		}
		if leg == nil {
			t.Fatalf("no prompt offering %s", pick)
		}
		if err := g.ResolveTheirPermanents(leg.ID, me.ID, []uuid.UUID{pick}); err != nil {
			t.Fatalf("ResolveTheirPermanents: %v", err)
		}
	}
	answerPermanentsPick(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	tokens := effectstest.S46Tokens(g, me.ID, "Elk")
	if len(tokens) != 1 {
		t.Fatalf("%d Elk copies, want 1", len(tokens))
	}
	id := tokens[0].InstanceID
	if p, tg := effectstest.PtOf(t, g, id); p != 5 || tg != 7 {
		t.Errorf("the copy is %d/%d, want 5/7 (the totals)", p, tg)
	}
	c := effectstest.LayeredCard(t, g, id)
	if !c.IsColorless() {
		t.Error("the copy is not colorless")
	}
	if !c.HasSubtypeWord("Eldrazi") || c.HasSubtypeWord("Elk") || !c.IsCreature() {
		t.Errorf("the copy is %q, want a creature that is an Eldrazi only", c.TypeLine)
	}
	if c.ManaValue() != 2 {
		t.Errorf("mana value %d, want 2", c.ManaValue())
	}
}

func TestGapB45SilverquillLecturerOffersDemonstrateOnCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Silverquill Lecturer", "Creature — Kor Wizard", gapB45Lecturer, false)
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AwaitTriggerPrompt(t, g, me.ID)
	if c == nil {
		t.Fatal("no demonstrate prompt for a creature spell")
	}
	if err := g.ResolveTriggerPrompt(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
}

func TestGapB45SatyaCopiesAnAttackerAndSacrificesItUnlessEnergyIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.M3cb1Opponent(g)
	satya := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Satya, Aetherflux Genius", TypeLine: "Legendary Creature — Human Artificer",
		OracleID: gapB45Satya, ManaCost: "{1}{U}{R}{W}", Power: 3, Toughness: 5, Keywords: []string{"menace", "haste"}})
	effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Grizzly", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2})
	var grizzly uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Grizzly" {
				grizzly = c.InstanceID
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(satya, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, grizzly)
	effectstest.PassPriorityAroundTable(t, g)
	copies := effectstest.S46Tokens(g, me.ID, "Grizzly")
	if len(copies) != 1 {
		t.Fatalf("%d Grizzly copies, want 1", len(copies))
	}
	if got := me.Counters["energy"]; got != 2 {
		t.Errorf("energy %d, want 2", got)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EnergyPrompt(g, me.ID)
	if c == nil {
		t.Fatal("no energy payment owed at the end step")
	}
	// The copy has mana value 3 and the controller has 2 energy: declining
	// (or being short) sacrifices it.
	if err := g.ResolvePayUnless(c.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Grizzly")); got != 0 {
		t.Errorf("%d Grizzly copies after declining, want 0", got)
	}
}

func TestGapB45AuroraShifterPaysTwoEnergyToCopyAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shifter := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Aurora Shifter", TypeLine: "Creature — Shapeshifter",
		OracleID: gapB45AuroraShifter, ManaCost: "{1}{U}", Power: 1, Toughness: 3})
	ogre := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 4, Toughness: 4})
	g.WithWriteLock(func() {
		_ = effects.GetEnergy{N: 2}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	for i := 0; i < 40 && g.Turn.Step != game.StepDeclareAttackers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			if c := effectstest.EnergyPrompt(g, me.ID); c != nil {
				if err := g.ResolvePayUnless(c.ID, me.ID, true); err != nil {
					t.Fatal(err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				effectstest.AnswerPickTarget(t, g, ogre)
				effectstest.PassPriorityAroundTable(t, g)
				continue
			}
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if c := effectstest.LayeredCard(t, g, shifter); c.Name != "Ogre" {
		t.Fatalf("the Shifter is %q, want a copy of the Ogre", c.Name)
	}
	if p, tg := effectstest.PtOf(t, g, shifter); p != 4 || tg != 4 {
		t.Errorf("the copy is %d/%d, want 4/4", p, tg)
	}
	if got := me.Counters["energy"]; got != 0 {
		t.Errorf("energy %d, want 0 after paying two", got)
	}
}

// A token copy of a creature with an enters trigger fires it (the old
// "token copy skips an on-enter hook" caveat Satya waited on).
func TestGapB45SatyasTokenCopyFiresItsEntersTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.M3cb1Opponent(g)
	satya := effectstest.M3cb1Push(g, me.ID, game.Card{Name: "Satya, Aetherflux Genius", TypeLine: "Legendary Creature — Human Artificer",
		OracleID: gapB45Satya, ManaCost: "{1}{U}{R}{W}", Power: 3, Toughness: 5, Keywords: []string{"menace", "haste"}})
	drowner := effectstest.PushCatalogPermanent(g, me.ID, "Drowner of Hope", "Creature — Eldrazi", gapB45DrownerOfHope, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(satya, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AnswerPickTarget(t, g, drowner)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Eldrazi Scion")); got != 2 {
		t.Errorf("%d Eldrazi Scions, want 2 from the copy's enters trigger", got)
	}
}
