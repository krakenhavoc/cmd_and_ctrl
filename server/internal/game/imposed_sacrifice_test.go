package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// imposed_sacrifice_test.go — #6205: a sacrifice another permanent adds
// to spells' and activated abilities' costs (CR 601.2f, 602.2b). The
// rules, not the cards: Drought and Brutal Suppression prove the same
// from the catalog side when their batches build them.

const (
	droughtOracle    = "test-drought"
	suppressorOracle = "test-brutal-suppression"
)

// droughtStatics is Drought's two cost clauses as the engine reads them.
func droughtStatics() []ImposedSacrifice {
	swamp := SacrificeKind{Label: "a Swamp", Subtypes: []string{"Swamp"}}
	return []ImposedSacrifice{
		{Label: "Spells cost an additional \"Sacrifice a Swamp\" to cast for each black mana symbol in their mana costs.",
			Kind: swamp, Spells: true, PerBlackSymbol: true},
		{Label: "Activated abilities cost an additional \"Sacrifice a Swamp\" to activate for each black mana symbol in their activation costs.",
			Kind: swamp, Activations: true, PerBlackSymbol: true},
	}
}

// suppressionStatic is Brutal Suppression's clause.
func suppressionStatic() ImposedSacrifice {
	return ImposedSacrifice{
		Label:       "Activated abilities of nontoken Rebels cost an additional \"Sacrifice a land\" to activate.",
		Kind:        SacrificeKind{Label: "a land", CardTypes: []string{"land"}},
		Activations: true,
		AppliesTo: func(q ImposedSacrificeQuery) bool {
			return !q.Object.IsToken() && q.Object.HasSubtype("Rebel")
		},
	}
}

// imposedTable is a two-seat game at seat 0's main phase with the given
// imposing permanent on seat 1's side (the clauses reach every player).
func imposedTable(t *testing.T, oracle string, statics ...ImposedSacrifice) (*Game, *Player) {
	t.Helper()
	withCatalogDefs(t, map[string]*CardDef{oracle: {ImposedSacrifices: statics}})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	modifierSource(t, g, g.Seats[1], "Imposer", oracle)
	return g, g.Seats[0]
}

func addBlack(p *Player, n int) {
	for range n {
		p.ManaPool.AddMana(ManaToken{Color: "B"})
	}
}

// TestDroughtSpellNeedsASwampPerBlackSymbol — the acceptance test: under
// Drought a {B}{B} spell needs two Swamps sacrificed (CR 601.2f), and a
// payment short of that, or naming a land that is not a Swamp, is refused
// with nothing paid (CR 118.3).
func TestDroughtSpellNeedsASwampPerBlackSymbol(t *testing.T) {
	g, me := imposedTable(t, droughtOracle, droughtStatics()...)
	spell := spellInHand(t, g, me, "Twin Shadows", "Sorcery", "{B}{B}")
	s1 := pushTestLand(g, me, "Swamp", "Basic Land — Swamp")
	s2 := pushTestLand(g, me, "Swamp", "Basic Land — Swamp")
	plains := pushTestLand(g, me, "Plains", "Basic Land — Plains")
	addBlack(me, 2)

	for name, ids := range map[string][]uuid.UUID{
		"nothing":            nil,
		"one Swamp":          {s1},
		"a Swamp and Plains": {s1, plains},
		"the same Swamp":     {s1, s1},
	} {
		if err := g.CastSpell(me.ID, spell, CastSpellParams{SacrificeIDs: ids, Strict: true}); err == nil {
			t.Fatalf("%s: cast accepted, want a refusal", name)
		}
		if !onBattlefield(g, s1) || !onBattlefield(g, s2) || !onBattlefield(g, plains) {
			t.Fatalf("%s: a refused cast sacrificed something", name)
		}
	}
	if err := g.CastSpell(me.ID, spell, CastSpellParams{SacrificeIDs: []uuid.UUID{s2, s1}, Strict: true}); err != nil {
		t.Fatalf("two Swamps: %v", err)
	}
	if onBattlefield(g, s1) || onBattlefield(g, s2) {
		t.Error("the Swamps were not sacrificed")
	}
	if !onBattlefield(g, plains) {
		t.Error("the Plains nobody named was sacrificed")
	}
}

// TestDroughtCountsHybridAndLeavesOtherSpellsAlone — a hybrid symbol that
// offers black is a black mana symbol; a spell with none owes nothing.
func TestDroughtCountsHybridAndLeavesOtherSpellsAlone(t *testing.T) {
	if got := BlackManaSymbols("{2}{B/G}{B}{R}"); got != 2 {
		t.Errorf("BlackManaSymbols({2}{B/G}{B}{R}) = %d, want 2", got)
	}
	if got := BlackManaSymbols("{B/P}"); got != 1 {
		t.Errorf("BlackManaSymbols({B/P}) = %d, want 1", got)
	}
	g, me := imposedTable(t, droughtOracle, droughtStatics()...)
	spell := spellInHand(t, g, me, "Shock", "Instant", "{R}")
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	pushTestLand(g, me, "Swamp", "Basic Land — Swamp")
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a red spell under Drought: %v", err)
	}
}

