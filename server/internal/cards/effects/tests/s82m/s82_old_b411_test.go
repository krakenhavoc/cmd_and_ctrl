package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b411_test.go — Gatecrash, batch 411 (#5565): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB411Unexpected  = "048bdc4a-4d26-420c-88ed-23a073c30fff"
	oldB411Gridlock    = "049e9963-1b74-476f-9a81-dbfae543e1c5"
	oldB411Bioshift    = "3da604fc-68e9-4749-98f9-6fbcebcab9b6"
	oldB411KnightWatch = "4965f5fe-a5fe-4976-822c-5cf40fea8c34"
	oldB411Slime       = "50e7f705-c347-440b-8159-81a52bbfbdfb"
	oldB411Mark        = "ad2e4ab0-d990-472d-8788-3f85215da51c"
	oldB411Hydroform   = "d8b6ea82-da4e-4763-84b1-4caa214ed945"
)

func TestOldB411CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB411Unexpected:  "Unexpected Results",
		oldB411Gridlock:    "Gridlock",
		oldB411Bioshift:    "Bioshift",
		oldB411KnightWatch: "Knight Watch",
		oldB411Slime:       "Miming Slime",
		oldB411Mark:        "Mark for Death",
		oldB411Hydroform:   "Hydroform",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Unexpected Results ---

// oldB411CastUnexpected empties the library, seeds `top`, casts the spell and
// resolves it up to its first question.
func oldB411CastUnexpected(t *testing.T, top game.Card) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	topID := effectstest.SeedLibraryTop(me, top.Name, top.TypeLine, top.ManaCost)
	spell := effectstest.CastCatalogSpell(t, g, "Unexpected Results", "Sorcery", oldB411Unexpected, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, spell, topID
}

func TestOldB411UnexpectedResultsCastsANonlandCardFree(t *testing.T) {
	g, me, spell, top := oldB411CastUnexpected(t, game.Card{Name: "Big Spell", TypeLine: "Creature — Giant", ManaCost: "{6}{G}"})
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the revealed card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneBattlefield {
		t.Errorf("the cast creature is in %v, want the battlefield", effectstest.B12ZoneOf(g, top))
	}
	if !me.Graveyard.Contains(spell) {
		t.Errorf("Unexpected Results is in %v, want the graveyard", effectstest.B12ZoneOf(g, spell))
	}
}

