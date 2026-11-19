package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB23GuardianIdol   = "829d2e11-192f-4093-99f8-e13767b321a2"
	bnsB23Puzzleknot     = "c19d2f54-2c3a-4cbe-8328-061d4b51f574"
	bnsB23Memento        = "887de1fc-8fbd-472a-bcf2-4f36ea9376d1"
	bnsB23Solvent        = "f34ba913-cf9a-48fe-b5fe-68bf504693b2"
	bnsB23Whirlermaker   = "bdf28770-0962-4317-822f-c187bc6706d1"
	bnsB23WitchingWell   = "baadfc17-42fd-4771-aabe-027ef7b3bf69"
	bnsB23AncestralBlade = "ab28a801-6ab1-4db2-8ae2-c24bb4c10ebb"
	bnsB23Kitesail       = "68fb8f21-4403-4829-8826-f2418664197e"
	bnsB23GorgonFlail    = "91dab3e2-039e-4db9-94e3-1146bced18ed"
	bnsB23LeoninScimitar = "cde26d69-f3e7-4dd0-a53b-cd0ec812d717"
	bnsB23RelicAxe       = "d851312f-8b4e-4d8e-bfee-c470853be930"
)

func TestBnsB23Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB23GuardianIdol: "Guardian Idol", bnsB23Puzzleknot: "Metalspinner's Puzzleknot",
		bnsB23Memento: "Starnheim Memento", bnsB23Solvent: "Universal Solvent",
		bnsB23Whirlermaker: "Whirlermaker", bnsB23WitchingWell: "Witching Well",
		bnsB23AncestralBlade: "Ancestral Blade", bnsB23Kitesail: "Cliffhaven Kitesail",
		bnsB23GorgonFlail: "Gorgon Flail", bnsB23LeoninScimitar: "Leonin Scimitar",
		bnsB23RelicAxe: "Relic Axe",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec.Completeness)
		}
	}
}

func bnsB23Pump(t *testing.T, oracle, name string, wantP, wantT int, keyword string) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	eq := pushCatalogPermanent(g, me.ID, name, "Artifact — Equipment", oracle, false)
	if p, tt := effectivePTOf(t, g, bear); p != 2 || tt != 2 {
		t.Fatalf("unequipped bear is %d/%d", p, tt)
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, eq, bear)
	if p, tt := effectivePTOf(t, g, bear); p != 2+wantP || tt != 2+wantT {
		t.Errorf("%s: equipped bear is %d/%d, want %d/%d", name, p, tt, 2+wantP, 2+wantT)
	}
	if keyword != "" && !hasKeywordOf(g, bear, keyword) {
		t.Errorf("%s: no %s on the equipped creature", name, keyword)
	}
}

func TestBnsB23GorgonFlail(t *testing.T) {
	bnsB23Pump(t, bnsB23GorgonFlail, "Gorgon Flail", 1, 1, "deathtouch")
}
func TestBnsB23LeoninScimitar(t *testing.T) {
	bnsB23Pump(t, bnsB23LeoninScimitar, "Leonin Scimitar", 1, 1, "")
}

func TestBnsB23AncestralBladeMakesASoldierAndEquipsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	blade := castCatalogSpell(t, g, "Ancestral Blade", "Artifact — Equipment", bnsB23AncestralBlade, nil)
	passPriorityAroundTable(t, g)
	soldier := avB1Named(g, me.ID, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier token")
	}
	if c := findBattlefieldCardForTest(g, blade); c == nil || c.AttachedTo.ID != soldier {
		t.Fatal("the Blade is not attached to the Soldier")
	}
	if p, tt := effectivePTOf(t, g, soldier); p != 2 || tt != 2 {
		t.Errorf("equipped Soldier is %d/%d, want 2/2", p, tt)
	}
}

