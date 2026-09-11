package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// cost_modifier_test.go — S28. The four things a bug here would hide:
//
//  1. ORDER. Increases before reductions is the whole rule. A
//     version that ran reductions first passes "the spell got
//     cheaper" and fails Sphere of Resistance + Goblin Electromancer
//     on a one-mana spell, which is the canonical interaction.
//  2. THE FLOOR. A reduction spends generic mana and stops at zero.
//     It must never eat a {B}, and an overshoot must not carry.
//  3. TRINISPHERE LAST. "Would cost less than three" has to mean the
//     price after everything else, and the shortfall has to be
//     generic so a {B} stays a {B}.
//  4. A BROKEN MODIFIER REFUSES THE CAST. #289's lesson: the failure
//     mode that matters is a spell that comes out free.

func withCatalogCostModifiers(t *testing.T, fn func(oracleID string) []CostModifier) {
	t.Helper()
	prev := CatalogCostModifiers
	CatalogCostModifiers = fn
	t.Cleanup(func() { CatalogCostModifiers = prev })
}

// modifiersFor wires one card's modifiers and returns the hook body.
func modifiersFor(oracle string, mods ...CostModifier) func(string) []CostModifier {
	return func(id string) []CostModifier {
		if id == oracle {
			return mods
		}
		return nil
	}
}

// spellInHand seeds a spell with the given printed cost in the
// seat's hand at a main phase and returns its instance ID.
func spellInHand(t *testing.T, g *Game, me *Player, name, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard(name, me.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	me.Hand.PushTop(c)
	return c.InstanceID
}

// modifierSource drops a permanent carrying `oracle` onto the
// battlefield under `owner`'s control and returns it.
func modifierSource(t *testing.T, g *Game, owner *Player, name, oracle string) uuid.UUID {
	t.Helper()
	c := NewCard(name, owner.ID)
	c.TypeLine = "Artifact"
	c.OracleID = oracle
	c.Controller = owner.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// fixed is the amount hook for a flat modifier.
func fixed(n int) func(CostQuery) int {
	return func(CostQuery) int { return n }
}

// priceOf runs the cast-cost computation the way CastSpell does,
// without casting anything.
func priceOf(t *testing.T, g *Game, p *Player, cardID uuid.UUID, params CastSpellParams) ParsedCost {
	t.Helper()
	card, ok := g.LookupCardForEffect(cardID)
	if !ok {
		t.Fatalf("card %s not found", cardID)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	cost, err := g.effectiveCostLocked(p, card, params)
	if err != nil {
		t.Fatalf("effectiveCostLocked: %v", err)
	}
	return cost
}

// The canonical interaction, and the reason the engine makes two
// passes: Sphere of Resistance makes a {R} Bolt cost {1}{R}, and
// Goblin Electromancer then has a generic symbol to eat. Run the
// reduction first and the Bolt stays at {1}{R}.
func TestCostModifierIncreasesApplyBeforeReductions(t *testing.T) {
	const sphere, mancer = "test-sphere", "test-mancer"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, func(id string) []CostModifier {
		switch id {
		case sphere:
			return []CostModifier{{Kind: CostIncrease, Label: "Spells cost {1} more", Amount: fixed(1)}}
		case mancer:
			return []CostModifier{{Kind: CostReduction, Label: "Instants cost {1} less", Amount: fixed(1)}}
		}
		return nil
	})
	bolt := spellInHand(t, g, me, "Test Bolt", "Instant", "{R}")

	// Neither out: the printed cost stands.
	if got := priceOf(t, g, me, bolt, CastSpellParams{}); got.Generic != 0 || len(got.Required) != 1 {
		t.Fatalf("unmodified cost = %+v, want {R}", got)
	}

	modifierSource(t, g, me, "Sphere of Resistance", sphere)
	if got := priceOf(t, g, me, bolt, CastSpellParams{}); got.Generic != 1 {
		t.Errorf("with the sphere: generic = %d, want 1", got.Generic)
	}

	modifierSource(t, g, me, "Goblin Electromancer", mancer)
	got := priceOf(t, g, me, bolt, CastSpellParams{})
	if got.Generic != 0 {
		t.Errorf("sphere + electromancer: generic = %d, want 0 (the increase gave the reduction something to eat)", got.Generic)
	}
	if len(got.Required) != 1 || got.Required[0].Options[0] != "R" {
		t.Errorf("sphere + electromancer: required = %+v, want one {R}", got.Required)
	}
}

// CR 601.2f's floor: a reduction spends GENERIC mana, stops at zero,
// and never touches a coloured requirement. The overshoot must not
// carry into the {B} and must not carry into a second reduction.
func TestCostReductionNeverEatsColoredMana(t *testing.T) {
	const summoning = "test-summoning"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(summoning, CostModifier{
		Kind:   CostReduction,
		Label:  "Creature spells you cast cost {2} less to cast.",
		Amount: fixed(2),
	}))
	modifierSource(t, g, me, "Heartless Summoning", summoning)

	// {B} — one coloured symbol, no generic. The reduction has
	// nothing to spend and the card still costs {B}.
	feeder := spellInHand(t, g, me, "Carrion Feeder", "Creature — Zombie", "{B}")
	got := priceOf(t, g, me, feeder, CastSpellParams{})
	if got.Generic != 0 {
		t.Errorf("{B} under a {2} reduction: generic = %d, want 0", got.Generic)
	}
	if len(got.Required) != 1 {
		t.Fatalf("{B} under a {2} reduction: required = %+v, want one {B} still standing", got.Required)
	}

	// {1}{B} — one generic to spend, one left over that must not
	// reach the {B}.
	brood := spellInHand(t, g, me, "Bloodghast", "Creature — Vampire", "{1}{B}")
	got = priceOf(t, g, me, brood, CastSpellParams{})
	if got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("{1}{B} under a {2} reduction = %+v, want {B}", got)
	}

	// {3}{B} — the reduction fully lands.
	giant := spellInHand(t, g, me, "Test Giant", "Creature — Giant", "{3}{B}")
	got = priceOf(t, g, me, giant, CastSpellParams{})
	if got.Generic != 1 || len(got.Required) != 1 {
		t.Errorf("{3}{B} under a {2} reduction = %+v, want {1}{B}", got)
	}
}