// TestDroughtJoinsTheSpellsOwnSacrifice — a spell that already asks for a
// creature asks for the creature AND the Swamp, and one permanent pays one
// of them (CR 118.3): two Swamps are not a creature and a Swamp.
func TestDroughtJoinsTheSpellsOwnSacrifice(t *testing.T) {
	const splinters = "test-bone-splinters"
	g, me := imposedTable(t, droughtOracle, droughtStatics()...)
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == splinters {
			return &AdditionalCost{Label: "Sacrifice a creature", Sacrifice: &TargetSpec{
				Mode: "permanent", Label: "a creature", Min: 1, Max: 1,
				Zones:  []ZoneKind{ZoneBattlefield},
				CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool { return c.IsCreature() },
			}}
		}
		return nil
	})
	spell := catalogSpellInHand(t, g, me, "Bone Splinters", "Sorcery", "{B}", splinters)
	bear := pushColoredCreature(g, me, "Bear", "{1}{G}")
	s1 := pushTestLand(g, me, "Swamp", "Basic Land — Swamp")
	s2 := pushTestLand(g, me, "Swamp", "Basic Land — Swamp")
	addBlack(me, 1)

	for name, ids := range map[string][]uuid.UUID{
		"the creature alone": {bear},
		"two Swamps":         {s1, s2},
	} {
		if err := g.CastSpell(me.ID, spell, CastSpellParams{SacrificeIDs: ids, Strict: true}); err == nil {
			t.Fatalf("%s: cast accepted, want a refusal", name)
		}
	}
	if err := g.CastSpell(me.ID, spell, CastSpellParams{SacrificeIDs: []uuid.UUID{s1, bear}, Strict: true}); err != nil {
		t.Fatalf("a Swamp and the creature: %v", err)
	}
	if onBattlefield(g, bear) || onBattlefield(g, s1) || !onBattlefield(g, s2) {
		t.Error("the payment did not sacrifice exactly the creature and the named Swamp")
	}
}

// pushAbilityCreature puts a creature of `typeLine` with one "{cost}:
// no-op" activated ability under `owner`.
func pushAbilityCreature(g *Game, owner *Player, typeLine, mana string, token bool) uuid.UUID {
	c := NewCard("Probe", owner.ID)
	c.TypeLine = typeLine
	c.Controller = owner.ID
	if token {
		c.TypeLine = "Token " + typeLine
	}
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label:  mana + ": Probe.",
		Cost:   AbilityCost{Mana: mana},
		Effect: func(*Game, *StackItem) error { return nil },
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// TestBrutalSuppressionTaxesNontokenRebelsWithALand — the acceptance
// test: a nontoken Rebel's ability needs a land sacrificed (CR 602.2b);
// a Rebel token's and a non-Rebel's do not.
func TestBrutalSuppressionTaxesNontokenRebelsWithALand(t *testing.T) {
	g, me := imposedTable(t, suppressorOracle, suppressionStatic())
	rebel := pushAbilityCreature(g, me, "Creature — Human Rebel", "{1}", false)
	token := pushAbilityCreature(g, me, "Creature — Human Rebel", "{1}", true)
	soldier := pushAbilityCreature(g, me, "Creature — Human Soldier", "{1}", false)
	land := pushTestLand(g, me, "Plains", "Basic Land — Plains")
	addMana(me, 3)

	if err := g.ActivateCatalogAbility(me.ID, rebel, 0, ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Rebel's ability was activated without a land")
	}
	if err := g.ActivateCatalogAbility(me.ID, rebel, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{soldier}, Strict: true}); !errors.Is(err, ErrIllegalTarget) {
		t.Fatalf("a creature for \"Sacrifice a land\" = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, rebel, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}, Strict: true}); err != nil {
		t.Fatalf("the Rebel's ability with a land: %v", err)
	}
	if onBattlefield(g, land) {
		t.Error("the land was not sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, token, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("a Rebel token's ability: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, soldier, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("a Soldier's ability: %v", err)
	}
}

// TestDroughtTaxesActivationsPerBlackSymbol — "for each black mana symbol
// in their activation costs": {B} costs one Swamp, {1} none.
func TestDroughtTaxesActivationsPerBlackSymbol(t *testing.T) {
	g, me := imposedTable(t, droughtOracle, droughtStatics()...)
	black := pushAbilityCreature(g, me, "Creature — Zombie", "{B}", false)
	plain := pushAbilityCreature(g, me, "Creature — Zombie", "{1}", false)
	swamp := pushTestLand(g, me, "Swamp", "Basic Land — Swamp")
	addBlack(me, 2)

	if err := g.ActivateCatalogAbility(me.ID, black, 0, ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a {B} ability was activated without a Swamp")
	}
	if err := g.ActivateCatalogAbility(me.ID, black, 0, ActivateAbilityParams{SacrificeIDs: []uuid.UUID{swamp}, Strict: true}); err != nil {
		t.Fatalf("a {B} ability with a Swamp: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, plain, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("a {1} ability under Drought: %v", err)
	}
}

// TestImposedSacrificeGoneWithItsSource — the cost is a static of a
// permanent on the battlefield: with Drought gone, the {B}{B} spell is
// paid with mana alone.
func TestImposedSacrificeGoneWithItsSource(t *testing.T) {
	withCatalogDefs(t, map[string]*CardDef{droughtOracle: {ImposedSacrifices: droughtStatics()}})
	g := newActiveGame(t)
	me := g.Seats[0]
	spell := spellInHand(t, g, me, "Twin Shadows", "Sorcery", "{B}{B}")
	addBlack(me, 2)
	if err := g.CastSpell(me.ID, spell, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("no Drought on the battlefield: %v", err)
	}
}
