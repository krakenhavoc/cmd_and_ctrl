package testscr

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_mp2_cards_test.go — Aetherdrift (DFT), cards from M to P, second
// batch (#4464). Each test plays the real catalog entry and asserts what
// its text says. The fixtures are dft_ab_cards_test.go's.

const (
	dftMP2MindspringOracle   = "35142b01-4b8d-4335-a943-a3172fa04573"
	dftMP2GearhulkOracle     = "aff0740f-51d5-41c7-ac13-97c298cfd743"
	dftMP2OviyaOracle        = "b01649b8-b0b5-4cf6-aea7-291f3407859e"
	dftMP2PitAutomatonOracle = "49bfecc3-6258-454e-a8fd-fc49098a0898"
	dftMP2PossessionOracle   = "ab7c0256-eea9-4e0d-932d-c25460eed836"
	dftMP2PushTheLimitOracle = "62e54c6d-d775-4597-bd4d-dc38460b33ec"
)

func TestDftMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftMP2MindspringOracle:   "Mindspring Merfolk",
		dftMP2GearhulkOracle:     "Oildeep Gearhulk",
		dftMP2OviyaOracle:        "Oviya, Automech Artisan",
		dftMP2PitAutomatonOracle: "Pit Automaton",
		dftMP2PossessionOracle:   "Possession Engine",
		dftMP2PushTheLimitOracle: "Push the Limit",
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
	// Mimeoplasm, Revered One is registered by batch 14 (dft_b13_cards_test.go).
}

// --- fixtures ------------------------------------------------------------------

// dftMP2Hand puts a card with printed P/T into p's hand.
func dftMP2Hand(p *game.Player, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

// --- Mindspring Merfolk -----------------------------------------------------------

func TestDftMP2MindspringMerfolkDrawsXAndCountersEachMerfolkOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	merfolk := effectstest.DftABPush(g, me.ID, "Mindspring Merfolk", "Creature — Merfolk Wizard", dftMP2MindspringOracle, 1, 1)
	other := effectstest.DftABPush(g, me.ID, "Another Merfolk", "Creature — Merfolk", "", 2, 2)
	bear := effectstest.DftABPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DftABPush(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Merfolk", "Creature — Merfolk", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{U}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, merfolk, 0, game.ActivateAbilityParams{XValue: 2, Strict: true}); err != nil {
		t.Fatalf("exhaust X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want X = 2 cards drawn", hand, got)
	}
	for id, want := range map[uuid.UUID]int{merfolk: 1, other: 1, bear: 0, theirs: 0} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != want {
			c, _ := effectstest.BattlefieldCard(g, id)
			t.Errorf("%s: %d +1/+1 counters, want %d", c.Name, got, want)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, merfolk); !c.Tapped {
		t.Error("the {T} in the cost did not tap the Merfolk")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == merfolk {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	effectstest.FloatMana(t, g, me, "{C}{U}{U}")
	err := g.ActivateCatalogAbility(me.ID, merfolk, 0, game.ActivateAbilityParams{XValue: 1})
	if !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second exhaust: err = %v, want ErrAbilityExhausted", err)
	}
}

// --- Oildeep Gearhulk -------------------------------------------------------------

func TestDftMP2OildeepGearhulkDiscardsTheChosenCardThenTheyDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	land := dftMP2Hand(opp, "Their Mountain", "Basic Land — Mountain", 0, 0)
	bolt := dftMP2Hand(opp, "Their Bolt", "Instant", 0, 0)
	hulk := effectstest.DftABCast(t, g, me, game.Card{Name: "Oildeep Gearhulk", TypeLine: "Artifact Creature — Construct",
		OracleID: dftMP2GearhulkOracle, Power: 4, Toughness: 4})
	if !effectstest.HasEffectiveKeyword(t, g, hulk, "lifelink") {
		t.Error("lifelink is printed")
	}
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if pick.Chooser != me.ID || !effectstest.HasID(pick.DiscardOptions, land) || !effectstest.HasID(pick.DiscardOptions, bolt) {
		t.Fatalf("pick chooser %v options %v, want me choosing among every card", pick.Chooser, pick.DiscardOptions)
	}
	if !pick.PickOptional {
		t.Error("the choice is not optional")
	}
	hand := opp.Hand.Size()
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B12ZoneOf(g, land); got != game.ZoneGraveyard {
		t.Errorf("the chosen card is in %q, want their graveyard", got)
	}
	if effectstest.RvDiscards(g, land) != 1 {
		t.Error("the card was not discarded")
	}
	if opp.Hand.Size() != hand {
		t.Errorf("their hand %d, want %d (one discarded, one drawn)", opp.Hand.Size(), hand)
	}
}

