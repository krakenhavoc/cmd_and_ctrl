package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (registered %v)", name, spec, ok)
		}
	}
}

func bnsActive(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestBnsB11DeputizedProtesterHasMenaceAndMelee(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Deputized Protester", "Creature — Human Warrior", bnsProtesterOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, id, "menace") {
		t.Error("no menace")
	}
	declareAttack(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tough != 2 {
		t.Errorf("after attacking one opponent: %d/%d, want 3/2", p, tough)
	}
}

func TestBnsB11EnforcerGriffinFlies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	id := b12Push(g, me.ID, "Enforcer Griffin", "Creature — Griffin", bnsGriffinOracle, 3, 4)
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
}

func TestBnsB11EnragedCeratokCantBeBlockedBySmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	rhino := b12Push(g, me.ID, "Enraged Ceratok", "Creature — Rhino", bnsCeratokOracle, 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	declareAttack(t, g, opp.ID, rhino)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, rhino); err == nil {
		t.Error("a power 2 creature blocked the Ceratok")
	}
	if err := g.DeclareBlocker(big, rhino); err != nil {
		t.Errorf("a power 3 creature could not block it: %v", err)
	}
}

func TestBnsB11ErraticVisionaryLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Erratic Visionary", "Creature — Human Wizard", bnsVisionaryOracle, 1, 3)
	pushLibraryCardForTest(me, game.Card{Name: "Drawn", TypeLine: "Instant"})
	fodder := handCard(me, "Fodder", "Instant")
	// Not enough mana: refused.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty pool")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after the draw")
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) || me.Hand.Size() != hand-1 {
		t.Errorf("hand %d -> %d, fodder still there %v", hand, me.Hand.Size(), me.Hand.Contains(fodder))
	}
	c, _ := battlefieldCard(g, id)
	if !c.Tapped {
		t.Error("the Visionary did not tap")
	}
}

func TestBnsB11FireUrchinPumpsOnInstantsAndSorceriesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Fire Urchin", "Creature — Elemental", bnsUrchinOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample")
	}
	castCatalogSpell(t, g, "Test Creature", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, id); p != 1 {
		t.Errorf("creature spell: power %d, want 1", p)
	}
	castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, id); p != 2 {
		t.Errorf("instant: power %d, want 2", p)
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, id); p != 3 {
		t.Errorf("sorcery: power %d, want 3", p)
	}
}

func TestBnsB11FirefistStrikerNeedsTwoOtherAttackers(t *testing.T) {
	for _, others := range []int{1, 2} {
		g := newCatalogGame(t)
		me, opp := bnsActive(g)
		advanceToMain(t, g)
		striker := b12Push(g, me.ID, "Firefist Striker", "Creature — Human Soldier", bnsStrikerOracle, 2, 1)
		attackers := []uuid.UUID{striker}
		for i := 0; i < others; i++ {
			attackers = append(attackers, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
		declareAttack(t, g, opp.ID, attackers...)
		if others == 2 {
			answerTriggerTargets(t, g, me.ID, blocker)
		}
		passPriorityAroundTable(t, g)
		advanceTo(t, g, game.StepDeclareBlockers)
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
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	av := b12Push(g, me.ID, "Firemane Avenger", "Creature — Angel", bnsAvengerOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	life, oppLife := me.Life, opp.Life
	declareAttack(t, g, opp.ID, av, a, b)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 || me.Life != life+3 {
		t.Errorf("life: me %d->%d, opp %d->%d; want +3 and -3", life, me.Life, oppLife, opp.Life)
	}
}

func TestBnsB11FiremaneAvengerDoesNothingWithOneOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	av := b12Push(g, me.ID, "Firemane Avenger", "Creature — Angel", bnsAvengerOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	life := me.Life
	declareAttack(t, g, opp.ID, av, a)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("battalion triggered with one other attacker")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life changed to %d", me.Life)
	}
}

func TestBnsB11FootlightFiendDiesForOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	fiend := b12Push(g, me.ID, "Footlight Fiend", "Creature — Devil", bnsFiendOracle, 1, 1)
	before := opp.Life
	killCreature(t, g, me.ID, fiend)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt on death")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("life %d -> %d, want -1", before, opp.Life)
	}
}

func TestBnsB11EcstaticElectromancerRummagesOnEntryAndAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	pushLibraryCardForTest(me, game.Card{Name: "Drawn 1", TypeLine: "Instant"})
	pushLibraryCardForTest(me, game.Card{Name: "Drawn 2", TypeLine: "Instant"})
	fodder := handCard(me, "Fodder", "Instant")
	id := castCatalogSpell(t, g, "Ecstatic Electromancer", "Creature — Goblin Wizard", bnsElectromancerOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, id, "prowess") {
		t.Error("no prowess")
	}
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt on entry")
	}
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder) {
		t.Error("fodder not discarded")
	}
	fodder2 := handCard(me, "Fodder 2", "Instant")
	attacker := b12Push(g, me.ID, "Ecstatic Electromancer", "Creature — Goblin Wizard", bnsElectromancerOracle, 2, 3)
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt on attack")
	}
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, fodder2)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(fodder2) || me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d after the attack rummage", hand, me.Hand.Size())
	}
}

