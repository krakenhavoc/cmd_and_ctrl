package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB211Chardalyn = "0bf1f2c1-2148-4302-a6e7-d509a005a85f"
	gapB211Javelin   = "3f244320-3e74-4eff-ba4e-a345700c40c9"
	gapB211Walrus    = "3fdc76b0-450a-4fcb-8a64-113edce67f3d"
	gapB211Robe      = "570de99a-7c65-4052-bc4b-cfb8caef40b3"
	gapB211Goggles   = "af27761d-88df-4d4f-a366-d55ed62d1d33"
	gapB211Lapis     = "c61ab8e1-ef23-465d-bcd3-f771434988b2"
	gapB211Greatsw   = "81a602b6-a4f9-4534-a84e-2fa7556c2309"
	gapB211Saddle    = "f6cd939e-6b45-4d47-b251-7a09a149dedd"
	gapB211Fox       = "eb73db77-4050-4617-a29b-f9a8c885b424"
	gapB211Mole      = "9fa3286b-ee10-4f68-bfc3-974cdc39d986"
)

func TestGapB211EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB211Chardalyn: "Chardalyn Dragon", gapB211Javelin: "Javelin of Lightning", gapB211Walrus: "Bronze Walrus",
		gapB211Robe: "Robe of the Archmagi", gapB211Goggles: "Goggles of Night", gapB211Lapis: "Lapis Orb of Dragonkind",
		gapB211Greatsw: "Greatsword of Tyr", gapB211Saddle: "Saddle of the Cavalier", gapB211Fox: "Clockwork Fox",
		gapB211Mole: "Drillworks Mole",
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
}

// gapB211Attach puts an Equipment on the battlefield attached to a new creature.
func gapB211Attach(g *game.Game, owner uuid.UUID, name, oracle string, power int) (equipment, host uuid.UUID) {
	host = effectstest.PushVanillaCreature(g, owner, "Host", power, 3)
	equipment = effectstest.PushCatalogPermanent(g, owner, name, "Artifact — Equipment", oracle, false)
	g.WithWriteLock(func() { _ = g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: host}) })
	return equipment, host
}

// gapB211Hit attacks the opponent with `host`, unblocked, through combat damage.
func gapB211Hit(t *testing.T, g *game.Game, host uuid.UUID, opp *game.Player) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB211ChardalynDragonFlies(t *testing.T) {
	spec, _ := effects.Lookup(gapB211Chardalyn)
	if !effectstest.HasKeyword(spec.PrintedKeywords, "flying") {
		t.Errorf("printed keywords %v, want flying", spec.PrintedKeywords)
	}
}

func TestGapB211JavelinAttachesOnEntryAndGrantsOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Javelin of Lightning", "Artifact — Equipment", gapB211Javelin, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("power %d on your turn, want 4", p)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "first strike") {
		t.Error("no first strike on your turn")
	}
}

func TestGapB211BronzeWalrusScriesOnEntryAndMakesAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	walrus := effectstest.CastHoldCreature(t, g, "Bronze Walrus", "Artifact Creature — Walrus", gapB211Walrus, 2, 2)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt on entry")
	} else {
		effectstest.AnswerScryKeepAll(t, g, me.ID)
	}
	_ = walrus
	ready := effectstest.PushCatalogPermanent(g, me.ID, "Bronze Walrus", "Artifact Creature — Walrus", gapB211Walrus, false)
	if err := g.ActivateManaAbility(me.ID, ready, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana %d, want 1", me.ManaPool.Total())
	}
}

func TestGapB211RobeDrawsThatManyCardsAndEquipsWizardsCheaply(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	_, host := gapB211Attach(g, me.ID, "Robe of the Archmagi", gapB211Robe, 3)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	hand := me.Hand.Size()
	gapB211Hit(t, g, host, opp)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

func TestGapB211RobeCheapEquipOnlyTargetsAShamanWarlockOrWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	robe := effectstest.PushCatalogPermanent(g, me.ID, "Robe of the Archmagi", "Artifact — Equipment", gapB211Robe, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mage := effectstest.B12Creature(g, me.ID, "Mage", "Creature — Human Wizard", 1, 1)
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, robe, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("the {1} equip targeted a Bear")
	}
	if err := g.ActivateCatalogAbility(me.ID, robe, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mage), Strict: true}); err != nil {
		t.Fatalf("the {1} equip on a Wizard: %v", err)
	}
}

func TestGapB211GogglesScryThenDrawOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	_, host := gapB211Attach(g, me.ID, "Goggles of Night", gapB211Goggles, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	gapB211Hit(t, g, host, opp)
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestGapB211GreatswordPutsACounterAndTapsADefendingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	_, host := gapB211Attach(g, me.ID, "Greatsword of Tyr", gapB211Greatsw, 2)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, mine) || !effectstest.HasID(pick.PickTargetCards, wall) {
		t.Errorf("offered %v; want only the defending player's creature", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, wall)
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, wall); !tapped {
		t.Error("the Wall was not tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, host); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters %d, want 1", c.Counters[game.CounterPlusOne])
	}
}

func TestGapB211SaddleGivesPlusThreeAndOnlyBigCreaturesBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	_, host := gapB211Attach(g, me.ID, "Saddle of the Cavalier", gapB211Saddle, 2)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Small", 3, 3)
	big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 4, 4)
	if p := effectstest.EffectivePower(t, g, host); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, host); err == nil {
		t.Error("a power-3 creature blocked the equipped creature")
	}
	if err := g.DeclareBlocker(big, host); err != nil {
		t.Errorf("a power-4 creature may block: %v", err)
	}
}

func TestGapB211ClockworkFoxDrawsTwoAndEachOpponentOneWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	fox := effectstest.PushCatalogPermanent(g, me.ID, "Clockwork Fox", "Artifact Creature — Fox", gapB211Fox, false)
	for _, s := range g.Seats {
		effectstest.SeedLibrary(s, "A", "B", "C")
	}
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Hand.Size()
	}
	effectstest.KillCreature(t, g, me.ID, fox)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := 1
		if s.ID == me.ID {
			want = 2
		}
		if got := s.Hand.Size() - before[s.ID]; got != want {
			t.Errorf("%s drew %d, want %d", s.Name, got, want)
		}
	}
}

func TestGapB211DrillworksMoleCountersItselfAndAnOptionalCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	mole := effectstest.PushCatalogPermanent(g, me.ID, "Drillworks Mole", "Artifact Creature — Mole", gapB211Mole, false)
	cmdr := effectstest.PushVanillaCreature(g, me.ID, "Commander", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, cmdr).IsCommander = true })
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, mole, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other), Strict: true}); err == nil {
		t.Fatal("targeted a creature that is not a commander")
	}
	if err := g.ActivateCatalogAbility(me.ID, mole, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(cmdr), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mole, cmdr} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", c.Name, c.Counters[game.CounterPlusOne])
		}
	}
}

func TestGapB211LapisOrbScriesOnlyWhenItsManaCastsADragon(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		scry           bool
	}{{"a Dragon creature spell", "Creature — Dragon", true}, {"a non-Dragon creature spell", "Creature — Bear", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.SeedLibrary(me, "A", "B", "C")
			orb := effectstest.PushCatalogPermanent(g, me.ID, "Lapis Orb of Dragonkind", "Artifact", gapB211Lapis, false)
			if err := g.ActivateManaAbility(me.ID, orb, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("mana ability: %v", err)
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: tc.typeLine, ManaCost: "{U}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.ScryChoiceFor(g, me.ID) != nil; got != tc.scry {
				t.Errorf("scry prompt open = %v, want %v", got, tc.scry)
			}
		})
	}
}
