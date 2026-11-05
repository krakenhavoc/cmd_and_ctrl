package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b2_cards_test.go — Star Trek, batch 2 (#4973). Each test plays the
// real catalog entry and asserts what its text says.

const (
	trekB2DeannaOracle     = "cf612cbe-7af8-48fd-875e-a669e05cfb49"
	trekB2Dot7Oracle       = "b860f64c-e830-49ea-a478-5dc87e8a3be1"
	trekB2CrusherOracle    = "5fe6f4ae-3fb1-4df6-8760-88c8666068df"
	trekB2McCoyOracle      = "df8a0737-5841-4c17-b567-e9296541c3c9"
	trekB2HologramOracle   = "8170dcbb-9879-414c-819d-ea2658ff770c"
	trekB2ExcalbianOracle  = "f409e869-ea1a-4e08-b19e-949436d01cc1"
	trekB2FieldMedicOracle = "3759f1de-eda1-4f09-b12b-ea5f5fc731d7"
	trekB2ResupplyOracle   = "4450d4e0-2825-4536-bb0f-f3a36f1f682c"
	trekB2FreeBorgOracle   = "28914782-92f4-407f-beb6-5e290a49000c"
	trekB2GhostOracle      = "e705b055-14b3-4de0-b4a7-39626839dea8"

	trekB2VampireOracle = "a2057118-9fd1-431d-9441-fef198d67ff2"
)

func TestTrekB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB2DeannaOracle:     "Deanna Troi, Psychic Counsel",
		trekB2Dot7Oracle:       "DOT-7 Repair Squad",
		trekB2CrusherOracle:    "Dr. Beverly Crusher",
		trekB2McCoyOracle:      "Dr. McCoy, Surly Surgeon",
		trekB2HologramOracle:   "Emergency Medical Hologram",
		trekB2ExcalbianOracle:  "Excalbian Gamemaster",
		trekB2FieldMedicOracle: "Federation Field Medic",
		trekB2ResupplyOracle:   "Federation Resupply",
		trekB2FreeBorgOracle:   "Free Borg Revolutionaries",
		trekB2GhostOracle:      "Ghost of Illyria",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// trekB2Cast puts a catalog creature with its printed P/T into the active
// seat's hand, casts it from a main phase and resolves the spell alone,
// leaving whatever it triggered for the caller.
func trekB2Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	for i := 0; i < 8 && g.Stack.Contains(id); i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(id) {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
	return id
}

// trekB2Counter puts n +1/+1 counters on a permanent.
func trekB2Counter(t *testing.T, g *game.Game, id uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(id, game.CounterPlusOne, n); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
}

// trekB2GainLife has p gain n life from no source and settles what it triggered.
func trekB2GainLife(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n); err != nil {
			t.Fatalf("ChangePlayerLifeForEffect: %v", err)
		}
	})
	settleOrderingTriggers(t, g)
}

// --- Deanna Troi, Psychic Counsel --------------------------------------------

func TestTrekB2DeannaTroiOthersEnterWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	deanna := b12Push(g, me.ID, "Deanna Troi, Psychic Counsel", "Legendary Creature — Betazoid Human Advisor", trekB2DeannaOracle, 2, 2)
	bear := trekB2Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("a creature cast under Deanna: %d +1/+1 counters, want 1", got)
	}
	if got := plusOnes(g, deanna); got != 0 {
		t.Errorf("Deanna herself has %d counters, want 0 (other creatures)", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 red Officer"), 1) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, findBattlefieldByName(g, "Officer")); got != 1 {
		t.Errorf("a token: %d counters, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, BlackZombieToken(), 1) })
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, findBattlefieldByName(g, "Zombie")); got != 0 {
		t.Errorf("an opponent's creature: %d counters, want 0", got)
	}
}

// --- DOT-7 Repair Squad ------------------------------------------------------

func TestTrekB2Dot7UntapsATargetAtSorcerySpeedOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	squad := b12Push(g, me.ID, "DOT-7 Repair Squad", "Artifact Creature — Robot", trekB2Dot7Oracle, 3, 3)
	assertKeywords(t, g, squad, "flying")
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	theirs := pushTappedVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	untap := func(target uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{
			Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	}

	// The harness starts in the draw step: not a sorcery-speed window.
	floatMana(t, g, me, "{C}{W}")
	if err := untap(tapped); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{W}")
	if err := untap(tapped); err != nil {
		t.Fatalf("untap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tmtB1Card(t, g, tapped).Tapped {
		t.Error("the target is still tapped")
	}
	floatMana(t, g, me, "{C}{W}")
	if err := untap(theirs); err == nil {
		t.Error("activated a second time in one turn")
	}

	// Any creature is a legal target, an opponent's included, on a later turn.
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	advanceToStepOf(t, g, 0, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	floatMana(t, g, me, "{C}{W}")
	if err := untap(theirs); err != nil {
		t.Fatalf("next turn's untap: %v", err)
	}
	passPriorityAroundTable(t, g)
	if tmtB1Card(t, g, theirs).Tapped {
		t.Error("the opponent's creature is still tapped")
	}
}

func TestTrekB2Dot7NeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	squad := b12Push(g, me.ID, "DOT-7 Repair Squad", "Artifact Creature — Robot", trekB2Dot7Oracle, 3, 3)
	tapped := pushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tapped}},
	}); err == nil {
		t.Fatal("activated with {1} for a {1}{W} cost")
	}
}

// --- Dr. Beverly Crusher ------------------------------------------------------

func TestTrekB2CrusherDoublesHerOwnLifeGainTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crusher := b12Push(g, me.ID, "Dr. Beverly Crusher", "Legendary Creature — Human Doctor", trekB2CrusherOracle, 2, 2)
	trekB2GainLife(t, g, me, 3)
	if got := plusOnes(g, crusher); got != 2 {
		t.Errorf("one life gain put %d counters on Crusher, want 2 (the trigger and its additional time)", got)
	}
}

func TestTrekB2CrusherDoublesAnotherPermanentsLifeGainTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Dr. Beverly Crusher", "Legendary Creature — Human Doctor", trekB2CrusherOracle, 2, 2)
	mine := b12Push(g, me.ID, "Bloodbond Vampire", "Creature — Vampire Shaman Ally", trekB2VampireOracle, 2, 2)
	theirs := b12Push(g, opp.ID, "Bloodbond Vampire", "Creature — Vampire Shaman Ally", trekB2VampireOracle, 2, 2)
	trekB2GainLife(t, g, me, 1)
	if got := plusOnes(g, mine); got != 2 {
		t.Errorf("your Vampire: %d counters, want 2", got)
	}
	// An opponent gaining life: their Vampire is not a permanent you
	// control, and it was not you gaining life.
	trekB2GainLife(t, g, opp, 1)
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("the opponent's Vampire: %d counters, want 1", got)
	}
}

