package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB021Dakkon   = "1aa78e3b-bf7e-4645-9521-4b3537c93f46"
	oldB021Dourbark = "3fd25479-b65c-48f7-bb74-95d49af65ac1"
	oldB021Horned   = "64553181-9852-459f-a6e0-54ce188dd937"
	oldB021Ruhan    = "82d109ea-3421-46eb-9bbf-f7f28b548ea6"
	oldB021Sire     = "b55d8cd6-037d-4d7c-90ec-51249d34b02e"
	oldB021Hokori   = "d2aaf485-8c74-4a27-b80b-f600c20d8b9c"
	oldB021Maro     = "ef32258a-d567-4fbc-ab85-8c00080f5b36"
)

func TestOldB021CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB021Dakkon:   "Dakkon Blackblade",
		oldB021Dourbark: "Dauntless Dourbark",
		oldB021Horned:   "Horned Sliver",
		oldB021Ruhan:    "Ruhan of the Fomori",
		oldB021Sire:     "Sire of Insanity",
		oldB021Hokori:   "Hokori, Dust Drinker",
		oldB021Maro:     "Maro",
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

func oldB021Size(g *game.Game, id uuid.UUID) (int, int) {
	c := effectstest.ApaLive(g, id)
	return c.PowerForComparison(), c.CurrentToughness()
}

func TestOldB021DakkonCountsOnlyYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.B12Push(g, me.ID, "Dakkon Blackblade", "Legendary Creature — Human Warrior", oldB021Dakkon, 0, 0)
	effectstest.PushLands(g, me.ID, 3)
	effectstest.PushLands(g, opp.ID, 2)
	if p, tough := oldB021Size(g, d); p != 3 || tough != 3 {
		t.Fatalf("Dakkon is %d/%d with three lands, want 3/3", p, tough)
	}
	effectstest.PushLands(g, me.ID, 2)
	if p, tough := oldB021Size(g, d); p != 5 || tough != 5 {
		t.Fatalf("Dakkon is %d/%d with five lands, want 5/5", p, tough)
	}
}

func TestOldB021MaroIsAsBigAsYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	m := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Maro", TypeLine: "Creature — Elemental", OracleID: oldB021Maro,
		Owner: me.ID, Controller: me.ID,
	})
	size := me.Hand.Size()
	if p, tough := oldB021Size(g, m); p != size || tough != size {
		t.Fatalf("Maro is %d/%d with %d cards in hand, want %d/%d", p, tough, size, size, size)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if p, tough := oldB021Size(g, m); p != size+1 || tough != size+1 {
		t.Fatalf("Maro is %d/%d after a draw, want %d/%d", p, tough, size+1, size+1)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("DrawCard (opponent): %v", err)
	}
	if p, _ := oldB021Size(g, m); p != size+1 {
		t.Fatalf("Maro has power %d after an opponent drew, want %d: only your hand counts", p, size+1)
	}
}

func TestOldB021DourbarkCountsForestsAndTreefolkAndTramplesWithAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := effectstest.B12Push(g, me.ID, "Dauntless Dourbark", "Creature — Treefolk Warrior", oldB021Dourbark, 0, 0)
	// Itself is a Treefolk: 1.
	if p, tough := oldB021Size(g, d); p != 1 || tough != 1 {
		t.Fatalf("alone, Dourbark is %d/%d, want 1/1", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, d, "trample") {
		t.Fatal("alone, Dourbark has trample")
	}
	effectstest.PushLands(g, me.ID, 2) // two Forests
	effectstest.B12Creature(g, opp.ID, "Opposing Ent", "Creature — Treefolk", 2, 2)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Dryad Arbor", TypeLine: "Land Creature — Forest Dryad", Power: 1, Toughness: 1})
	if p, _ := oldB021Size(g, d); p != 4 {
		t.Fatalf("with two Forests, a Forest Dryad and itself, Dourbark has power %d, want 4 (an opponent's Treefolk does not count)", p)
	}
	if effectstest.HasEffectiveKeyword(t, g, d, "trample") {
		t.Fatal("an opponent's Treefolk gave Dourbark trample")
	}
	ent := effectstest.B12Creature(g, me.ID, "Ent", "Creature — Treefolk", 3, 3)
	if p, _ := oldB021Size(g, d); p != 5 {
		t.Fatalf("with another Treefolk, Dourbark has power %d, want 5", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, d, "trample") {
		t.Fatal("with another Treefolk Dourbark has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, ent, "trample") {
		t.Fatal("the other Treefolk got trample too")
	}
}

func TestOldB021HornedSliverGivesEverySliverTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := effectstest.B12Push(g, me.ID, "Horned Sliver", "Creature — Sliver", oldB021Horned, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Sliver", "Creature — Sliver", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Sliver", "Creature — Sliver", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{h, mine, theirs} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Errorf("Sliver %s has no trample", id)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a non-Sliver has trample")
	}
}

func TestOldB021SireOfInsanityEmptiesEveryHandAtEachEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Sire of Insanity", "Creature — Demon", oldB021Sire, 6, 4)
	for _, p := range g.Seats {
		for i := 0; i < 2; i++ {
			g.WithWriteLock(func() {
				p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
			})
		}
	}
	effectstest.AdvanceToEndStepOf(t, g, active.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, opp, g.Seats[2], g.Seats[3]} {
		if n := p.Hand.Size(); n != 0 {
			t.Errorf("seat %d still holds %d cards after the end step", p.Seat, n)
		}
	}
	if opp.Graveyard.Size() < 2 {
		t.Errorf("the opponent's discarded cards are not in their graveyard")
	}
}

