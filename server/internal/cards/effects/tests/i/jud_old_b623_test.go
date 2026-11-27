package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jud_old_b623_test.go — S82 OLD batch 623 (#5611): Judgment, damage. Each test
// plays the real catalog card through the engine.

const (
	oldB623Jeska      = "3186fddd-23fd-440c-ad61-b4130e00f765"
	oldB623Firebeast  = "853ce64f-50b4-49a6-9aa3-f07b11cd8b69"
	oldB623Scalpelex  = "99eb50ef-352f-47c4-91e3-32813cbe0649"
	oldB623Constable  = "a5236bf7-c84f-4c64-bf6b-3a89c05e161a"
	oldB623Scorcher   = "a896cedb-f10a-48ac-82e7-6ae1e8b5a439"
	oldB623Swelter    = "e1169108-1bc8-42ca-a056-a9358fba10cb"
	oldB623AncChosen  = "fc2ccab7-cab1-4463-b73d-898070136d74"
	oldB623Typeline   = "Creature — Test"
	oldB623LibraryCap = 8
)

func TestOldB623EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB623Jeska: "Jeska, Warrior Adept", oldB623Firebeast: "Goretusk Firebeast", oldB623Scalpelex: "Scalpelexis",
		oldB623Constable: "Cephalid Constable", oldB623Scorcher: "Dwarven Scorcher", oldB623Swelter: "Swelter",
		oldB623AncChosen: "Ancestor's Chosen",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB623JeskaHasFirstStrikeAndHasteAndPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	jeska := effectstest.BlbB1Push(g, me.ID, "Jeska, Warrior Adept", "Legendary Creature — Human Barbarian Warrior", oldB623Jeska, 3, 1)
	for _, kw := range []string{"first strike", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, jeska, kw) {
			t.Errorf("no %s", kw)
		}
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, jeska, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("Jeska dealt %d, want 1", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, jeska, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Error("a tapped Jeska activated again")
	}
}

func TestOldB623GoretuskFirebeastDealsFourToAPlayerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	life := opp.Life
	effectstest.B12PlayFromHand(t, g, "Goretusk Firebeast", "Creature — Elemental Boar Beast", oldB623Firebeast, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, bear) {
			t.Error("a creature was offered as the target")
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 4 {
		t.Errorf("the Firebeast dealt %d, want 4", got)
	}
}

func TestOldB623AncestorsChosenGainsLifePerCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbMPFillGraveyard(g, me, 5)
	effectstest.BlbMPFillGraveyard(g, opp, 3)
	life := me.Life
	effectstest.B12PlayFromHand(t, g, "Ancestor's Chosen", "Creature — Human Cleric", oldB623AncChosen, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 5 {
		t.Errorf("gained %d life, want 5", got)
	}
}

func TestOldB623SwelterDealsTwoToEachOfExactlyTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, _ := effectstest.BlbB1Seats(g)
	a := effectstest.BlbB1Push(g, opp.ID, "A", "Creature — Bear", "", 3, 3)
	b := effectstest.BlbB1Push(g, opp.ID, "B", "Creature — Bear", "", 3, 3)
	c := effectstest.BlbB1Push(g, opp.ID, "C", "Creature — Bear", "", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Swelter", "Sorcery", oldB623Swelter, effectstest.CardTarget(a)); err == nil {
		t.Fatal("one target was accepted")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Swelter", "Sorcery", oldB623Swelter, append(effectstest.CardTarget(a), append(effectstest.CardTarget(b), effectstest.CardTarget(c)...)...)); err == nil {
		t.Fatal("three targets were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Swelter", "Sorcery", oldB623Swelter, append(effectstest.CardTarget(a), effectstest.CardTarget(b)...))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, a) != 2 || effectstest.DamageMarkedOn(g, b) != 2 || effectstest.DamageMarkedOn(g, c) != 0 {
		t.Errorf("damage: A=%d B=%d C=%d, want 2 2 0", effectstest.DamageMarkedOn(g, a), effectstest.DamageMarkedOn(g, b), effectstest.DamageMarkedOn(g, c))
	}
}

