package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestBnsB23Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.BnsB23GuardianIdol: "Guardian Idol", effectstest.BnsB23Puzzleknot: "Metalspinner's Puzzleknot",
		effectstest.BnsB23Memento: "Starnheim Memento", effectstest.BnsB23Solvent: "Universal Solvent",
		effectstest.BnsB23Whirlermaker: "Whirlermaker", effectstest.BnsB23WitchingWell: "Witching Well",
		effectstest.BnsB23AncestralBlade: "Ancestral Blade", effectstest.BnsB23Kitesail: "Cliffhaven Kitesail",
		effectstest.BnsB23GorgonFlail: "Gorgon Flail", effectstest.BnsB23LeoninScimitar: "Leonin Scimitar",
		effectstest.BnsB23RelicAxe: "Relic Axe",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec.Completeness)
		}
	}
}

func TestBnsB23GorgonFlail(t *testing.T) {
	effectstest.BnsB23Pump(t, effectstest.BnsB23GorgonFlail, "Gorgon Flail", 1, 1, "deathtouch")
}
func TestBnsB23LeoninScimitar(t *testing.T) {
	effectstest.BnsB23Pump(t, effectstest.BnsB23LeoninScimitar, "Leonin Scimitar", 1, 1, "")
}

func TestBnsB23AncestralBladeMakesASoldierAndEquipsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	blade := effectstest.CastCatalogSpell(t, g, "Ancestral Blade", "Artifact — Equipment", effectstest.BnsB23AncestralBlade, nil)
	effectstest.PassPriorityAroundTable(t, g)
	soldier := effectstest.AvB1Named(g, me.ID, "Soldier")
	if soldier == uuid.Nil {
		t.Fatal("no Soldier token")
	}
	if c := effectstest.FindBattlefieldCardForTest(g, blade); c == nil || c.AttachedTo.ID != soldier {
		t.Fatal("the Blade is not attached to the Soldier")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, soldier); p != 2 || tt != 2 {
		t.Errorf("equipped Soldier is %d/%d, want 2/2", p, tt)
	}
}

func bnsB23CastAttach(t *testing.T, name, oracle string) (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	foe := effectstest.PushVanillaCreature(g, opp.ID, "Foe", 2, 2)
	eq := effectstest.CastCatalogSpell(t, g, name, "Artifact — Equipment", oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, foe) {
		t.Error("an opponent's creature is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, eq); c == nil || c.AttachedTo.ID != bear {
		t.Fatalf("%s is not attached to the target", name)
	}
	return g, me.ID, eq, bear
}

func TestBnsB23KitesailAttachesAndGrantsFlying(t *testing.T) {
	g, _, _, bear := bnsB23CastAttach(t, "Cliffhaven Kitesail", effectstest.BnsB23Kitesail)
	if !effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("equipped creature lacks flying")
	}
	if p, tt := effectstest.EffectivePTOf(t, g, bear); p != 2 || tt != 2 {
		t.Errorf("the Kitesail changes size: %d/%d", p, tt)
	}
}

func TestBnsB23RelicAxePumpsAndFavoursWarriors(t *testing.T) {
	g, _, _, bear := bnsB23CastAttach(t, "Relic Axe", effectstest.BnsB23RelicAxe)
	if p, tt := effectstest.EffectivePTOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("non-Warrior: %d/%d, want 3/3", p, tt)
	}

	g = effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warrior := effectstest.PushDiesCreatureForTest(g, me.ID, "Warrior", "", "Creature — Human Warrior", 2, 2)
	axe := effectstest.PushCatalogPermanent(g, me.ID, "Relic Axe", "Artifact — Equipment", effectstest.BnsB23RelicAxe, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, axe, warrior)
	if p, tt := effectstest.EffectivePTOf(t, g, warrior); p != 4 || tt != 3 {
		t.Errorf("Warrior: %d/%d, want 4/3 (+2/+1, not +3/+2)", p, tt)
	}
}

func TestBnsB23WhirlermakerMakesAThopter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	w := effectstest.PushCatalogPermanent(g, me.ID, "Whirlermaker", "Artifact", effectstest.BnsB23Whirlermaker, false)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	th := effectstest.AvB1Named(g, me.ID, "Thopter")
	if th == uuid.Nil || !effectstest.HasKeywordOf(g, th, "flying") {
		t.Fatal("no flying Thopter")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, w); !tapped {
		t.Error("{T} is part of the cost")
	}
}

func TestBnsB23UniversalSolventDestroysAnyPermanentAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	s := effectstest.PushCatalogPermanent(g, me.ID, "Universal Solvent", "Artifact", effectstest.BnsB23Solvent, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("activated with only six mana")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(s) {
		t.Error("the Solvent is still there: sacrifice is a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the targeted land survived")
	}
}

func TestBnsB23WitchingWellScriesThenDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		}
	})
	w := effectstest.CastCatalogSpell(t, g, "Witching Well", "Artifact", effectstest.BnsB23WitchingWell, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Fatal("no scry prompt")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	hand := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if g.Battlefield.Contains(w) {
		t.Error("the Well is still there")
	}
}

func TestBnsB23PuzzleknotDrawsAndLosesLifeOnEntryAndSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for i := 0; i < 4; i++ {
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
		}
	})
	life, hand := me.Life, me.Hand.Size()
	k := effectstest.CastCatalogSpell(t, g, "Metalspinner's Puzzleknot", "Artifact", effectstest.BnsB23Puzzleknot, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 { // the cast card left the hand, one drawn
		t.Errorf("entry: hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+1, me.Life, life-1)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("sacrifice: hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+2, me.Life, life-2)
	}
	if g.Battlefield.Contains(k) {
		t.Error("the Puzzleknot is still there")
	}
}

func TestBnsB23StarnheimMementoManaAndSorceryPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	m := effectstest.PushCatalogPermanent(g, me.ID, "Starnheim Memento", "Artifact", effectstest.BnsB23Memento, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePTOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("no flying")
	}
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 || effectstest.HasKeywordOf(g, bear, "flying") {
		t.Error("the pump lasts past end of turn")
	}
}

func TestBnsB23StarnheimMementoRefusesOutsideSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	m := effectstest.PushCatalogPermanent(g, me.ID, "Starnheim Memento", "Artifact", effectstest.BnsB23Memento, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("activated during combat")
	}
}

func TestBnsB23GuardianIdolEntersTappedAndAnimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	idol := effectstest.CastCatalogSpell(t, g, "Guardian Idol", "Artifact", effectstest.BnsB23GuardianIdol, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, ok := effectstest.BattlefieldCardTapped(g, idol); !ok || !tapped {
		t.Fatalf("the Idol must enter tapped (on battlefield %v, tapped %v)", ok, tapped)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, idol)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Golem") || c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("not a 2/2 Golem artifact creature: %+v", c.Effective())
	}
	effectstest.EndTurn(t, g)
	if c := effectstest.LayeredCard(t, g, idol); c.IsCreature() {
		t.Error("still a creature next turn")
	}
}
