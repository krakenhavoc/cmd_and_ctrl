package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b124_cards_test.go — The Lord of the Rings: Tales of Middle-earth
// creatures (S82 GAP batch 124, #5948). Every test plays the real catalog
// entry through the engine.

const (
	gapB124UglukOracle      = "ebd6a3a0-ae8d-44e6-8d0b-048da6dd1f03"
	gapB124MouthOracle      = "49e3396a-a125-4cca-b02a-4a96102bcb16"
	gapB124DenethorOracle   = "1b82312c-f748-4c36-8934-ecbcfc10cc7d"
	gapB124KingOracle       = "3ee451e7-2b2f-4ce7-af96-540bba555c60"
	gapB124FarmerOracle     = "0fea3406-e5c7-4d5c-9890-6f1bc04a7cb8"
	gapB124GaladrielOracle  = "d0c0a2ef-27b4-4186-8538-6ba5d35d85bb"
	gapB124GenerousOracle   = "b6859c7a-4c4d-4652-938f-ecedce29d506"
	gapB124FangornOracle    = "04166d20-88c9-4c89-be9d-3e422f90959a"
	gapB124LegolasOracle    = "5e99497f-7c73-49ae-88b3-78d5d74ba9ea"
	gapB124ElvesOracle      = "8958a6fe-76ef-45e3-8522-6bc13486b447"
	gapB124GlorfindelOracle = "93842030-2017-4233-a7ac-7112361c019f"
	gapB124PorterOracle     = "b80473f7-f1d4-4aa3-988e-1a8610ffbeee"
	gapB124LegendType       = "Legendary Creature — Elf Noble"
)

func TestGapB124EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB124UglukOracle:      "Uglúk of the White Hand",
		gapB124MouthOracle:      "The Mouth of Sauron",
		gapB124DenethorOracle:   "Denethor, Ruling Steward",
		gapB124KingOracle:       "King of the Oathbreakers",
		gapB124FarmerOracle:     "Brandywine Farmer",
		gapB124GaladrielOracle:  "Galadriel, Gift-Giver",
		gapB124GenerousOracle:   "Generous Ent",
		gapB124FangornOracle:    "Fangorn, Tree Shepherd",
		gapB124LegolasOracle:    "Legolas, Master Archer",
		gapB124ElvesOracle:      "Chance-Met Elves",
		gapB124GlorfindelOracle: "Glorfindel, Dauntless Rescuer",
		gapB124PorterOracle:     "Bag End Porter",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func gapB124Push(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner})
}

func gapB124Count(g *game.Game, owner uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			n++
		}
	}
	return n
}

// gapB124Scry scries 1 for `me` and keeps the card on top.
func gapB124Scry(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	if c := effectstest.ScryChoiceFor(g, me.ID); c != nil {
		if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
			t.Fatalf("ResolveScry: %v", err)
		}
	}
}

func TestGapB124BrandywineFarmerMakesAFoodWhenItEntersAndWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	farmer := effectstest.TdCastFromHand(t, g, game.Card{Name: "Brandywine Farmer", TypeLine: "Creature — Halfling Peasant", OracleID: gapB124FarmerOracle,
		ManaCost: "{2}{G}", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB124Count(g, me.ID, "Food"); n != 1 {
		t.Fatalf("%d Food after it entered, want 1", n)
	}
	effectstest.P1Destroy(t, g, farmer)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB124Count(g, me.ID, "Food"); n != 2 {
		t.Errorf("%d Food after it left, want 2", n)
	}
}

func TestGapB124GenerousEntMakesFoodAndForestcycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Generous Ent", TypeLine: "Creature — Treefolk", OracleID: gapB124GenerousOracle,
		ManaCost: "{5}{G}", Power: 5, Toughness: 7}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB124Count(g, me.ID, "Food"); n != 1 {
		t.Errorf("%d Food, want 1", n)
	}
	spec, _ := effects.Lookup(gapB124GenerousOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "reach" {
		t.Errorf("keywords = %v, want reach", spec.PrintedKeywords)
	}
	if len(spec.Activated) != 1 {
		t.Fatalf("%d activated abilities, want the forestcycling", len(spec.Activated))
	}
}

func TestGapB124BagEndPorterGetsPlusXPerLegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	porter := effectstest.PushCatalogPermanent(g, me.ID, "Bag End Porter", "Creature — Dwarf", gapB124PorterOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == porter {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 4
			}
		}
	})
	gapB124Push(g, me.ID, "Legend One", gapB124LegendType, 1, 1)
	gapB124Push(g, me.ID, "Legend Two", gapB124LegendType, 1, 1)
	gapB124Push(g, opp.ID, "Their Legend", gapB124LegendType, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, porter)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, porter); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("Porter is %d/%d after attacking with two legendary creatures of mine, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestGapB124UglukGrowsWhenAnotherCreatureOfYoursDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ugluk := effectstest.PushCatalogPermanent(g, me.ID, "Uglúk of the White Hand", "Legendary Creature — Orc Soldier", gapB124UglukOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	orc := gapB124Push(g, me.ID, "Orc Raider", "Creature — Orc Soldier", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	counters := func() int { return effectstest.CountersOn(g, ugluk, game.CounterPlusOne) }

	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 0 {
		t.Fatalf("an opponent's creature dying gave %d counters", counters())
	}
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 1 {
		t.Fatalf("a Bear dying gave %d counters, want 1", counters())
	}
	effectstest.P1Destroy(t, g, orc)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 3 {
		t.Errorf("%d counters after an Orc died too, want 3 (two for the Orc)", counters())
	}
}