func TestTrekB2CrusherDoesNotDoubleOtherCauses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crusher := b12Push(g, me.ID, "Dr. Beverly Crusher", "Legendary Creature — Human Doctor", trekB2CrusherOracle, 2, 2)
	// The Hologram's enters trigger is caused by a creature entering, not
	// by life gain, so it triggers once: one counter on its target and 1
	// life. Its own life gain then doubles Crusher's trigger.
	life := me.Life
	holo := trekB2Cast(t, g, "Emergency Medical Hologram", "Artifact Creature — Illusion Doctor", trekB2HologramOracle, 0, 1)
	pickIfAsked(t, g, me.ID, holo)
	settleOrderingTriggers(t, g)
	if got := plusOnes(g, holo); got != 1 {
		t.Errorf("the Hologram's enters trigger put %d counters, want 1", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
	if got := plusOnes(g, crusher); got != 2 {
		t.Errorf("Crusher has %d counters after the 1 life, want 2", got)
	}
}

// --- Dr. McCoy, Surly Surgeon -------------------------------------------------

func TestTrekB2McCoyEntersPuttingACounterOnAnother(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	mccoy := trekB2Cast(t, g, "Dr. McCoy, Surly Surgeon", "Legendary Creature — Human Doctor", trekB2McCoyOracle, 2, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(mccoy)); err == nil {
		t.Error("McCoy targeted himself (another)")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(theirs)); err == nil {
		t.Error("McCoy targeted an opponent's creature")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 || plusOnes(g, mccoy) != 0 {
		t.Errorf("counters: Bear %d, McCoy %d; want 1, 0", plusOnes(g, bear), plusOnes(g, mccoy))
	}
}

func TestTrekB2McCoyReplacesDestructionOfACreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Dr. McCoy, Surly Surgeon", "Legendary Creature — Human Doctor", trekB2McCoyOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB2Counter(t, g, bear, 2)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the Bear with +1/+1 counters was destroyed")
	}
	if got := plusOnes(g, bear); got != 0 {
		t.Errorf("the Bear kept %d +1/+1 counters, want all removed", got)
	}
	// With no counter left, a second destruction is not replaced.
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a Bear without +1/+1 counters survived destruction")
	}
}

func TestTrekB2McCoyHealsLethalDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mccoy := b12Push(g, me.ID, "Dr. McCoy, Surly Surgeon", "Legendary Creature — Human Doctor", trekB2McCoyOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	trekB2Counter(t, g, bear, 1)
	g.WithWriteLock(func() {
		if err := (DealDamage{Source: mccoy, Target: bear, Amount: 3}).Apply(NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: mccoy})); err != nil {
			t.Fatalf("DealDamage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("lethal damage destroyed the Bear despite its +1/+1 counter")
	}
	c := tmtB1Card(t, g, bear)
	if c.DamageMarked != 0 || plusOnes(g, bear) != 0 {
		t.Errorf("the Bear: %d damage, %d counters; want 0 and 0", c.DamageMarked, plusOnes(g, bear))
	}
}

func TestTrekB2McCoyLeavesOthersAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mccoy := b12Push(g, me.ID, "Dr. McCoy, Surly Surgeon", "Legendary Creature — Human Doctor", trekB2McCoyOracle, 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB2Counter(t, g, theirs, 1)
	b27Kill(g, plain)
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plain) {
		t.Error("a creature with no +1/+1 counter survived")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("an opponent's creature with a counter survived")
	}
	// McCoy protects only other creatures.
	trekB2Counter(t, g, mccoy, 1)
	b27Kill(g, mccoy)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mccoy) {
		t.Error("McCoy saved himself")
	}
}

// --- Emergency Medical Hologram -----------------------------------------------

func TestTrekB2HologramPutsACounterAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	trekB2Cast(t, g, "Emergency Medical Hologram", "Artifact Creature — Illusion Doctor", trekB2HologramOracle, 0, 1)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(theirs)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("the target has %d counters, want 1 (any creature)", got)
	}
	if me.Life != life+1 {
		t.Errorf("life %d → %d, want +1", life, me.Life)
	}
}

func TestTrekB2HologramGainsNothingIfItsTargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	trekB2Cast(t, g, "Emergency Medical Hologram", "Artifact Creature — Illusion Doctor", trekB2HologramOracle, 0, 1)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d: the trigger with an illegal target still resolved", life, me.Life)
	}
}

// --- Excalbian Gamemaster -----------------------------------------------------

func TestTrekB2ExcalbianMakesAnOfficerPerLandName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp")
	trekB2Cast(t, g, "Excalbian Gamemaster", "Creature — Excalbian", trekB2ExcalbianOracle, 6, 6)
	passPriorityAroundTable(t, g)
	officers := tmtB1Named(g, "Officer")
	if len(officers) != 2 {
		t.Fatalf("made %d Officers, want 2 (Plains and Island)", len(officers))
	}
	c := tmtB1Card(t, g, officers[0])
	if !c.IsToken() || !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "R" || !c.HasSubtype("Officer") {
		t.Errorf("the token is %+v, want a 1/1 red Officer creature token", c)
	}
}

