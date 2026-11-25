package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theros Beyond Death batch 136 (S82 OLD, #5514).

const (
	oldB136Cyclops  = "789a94be-020e-4f65-8f0d-0a170d4cd2f5"
	oldB136Daybreak = "a2ea435c-5a61-4c98-bb76-8ff4ef0a0da9"
	oldB136Warden   = "d2f33d3a-afc6-4d15-9049-ecd33f85eb94"
	oldB136Giant    = "d601a28d-a02f-45eb-9dab-a3cf50a87ca4"
	oldB136Alirios  = "7dbf9e96-e61f-43db-8767-63a351cf4ef8"
	oldB136Aphemia  = "820dab0f-3b6b-437c-ba4b-2311437709c3"
	oldB136Hoplite  = "8cb6dbb5-9ea2-4c19-8173-148080fb898e"
	oldB136Heroes   = "d4f72eac-2161-4fe8-8590-436e241c3b22"
	oldB136Eidolon  = "15d38195-b661-42c1-82c9-1af2a3ba2ea5"
	oldB136Shimmer  = "21c768fa-4f1a-4a1a-abb9-e69e17426233"
	oldB136Callaphe = "faf081dc-1fb2-47ac-a9c7-8001b4904120"
)

func TestOldB136CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB136Cyclops: "Flummoxed Cyclops", oldB136Daybreak: "Daybreak Chimera", oldB136Warden: "Warden of the Chained",
		oldB136Giant: "Grasping Giant", oldB136Alirios: "Alirios, Enraptured", oldB136Aphemia: "Aphemia, the Cacophony",
		oldB136Hoplite: "Reverent Hoplite", oldB136Heroes: "Heroes of the Revel", oldB136Eidolon: "Eidolon of Obstruction",
		oldB136Shimmer: "Shimmerwing Chimera", oldB136Callaphe: "Callaphe, Beloved of the Sea",
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

// --- Flummoxed Cyclops --------------------------------------------------------

func TestOldB136CyclopsCantBlockWhenAnOpponentAttacksWithTwoOrMoreCreatures(t *testing.T) {
	for _, n := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		attacker, me := g.Seats[0], g.Seats[1]
		cyclops := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Flummoxed Cyclops", OracleID: oldB136Cyclops, TypeLine: "Creature — Cyclops", ManaCost: "{3}{R}", Power: 4, Toughness: 4})
		var attackers []uuid.UUID
		for i := 0; i < n; i++ {
			attackers = append(attackers, effectstest.PushVanillaCreature(g, attacker.ID, "Attacker", 1, 1))
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range attackers {
			if err := g.DeclareAttacker(id, me.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		cantBlock := effectstest.AuraRestrictions(t, g, cyclops)&game.CantBlock != 0
		if want := n >= 2; cantBlock != want {
			t.Errorf("%d attacker(s): the Cyclops can't block = %v, want %v", n, cantBlock, want)
		}
	}
}

// --- Daybreak Chimera ---------------------------------------------------------

func TestOldB136DaybreakChimeraCostsLessByDevotionToWhite(t *testing.T) {
	cast := func(devotion string) error {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if devotion != "" {
			effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "White Thing", TypeLine: "Creature — Soldier", ManaCost: devotion, Power: 1, Toughness: 1})
		}
		effectstest.ToMainForCost(t, g)
		effectstest.FloatForTest(g, me, "WW")
		id := effectstest.RfdHandCard(me, "Daybreak Chimera", "Creature — Chimera", "{3}{W}{W}", oldB136Daybreak, 3, 3)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast("{W}{W}{W}"); err != nil {
		t.Errorf("with devotion 3 the Chimera should cost {W}{W}: %v", err)
	}
	if err := cast(""); err == nil {
		t.Error("with no devotion {W}{W} paid for a {3}{W}{W} spell")
	}
}

// --- Warden of the Chained ----------------------------------------------------

func TestOldB136WardenCantAttackUnlessYouControlAnotherCreatureWithPowerFourOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warden := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Warden of the Chained", OracleID: oldB136Warden, TypeLine: "Creature — Minotaur Warrior", ManaCost: "{1}{R}{G}", Power: 4, Toughness: 4})
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(warden, opp.ID); err == nil {
		t.Fatal("the Warden attacked with no other creature of power 4 or greater (its own power does not count)")
	}
	_ = small
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	if err := g.DeclareAttacker(warden, opp.ID); err != nil {
		t.Errorf("the Warden could not attack with another 4-power creature: %v", err)
	}
}

// --- Grasping Giant -----------------------------------------------------------

func TestOldB136GraspingGiantExilesEachCreatureThatBlocksItUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Grasping Giant", OracleID: oldB136Giant, TypeLine: "Creature — Giant", ManaCost: "{5}{W}", Power: 5, Toughness: 7})
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(giant, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.DeclareBlocker(blocker, giant); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInCombatDeclaration(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AdZone(g, blocker); got != game.ZoneExile {
		t.Fatalf("the blocker is in zone %v, want exile", got)
	}
	effectstest.RadDestroy(t, g, giant)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Blocker"); n != 1 {
		t.Errorf("after the Giant left the blocker was on the battlefield %d times, want 1 (a new object)", n)
	}
}

// --- Alirios, Enraptured ------------------------------------------------------

func TestOldB136AliriosEntersTappedMakesAReflectionAndStaysTappedWhileYouControlOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.RfdCast(t, g, "Alirios, Enraptured", "Legendary Creature — Human", oldB136Alirios, 2, 3)
	var tapped bool
	var reflections int
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			tapped = c.Tapped
		}
		if c.Controller == me.ID && c.Name == "Reflection" && c.Power == 3 && c.Toughness == 2 {
			reflections++
		}
	}
	if !tapped || reflections != 1 {
		t.Fatalf("Alirios tapped = %v, Reflections = %d, want tapped and one 3/2 Reflection", tapped, reflections)
	}
	// Through your next untap step Alirios stays tapped while the Reflection lives.
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && !c.Tapped {
			t.Error("Alirios untapped while you control a Reflection")
		}
	}
	// With the Reflection gone it untaps.
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Reflection" {
			effectstest.RadDestroy(t, g, c.InstanceID)
			break
		}
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, 0)
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id && c.Tapped {
			t.Error("Alirios stayed tapped with no Reflection")
		}
	}
}