func TestGapB124DenethorMakesASoldierOnlyAfterACreatureDiedAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	den := effectstest.PushCatalogPermanent(g, me.ID, "Denethor, Ruling Steward", "Legendary Creature — Human Noble", gapB124DenethorOracle, false)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	effectstest.AdvanceToMain(t, g)
	// The drain: {2}, sacrifice another creature.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatalf("AddMana: %v", err)
	}
	myLife, theirLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, den, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{den}}); err == nil {
		t.Fatal("Denethor sacrificed himself")
	}
	if err := g.ActivateCatalogAbility(me.ID, den, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, myLife+1, theirLife-1)
	}
	// A creature died under my control this turn: the end step makes a Soldier.
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB124Count(g, me.ID, "Human Soldier"); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", n)
	}
}

func TestGapB124DenethorMakesNoSoldierWhenNothingDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Denethor, Ruling Steward", "Legendary Creature — Human Noble", gapB124DenethorOracle, false)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB124Count(g, me.ID, "Human Soldier"); n != 0 {
		t.Errorf("%d Human Soldier tokens with no death, want 0", n)
	}
}

func TestGapB124ChanceMetElvesGrowsOncePerTurnOnScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.PushCatalogPermanent(g, me.ID, "Chance-Met Elves", "Creature — Elf Warrior", gapB124ElvesOracle, false)
	effectstest.SeedLibraryTop(me, "A", "Instant", "")
	effectstest.SeedLibraryTop(me, "B", "Instant", "")
	gapB124Scry(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	gapB124Scry(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, elves, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after two scries in one turn, want 1", n)
	}
}

func TestGapB124GlorfindelPumpsAndChoosesAMode(t *testing.T) {
	for mode := 0; mode < 2; mode++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		glor := effectstest.PushCatalogPermanent(g, me.ID, "Glorfindel, Dauntless Rescuer", gapB124LegendType, gapB124GlorfindelOracle, false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == glor {
					g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 2
				}
			}
		})
		effectstest.SeedLibraryTop(me, "A", "Instant", "")
		a := effectstest.PushVanillaCreature(g, opp.ID, "Blocker A", 2, 2)
		b := effectstest.PushVanillaCreature(g, opp.ID, "Blocker B", 2, 2)
		gapB124Scry(t, g, me)
		effectstest.PassPriorityAroundTable(t, g)
		mp := effectstest.ModePickChoiceFor(g, me.ID)
		if mp == nil {
			t.Fatalf("mode %d: no mode prompt", mode)
		}
		if err := g.ResolveModePick(mp.ID, me.ID, []int{mode}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.LayeredCard(t, g, glor); c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
			t.Errorf("mode %d: Glorfindel is %d/%d, want 4/3", mode, c.CurrentPower(), c.CurrentToughness())
		}
		effectstest.DeclareAttack(t, g, opp.ID, glor)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if mode == 1 {
			if err := g.DeclareBlocker(a, glor); err != nil {
				t.Fatalf("first block: %v", err)
			}
			if err := g.DeclareBlocker(b, glor); err == nil {
				t.Error("two creatures blocked Glorfindel under the one-blocker cap")
			}
		} else {
			// Must be blocked if able: the lock-in is refused with able blockers
			// who didn't block.
			if err := tryLockInBlocks(g); err == nil {
				t.Error("blocks were locked in with no block on a must-be-blocked creature")
			}
		}
	}
}

func tryLockInBlocks(g *game.Game) error {
	_, err := g.AdvanceStep()
	return err
}

func TestGapB124LegolasGrowsWhenTargetedAndShootsWhenYouTargetAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	legolas := effectstest.PushCatalogPermanent(g, me.ID, "Legolas, Master Archer", "Legendary Creature — Elf Archer", gapB124LegolasOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == legolas {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 4
			}
		}
	})
	// A spell that targets Legolas: a counter, and no shot (he is mine).
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: legolas}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, legolas, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters after being targeted, want 1", n)
	}
	// A spell that targets an opponent's creature: Legolas shoots (power 2 now).
	targeted := effectstest.PushVanillaCreature(g, opp.ID, "Targeted", 5, 5)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: targeted}})
	answerPendingTargetIfAny(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectsBattlefieldHas(g, victim) {
		t.Error("Legolas did not deal his power in damage to the chosen creature")
	}
}

