package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b11_cards_test.go — S82 bonus-sheet batch 11 (#5929), the
// Ravnica: Clue Edition creatures. Each test plays the catalog entry
// through the engine and asserts what its printed text says.

const (
	bnsProtesterOracle     = "09947592-1730-4c1c-bfa5-be53a68d6919"
	bnsStrandcatcherOracle = "4689206f-7907-46fb-82a7-2716687a58fc"
	bnsSeerOracle          = "d8d9c7e2-8ae7-4e1c-9859-bb3b85a48117"
	bnsElectromancerOracle = "2ad9f28c-94a9-4884-9e3b-42341adfaa5d"
	bnsGriffinOracle       = "4e1f4211-5865-492d-b850-0d36ac12e857"
	bnsCeratokOracle       = "e1f3f377-c7b1-474c-aed6-78b061d138c1"
	bnsVisionaryOracle     = "334c0913-0ce0-4934-979d-64dbc1ec4aa2"
	bnsUrchinOracle        = "4000e119-3372-431c-bc3a-0db0639311ce"
	bnsStrikerOracle       = "a017ab43-7d73-441a-8931-2a038306083f"
	bnsAvengerOracle       = "7ea71a36-8fa8-4ba3-9cb1-7fc6917c3ddd"
	bnsFiendOracle         = "77a5305a-2611-4a64-adb4-da131dcea233"
	bnsGorespawnOracle     = "197809b7-1ddb-4329-8d82-7651696538a8"
)

func TestBnsB11EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		bnsProtesterOracle: "Deputized Protester", bnsStrandcatcherOracle: "Dimir Strandcatcher",
		bnsSeerOracle: "Duskmantle Seer", bnsElectromancerOracle: "Ecstatic Electromancer",
		bnsGriffinOracle: "Enforcer Griffin", bnsCeratokOracle: "Enraged Ceratok",
		bnsVisionaryOracle: "Erratic Visionary", bnsUrchinOracle: "Fire Urchin",
		bnsStrikerOracle: "Firefist Striker", bnsAvengerOracle: "Firemane Avenger",
		bnsFiendOracle: "Footlight Fiend", bnsGorespawnOracle: "Frenzied Gorespawn",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (registered %v)", name, spec, ok)
		}
	}
}

func TestBnsB11DeputizedProtesterHasMenaceAndMelee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Deputized Protester", "Creature — Human Warrior", bnsProtesterOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
		t.Error("no menace")
	}
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tough != 2 {
		t.Errorf("after attacking one opponent: %d/%d, want 3/2", p, tough)
	}
}

func TestBnsB11EnforcerGriffinFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.BnsActive(g)
	id := effectstest.B12Push(g, me.ID, "Enforcer Griffin", "Creature — Griffin", bnsGriffinOracle, 3, 4)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
}

func TestBnsB11EnragedCeratokCantBeBlockedBySmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	rhino := effectstest.B12Push(g, me.ID, "Enraged Ceratok", "Creature — Rhino", bnsCeratokOracle, 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, rhino)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, rhino); err == nil {
		t.Error("a power 2 creature blocked the Ceratok")
	}
	if err := g.DeclareBlocker(big, rhino); err != nil {
		t.Errorf("a power 3 creature could not block it: %v", err)
	}
}

func TestBnsB11ErraticVisionaryLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Erratic Visionary", "Creature — Human Wizard", bnsVisionaryOracle, 1, 3)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant"})
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	// Not enough mana: refused.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after the draw")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) || me.Hand.Size() != hand-1 {
		t.Errorf("hand %d -> %d, fodder still there %v", hand, me.Hand.Size(), me.Hand.Contains(fodder))
	}
	c, _ := effectstest.BattlefieldCard(g, id)
	if !c.Tapped {
		t.Error("the Visionary did not tap")
	}
}

func TestBnsB11FireUrchinPumpsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Fire Urchin", "Creature — Elemental", bnsUrchinOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample")
	}
	effectstest.CastCatalogSpell(t, g, "Test Creature", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 1 {
		t.Errorf("creature spell: power %d, want 1", p)
	}
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("instant: power %d, want 2", p)
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, id); p != 3 {
		t.Errorf("sorcery: power %d, want 3", p)
	}
}

func TestBnsB11FirefistStrikerNeedsTwoOtherAttackers(t *testing.T) {
	for _, others := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.BnsActive(g)
		effectstest.AdvanceToMain(t, g)
		striker := effectstest.B12Push(g, me.ID, "Firefist Striker", "Creature — Human Soldier", bnsStrikerOracle, 2, 1)
		attackers := []uuid.UUID{striker}
		for i := 0; i < others; i++ {
			attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, attackers...)
		if others == 2 {
			effectstest.AnswerTriggerTargets(t, g, me.ID, blocker)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(blocker, striker)
		if others == 2 && err == nil {
			t.Error("battalion: the targeted creature blocked")
		}
		if others == 1 && err != nil {
			t.Errorf("no battalion: blocker refused: %v", err)
		}
	}
}

func TestBnsB11FiremaneAvengerBattalionDamagesAndGains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	av := effectstest.B12Push(g, me.ID, "Firemane Avenger", "Creature — Angel", bnsAvengerOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	life, oppLife := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, av, a, b)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != life+3 {
		t.Errorf("life: me %d->%d, opp %d->%d; want +3 and -3", life, me.Life, oppLife, opp.Life)
	}
}

