package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_sz_cards_test.go — Final Fantasy (FIN), the cards S–Z, batch 1
// (#4462). Each test plays the real catalog entry and asserts what its
// text says. The fixtures (finABOpp, finABAttack, finABCard) are the
// A–B batch's.

const (
	finSZVectorOracle      = "8dcba63c-4701-4dc9-81f5-ca8e8933a3ba"
	finSZWindurstOracle    = "e198126a-f280-47f7-8bd0-3dc9f5ff05a0"
	finSZDireRatOracle     = "f09811d0-c5da-4061-aaa7-835ccee3b05a"
	finSZNotAloneOracle    = "9dbc2fc8-ed42-4534-9676-bf0695cd518c"
	finSZWarriorSwordOracl = "a339961e-3555-467f-a4f8-c5eaac7e7ed1"
	finSZWhiteStaffOracle  = "af1acf07-39e2-441e-a1e0-1db535535e2e"
	finSZXandeOracle       = "36b5c793-30f6-4f21-8bf6-c18d1c46d732"
	finSZOverworldOracle   = "61e3709e-7ecd-4a59-bca1-019d16795740"
	finSZValkyrieOracle    = "1333d690-a6dc-4ef8-8c57-2ba16727c0a5"
	finSZSlashOracle       = "3cc4d6f4-79b2-4210-84de-a0febf81eb67"
	finSZVayneOracle       = "93af8f9d-6877-4cb8-a744-c89186db22e1"
	finSZTidusOracle       = "ef827c62-a5c5-4a7a-92a8-edefb081fae8"
	finSZSolRingOracle     = "6ad8011d-3471-4369-9d68-b264cc027487"
)

func TestFinSZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finSZVectorOracle:      "Vector, Imperial Capital",
		finSZWindurstOracle:    "Windurst, Federation Center",
		finSZDireRatOracle:     "Undercity Dire Rat",
		finSZNotAloneOracle:    "You're Not Alone",
		finSZWarriorSwordOracl: "Warrior's Sword",
		finSZWhiteStaffOracle:  "White Mage's Staff",
		finSZXandeOracle:       "Xande, Dark Mage",
		finSZOverworldOracle:   "Travel the Overworld",
		finSZValkyrieOracle:    "Valkyrie Aerial Unit",
		finSZSlashOracle:       "Slash of Light",
		finSZVayneOracle:       "Vayne's Treachery",
		finSZTidusOracle:       "Tidus, Blitzball Star",
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

// --- the Town taplands ----------------------------------------------------------

func TestFinSZTownsEnterTappedAndTapForEitherColour(t *testing.T) {
	for _, row := range []struct {
		name, oracle string
		colours      [2]string
	}{
		{"Vector, Imperial Capital", finSZVectorOracle, [2]string{"B", "R"}},
		{"Windurst, Federation Center", finSZWindurstOracle, [2]string{"G", "W"}},
	} {
		for _, colour := range row.colours {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.B12PlayFromHand(t, g, row.name, "Land — Town", row.oracle, game.CastSpellParams{})
			if !effectstest.B16Tapped(t, g, land) {
				t.Fatalf("%s should enter tapped", row.name)
			}
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
			effectstest.B28TapForMana(t, g, me.ID, land, colour)
			if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
				t.Errorf("%s tapped for {%s}: pool %v", row.name, colour, got)
			}
		}
	}
}

// --- Undercity Dire Rat ----------------------------------------------------------

func TestFinSZDireRatLeavesATreasureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rat := effectstest.PushDiesCreatureForTest(g, me.ID, "Undercity Dire Rat", finSZDireRatOracle, "Creature — Rat", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rat) })
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasure tokens, want 1", n)
	}
}

// --- You're Not Alone ------------------------------------------------------------

func TestFinSZNotAloneIsBiggerWithThreeCreatures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		others int
		want   int
	}{
		{"alone", 0, 4},
		{"two others", 1, 4},
		{"three creatures", 2, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			for i := 0; i < tc.others; i++ {
				effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
			}
			effectstest.CastCatalogSpell(t, g, "You're Not Alone", "Instant", finSZNotAloneOracle, effectstest.FinABCard(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if v := effectstest.S58p6View(t, g, bear); v.CurrentPower() != tc.want || v.CurrentToughness() != tc.want {
				t.Errorf("the target is %d/%d, want %d/%d", v.CurrentPower(), v.CurrentToughness(), tc.want, tc.want)
			}
		})
	}
}

// --- the job-select Equipment ----------------------------------------------------

func TestFinSZWarriorsSwordMakesAFourThreeWarriorHero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	sword := effectstest.CastCatalogSpell(t, g, "Warrior's Sword", "Artifact — Equipment", finSZWarriorSwordOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, sword)
	v := effectstest.S58p6View(t, g, hero)
	if v.CurrentPower() != 4 || v.CurrentToughness() != 3 || !v.HasSubtype("Warrior") || !v.HasSubtype("Hero") {
		t.Errorf("Hero is %d/%d %v, want a 4/3 Hero Warrior", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
}

func TestFinSZWhiteMagesStaffGainsOneOnTheAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	staff := effectstest.CastCatalogSpell(t, g, "White Mage's Staff", "Artifact — Equipment", finSZWhiteStaffOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hero := effectstest.YsBHero(t, g, staff)
	if v := effectstest.S58p6View(t, g, hero); v.CurrentPower() != 2 || v.CurrentToughness() != 2 || !v.HasSubtype("Cleric") {
		t.Fatalf("Hero is %d/%d %v, want a 2/2 Hero Cleric", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AttachForEffect(staff, game.TargetRef{Kind: game.TargetCard, ID: bear}) })
	life := me.Life
	effectstest.FinABAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1 from the attack", life, me.Life)
	}
}