func effectsBattlefieldHas(g *game.Game, id uuid.UUID) bool { return effectstest.BattlefieldHas(g, id) }

func TestGapB124GaladrielChoosesAModeWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	galadriel := effectstest.TdCastFromHand(t, g, game.Card{Name: "Galadriel, Gift-Giver", TypeLine: "Legendary Creature — Elf Noble", OracleID: gapB124GaladrielOracle,
		ManaCost: "{3}{G}{G}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	mp := effectstest.ModePickChoiceFor(g, me.ID)
	if mp == nil {
		t.Fatal("no mode prompt when Galadriel entered")
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	answerPendingTargetIfAny(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters on the other creature, want 1", n)
	}
	if n := effectstest.CountersOn(g, galadriel, game.CounterPlusOne); n != 0 {
		t.Errorf("Galadriel put a counter on herself (%d)", n)
	}
}

func TestGapB124GaladrielChoosesAModeWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	attacker := effectstest.PushCatalogPermanent(g, me.ID, "Galadriel, Gift-Giver", "Legendary Creature — Elf Noble", gapB124GaladrielOracle, false)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	mp := effectstest.ModePickChoiceFor(g, me.ID)
	if mp == nil {
		t.Fatal("no mode prompt when Galadriel attacked")
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB124Count(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

func TestGapB124TheMouthOfSauronMillsThenAmassesByInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	old := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Sorcery", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.SeedLibraryTop(opp, "Top Creature", "Creature — Bear", "")
	effectstest.SeedLibraryTop(opp, "Top Land", "Basic Land — Forest", "")
	effectstest.SeedLibraryTop(opp, "Top Instant", "Instant", "")
	effectstest.SeedLibraryTop(opp, "Top Sorcery", "Sorcery", "")
	gyBefore := opp.Graveyard.Size()
	effectstest.TdCastFromHand(t, g, game.Card{Name: "The Mouth of Sauron", TypeLine: "Legendary Creature — Human Advisor", OracleID: gapB124MouthOracle,
		ManaCost: "{3}{U}{B}", Power: 3, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - gyBefore; got != 3 {
		t.Fatalf("the target milled %d cards, want 3 (targeting needs the opponent)", got)
	}
	var army uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Army") {
			army = c.InstanceID
		}
	}
	if army == uuid.Nil {
		t.Fatal("no Army was created")
	}
	// Graveyard: Old Sorcery, plus the milled Sorcery and Instant and Land: 3 spells.
	if n := effectstest.CountersOn(g, army, game.CounterPlusOne); n != 3 {
		t.Errorf("Army has %d counters, want 3 instants and sorceries in the graveyard", n)
	}
}

func TestGapB124FangornGrantsVigilanceAndAddsGreenForAttackingTreefolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	fangorn := effectstest.PushCatalogPermanent(g, me.ID, "Fangorn, Tree Shepherd", "Legendary Creature — Treefolk", gapB124FangornOracle, false)
	ent := gapB124Push(g, me.ID, "Ent", "Creature — Treefolk", 3, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	for _, id := range []uuid.UUID{fangorn, ent} {
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "vigilance") {
			t.Errorf("a Treefolk lacks vigilance")
		}
	}
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("a non-Treefolk gained vigilance")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("G"); got != 0 {
		t.Fatalf("%d {G} for a Bear attack, want 0", got)
	}
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, fangorn, ent)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("G"); got != 4 {
		t.Errorf("%d {G} after two Treefolk attacked, want twice that much (4)", got)
	}
	// The mana stays as steps end.
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if got := me.ManaPool.CountColor("G"); got != 4 {
		t.Errorf("%d {G} after the step ended, want 4 (green mana doesn't empty)", got)
	}
}

func TestGapB124KingOfTheOathbreakersPhasesOutWhenTargetedAndTokensOnPhaseIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := effectstest.PushCatalogPermanent(g, me.ID, "King of the Oathbreakers", "Legendary Creature — Spirit Noble", gapB124KingOracle, false)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetCard, ID: king}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, king) {
		t.Fatal("the King stayed on the battlefield after becoming the target of a spell")
	}
	// It phases in during its controller's next untap step, making a tapped Spirit.
	for i := 0; i < 4; i++ {
		effectstest.EndTurn(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.BattlefieldHas(g, king) {
		t.Fatal("the King did not phase back in")
	}
	spirits := effectstest.BattlefieldIDsNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(spirits))
	}
	c, _ := effectstest.BattlefieldCard(g, spirits[0])
	if c.Power != 1 || c.Toughness != 1 || !effectstest.EffectiveAbilitiesContain(t, g, spirits[0], "flying") {
		t.Errorf("Spirit token %d/%d, want a 1/1 flyer", c.Power, c.Toughness)
	}
}