func TestDftMP2OildeepGearhulkChoosingNothingDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	bolt := dftMP2Hand(opp, "Their Bolt", "Instant", 0, 0)
	effectstest.DftABCast(t, g, me, game.Card{Name: "Oildeep Gearhulk", TypeLine: "Artifact Creature — Construct",
		OracleID: dftMP2GearhulkOracle, Power: 4, Toughness: 4})
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	hand := opp.Hand.Size()
	if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bolt) || opp.Hand.Size() != hand {
		t.Errorf("their hand changed (%d, want %d)", opp.Hand.Size(), hand)
	}
}

// --- Oviya, Automech Artisan --------------------------------------------------------

// Trample goes to every creature attacking one of Oviya's controller's
// opponents, whoever controls it, and not to one attacking her controller.
func TestDftMP2OviyaGivesTrampleToCreaturesAttackingHerOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := g.Turn.ActiveSeat
	attacker, oviyaSide, other := g.Seats[a], g.Seats[(a+1)%4], g.Seats[(a+2)%4]
	effectstest.DftABPush(g, oviyaSide.ID, "Oviya, Automech Artisan", "Legendary Creature — Human Artificer", dftMP2OviyaOracle, 1, 2)
	atOther := effectstest.DftABPush(g, attacker.ID, "Bear A", "Creature — Bear", "", 2, 2)
	atOviya := effectstest.DftABPush(g, attacker.ID, "Bear B", "Creature — Bear", "", 2, 2)
	idle := effectstest.DftABPush(g, attacker.ID, "Bear C", "Creature — Bear", "", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, atOther, "trample") {
		t.Fatal("a creature that is not attacking has trample")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atOther, other.ID); err != nil {
		t.Fatalf("attack %v: %v", other.ID, err)
	}
	if err := g.DeclareAttacker(atOviya, oviyaSide.ID); err != nil {
		t.Fatalf("attack %v: %v", oviyaSide.ID, err)
	}
	effectstest.LockInAttacks(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, atOther, "trample") {
		t.Error("a creature attacking one of Oviya's controller's opponents lacks trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, atOviya, "trample") {
		t.Error("a creature attacking Oviya's controller has trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, idle, "trample") {
		t.Error("a creature that did not attack has trample")
	}
}

func TestDftMP2OviyaPutsACreatureOrVehicleAndCountersOnlyAnArtifact(t *testing.T) {
	for _, tc := range []struct {
		name     string
		counters int
		pick     func(vehicle, artifactCreature, bear uuid.UUID) uuid.UUID
	}{
		{"a Vehicle", 2, func(v, _, _ uuid.UUID) uuid.UUID { return v }},
		{"an artifact creature", 2, func(_, a, _ uuid.UUID) uuid.UUID { return a }},
		{"a nonartifact creature", 0, func(_, _, b uuid.UUID) uuid.UUID { return b }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			oviya := effectstest.DftABPush(g, me.ID, "Oviya, Automech Artisan", "Legendary Creature — Human Artificer", dftMP2OviyaOracle, 1, 2)
			effectstest.AdvanceToMain(t, g)
			vehicle := dftMP2Hand(me, "A Vehicle", effectstest.DftABVehicle, 3, 3)
			artifactCreature := dftMP2Hand(me, "A Construct", "Artifact Creature — Construct", 1, 1)
			bear := dftMP2Hand(me, "A Bear", "Creature — Bear", 2, 2)
			rock := dftMP2Hand(me, "A Rock", "Artifact", 0, 0)
			effectstest.FloatMana(t, g, me, "{G}")
			if err := g.ActivateCatalogAbility(me.ID, oviya, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("no put-from-hand prompt")
			}
			for _, id := range []uuid.UUID{vehicle, artifactCreature, bear} {
				if !slices.Contains(pick.ChooseCards, id) {
					t.Errorf("%v was not offered", id)
				}
			}
			if slices.Contains(pick.ChooseCards, rock) {
				t.Error("a noncreature, non-Vehicle artifact was offered")
			}
			chosen := tc.pick(vehicle, artifactCreature, bear)
			effectstest.AnswerChooseCards(t, g, me.ID, chosen)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(chosen) {
				t.Fatal("the chosen card did not enter")
			}
			if got := effectstest.CountersOn(g, chosen, game.CounterPlusOne); got != tc.counters {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.counters)
			}
		})
	}
}

