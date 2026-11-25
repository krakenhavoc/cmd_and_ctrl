package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b325_test.go — Dragons of Tarkir, batch 325 (#5551): the pump and
// combat-trick cards. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	oldB325TreadUpon  = "0b3bdbc6-cde7-4187-b13c-82fd02ae4037"
	oldB325ShapeSands = "d96506b1-845d-4a9b-8414-456bb457d80f"
	oldB325Press      = "7a021af0-d54d-4220-ad01-1cfed9a65e98"
	oldB325Volcanic   = "6944261c-9f35-4a51-9b1a-c29e24f292e1"
	oldB325Monk       = "8711b732-8dab-41d4-a85b-b10edcd25e97"
	oldB325Artisan    = "2fa4e241-6b2f-4372-8bb6-a876e5ae7c50"
	oldB325Plague     = "adfba95e-c73c-415d-adea-edcd04372ba4"
	oldB325Spellfist  = "b9ba00f2-b11c-45d8-9c38-f9ade5d07d0f"
	oldB325Scalelords = "4f868b9b-42f4-4f74-bce2-3f9c063a572c"
)

func TestOldB325CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB325TreadUpon: "Tread Upon", oldB325ShapeSands: "Shape the Sands", oldB325Press: "Press the Advantage",
		oldB325Volcanic: "Volcanic Rush", oldB325Monk: "Strongarm Monk", oldB325Artisan: "Graceblade Artisan",
		oldB325Plague: "Virulent Plague", oldB325Spellfist: "Elusive Spellfist", oldB325Scalelords: "Sight of the Scalelords",
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

func TestOldB325SingleTargetTricks(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		power, tough          int
	}{
		{"Tread Upon", oldB325TreadUpon, "trample", 4, 4},
		{"Shape the Sands", oldB325ShapeSands, "reach", 2, 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			effectstest.CastCatalogSpell(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tou := effectstest.PtOf(t, g, bear); p != tc.power || tou != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tou, tc.power, tc.tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
			effectstest.AdvanceToNextSeatsTurn(t, g)
			if p, tou := effectstest.PtOf(t, g, bear); p != 2 || tou != 2 {
				t.Errorf("after cleanup the Bear is %d/%d, want 2/2", p, tou)
			}
			if effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear still has %s after cleanup", tc.keyword)
			}
		})
	}
}

func TestOldB325SingleTargetTricksNeedACreature(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Tread Upon", oldB325TreadUpon}, {"Shape the Sands", oldB325ShapeSands},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
		if err := effectstest.CastCatalogSpellErr(t, g, tc.name, "Instant", tc.oracle, effectstest.CardTarget(rock)); err == nil {
			t.Errorf("%s accepted a noncreature target", tc.name)
		}
	}
}

func TestOldB325PressTheAdvantagePumpsUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 3, 3)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Press the Advantage", "Instant", oldB325Press, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{a: {4, 4}, b: {5, 5}, c: {1, 1}} {
		if p, tou := effectstest.PtOf(t, g, id); p != want[0] || tou != want[1] {
			t.Errorf("creature is %d/%d, want %d/%d", p, tou, want[0], want[1])
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "trample") || !effectstest.HasEffectiveKeyword(t, g, b, "trample") {
		t.Error("both targets should have trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, c, "trample") {
		t.Error("the untargeted creature should not have trample")
	}
}

func TestOldB325PressTheAdvantageTakesOneOrNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Press the Advantage", "Instant", oldB325Press, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, a); p != 4 || tou != 4 {
		t.Errorf("the lone target is %d/%d, want 4/4", p, tou)
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Press the Advantage", "Instant", oldB325Press, effectstest.CardTargets(a, a, a)); err == nil {
		t.Error("three targets were accepted")
	}
}

func TestOldB325VolcanicRushPumpsOnlyAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	effectstest.ArnSetCombat(g, mine, opp.ID, uuid.Nil)
	effectstest.ArnSetCombat(g, theirs, me.ID, uuid.Nil)
	effectstest.CastCatalogSpell(t, g, "Volcanic Rush", "Instant", oldB325Volcanic, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if p, tou := effectstest.PtOf(t, g, id); p != 4 || tou != 2 {
			t.Errorf("attacker %d/%d, want 4/2", p, tou)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Error("an attacker has no trample")
		}
	}
	if p, tou := effectstest.PtOf(t, g, home); p != 2 || tou != 2 || effectstest.HasEffectiveKeyword(t, g, home, "trample") {
		t.Errorf("a creature at home is %d/%d, want an unchanged 2/2 without trample", p, tou)
	}
}