func TestOldB021HokoriKeepsLandsTappedButUntapsOneAtEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Hokori, Dust Drinker", "Legendary Creature — Spirit", oldB021Hokori, 2, 2)
	var mine []uuid.UUID
	for i := 0; i < 2; i++ {
		mine = append(mine, effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest"))
	}
	for _, id := range mine {
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Tapped = true })
	}
	creature := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, creature).Tapped = true })
	effectstest.AdvanceToUpkeepOf(t, g, opp.Seat)
	if !effectstest.IsTapped(g, mine[0]) || !effectstest.IsTapped(g, mine[1]) {
		t.Fatal("a land untapped during its controller's untap step")
	}
	if effectstest.IsTapped(g, creature) {
		t.Fatal("a creature stayed tapped: Hokori only holds lands")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, mine[1])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, mine[1]) {
		t.Error("the chosen land is still tapped")
	}
	if !effectstest.IsTapped(g, mine[0]) {
		t.Error("the land that was not chosen untapped")
	}
}

func TestOldB021RuhanMustAttackARandomOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	r := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Ruhan of the Fomori", OracleID: oldB021Ruhan,
		TypeLine: "Legendary Creature — Giant Warrior", Power: 7, Toughness: 7})
	effectstest.AdvanceDecliningTriggerPrompts(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	var target uuid.UUID
	for _, rec := range g.ScopedEffects {
		for _, m := range rec.Mods {
			if m.Kind == game.ModAttackPlayerIfAble && m.Player != uuid.Nil {
				target = m.Player
			}
		}
	}
	if target == uuid.Nil || target == me.ID {
		t.Fatalf("no opponent was chosen (got %s)", target)
	}
	for _, opp := range g.Seats[1:] {
		err := effectstest.DeclareResult(g, r, opp.ID)
		if opp.ID == target && err != nil {
			t.Fatalf("attacking the chosen player: %v", err)
		}
		if opp.ID != target && err == nil {
			t.Fatalf("attacking %s was allowed while Ruhan must attack %s", opp.ID, target)
		}
	}
}

const oldB021FourteenthDoctor = "0afcf1eb-ac2c-4bb0-8823-9fe45968c2d9"

// The cast trigger reveals fourteen, puts the Doctors among them into the
// graveyard and the rest under the library; the spell then offers only those
// Doctors (not one that was already in the graveyard, nor one past the
// fourteenth card) and the copy has haste only until end of turn.
func TestOldB021FourteenthDoctorRevealsThenCopiesADoctorWithTemporaryHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	var inTop, beyond []uuid.UUID
	// Pushed bottom first: index 19 is the top, so indices 6..19 are revealed.
	for i := 0; i < 20; i++ {
		c := game.Card{InstanceID: uuid.New(), Name: "Plain Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		if i == 2 || i == 10 || i == 15 {
			c.Name, c.TypeLine, c.Power, c.Toughness = "Test Doctor", "Legendary Creature — Time Lord Doctor", 2, 3
			if i == 2 {
				beyond = append(beyond, c.InstanceID)
			} else {
				inTop = append(inTop, c.InstanceID)
			}
		}
		g.WithWriteLock(func() { me.Library.PushTop(c) })
	}
	older := effectstest.GraveyardCreature(me, "Old Doctor", "Time Lord Doctor", 1, 1)

	doc := effectstest.CastGraveyardCopier(t, g, "The Fourteenth Doctor", "Legendary Creature — Time Lord Doctor", oldB021FourteenthDoctor, 3, 4)
	prompt := effectstest.WaitForCopyPrompt(t, g)
	for _, id := range inTop {
		if !me.Graveyard.Contains(id) {
			t.Errorf("a revealed Doctor %s was not put into the graveyard", id)
		}
		if !slices.Contains(prompt.CopyOptions, id) {
			t.Errorf("a Doctor put there from the library %s is not offered: %v", id, prompt.CopyOptions)
		}
	}
	if slices.Contains(prompt.CopyOptions, older) {
		t.Error("a Doctor that was already in the graveyard is offered")
	}
	if slices.Contains(prompt.CopyOptions, beyond[0]) || me.Graveyard.Contains(beyond[0]) {
		t.Error("the Doctor past the fourteenth card was revealed")
	}
	if n := me.Library.Size(); n != 18 {
		t.Errorf("library holds %d cards, want 18: twenty minus two Doctors", n)
	}
	if err := g.ResolveCopyTarget(prompt.ID, prompt.Chooser, inTop[0]); err != nil {
		t.Fatalf("ResolveCopyTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.CopyBattlefieldCard(t, g, doc)
	if got.Name != "Test Doctor" || got.CurrentPower() != 2 || got.CurrentToughness() != 3 {
		t.Fatalf("entered as %q %d/%d, want a 2/3 Test Doctor", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g, doc, "haste") {
		t.Error("the copy has no haste the turn it entered")
	}
	effectstest.AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%4)
	if effectstest.HasEffectiveKeyword(t, g, doc, "haste") {
		t.Error("the copy kept haste past the end of the turn")
	}
}