func TestDftMP2OviyaMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oviya := effectstest.DftABPush(g, me.ID, "Oviya, Automech Artisan", "Legendary Creature — Human Artificer", dftMP2OviyaOracle, 1, 2)
	effectstest.AdvanceToMain(t, g)
	bear := dftMP2Hand(me, "A Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, oviya, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("declining still put the creature onto the battlefield")
	}
}

// --- Pit Automaton --------------------------------------------------------------

func TestDftMP2PitAutomatonManaPaysOnlyForAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pit := effectstest.DftABPush(g, me.ID, "Pit Automaton", "Artifact Creature — Construct", dftMP2PitAutomatonOracle, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, pit, "defender") {
		t.Error("defender is printed")
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, pit, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool %d, want {C}{C}", got)
	}
	cost, _ := game.ParseCost("{2}")
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}"}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana refused an activated ability")
	}
}

// The next exhaust ability you activate is copied, X and all, and the
// trigger is then used up.
func TestDftMP2PitAutomatonCopiesTheNextExhaustAbilityOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pit := effectstest.DftABPush(g, me.ID, "Pit Automaton", "Artifact Creature — Construct", dftMP2PitAutomatonOracle, 0, 4)
	first := effectstest.DftABPush(g, me.ID, "Mindspring Merfolk", "Creature — Merfolk Wizard", dftMP2MindspringOracle, 1, 1)
	second := effectstest.DftABPush(g, me.ID, "Mindspring Merfolk", "Creature — Merfolk Wizard", dftMP2MindspringOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, pit, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Automaton: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.FloatMana(t, g, me, "{C}{U}{U}")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{XValue: 1, Strict: true}); err != nil {
		t.Fatalf("first exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want 2: the ability and its copy each draw X = 1", hand, got)
	}
	if got := effectstest.CountersOn(g, first, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters on the Merfolk, want 2 (ability and copy)", got)
	}

	effectstest.FloatMana(t, g, me, "{C}{U}{U}")
	hand = me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, second, 0, game.ActivateAbilityParams{XValue: 1, Strict: true}); err != nil {
		t.Fatalf("second exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want 1: the trigger was used up", hand, got)
	}
}

// --- Possession Engine ----------------------------------------------------------