func TestOldB325StrongarmMonkPumpsTheTeamOnANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	monk := effectstest.B12Push(g, me.ID, "Strongarm Monk", "Creature — Human Monk", oldB325Monk, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.SettleProwess(t, g)
	if p, tou := effectstest.PtOf(t, g, monk); p != 3 || tou != 3 {
		t.Errorf("a creature spell pumped the Monk to %d/%d", p, tou)
	}

	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	for _, id := range []uuid.UUID{monk, bear} {
		if p, tou := effectstest.PtOf(t, g, id); (id == monk && (p != 4 || tou != 4)) || (id == bear && (p != 3 || tou != 3)) {
			t.Errorf("creature %d/%d after a noncreature spell", p, tou)
		}
	}
	if p, tou := effectstest.PtOf(t, g, theirs); p != 2 || tou != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tou)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tou := effectstest.PtOf(t, g, monk); p != 3 || tou != 3 {
		t.Errorf("after cleanup the Monk is %d/%d, want 3/3", p, tou)
	}
}

func TestOldB325GracebladeArtisanGrowsPerAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.B12Push(g, me.ID, "Graceblade Artisan", "Creature — Human Monk", oldB325Artisan, 2, 3)
	if p, tou := effectstest.PtOf(t, g, monk); p != 2 || tou != 3 {
		t.Fatalf("bare Artisan is %d/%d, want 2/3", p, tou)
	}
	for i := 1; i <= 2; i++ {
		aura := effectstest.B12Permanent(g, me.ID, "Aura", "Enchantment — Aura")
		effectstest.AttachForTest(t, g, aura, monk)
		if p, tou := effectstest.PtOf(t, g, monk); p != 2+2*i || tou != 3+2*i {
			t.Errorf("with %d Auras the Artisan is %d/%d, want %d/%d", i, p, tou, 2+2*i, 3+2*i)
		}
	}
}

func TestOldB325VirulentPlagueShrinksCreatureTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Virulent Plague", "Enchantment", oldB325Plague, 0, 0)
	mine := effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 3, Toughness: 3})
	theirs := effectstest.PushToken(g, opp.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 2, Toughness: 2})
	card := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	if p, tou := effectstest.PtOf(t, g, mine); p != 1 || tou != 1 {
		t.Errorf("my token is %d/%d, want 1/1", p, tou)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("their 2/2 token survived -2/-2")
	}
	if p, tou := effectstest.PtOf(t, g, card); p != 3 || tou != 3 {
		t.Errorf("a nontoken creature is %d/%d, want 3/3", p, tou)
	}
}

func TestOldB325ElusiveSpellfistGrowsAndSlipsPastBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fist := effectstest.B12Push(g, me.ID, "Elusive Spellfist", "Creature — Human Monk", oldB325Spellfist, 1, 3)
	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.SettleProwess(t, g)
	if p, _ := effectstest.PtOf(t, g, fist); p != 1 {
		t.Errorf("a creature spell pumped the Spellfist to %d power", p)
	}
	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if p, tou := effectstest.PtOf(t, g, fist); p != 2 || tou != 3 {
		t.Errorf("after a noncreature spell the Spellfist is %d/%d, want 2/3", p, tou)
	}
	blocker := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, fist)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, fist); err == nil {
		t.Error("the Spellfist could be blocked this turn")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, fist); p != 1 {
		t.Errorf("after cleanup the Spellfist has %d power, want 1", p)
	}
}

func TestOldB325SightOfTheScalelordsPumpsBigCreaturesAtBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Sight of the Scalelords", "Enchantment", oldB325Scalelords, 0, 0)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Wurm", 3, 4)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Wurm", 3, 5)
	effectstest.AdvanceToStepInTurn(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tou := effectstest.PtOf(t, g, big); p != 5 || tou != 6 || !effectstest.HasEffectiveKeyword(t, g, big, "vigilance") {
		t.Errorf("the 3/4 is %d/%d, want 5/6 with vigilance", p, tou)
	}
	if p, tou := effectstest.PtOf(t, g, small); p != 2 || tou != 3 || effectstest.HasEffectiveKeyword(t, g, small, "vigilance") {
		t.Errorf("the 2/3 is %d/%d, want it unchanged", p, tou)
	}
	if p, tou := effectstest.PtOf(t, g, theirs); p != 3 || tou != 5 {
		t.Errorf("an opponent's 3/5 is %d/%d, want it unchanged", p, tou)
	}
}