func TestOldB623DwarvenScorcherDealsOneUnlessTheControllerTakesTwo(t *testing.T) {
	for _, take := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := effectstest.BlbB1Seats(g)
		sc := effectstest.BlbB1Push(g, me.ID, "Dwarven Scorcher", "Creature — Dwarf", oldB623Scorcher, 1, 1)
		bear := effectstest.BlbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 3, 3)
		life := opp.Life
		effectstest.B16Activate(t, g, me.ID, sc, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(sc) {
			t.Fatal("the Scorcher was not sacrificed as a cost")
		}
		if effectstest.AnswerPayUnless(t, g, opp.ID, take) == nil {
			t.Fatal("the creature's controller was not asked")
		}
		effectstest.PassPriorityAroundTable(t, g)
		gotLife := life - opp.Life
		gotDamage := effectstest.DamageMarkedOn(g, bear)
		if take && (gotLife != 2 || gotDamage != 0) {
			t.Errorf("took the damage: life -%d, creature damage %d, want -2 and 0", gotLife, gotDamage)
		}
		if !take && (gotLife != 0 || gotDamage != 1) {
			t.Errorf("declined: life -%d, creature damage %d, want 0 and 1", gotLife, gotDamage)
		}
	}
}

func oldB623Attack(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	effectstest.BrAttack(t, g, id)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
}

func TestOldB623CephalidConstableReturnsUpToAsManyPermanentsAsDamageDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	cc := effectstest.B12Push(g, me.ID, "Cephalid Constable", "Creature — Octopus Wizard", oldB623Constable, 2, 1)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 0, 9)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 0, 9)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 0, 9)
	elsewhere := effectstest.B12Creature(g, other.ID, "Elsewhere", "Creature — Bear", 0, 9)
	oldB623Attack(t, g, cc)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, elsewhere) || !effectstest.HasID(p.PickTargetCards, a) {
		t.Fatalf("offered %v: only the damaged player's permanents", p.PickTargetCards)
	}
	three := []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c},
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, three); err == nil {
		t.Fatal("three targets were accepted for two damage")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, three[:2]); err != nil {
		t.Fatalf("two targets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if g.Battlefield.Contains(id) {
			t.Error("a target is still on the battlefield")
		}
	}
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(elsewhere) {
		t.Error("an untargeted permanent was returned")
	}
	if opp.Hand.Size() < 2 {
		t.Errorf("the opponent's hand holds %d, want the two creatures", opp.Hand.Size())
	}
}

func oldB623Library(p *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestOldB623ScalpelexisRepeatsWhileTwoCardsShareAName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	sc := effectstest.B12Push(g, me.ID, "Scalpelexis", "Creature — Beast", oldB623Scalpelex, 1, 5)
	lib := opp.Library.Size()
	// The last card pushed is on top. The top four (two Alphas) repeat the
	// process; the next four (Z1, S4, S3, S2) share nothing, so it stops and S1,
	// the ninth card, stays.
	bottom := oldB623Library(opp, "S1", "S2", "S3", "S4", "Z1")
	oldB623Library(opp, "Alpha", "Beta", "Alpha", "Gamma")
	oldB623Attack(t, g, sc)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib+1 {
		t.Errorf("library %d, want %d (nine pushed, eight exiled)", got, lib+1)
	}
	if !opp.Library.Contains(bottom[0]) {
		t.Error("the ninth card was exiled")
	}
	if effectstest.InExile(g, bottom[1]) == false {
		t.Error("the second batch was not exiled")
	}
}

func TestOldB623ScalpelexisStopsWhenNoNamesMatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	sc := effectstest.B12Push(g, me.ID, "Scalpelexis", "Creature — Beast", oldB623Scalpelex, 1, 5)
	lib := opp.Library.Size()
	oldB623Library(opp, "Dup", "Dup", "X1", "X2") // bottom: not reached
	oldB623Library(opp, "Alpha", "Beta", "Gamma", "Delta")
	oldB623Attack(t, g, sc)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib+8-4 {
		t.Errorf("library %d, want %d (four exiled, no repeat)", got, lib+8-4)
	}
}