func bnsB23CastAttach(t *testing.T, name, oracle string) (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	foe := pushVanillaCreature(g, opp.ID, "Foe", 2, 2)
	eq := castCatalogSpell(t, g, name, "Artifact — Equipment", oracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if hasID(pick.PickTargetCards, foe) {
		t.Error("an opponent's creature is not a legal target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, eq); c == nil || c.AttachedTo.ID != bear {
		t.Fatalf("%s is not attached to the target", name)
	}
	return g, me.ID, eq, bear
}

func TestBnsB23KitesailAttachesAndGrantsFlying(t *testing.T) {
	g, _, _, bear := bnsB23CastAttach(t, "Cliffhaven Kitesail", bnsB23Kitesail)
	if !hasKeywordOf(g, bear, "flying") {
		t.Error("equipped creature lacks flying")
	}
	if p, tt := effectivePTOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("the Kitesail changes size: %d/%d", p, tt)
	}
}

func TestBnsB23RelicAxePumpsAndFavoursWarriors(t *testing.T) {
	g, _, _, bear := bnsB23CastAttach(t, "Relic Axe", bnsB23RelicAxe)
	if p, tt := effectivePTOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("non-Warrior: %d/%d, want 3/3", p, tt)
	}

	g = newCatalogGame(t)
	me := g.Seats[0]
	warrior := pushDiesCreatureForTest(g, me.ID, "Warrior", "", "Creature — Human Warrior", 2, 2)
	axe := pushCatalogPermanent(g, me.ID, "Relic Axe", "Artifact — Equipment", bnsB23RelicAxe, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	equipTo(t, g, me.ID, axe, warrior)
	if p, tt := effectivePTOf(t, g, warrior); p != 4 || tt != 3 {
		t.Errorf("Warrior: %d/%d, want 4/3 (+2/+1, not +3/+2)", p, tt)
	}
}

func TestBnsB23WhirlermakerMakesAThopter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := pushCatalogPermanent(g, me.ID, "Whirlermaker", "Artifact", bnsB23Whirlermaker, false)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	th := avB1Named(g, me.ID, "Thopter")
	if th == uuid.Nil || !hasKeywordOf(g, th, "flying") {
		t.Fatal("no flying Thopter")
	}
	if tapped, _ := battlefieldCardTapped(g, w); !tapped {
		t.Error("{T} is part of the cost")
	}
}

func TestBnsB23UniversalSolventDestroysAnyPermanentAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	s := pushCatalogPermanent(g, me.ID, "Universal Solvent", "Artifact", bnsB23Solvent, false)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("activated with only six mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(s) {
		t.Error("the Solvent is still there: sacrifice is a cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the targeted land survived")
	}
}

func TestBnsB23WitchingWellScriesThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		}
	})
	w := castCatalogSpell(t, g, "Witching Well", "Artifact", bnsB23WitchingWell, nil)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Fatal("no scry prompt")
	}
	answerScryKeepAll(t, g, me.ID)
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if g.Battlefield.Contains(w) {
		t.Error("the Well is still there")
	}
}

func TestBnsB23PuzzleknotDrawsAndLosesLifeOnEntryAndSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		}
	})
	life, hand := me.Life, me.Hand.Size()
	k := castCatalogSpell(t, g, "Metalspinner's Puzzleknot", "Artifact", bnsB23Puzzleknot, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 { // the cast card left the hand, one drawn
		t.Errorf("entry: hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+1, me.Life, life-1)
	}
	floatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("sacrifice: hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+2, me.Life, life-2)
	}
	if g.Battlefield.Contains(k) {
		t.Error("the Puzzleknot is still there")
	}
}

func TestBnsB23StarnheimMementoManaAndSorceryPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	m := pushCatalogPermanent(g, me.ID, "Starnheim Memento", "Artifact", bnsB23Memento, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePTOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, tt)
	}
	if !hasKeywordOf(g, bear, "flying") {
		t.Error("no flying")
	}
	endTurn(t, g)
	if p, _ := effectivePTOf(t, g, bear); p != 2 || hasKeywordOf(g, bear, "flying") {
		t.Error("the pump lasts past end of turn")
	}
}

func TestBnsB23StarnheimMementoRefusesOutsideSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	m := pushCatalogPermanent(g, me.ID, "Starnheim Memento", "Artifact", bnsB23Memento, false)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("activated during combat")
	}
}

func TestBnsB23GuardianIdolEntersTappedAndAnimates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := castCatalogSpell(t, g, "Guardian Idol", "Artifact", bnsB23GuardianIdol, nil)
	passPriorityAroundTable(t, g)
	if tapped, ok := battlefieldCardTapped(g, idol); !ok || !tapped {
		t.Fatalf("the Idol must enter tapped (on battlefield %v, tapped %v)", ok, tapped)
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, idol)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Golem") || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("not a 2/2 Golem artifact creature: %+v", c.Effective())
	}
	endTurn(t, g)
	if c := layeredCard(t, g, idol); c.IsCreature() {
		t.Error("still a creature next turn")
	}
}