// --- Aphemia, the Cacophony ---------------------------------------------------

func TestOldB136AphemiaExilesAnEnchantmentCardAtYourEndStepForAZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Aphemia, the Cacophony", OracleID: oldB136Aphemia, TypeLine: "Legendary Enchantment Creature — Harpy", ManaCost: "{1}{B}", Power: 2, Toughness: 1})
	ench := effectstest.ApaCreature("Dead Enchantment", "", 0, 0)
	ench.TypeLine = "Enchantment"
	var id uuid.UUID
	g.WithWriteLock(func() {
		ench.InstanceID = uuid.New()
		ench.Owner, ench.Controller = me.ID, me.ID
		me.Graveyard.PushTop(ench)
		id = ench.InstanceID
	})
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.AdZone(g, id); got != game.ZoneExile {
		t.Errorf("the enchantment is in zone %v, want exile", got)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Zombie") {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Zombies, want 1", n)
	}
}

// --- Reverent Hoplite ---------------------------------------------------------

func TestOldB136HopliteMakesATokenForEachWhiteSymbolYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "White Thing", TypeLine: "Creature — Soldier", ManaCost: "{W}{W}", Power: 1, Toughness: 1})
	id := effectstest.RfdHandCard(me, "Reverent Hoplite", "Creature — Human Soldier", "{4}{W}", oldB136Hoplite, 1, 2)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Human Soldier" {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d Human Soldier tokens, want 3 (devotion 2 + the Hoplite's {W})", n)
	}
}

// --- Heroes of the Revel ------------------------------------------------------