func TestTrekB2ExcalbianCountersOnDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gm := b12Push(g, me.ID, "Excalbian Gamemaster", "Creature — Excalbian", trekB2ExcalbianOracle, 6, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	keeper := b12Creature(g, me.ID, "Keeper", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's creature dying triggered the Gamemaster")
	}

	b27Kill(g, bear)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt when a creature you control died")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(theirs)); err == nil {
		t.Error("targeted a creature you don't control")
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(keeper)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, keeper); got != 1 {
		t.Errorf("Keeper has %d counters, want 1", got)
	}

	// The Gamemaster's own death triggers it too.
	b27Kill(g, gm)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(keeper)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, keeper); got != 2 {
		t.Errorf("after the Gamemaster died, Keeper has %d counters, want 2", got)
	}
}

// --- Federation Field Medic ---------------------------------------------------

func trekB2CastMedic(t *testing.T, g *game.Game, mode int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	medic := trekB2Cast(t, g, "Federation Field Medic", "Creature — Human Doctor", trekB2FieldMedicOracle, 3, 2)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	return medic
}

func TestTrekB2FieldMedicCounterAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	medic := trekB2CastMedic(t, g, 0)
	assertKeywords(t, g, medic, "vigilance")
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the Bear did not gain vigilance")
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if hasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("the vigilance outlasted the turn")
	}
}

func TestTrekB2FieldMedicDrawsOnlyWithACounterAround(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB2CastMedic(t, g, 1)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("drew with no permanent with a counter: hand %d → %d", hand, got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	rock := b12Permanent(g2, me2.ID, "Charged Rock", "Artifact")
	g2.WithWriteLock(func() { _ = g2.AddCounterForEffect(rock, "charge", 1) })
	trekB2CastMedic(t, g2, 1)
	hand2 := me2.Hand.Size()
	passPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != hand2+1 {
		t.Errorf("with a charge counter on a permanent you control: hand %d → %d, want one drawn", hand2, got)
	}

	// An opponent's permanent with a counter does not count.
	g3 := newCatalogGame(t)
	me3 := g3.Seats[0]
	theirs := b12Creature(g3, g3.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB2Counter(t, g3, theirs, 1)
	trekB2CastMedic(t, g3, 1)
	hand3 := me3.Hand.Size()
	passPriorityAroundTable(t, g3)
	if got := me3.Hand.Size(); got != hand3 {
		t.Errorf("drew off an opponent's counter: hand %d → %d", hand3, got)
	}
}

// --- Federation Resupply ------------------------------------------------------

func TestTrekB2ResupplyFindsASpacecraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := stapleLibraryCard(me, "Starship", "Artifact — Spacecraft")
	castCatalogSpell(t, g, "Federation Resupply", "Sorcery", trekB2ResupplyOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ship)
	if !me.Hand.Contains(ship) {
		t.Error("the Spacecraft did not reach the hand")
	}
}

func TestTrekB2ResupplyRefusesOtherCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stapleLibraryCard(me, "Sword", "Artifact — Equipment")
	bear := stapleLibraryCard(me, "Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Federation Resupply", "Sorcery", trekB2ResupplyOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	beginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Error("the search took a creature card")
	}
}

func TestTrekB2ResupplyBeamsUpFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	aura := stapleLibraryCard(me, "Rancor", "Enchantment — Aura")
	spell := seedGraveyardCard(t, g, "Federation Resupply", "Sorcery", trekB2ResupplyOracle)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == spell {
				me.Graveyard.Cards[i].ManaCost = "{1}{W}"
			}
		}
	})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	beam := func(ids ...uuid.UUID) error {
		return g.CastSpell(me.ID, spell, game.CastSpellParams{
			Strict: true, FromZone: "graveyard", AlternativeCost: "beam_me_up", AltCostIDs: ids,
		})
	}

	floatMana(t, g, me, "{C}{C}{C}")
	if err := beam(bear); err == nil {
		t.Fatal("beamed up for {3}")
	}
	floatMana(t, g, me, "{W}")
	if err := beam(); err == nil {
		t.Fatal("beamed up without returning a creature")
	}
	if err := beam(theirs); err == nil {
		t.Fatal("beamed up returning an opponent's creature")
	}
	if err := beam(bear); err != nil {
		t.Fatalf("beam me up: %v", err)
	}
	if !me.Hand.Contains(bear) {
		t.Error("the creature was not returned to hand as the cost")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, aura)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not reach the hand")
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("the beamed-up spell was not exiled")
	}
}