// Two reducers on the board are two separate spends against the same
// generic pool, each clamped on its own. Three total reduction on a
// {1}{B} spell leaves {B}, not a negative anything.
func TestStackedReductionsClampIndependently(t *testing.T) {
	const a, b = "test-red-a", "test-red-b"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, func(id string) []CostModifier {
		switch id {
		case a:
			return []CostModifier{{Kind: CostReduction, Label: "-2", Amount: fixed(2)}}
		case b:
			return []CostModifier{{Kind: CostReduction, Label: "-1", Amount: fixed(1)}}
		}
		return nil
	})
	modifierSource(t, g, me, "Reducer A", a)
	modifierSource(t, g, me, "Reducer B", b)
	spell := spellInHand(t, g, me, "Test Spell", "Creature", "{1}{B}")
	got := priceOf(t, g, me, spell, CastSpellParams{})
	if got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("{1}{B} under {2} + {1} of reduction = %+v, want {B}", got)
	}
}

// Trinisphere: applied last, makes up the shortfall in GENERIC mana,
// and leaves an already-expensive spell alone.
func TestCostFloorAppliesLastAndInGenericMana(t *testing.T) {
	const trinisphere, mancer = "test-trinisphere", "test-mancer"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, func(id string) []CostModifier {
		switch id {
		case trinisphere:
			return []CostModifier{{
				Kind:      CostFloor,
				Label:     "each spell that would cost less than three mana costs three",
				AppliesTo: func(q CostQuery) bool { return !q.Source.Tapped },
				Amount:    fixed(3),
			}}
		case mancer:
			return []CostModifier{{Kind: CostReduction, Label: "-1", Amount: fixed(1)}}
		}
		return nil
	})
	sphereID := modifierSource(t, g, me, "Trinisphere", trinisphere)

	// The reminder text's own example: {1}{B} becomes {2}{B}.
	ritual := spellInHand(t, g, me, "Dark Ritual", "Instant", "{1}{B}")
	got := priceOf(t, g, me, ritual, CastSpellParams{})
	if got.Generic != 2 || len(got.Required) != 1 || got.Required[0].Options[0] != "B" {
		t.Errorf("{1}{B} under a three-floor = %+v, want {2}{B}", got)
	}

	// Already three or more: untouched.
	big := spellInHand(t, g, me, "Test Wrath", "Sorcery", "{2}{W}{W}")
	got = priceOf(t, g, me, big, CastSpellParams{})
	if got.Generic != 2 {
		t.Errorf("{2}{W}{W} under a three-floor: generic = %d, want 2 (untouched)", got.Generic)
	}

	// The floor runs AFTER reductions, so a reducer can't duck under
	// it — the spell still ends up at three.
	modifierSource(t, g, me, "Goblin Electromancer", mancer)
	got = priceOf(t, g, me, ritual, CastSpellParams{})
	if got.ManaValue() != 3 {
		t.Errorf("{1}{B} with a reducer AND a three-floor: total = %d, want 3", got.ManaValue())
	}

	// Tapped Trinisphere stops applying — the predicate reads the
	// source's live battlefield state, so nothing needs invalidating.
	tapBattlefieldCard(t, g, sphereID)
	got = priceOf(t, g, me, ritual, CastSpellParams{})
	if got.ManaValue() >= 3 {
		t.Errorf("tapped three-floor still applied: %+v", got)
	}
}