func TestOldB411UnexpectedResultsDeclinedNonlandStaysOnTop(t *testing.T) {
	g, me, _, top := oldB411CastUnexpected(t, game.Card{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{6}{G}"})
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, top) != game.ZoneLibrary {
		t.Errorf("the declined card is in %v, want the library", effectstest.B12ZoneOf(g, top))
	}
}

func TestOldB411UnexpectedResultsPutsALandAndReturnsToHand(t *testing.T) {
	g, me, spell, land := oldB411CastUnexpected(t, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("the land was not offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(land) {
		t.Errorf("the land is in %v, want the battlefield", effectstest.B12ZoneOf(g, land))
	}
	if !me.Hand.Contains(spell) {
		t.Errorf("Unexpected Results is in %v, want its owner's hand", effectstest.B12ZoneOf(g, spell))
	}
}

func TestOldB411UnexpectedResultsDeclinedLandKeepsTheSpellInTheGraveyard(t *testing.T) {
	g, me, spell, land := oldB411CastUnexpected(t, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("a declined land was put onto the battlefield")
	}
	if !me.Graveyard.Contains(spell) {
		t.Errorf("Unexpected Results is in %v, want the graveyard", effectstest.B12ZoneOf(g, spell))
	}
}

// --- Gridlock ---

func TestOldB411GridlockTapsExactlyXNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Permanent(g, opp.ID, "B", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	spare := effectstest.B12Creature(g, opp.ID, "Spare", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Gridlock", "Instant", oldB411Gridlock, effectstest.CardRefs(land)); err == nil {
		t.Fatal("targeted a land")
	}
	effectstest.D1658Cast(t, g, "Gridlock", "Instant", oldB411Gridlock, game.CastSpellParams{XValue: 2, Targets: effectstest.CardRefs(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.B16Tapped(t, g, id) {
			t.Errorf("%s was not tapped", id)
		}
	}
	for _, id := range []uuid.UUID{land, spare} {
		if effectstest.B16Tapped(t, g, id) {
			t.Errorf("%s was tapped", id)
		}
	}
}

// --- Bioshift ---

func TestOldB411BioshiftMovesTheChosenNumberOfCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	from := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "From", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Counters: map[string]int{game.CounterPlusOne: 3}})
	to := effectstest.B12Creature(g, me.ID, "To", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Bioshift", "Instant", oldB411Bioshift, effectstest.CardRefs(from, theirs)); err == nil {
		t.Fatal("moved counters to a creature with another controller")
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Bioshift", "Instant", oldB411Bioshift, effectstest.CardRefs(from, from)); err == nil {
		t.Fatal("moved counters onto the same creature")
	}
	effectstest.CastCatalogSpell(t, g, "Bioshift", "Instant", oldB411Bioshift, effectstest.CardRefs(from, to))
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.PayAmountPrompt(g, me.ID)
	if p == nil {
		t.Fatal("no number prompt")
	}
	if err := g.ResolvePayAmount(p.ID, me.ID, 2); err != nil {
		t.Fatalf("answering: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, from, game.CounterPlusOne); n != 1 {
		t.Errorf("From has %d counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, to, game.CounterPlusOne); n != 2 {
		t.Errorf("To has %d counters, want 2", n)
	}
}

// --- Knight Watch ---

func TestOldB411KnightWatchMakesTwoVigilantKnights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Knight Watch", "Sorcery", oldB411KnightWatch, nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Knight" && c.Controller == me.ID {
			n++
			if c.Power != 2 || c.Toughness != 2 || !effectstest.HasKeywordOnBattlefield(t, g, c.InstanceID, "vigilance") {
				t.Errorf("Knight is %d/%d without vigilance", c.Power, c.Toughness)
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Knights, want 2", n)
	}
}

// --- Miming Slime ---

func TestOldB411MimingSlimeIsAsBigAsYourBiggestCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 5, 1)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Giant", 9, 9)
	effectstest.CastCatalogSpell(t, g, "Miming Slime", "Sorcery", oldB411Slime, nil)
	effectstest.PassPriorityAroundTable(t, g)
	found := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Ooze" && c.Controller == me.ID {
			found = true
			if p, tg := effectstest.EffectivePTOf(t, g, c.InstanceID); p != 5 || tg != 5 {
				t.Errorf("Ooze is %d/%d, want 5/5", p, tg)
			}
		}
	}
	if !found {
		t.Fatal("no Ooze token")
	}
}

// --- Mark for Death ---

func TestOldB411MarkForDeathForcesOneBlockerAndBarsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, third.ID, "Bystander", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, target).Tapped = true })
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mark for Death", "Sorcery", oldB411Mark, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("targeted a creature the caster controls")
	}
	effectstest.CastCatalogSpell(t, g, "Mark for Death", "Sorcery", oldB411Mark, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B16Tapped(t, g, target) {
		t.Error("the target was not untapped")
	}
	late := effectstest.B12Creature(g, opp.ID, "Late", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for name, c := range map[string]struct {
		id   uuid.UUID
		bars bool
	}{
		"the target":                 {target, false},
		"another creature of theirs": {other, true},
		"one that arrived later":     {late, true},
		"another player's creature":  {bystander, false},
	} {
		if got := effectstest.EffectiveOf(t, g, c.id).Restrictions.Has(game.CantBlock); got != c.bars {
			t.Errorf("%s: can't block = %v, want %v", name, got, c.bars)
		}
	}
}

// --- Hydroform ---

func TestOldB411HydroformMakesALandAThreeThreeFlyingElementalUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Hydroform", "Instant", oldB411Hydroform, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("targeted a nonland permanent")
	}
	effectstest.CastCatalogSpell(t, g, "Hydroform", "Instant", oldB411Hydroform, effectstest.CardRefs(land))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePTOf(t, g, land); p != 3 || tg != 3 {
		t.Errorf("land is %d/%d, want 3/3", p, tg)
	}
	types := effectstest.EffectiveTypes(t, g, land)
	hasLand, hasCreature := false, false
	for _, ty := range types {
		hasLand = hasLand || ty == "land" || ty == "Land"
		hasCreature = hasCreature || ty == "creature" || ty == "Creature"
	}
	if !hasLand || !hasCreature {
		t.Errorf("types %v, want a land that is also a creature", types)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, land, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, land, "flying") {
		t.Error("flying lasted past the turn")
	}
}