func TestBnsB11DimirStrandcatcherSurveilsPerOpponentAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Dimir Strandcatcher", "Creature — Faerie Rogue", bnsStrandcatcherOracle, 3, 3)
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	declareAttack(t, g, opp.ID, id)
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("no flying")
	}
	passPriorityAroundTable(t, g)
	// One opponent attacked: a surveil 1 prompt is open or has been answered.
	if len(g.PendingChoices) == 0 {
		t.Error("no surveil prompt after attacking one opponent")
	}
}

func TestBnsB11DimirStrandcatcherDrawsAfterThreeCardsHitTheGraveyard(t *testing.T) {
	for _, n := range []int{2, 3} {
		g := newCatalogGame(t)
		me, _ := bnsActive(g)
		advanceToMain(t, g)
		b12Push(g, me.ID, "Dimir Strandcatcher", "Creature — Faerie Rogue", bnsStrandcatcherOracle, 3, 3)
		for i := 0; i < n; i++ {
			id := handCard(me, "Discarded", "Instant")
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID},
				game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
				t.Fatal(err)
			}
		}
		hand := me.Hand.Size()
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Dimir Strandcatcher", "Creature — Faerie Rogue", bnsStrandcatcherOracle, 3, 3)
	for i := 0; i < 3; i++ {
		id := b12Creature(g, me.ID, "Doomed", "Creature — Bear", 1, 1)
		killCreature(t, g, me.ID, id)
	}
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d: creatures dying are not cards from elsewhere", hand, me.Hand.Size())
	}
}

func TestBnsB11DuskmantleSeerEachPlayerRevealsLosesLifeAndTakesTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Duskmantle Seer", bnsSeerOracle, "Creature — Vampire Wizard")
	tops := map[uuid.UUID]uuid.UUID{}
	lifes := map[uuid.UUID]int{}
	for i, p := range g.Seats {
		cost := []string{"{3}", "", "{1}{R}", "{2}"}[i]
		top := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Artifact", ManaCost: cost, Owner: p.ID, Controller: p.ID})
		tops[p.ID] = top
		lifes[p.ID] = p.Life
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, _ := bnsActive(g)
	var picks []uuid.UUID
	for i, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		picks = append(picks, tdCreature(g, p.ID, "Foe", 2+i, 2))
	}
	mine := tdCreature(g, me.ID, "Mine", 1, 1)
	castCatalogSpell(t, g, "Frenzied Gorespawn", "Creature — Horror", bnsGorespawnOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range picks {
		if pick := latestPickTarget(g, me.ID); pick != nil {
			pickCard(t, g, me.ID, id)
		}
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := bnsActive(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Frenzied Gorespawn", "Creature — Horror", bnsGorespawnOracle, 4, 4)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !hasEffectiveKeyword(t, g, id, "menace") {
			t.Error("an attacker did not gain menace")
		}
	}
}
