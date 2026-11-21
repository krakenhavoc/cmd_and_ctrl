package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Orrery's whole point: its own colourless pays a three-colour
// spell.
func TestChromaticOrreryPaysAColoredSpellWithItsColorless(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	orrery := effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)
	if err := g.ActivateManaAbility(me.ID, orrery, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Orrery: %v", err)
	}
	spell := effectstest.HandSpell(me, "Three Colours", "Sorcery", "{U}{U}{B}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{U}{U}{B} off the Orrery's five colorless: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 {
		t.Errorf("pool = %v, want two colorless left", got)
	}
}

// An activated ability's coloured cost — Pestilence's {B} — is a cost
// its controller pays, so the Orrery reaches it.
func TestChromaticOrreryPaysAnActivatedAbilitysColoredCost(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)
	pest := effectstest.PushCatalogPermanent(g, me.ID, "Pestilence", "Enchantment", effectstest.PestilenceOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pest, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("colorless for Pestilence's {B}: %v", err)
	}
}

// The Orrery's draw, end to end through the activation: {5}, {T}, one
// card for each colour among its controller's permanents — red and
// green here, and the colourless Orrery adds none.
func TestChromaticOrreryDrawsOneCardPerColorAmongYourPermanents(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	orrery := effectstest.PushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)
	effectstest.B31Push(g, me.ID, "Goblin Guide", "Creature — Goblin Scout", "", "{R}", 2, 2, "R")
	effectstest.B31Push(g, me.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	effectstest.B31Push(g, opp.ID, "Opposing Blue", "Creature — Merfolk", "", "{U}", 1, 1, "U")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, orrery, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d → %d, want +2 (red and green; the opponent's blue is not yours)", hand, got)
	}
}

// "You may spend": an opponent's Orrery is no help to me.
func TestChromaticOrreryIsOnlyItsControllers(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "Chromatic Orrery", "Legendary Artifact", effectstest.ChromaticOrreryOracle, false)
	spell := effectstest.HandSpell(me, "Blue Spell", "Sorcery", "{U}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}), "red for {U} under an opponent's Orrery")
}

// "Players may": Mycosynth Lattice covers every player, so an opponent's
// Lattice lets me pay a {U} with red. Its colourless line reaches every
// zone since #3862, so the card is complete.
func TestMycosynthLatticeLetsEveryPlayerSpendManaAsAnyColor(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "Mycosynth Lattice", "Artifact", effectstest.MycosynthLatticeOracle, false)
	spell := effectstest.HandSpell(me, "Blue Spell", "Sorcery", "{U}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red for {U} under an opponent's Lattice: %v", err)
	}
	spec, _ := effects.Lookup(effectstest.MycosynthLatticeOracle)
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("the Lattice is complete since #3862; caveats %v", spec.Caveats)
	}
}

// Oath of Nissa, both halves: the entry dig offers only the creature,
// land and planeswalker cards of the three, and the spend grant
// reaches a planeswalker spell and not an instant.
func TestOathOfNissaDigsAndPaysForPlaneswalkers(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	// The top three, top-first: a land, an instant, a planeswalker.
	names := []struct{ Name, TypeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Shock", "Instant"},
		{"Garruk Relentless", "Legendary Planeswalker — Garruk"},
	}
	top := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: names[i].Name, TypeLine: names[i].TypeLine, Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Oath of Nissa", "Legendary Enchantment", effectstest.OathOfNissaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("Oath of Nissa's entry raised no pick")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[0]] || !offered[top[2]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the Forest and the planeswalker, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[2]) {
		t.Fatal("the planeswalker card is not in hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{top[1], top[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if me.Library.Cards[0].InstanceID != top[0] {
		t.Error("the rest did not go to the bottom in the chosen order")
	}

	walker := effectstest.HandSpell(me, "Blue Walker", "Legendary Planeswalker — Jace", "{U}{U}")
	instant := effectstest.HandSpell(me, "Blue Instant", "Instant", "{U}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}), "red for an instant's {U}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, walker, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("red and green for a planeswalker's {U}{U}: %v", err)
	}
}