// X counts at its announced value while the total cost is being
// determined, so a {X}{U} announced with X=2 already clears a
// three-floor.
func TestCostFloorCountsAnnouncedX(t *testing.T) {
	const trinisphere = "test-trinisphere"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(trinisphere, CostModifier{
		Kind: CostFloor, Label: "three-floor", Amount: fixed(3),
	}))
	modifierSource(t, g, me, "Trinisphere", trinisphere)
	drain := spellInHand(t, g, me, "Test Drain", "Sorcery", "{X}{U}")

	got := priceOf(t, g, me, drain, CastSpellParams{XValue: 2})
	if got.Generic != 0 {
		t.Errorf("{X=2}{U} under a three-floor: generic = %d, want 0 (already costs three)", got.Generic)
	}
	got = priceOf(t, g, me, drain, CastSpellParams{XValue: 0})
	if got.Generic != 2 {
		t.Errorf("{X=0}{U} under a three-floor: generic = %d, want 2", got.Generic)
	}
}

// The #289 rule, one cost component over: a modifier the engine
// can't price REFUSES THE CAST. It must not clamp to zero, and it
// must not fall through to the printed cost — both of those ship a
// spell cheaper than it should be.
func TestNegativeCostModifierRefusesTheCast(t *testing.T) {
	const broken = "test-broken"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(broken, CostModifier{
		Kind: CostReduction, Label: "sign error", Amount: fixed(-3),
	}))
	modifierSource(t, g, me, "Broken Rock", broken)
	spell := spellInHand(t, g, me, "Test Spell", "Sorcery", "{2}{R}")

	err := g.CastSpell(me.ID, spell, CastSpellParams{})
	if !errors.Is(err, ErrCostModifier) {
		t.Fatalf("cast under a broken modifier: got %v, want ErrCostModifier", err)
	}
	if !me.Hand.Contains(spell) {
		t.Errorf("refused cast left the card out of hand")
	}
	if len(g.Stack.Cards) != 0 {
		t.Errorf("refused cast reached the stack: %d items", len(g.Stack.Cards))
	}
}

// A modifier is consulted per cast against the live board, so the
// predicate sees who is casting and what. "Spells your opponents
// cast" must not tax its own controller.
func TestCostModifierPredicateSeesCasterAndSource(t *testing.T) {
	const aura = "test-aura"
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	withCatalogCostModifiers(t, modifiersFor(aura, CostModifier{
		Kind:  CostIncrease,
		Label: "Artifact spells your opponents cast cost {2} more to cast.",
		AppliesTo: func(q CostQuery) bool {
			return q.Controller != q.Source.Controller && q.Card.IsArtifact()
		},
		Amount: fixed(2),
	}))
	modifierSource(t, g, me, "Aura of Silence", aura)

	mine := spellInHand(t, g, me, "My Signet", "Artifact", "{2}")
	if got := priceOf(t, g, me, mine, CastSpellParams{}); got.Generic != 2 {
		t.Errorf("own artifact taxed: generic = %d, want 2", got.Generic)
	}
	theirs := spellInHand(t, g, them, "Their Signet", "Artifact", "{2}")
	if got := priceOf(t, g, them, theirs, CastSpellParams{}); got.Generic != 4 {
		t.Errorf("opponent's artifact: generic = %d, want 4", got.Generic)
	}
	theirCreature := spellInHand(t, g, them, "Their Bear", "Creature — Bear", "{1}{G}")
	if got := priceOf(t, g, them, theirCreature, CastSpellParams{}); got.Generic != 1 {
		t.Errorf("opponent's creature taxed by an artifact clause: generic = %d, want 1", got.Generic)
	}
}

// Cost modifiers are layered on top of the commander tax and on top
// of an alternative cost, never underneath: CR 903.8 and CR 118.9
// both settle what the spell "would cost", and a modifier modifies
// that.
func TestCostModifierStacksOnCommanderTax(t *testing.T) {
	const sphere = "test-sphere"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(sphere, CostModifier{
		Kind: CostIncrease, Label: "Spells cost {1} more to cast.", Amount: fixed(1),
	}))
	modifierSource(t, g, me, "Sphere of Resistance", sphere)

	advanceTo(t, g, StepPrecombatMain)
	cmdr := NewCard("Test Commander", me.ID)
	cmdr.TypeLine = "Legendary Creature — Elemental"
	cmdr.ManaCost = "{2}{G}"
	me.Command.PushTop(cmdr)
	if me.CommanderCasts == nil {
		me.CommanderCasts = make(map[uuid.UUID]int)
	}
	me.CommanderCasts[cmdr.InstanceID] = 2 // two prior casts ⇒ +{4}

	got := priceOf(t, g, me, cmdr.InstanceID, CastSpellParams{FromZone: "command"})
	if got.Generic != 2+4+1 {
		t.Errorf("taxed commander under a sphere: generic = %d, want 7 ({2} printed + {4} tax + {1} sphere)", got.Generic)
	}
}

// tapBattlefieldCard taps a permanent in place.
func tapBattlefieldCard(t *testing.T, g *Game, id uuid.UUID) {
	t.Helper()
	g.mu.Lock()
	defer g.mu.Unlock()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
			return
		}
	}
	t.Fatalf("card %s not on the battlefield", id)
}