func TestTrekB2ResupplyBeamMeUpIsGraveyardOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	spell := handCardFull(me, "Federation Resupply", "Sorcery", "{1}{W}", trekB2ResupplyOracle, []string{"W"})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	floatMana(t, g, me, "{C}{C}{C}{W}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, AlternativeCost: "beam_me_up", AltCostIDs: []uuid.UUID{bear},
	}); err == nil {
		t.Error("beam me up was paid from the hand")
	}
}

// --- Free Borg Revolutionaries ------------------------------------------------

func TestTrekB2FreeBorgPumpsTheTeamUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	borg := trekB2Cast(t, g, "Free Borg Revolutionaries", "Artifact Creature — Borg Rebel", trekB2FreeBorgOracle, 4, 5)
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, bear) != 3 || effectiveToughness(t, g, bear) != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
	}
	if effectivePower(t, g, borg) != 5 || effectiveToughness(t, g, borg) != 6 {
		t.Errorf("the Revolutionaries are %d/%d, want 5/6", effectivePower(t, g, borg), effectiveToughness(t, g, borg))
	}
	if effectivePower(t, g, theirs) != 2 {
		t.Error("an opponent's creature was pumped")
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if effectivePower(t, g, bear) != 2 {
		t.Error("the pump outlasted the turn")
	}
}

func TestTrekB2FreeBorgBasicLandcycles(t *testing.T) {
	spec, ok := Lookup(trekB2FreeBorgOracle)
	if !ok || len(spec.Activated) != 1 || !spec.Activated[0].Cycling {
		t.Fatalf("Free Borg Revolutionaries: want one cycling ability, got %+v", spec.Activated)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	plains := stapleLibraryCard(me, "Plains", "Basic Land — Plains")
	stapleLibraryCard(me, "Bear", "Creature — Bear")
	borg := handCardFull(me, "Free Borg Revolutionaries", "Artifact Creature — Borg Rebel", "{3}{W}{W}", trekB2FreeBorgOracle, []string{"W"})
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, borg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("basic landcycling: %v", err)
	}
	if !me.Graveyard.Contains(borg) {
		t.Error("the card was not discarded as the cost")
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, plains)
	if !me.Hand.Contains(plains) {
		t.Error("the basic land did not reach the hand")
	}
}

// --- Ghost of Illyria ---------------------------------------------------------

func TestTrekB2GhostBlinksAnotherNonlandPermanentAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	ghost := trekB2Cast(t, g, "Ghost of Illyria", "Creature — Elemental Spirit", trekB2GhostOracle, 3, 3)
	assertKeywords(t, g, ghost, "flying")
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for name, bad := range map[string]uuid.UUID{"itself": ghost, "a land": land, "an opponent's permanent": theirs} {
		if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(bad)); err == nil {
			t.Errorf("the Ghost targeted %s", name)
		}
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(rock)...)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !g.Exile.Contains(rock) {
		t.Fatal("the Rock was not exiled")
	}
	for i := 0; g.Turn.Step != game.StepEnd; i++ {
		if i > 20 {
			t.Fatal("never reached the end step")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(rock) || len(tmtB1Named(g, "Rock")) != 1 {
		t.Error("the Rock did not return at the next end step")
	}
}

func TestTrekB2GhostMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	trekB2Cast(t, g, "Ghost of Illyria", "Creature — Elemental Spirit", trekB2GhostOracle, 3, 3)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the Rock left with no target chosen")
	}
}