func TestOldB136HeroesMakesASatyrAndPumpsWhenYouTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	heroes := effectstest.RfdCast(t, g, "Heroes of the Revel", "Creature — Satyr Soldier", oldB136Heroes, 4, 4)
	var satyr uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Satyr" {
			satyr = c.InstanceID
		}
	}
	if satyr == uuid.Nil {
		t.Fatal("no Satyr token")
	}
	if effectstest.AuraRestrictions(t, g, satyr)&game.CantBlock == 0 {
		t.Error("the Satyr can block")
	}
	// A spell that targets Heroes: creatures you control get +1/+0.
	effectstest.FloatMana(t, g, me, "{G}")
	id := effectstest.RfdHandCard(me, "Giant Growth", "Instant", "{G}", "", 0, 0)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: heroes}}}); err != nil {
		t.Skipf("the stand-in spell could not be cast here: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, satyr); p != 2 {
		t.Errorf("the Satyr has power %d, want 2", p)
	}
}

// --- Eidolon of Obstruction ---------------------------------------------------

func TestOldB136EidolonTaxesAnOpponentsPlaneswalkerLoyaltyAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Eidolon of Obstruction", OracleID: oldB136Eidolon, TypeLine: "Enchantment Creature — Spirit", ManaCost: "{1}{W}", Power: 2, Toughness: 1})
	walker := func(controller uuid.UUID) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test", Owner: controller, Controller: controller}
	}
	price := func(c game.Card, loyalty bool) string {
		base, _ := game.ParseCost("{0}")
		out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: c, Controller: c.Controller, FromZone: game.ZoneBattlefield,
			Ability: &game.AbilityCostSubject{Label: "ability", Loyalty: loyalty}})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
		return out.String()
	}
	if got := price(walker(opp.ID), true); got != "{1}" {
		t.Errorf("an opponent's loyalty ability costs %q more, want {1}", got)
	}
	if got := price(walker(opp.ID), false); got != "{0}" && got != "" {
		t.Errorf("an opponent's non-loyalty ability is taxed: %q", got)
	}
	if got := price(walker(me.ID), true); got != "{0}" && got != "" {
		t.Errorf("your own loyalty ability is taxed: %q", got)
	}
}

// --- Shimmerwing Chimera ------------------------------------------------------

func TestOldB136ShimmerwingReturnsAnotherEnchantmentAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chimera := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Shimmerwing Chimera", OracleID: oldB136Shimmer, TypeLine: "Enchantment Creature — Chimera", ManaCost: "{3}{U}", Power: 3, Toughness: 2})
	ench := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My Enchantment", TypeLine: "Enchantment"})
	for i := 0; i < 400 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the upkeep trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, chimera) || !effectstest.HasID(p.PickTargetCards, ench) {
		t.Errorf("targets %v: another enchantment only", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, ench)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment was not returned to hand")
	}
}

// --- Callaphe, Beloved of the Sea --------------------------------------------

func TestOldB136CallapheHasPowerEqualToDevotionAndTaxesSpellsThatTargetYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	callaphe := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Callaphe, Beloved of the Sea", OracleID: oldB136Callaphe, TypeLine: "Legendary Enchantment Creature — Demigod", ManaCost: "{1}{U}{U}", Toughness: 3})
	other := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Blue Thing", TypeLine: "Creature — Merfolk", ManaCost: "{U}", Power: 1, Toughness: 1})
	if p, _ := effectstest.PtOf(t, g, callaphe); p != 3 {
		t.Errorf("Callaphe's power is %d, want 3 (devotion to blue)", p)
	}
	price := func(who uuid.UUID, target uuid.UUID) string {
		spell := game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: who, Controller: who}
		base, _ := game.ParseCost("{R}")
		out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: spell, Controller: who, FromZone: game.ZoneHand,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
		if err != nil {
			t.Fatalf("ApplyCostModifiers: %v", err)
		}
		return out.String()
	}
	if got := price(opp.ID, other); got != "{1}{R}" {
		t.Errorf("an opponent's spell targeting your creature costs %q, want {1}{R}", got)
	}
	if got := price(opp.ID, callaphe); got != "{1}{R}" {
		t.Errorf("an opponent's spell targeting Callaphe costs %q, want {1}{R}", got)
	}
	if got := price(me.ID, other); got != "{R}" {
		t.Errorf("your own spell costs %q, want {R}", got)
	}
}
