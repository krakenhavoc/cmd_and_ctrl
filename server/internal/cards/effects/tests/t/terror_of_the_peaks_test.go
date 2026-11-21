package testst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const dismemberOracle = "fd74f8eb-0253-42dd-8277-186d4934da38"

func TestTerrorOfThePeaksTaxesAnOpponentsSpellThatTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	effectstest.AdvanceToMain(t, g)

	before := me.Life
	if _, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(terror)); err != nil {
		t.Fatalf("a Bolt at Terror of the Peaks is castable by a caster with the life: %v", err)
	}
	if got := before - me.Life; got != 3 {
		t.Errorf("the caster paid %d life, want the additional 3", got)
	}
}

func TestTerrorOfThePeaksDoesNotTaxItsControllersSpellsOrOtherTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, me.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	other := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)

	before := me.Life
	if _, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(terror)); err != nil {
		t.Fatalf("its own controller's Bolt: %v", err)
	}
	if me.Life != before {
		t.Errorf("its controller's own spell costs no life: %d -> %d", before, me.Life)
	}
	if _, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(other)); err != nil {
		t.Fatalf("a Bolt at another creature: %v", err)
	}
	if me.Life != before {
		t.Errorf("a spell that doesn't target the Dragon costs no life: %d -> %d", before, me.Life)
	}
}

// A player who can't pay can't target it (CR 119.4): the cast is
// refused whole, nothing is paid, and the spell stays in hand.
func TestTerrorOfThePeaksRefusesACasterWhoCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	effectstest.AdvanceToMain(t, g)
	me.Life = 2

	bolt, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(terror))
	if !errors.Is(err, game.ErrCantPayTargetTax) {
		t.Fatalf("with 2 life the cast is refused: got %v, want ErrCantPayTargetTax", err)
	}
	if me.Life != 2 || !me.Hand.Contains(bolt) {
		t.Errorf("a refused cast pays and moves nothing: life %d, in hand %v", me.Life, me.Hand.Contains(bolt))
	}
	// The same Bolt at a creature that isn't taxed is fine at 2 life.
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(bear)}}); err != nil {
		t.Errorf("the untaxed target is still castable: %v", err)
	}
}

// CR 119.4 lets a player pay life equal to their life total: exactly 3
// pays.
func TestTerrorOfThePeaksAllowsPayingExactlyYourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	effectstest.AdvanceToMain(t, g)
	me.Life = 3
	if _, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(terror)); err != nil {
		t.Fatalf("3 life pays the 3: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0", me.Life)
	}
}

// The tax is on the SPELL: a spell that names the Dragon in two of its
// slots pays once; two Dragons named pay twice.
func TestTerrorOfThePeaksChargesOncePerDragonNotPerTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	two := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	effectstest.AdvanceToMain(t, g)

	// Azorius Ploy: two instances of "target creature", no "another".
	ploy := effectstest.PushCatalogHandCard(me, "Azorius Ploy", "Instant", effectstest.Pr7bAzoriusPloy)
	before := me.Life
	if err := g.CastSpell(me.ID, ploy, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: one, Slot: 0}, {Kind: game.TargetCard, ID: one, Slot: 1},
	}}); err != nil {
		t.Fatalf("Ploy at one Dragon twice: %v", err)
	}
	if got := before - me.Life; got != 3 {
		t.Errorf("one Dragon named twice costs 3, paid %d", got)
	}

	ploy2 := effectstest.PushCatalogHandCard(me, "Azorius Ploy", "Instant", effectstest.Pr7bAzoriusPloy)
	before = me.Life
	if err := g.CastSpell(me.ID, ploy2, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: one, Slot: 0}, {Kind: game.TargetCard, ID: two, Slot: 1},
	}}); err != nil {
		t.Fatalf("Ploy at two Dragons: %v", err)
	}
	if got := before - me.Life; got != 6 {
		t.Errorf("two Dragons named cost 3 each, paid %d", got)
	}
}

// Sum, not each: the cast must afford every life cost of the one cast
// (CR 119.4 is asked once, of the total).
func TestTerrorOfThePeaksTaxJoinsTheCastsOtherLifeCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	effectstest.AdvanceToMain(t, g)
	// A Phyrexian symbol paid with life (2) plus the tax (3) is 5.
	dismember := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: dismember, Name: "Dismember", TypeLine: "Instant", OracleID: dismemberOracle,
		ManaCost: "{1}{B/P}{B/P}", Owner: me.ID, Controller: me.ID,
	})
	me.Life = 4
	err := g.CastSpell(me.ID, dismember, game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.CardRef(terror)}, PhyrexianLife: 1,
	})
	if !errors.Is(err, game.ErrCantPayTargetTax) {
		t.Fatalf("4 life can't pay 2 for the symbol and 3 for the tax: got %v", err)
	}
	me.Life = 5
	if err := g.CastSpell(me.ID, dismember, game.CastSpellParams{
		Targets: []game.TargetRef{effectstest.CardRef(terror)}, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("5 life pays both: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0 (2 + 3 paid)", me.Life)
	}
}

// An ability that targets the Dragon is not a spell cast, so it pays
// nothing here (ward is the other half of the card's protection family).
func TestTerrorOfThePeaksDoesNotTaxAnAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Bombardment", "Enchantment", effectstest.GoblinBombardmentOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.CardRef(terror)}, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("Goblin Bombardment at the Dragon: %v", err)
	}
	if me.Life != before {
		t.Errorf("an ability that targets it costs no life: %d -> %d", before, me.Life)
	}
}

// A Dragon that has lost its abilities taxes nothing (CR 613.1f): the
// tax is read through the same key the rest of its text is.
func TestTerrorOfThePeaksLosingAbilitiesLosesTheTax(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	terror := effectstest.PushCatalogPermanent(g, opp.ID, "Terror of the Peaks", "Creature — Dragon", effectstest.B04TerrorOfThePeaksOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(terror)
		g.RegisterScopedEffectForEffect(uuid.Nil, []game.AffectedObject{game.PinObject(terror, c.EnteredBattlefieldAt)},
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "test: loses all abilities")
	})
	before := me.Life
	if _, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(terror)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if me.Life != before {
		t.Errorf("a Dragon with no abilities charges nothing: %d -> %d", before, me.Life)
	}
}

func TestTerrorOfThePeaksIsComplete(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.B04TerrorOfThePeaksOracle)
	if !ok || spec.Completeness != effects.CompletenessFull {
		t.Errorf("Terror of the Peaks: %+v, want a Full catalog entry (#3860)", spec.Completeness)
	}
}