// --- Xande, Dark Mage ------------------------------------------------------------

func TestFinSZXandeCountsNoncreatureNonlandCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	xande := effectstest.B12Push(g, me.ID, "Xande, Dark Mage", "Legendary Creature — Human Wizard", finSZXandeOracle, 3, 3)
	effectstest.PushGraveyardCardTyped(me, "Shock", "Instant")
	effectstest.PushGraveyardCardTyped(me, "Divination", "Sorcery")
	effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(me, "Island", "Basic Land — Island")
	effectstest.PushGraveyardCardTyped(me, "Ornithopter", "Artifact Creature — Thopter")
	effectstest.PushGraveyardCardTyped(opp, "Their Shock", "Instant")
	if v := effectstest.S58p6View(t, g, xande); v.CurrentPower() != 5 || v.CurrentToughness() != 5 {
		t.Errorf("Xande is %d/%d with two noncreature, nonland cards, want 5/5", v.CurrentPower(), v.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, xande, "menace") {
		t.Error("Xande has no menace")
	}
}

// --- Travel the Overworld --------------------------------------------------------

func TestFinSZTravelTheOverworldCostsLessPerTownAndDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Starting Town", TypeLine: "Land — Town"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Treno, Dark City", TypeLine: "Land — Town"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Town", TypeLine: "Land — Town"})
	if got := effectstest.SelfPricedMV(t, g, me, finSZOverworldOracle, "Sorcery", "{5}{U}{U}"); got != 5 {
		t.Errorf("with two Towns it costs %d, want 5", got)
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Travel the Overworld", "Sorcery", finSZOverworldOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4", got)
	}
}

// --- Valkyrie Aerial Unit --------------------------------------------------------

func TestFinSZValkyrieCostsLessPerArtifactAndSurveilsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	if got := effectstest.SelfPricedMV(t, g, me, finSZValkyrieOracle, "Artifact Creature — Construct", "{5}{U}{U}"); got != 4 {
		t.Errorf("with three artifacts it costs %d, want 4", got)
	}
	effectstest.SeedLibrary(me, "Top A", "Top B", "Third")
	unit := effectstest.CastCatalogSpell(t, g, "Valkyrie Aerial Unit", "Artifact Creature — Construct", finSZValkyrieOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("surveil = %+v, want a look at two cards", c)
	}
	if !effectstest.HasEffectiveKeyword(t, g, unit, "flying") {
		t.Error("Valkyrie Aerial Unit has no flying")
	}
}

// --- Slash of Light --------------------------------------------------------------

func TestFinSZSlashOfLightCountsCreaturesAndEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FinABEquipment(g, me.ID, "Sword")
	effectstest.FinABEquipment(g, opp.ID, "Their Sword")
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	// Two creatures plus one Equipment is three: a 3/3 dies, a 3/4 lives.
	small := effectstest.B12Creature(g, opp.ID, "Three-Three", "Creature — Beast", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Slash of Light", "Instant", finSZSlashOracle, effectstest.FinABCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 3/3 survived three damage")
	}
	big := effectstest.B12Creature(g, opp.ID, "Three-Four", "Creature — Beast", 3, 4)
	effectstest.CastCatalogSpell(t, g, "Slash of Light", "Instant", finSZSlashOracle, effectstest.FinABCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("the 3/4 died to what should have been three damage")
	}
}

// --- Vayne's Treachery -----------------------------------------------------------

func TestFinSZVaynesTreacheryShrinksByTwoUnkicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.FinABOpp(g)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Vayne's Treachery", "Instant", finSZVayneOracle, effectstest.FinABCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 2 || tough != 2 {
		t.Errorf("unkicked: %d/%d, want 2/2", p, tough)
	}
}

func TestFinSZVaynesTreacheryKickedBySacrificeShrinksBySix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	effectstest.AdvanceToMain(t, g)
	huge := effectstest.B12Creature(g, opp.ID, "Huge Bear", "Creature — Bear", 7, 7)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.Pr7bCastNow(t, g, "Vayne's Treachery", "Instant", finSZVayneOracle, game.CastSpellParams{
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{fodder},
		Targets:       effectstest.FinABCard(huge),
	})
	if g.Battlefield.Contains(fodder) {
		t.Error("the kicker's sacrifice was not paid")
	}
	if p, tough := effectstest.EffectivePower(t, g, huge), effectstest.EffectiveToughness(t, g, huge); p != 1 || tough != 1 {
		t.Errorf("kicked: %d/%d, want 1/1 (-6/-6 instead of -2/-2)", p, tough)
	}
}

// --- Tidus, Blitzball Star -------------------------------------------------------

func TestFinSZTidusGrowsWhenYourArtifactEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tidus := effectstest.B12Push(g, me.ID, "Tidus, Blitzball Star", "Legendary Creature — Human Warrior", finSZTidusOracle, 2, 1)
	effectstest.CastCatalogSpell(t, g, "Sol Ring", "Artifact", finSZSolRingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, tidus, game.CounterPlusOne); got != 1 {
		t.Errorf("Tidus has %d +1/+1 counters after your artifact entered, want 1", got)
	}
}

func TestFinSZTidusTapsAnOpponentsCreatureWhenHeAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.FinABOpp(g)
	tidus := effectstest.B12Push(g, me.ID, "Tidus, Blitzball Star", "Legendary Creature — Human Warrior", finSZTidusOracle, 2, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWithTrigger(t, g, tidus, opp.ID, victim)
	if !effectstest.TappedNow(g, victim) {
		t.Error("Tidus's attack did not tap the opponent's creature")
	}
	if effectstest.TappedNow(g, mine) {
		t.Error("a creature of Tidus's controller was tapped")
	}
}