func TestDftMP2PossessionEngineStealsAndBenchesWhileYouControlIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.CtrlPushCreature(g, opp.ID, "Bear")
	mine := effectstest.CtrlPushCreature(g, me.ID, "My Bear")
	engine := effectstest.DftABCast(t, g, me, game.Card{Name: "Possession Engine", TypeLine: effectstest.DftABVehicle,
		OracleID: dftMP2PossessionOracle, Power: 5, Toughness: 5})
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && slices.Contains(p.PickTargetCards, mine) {
		t.Error("my own creature was offered as the target")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %v after the Engine entered, want %v", got, me.ID)
	}
	if r := effectstest.AuraRestrictions(t, g, victim); r&game.CantAttack == 0 || r&game.CantBlock == 0 {
		t.Errorf("restrictions %v, want can't attack and can't block", r)
	}
	if r := effectstest.AuraRestrictions(t, g, mine); r&(game.CantAttack|game.CantBlock) != 0 {
		t.Errorf("my other creature is restricted: %v", r)
	}
	// Losing control of the Engine ends both effects.
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.New(), engine, opp.ID, game.IndefiniteDuration(), "test — take the Engine")
	})
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Errorf("the Engine changed hands and the creature stayed: controller %v, want %v", got, opp.ID)
	}
	if r := effectstest.AuraRestrictions(t, g, victim); r&(game.CantAttack|game.CantBlock) != 0 {
		t.Errorf("the restriction outlived the Engine's control: %v", r)
	}
}

func TestDftMP2PossessionEngineCrewsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	engine := effectstest.DftABPush(g, me.ID, "Possession Engine", effectstest.DftABVehicle, dftMP2PossessionOracle, 5, 5)
	two := effectstest.DftABPush(g, me.ID, "Two", "Creature — Bear", "", 2, 2)
	one := effectstest.DftABPush(g, me.ID, "One", "Creature — Bear", "", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, engine, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{two}}); err == nil {
		t.Fatal("crewed with total power 2; crew 3 wants 3")
	}
	effectstest.DftABCrew(t, g, me, engine, 0, two, one)
}

// --- Push the Limit ---------------------------------------------------------------

func TestDftMP2PushTheLimitReturnsMountsAndVehiclesAnimatesAndHastes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	vehicle := effectstest.DftABInGraveyard(me, "A Dead Vehicle", effectstest.DftABVehicle)
	mount := effectstest.DftABInGraveyard(me, "A Dead Mount", "Creature — Horse Mount")
	bear := effectstest.DftABInGraveyard(me, "A Dead Bear", "Creature — Bear")
	already := effectstest.DftABPush(g, me.ID, "A Parked Vehicle", effectstest.DftABVehicle, "", 4, 4)
	mine := effectstest.DftABPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DftABPush(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Vehicle", effectstest.DftABVehicle, "", 3, 3)

	effectstest.CastCatalogSpell(t, g, "Push the Limit", "Sorcery", dftMP2PushTheLimitOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{vehicle, mount} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("%v did not return", id)
		}
	}
	if got := effectstest.B12ZoneOf(g, bear); got != game.ZoneGraveyard {
		t.Errorf("a non-Mount creature card left the graveyard (now %q)", got)
	}
	for _, id := range []uuid.UUID{vehicle, already} {
		if c, _ := effectstest.BattlefieldCard(g, id); !c.IsCreature() || !c.IsArtifact() {
			t.Errorf("%s is not an artifact creature: %q", c.Name, c.TypeLine)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.IsCreature() {
		t.Error("an opponent's Vehicle was animated")
	}
	for _, id := range []uuid.UUID{vehicle, mount, already, mine} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			c, _ := effectstest.BattlefieldCard(g, id)
			t.Errorf("%s lacks haste", c.Name)
		}
	}

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{vehicle, mount} {
		if got := effectstest.B12ZoneOf(g, id); got != game.ZoneGraveyard {
			t.Errorf("a returned card is in %q at the end step, want sacrificed", got)
		}
	}
	if !g.Battlefield.Contains(already) || !g.Battlefield.Contains(mine) {
		t.Error("a permanent that was not returned was sacrificed")
	}
}

func TestDftMP2PushTheLimitWithAnEmptyGraveyardStillHastes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.Graveyard.Cards = nil })
	mine := effectstest.DftABPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Push the Limit", "Sorcery", dftMP2PushTheLimitOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "haste") {
		t.Error("creatures you control did not gain haste")
	}
}