func TestBnsB11FiremaneAvengerDoesNothingWithOneOtherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	av := effectstest.B12Push(g, me.ID, "Firemane Avenger", "Creature — Angel", bnsAvengerOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DeclareAttack(t, g, opp.ID, av, a)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("battalion triggered with one other attacker")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life changed to %d", me.Life)
	}
}

func TestBnsB11FootlightFiendDiesForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	fiend := effectstest.B12Push(g, me.ID, "Footlight Fiend", "Creature — Devil", bnsFiendOracle, 1, 1)
	before := opp.Life
	effectstest.KillCreature(t, g, me.ID, fiend)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt on death")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d -> %d, want -1", before, opp.Life)
	}
}

func TestBnsB11EcstaticElectromancerRummagesOnEntryAndAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn 1", TypeLine: "Instant"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Drawn 2", TypeLine: "Instant"})
	fodder := effectstest.HandCard(me, "Fodder", "Instant")
	id := effectstest.CastCatalogSpell(t, g, "Ecstatic Electromancer", "Creature — Goblin Wizard", bnsElectromancerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "prowess") {
		t.Error("no prowess")
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt on entry")
	}
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) {
		t.Error("fodder not discarded")
	}
	fodder2 := effectstest.HandCard(me, "Fodder 2", "Instant")
	attacker := effectstest.B12Push(g, me.ID, "Ecstatic Electromancer", "Creature — Goblin Wizard", bnsElectromancerOracle, 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt on attack")
	}
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, fodder2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder2) || me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d after the attack rummage", hand, me.Hand.Size())
	}
}

func TestBnsB11DimirStrandcatcherSurveilsPerOpponentAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Dimir Strandcatcher", "Creature — Faerie Rogue", bnsStrandcatcherOracle, 3, 3)
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	effectstest.DeclareAttack(t, g, opp.ID, id)
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// One opponent attacked: a surveil 1 prompt is open or has been answered.
	if len(g.PendingChoices) == 0 {
		t.Error("no surveil prompt after attacking one opponent")
	}
}

func TestBnsB11DimirStrandcatcherDrawsAfterThreeCardsHitTheGraveyard(t *testing.T) {
	for _, n := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, _ := effectstest.BnsActive(g)
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Dimir Strandcatcher", "Creature — Faerie Rogue", bnsStrandcatcherOracle, 3, 3)
		for i := 0; i < n; i++ {
			id := effectstest.HandCard(me, "Discarded", "Instant")
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
				game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
				t.Fatal(err)
			}
		}
		hand := me.Hand.Size()
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := hand
		if n >= 3 {
			want = hand + 1
		}
		if me.Hand.Size() != want {
			t.Errorf("%d cards to the graveyard: hand %d -> %d, want %d", n, hand, me.Hand.Size(), want)
		}
	}
}

func TestBnsB11DimirStrandcatcherIgnoresCardsFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Dimir Strandcatcher", "Creature — Faerie Rogue", bnsStrandcatcherOracle, 3, 3)
	for i := 0; i < 3; i++ {
		id := effectstest.B12Creature(g, me.ID, "Doomed", "Creature — Bear", 1, 1)
		effectstest.KillCreature(t, g, me.ID, id)
	}
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: creatures dying are not cards from elsewhere", hand, me.Hand.Size())
	}
}

func TestBnsB11DuskmantleSeerEachPlayerRevealsLosesLifeAndTakesTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Duskmantle Seer", bnsSeerOracle, "Creature — Vampire Wizard")
	tops := map[uuid.UUID]uuid.UUID{}
	lifes := map[uuid.UUID]int{}
	for i, p := range g.Seats {
		cost := []string{"{3}", "", "{1}{R}", "{2}"}[i]
		top := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Artifact", ManaCost: cost, Owner: p.ID, Controller: p.ID})
		tops[p.ID] = top
		lifes[p.ID] = p.Life
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		loss := []int{3, 0, 2, 2}[i]
		if p.Life != lifes[p.ID]-loss {
			t.Errorf("seat %d life %d -> %d, want -%d", i, lifes[p.ID], p.Life, loss)
		}
		if !p.Hand.Contains(tops[p.ID]) {
			t.Errorf("seat %d did not take its revealed card", i)
		}
	}
}

func TestBnsB11FrenziedGorespawnGoadsOneCreaturePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.BnsActive(g)
	var picks []uuid.UUID
	for i, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		picks = append(picks, effectstest.TdCreature(g, p.ID, "Foe", 2+i, 2))
	}
	mine := effectstest.TdCreature(g, me.ID, "Mine", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Frenzied Gorespawn", "Creature — Horror", bnsGorespawnOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range picks {
		if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
			effectstest.PickCard(t, g, me.ID, id)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range picks {
		c, _ := g.LookupCardForEffect(id)
		if !c.IsGoadedBy(me.ID) {
			t.Errorf("%s was not goaded", c.Name)
		}
	}
	if c, _ := g.LookupCardForEffect(mine); c.IsGoaded() {
		t.Error("my own creature was goaded")
	}
}

func TestBnsB11FrenziedGorespawnGivesMenaceToCreaturesAttackingAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Frenzied Gorespawn", "Creature — Horror", bnsGorespawnOracle, 4, 4)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Error("an attacker did not gain menace")
		}
	}
}
