package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b21_cards_test.go — Star Trek, batch 21 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB21StalkerOracle   = "8f91e7aa-a2d4-42ed-a959-3acd621679a5"
	trekB21DoctorOracle    = "ad1d8eca-e3bf-479f-a442-bcd8c31bfe7f"
	trekB21JadziaOracle    = "67601646-f906-4824-9269-3cf89766d25f"
	trekB21JakeOracle      = "f0c1c97f-b8e1-49ec-b369-3d2694c33f29"
	trekB21MacoOracle      = "0d60cfb1-dd77-4013-b862-41e95fdb53ae"
	trekB21MoopsyOracle    = "89b9efc8-d1cd-4b2a-acf3-4f0dd1629167"
	trekB21MugatoOracle    = "eb9a3b52-62ad-467f-b62a-0273ceee48e5"
	trekB21NeelixOracle    = "7fba522b-a292-4d1f-9999-f787ba2d5fd1"
	trekB21PhysicianOracle = "08787942-456f-4c71-a821-1c90b5639ca1"
	trekB21RokTahkOracle   = "ef919269-043f-45ab-a5da-403cdaa4da23"
)

func TestTrekB21EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB21StalkerOracle:   "Hologram Stalker",
		trekB21DoctorOracle:    "I'm a Doctor, Not a . . .",
		trekB21JadziaOracle:    "Jadzia Dax, Living Legacy",
		trekB21JakeOracle:      "Jake Sisko, War Correspondent",
		trekB21MacoOracle:      "MACO Strike Team",
		trekB21MoopsyOracle:    "Moopsy",
		trekB21MugatoOracle:    "Mugato",
		trekB21NeelixOracle:    "Neelix, Moral Support",
		trekB21PhysicianOracle: "Photosynthetic Physician",
		trekB21RokTahkOracle:   "Rok-Tahk, Budding Scientist",
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

// --- shared helpers ------------------------------------------------------------

// trekB21Tokens is the tokens named `name` that p controls, through the
// layered view.
func trekB21Tokens(t *testing.T, g *game.Game, p uuid.UUID, name string) []game.Card {
	t.Helper()
	var ids []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == p && IsToken(c) {
				ids = append(ids, c.InstanceID)
			}
		}
	})
	out := make([]game.Card, 0, len(ids))
	for _, id := range ids {
		out = append(out, layeredCard(t, g, id))
	}
	return out
}

// trekB21CastStrict puts a catalog card in the active seat's hand, floats
// `mana` and casts it strictly at `targets`, returning the cast error.
func trekB21CastStrict(t *testing.T, g *game.Game, name, typeLine, cost, oracle, mana string, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Owner: me.ID, Controller: me.ID,
		})
	})
	floatMana(t, g, me, mana)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

func trekB21GreenInPool(p *game.Player) int {
	n := 0
	for _, c := range batch01PoolColors(p) {
		if c == "G" {
			n++
		}
	}
	return n
}

// --- Hologram Stalker ------------------------------------------------------------

func TestTrekB21HologramStalkerReturnsAPermanentCardAndMakesOneIllusionATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[next]
	bear := pushToGraveyard(g, me, graveBear())
	bolt := pushToGraveyard(g, me, game.Card{Name: "Grave Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	other := pushToGraveyard(g, me, graveBear())
	later := pushToGraveyard(g, me, graveBear())

	castHoldCreature(t, g, "Hologram Stalker", "Creature — Hirogen Warrior", trekB21StalkerOracle, 2, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enter trigger")
	}
	if hasID(pick.PickTargetCards, bolt) || !hasID(pick.PickTargetCards, bear) {
		t.Fatalf("offered %v: want the permanent cards only", pick.PickTargetCards)
	}
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the permanent card did not return to hand")
	}
	toks := trekB21Tokens(t, g, me.ID, "Illusion")
	if len(toks) != 1 {
		t.Fatalf("%d Illusions after the return, want 1", len(toks))
	}
	if il := toks[0]; il.CurrentPower() != 3 || il.CurrentToughness() != 3 || len(il.Colors) != 0 || !il.IsArtifact() || !il.IsCreature() {
		t.Errorf("the Illusion is %d/%d colors %v type %q, want a 3/3 colorless artifact creature", il.CurrentPower(), il.CurrentToughness(), il.Colors, il.TypeLine)
	}

	// A second card leaving this turn makes nothing.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(other) })
	passPriorityAroundTable(t, g)
	if n := len(trekB21Tokens(t, g, me.ID, "Illusion")); n != 1 {
		t.Fatalf("%d Illusions after a second card left this turn, want 1", n)
	}

	// An opponent's graveyard does not count; mine does again next turn.
	powerUpPassTo(t, g, next)
	theirs := pushToGraveyard(g, opp, graveBear())
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	passPriorityAroundTable(t, g)
	if n := len(trekB21Tokens(t, g, me.ID, "Illusion")); n != 1 {
		t.Fatalf("an opponent's card leaving their graveyard made an Illusion (%d)", n)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(later) })
	passPriorityAroundTable(t, g)
	if n := len(trekB21Tokens(t, g, me.ID, "Illusion")); n != 2 {
		t.Errorf("%d Illusions after a card left on the next turn, want 2", n)
	}
}

// --- I'm a Doctor, Not a . . . -----------------------------------------------------

func TestTrekB21ImADoctorCountersUntapsPumpsAndMakesItOnlyADoctor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Tapped = true })

	if _, err := trekB21CastStrict(t, g, "I'm a Doctor, Not a . . .", "Instant", "{G}", trekB21DoctorOracle, "{G}", cardRefs(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := counterOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	c := layeredCard(t, g, bear)
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if subs := c.Effective().Subtypes; !reflect.DeepEqual(subs, []string{"Doctor"}) || !c.IsCreature() {
		t.Errorf("the creature is %v (creature %v), want a creature that is only a Doctor", subs, c.IsCreature())
	}

	endTurn(t, g)
	c = layeredCard(t, g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 || !c.HasSubtype("Bear") || c.HasSubtype("Doctor") {
		t.Errorf("after the turn: %d/%d %v, want a 3/3 Bear", c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
	}
}

func TestTrekB21ImADoctorRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if _, err := trekB21CastStrict(t, g, "I'm a Doctor, Not a . . .", "Instant", "{G}", trekB21DoctorOracle, "{G}", cardRefs(theirs)); err == nil {
		t.Fatal("targeted a creature an opponent controls")
	}
}

// --- Jadzia Dax, Living Legacy -----------------------------------------------------

func TestTrekB21JadziaGrowsAndAddsGreenForEveryCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step == game.StepPrecombatMain {
		t.Fatal("the fixture starts in the precombat main")
	}
	jadzia := b27Push(g, me.ID, "Jadzia Dax, Living Legacy", "Legendary Creature — Trill Scientist", trekB21JadziaOracle, "{1}{G}", 0, 1, "G")
	addCountersForTest(g, jadzia, game.CounterStun, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	if n := counterOn(g, jadzia, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d +1/+1 counters, want 1", n)
	}
	// One +1/+1 and one stun counter: every counter counts.
	if n := trekB21GreenInPool(me); n != 2 {
		t.Errorf("pool has %d green, want 2", n)
	}
}

func TestTrekB21JadziaDoesNotTriggerInAnOpponentsMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	jadzia := b27Push(g, opp.ID, "Jadzia Dax, Living Legacy", "Legendary Creature — Trill Scientist", trekB21JadziaOracle, "{1}{G}", 0, 1, "G")
	advanceTo(t, g, game.StepPrecombatMain)
	passPriorityAroundTable(t, g)
	if n := counterOn(g, jadzia, game.CounterPlusOne); n != 0 || trekB21GreenInPool(opp) != 0 {
		t.Errorf("triggered in another player's main phase: %d counters", n)
	}
}

func TestTrekB21JadziaTransferSymbiontFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	jadzia := handCardFull(me, "Jadzia Dax, Living Legacy", "Legendary Creature — Trill Scientist", "{1}{G}", trekB21JadziaOracle, []string{"G"})
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{G}{C}")
	if err := g.ActivateCatalogAbility(me.ID, jadzia, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("activated with two mana")
	}
	if !me.Hand.Contains(jadzia) {
		t.Fatal("a refused activation discarded her")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, jadzia, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(jadzia) {
		t.Error("the discard was not the cost")
	}
	passPriorityAroundTable(t, g)
	if n := counterOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2", n)
	}
}

// --- Jake Sisko, War Correspondent ---------------------------------------------------

func TestTrekB21JakeGainsOnEveryLandAndInvestigatesOnTheSecond(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b27Push(g, me.ID, "Jake Sisko, War Correspondent", "Legendary Creature — Human Citizen", trekB21JakeOracle, "{1}{G}", 2, 2, "G")
	life := me.Life
	for i, wantClues := range []int{0, 1, 1} {
		playLandFromHand(t, g, "Forest", "")
		passPriorityAroundTable(t, g)
		if me.Life != life+i+1 {
			t.Errorf("land %d: life %d, want %d", i+1, me.Life, life+i+1)
		}
		if n := len(trekB21Tokens(t, g, me.ID, "Clue")); n != wantClues {
			t.Errorf("land %d: %d Clues, want %d", i+1, n, wantClues)
		}
	}
}

// --- MACO Strike Team ------------------------------------------------------------

func TestTrekB21MacoMakesATappedOfficerWhenAnOpponentTargetsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b27Push(g, me.ID, "MACO Strike Team", "Creature — Human Soldier", trekB21MacoOracle, "{G}", 2, 1, "G")
	big := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)

	// The opponent targeting their own creature: nothing.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if n := len(trekB21Tokens(t, g, me.ID, "Officer")); n != 0 {
		t.Fatalf("%d Officers after an opponent targeted their own creature", n)
	}
	// The opponent targeting a player I am: nothing either.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if n := len(trekB21Tokens(t, g, me.ID, "Officer")); n != 0 {
		t.Fatalf("%d Officers after an opponent targeted me, not a creature", n)
	}

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	toks := trekB21Tokens(t, g, me.ID, "Officer")
	if len(toks) != 1 {
		t.Fatalf("%d Officers, want 1", len(toks))
	}
	o := toks[0]
	if !o.Tapped || o.CurrentPower() != 1 || o.CurrentToughness() != 1 || !reflect.DeepEqual(o.Colors, []string{"R"}) {
		t.Errorf("the Officer: tapped %v %d/%d %v, want a tapped 1/1 red", o.Tapped, o.CurrentPower(), o.CurrentToughness(), o.Colors)
	}
}

func TestTrekB21MacoIgnoresItsControllersOwnSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b27Push(g, me.ID, "MACO Strike Team", "Creature — Human Soldier", trekB21MacoOracle, "{G}", 2, 1, "G")
	big := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 5, 5)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if n := len(trekB21Tokens(t, g, me.ID, "Officer")); n != 0 {
		t.Errorf("%d Officers from my own spell, want 0", n)
	}
}

// --- Moopsy -----------------------------------------------------------------------

func TestTrekB21MoopsyHasDeathtouchAndBuysReach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	moopsy := b27Push(g, me.ID, "Moopsy", "Creature — Beast", trekB21MoopsyOracle, "{G}", 1, 1, "G")
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, moopsy, "deathtouch") || hasEffectiveKeyword(t, g, moopsy, "reach") {
		t.Fatal("want deathtouch and no reach before the activation")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, moopsy, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with one mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, moopsy, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, moopsy, "reach") {
		t.Fatal("no reach after the activation")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, moopsy, "reach") {
		t.Error("reach outlasted the turn")
	}
}

// --- Mugato -----------------------------------------------------------------------

func TestTrekB21MugatoHasReachVigilanceAndDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mugato := b27Push(g, me.ID, "Mugato", "Creature — Ape Beast", trekB21MugatoOracle, "{3}{G}{G}", 6, 6, "G")
	for _, kw := range []string{"reach", "vigilance", "deathtouch"} {
		if !hasEffectiveKeyword(t, g, mugato, kw) {
			t.Errorf("Mugato lacks %s", kw)
		}
	}
}

// --- Neelix, Moral Support ----------------------------------------------------------

func TestTrekB21NeelixMakesFoodAndGrowsWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	neelix := castHoldCreature(t, g, "Neelix, Moral Support", "Legendary Creature — Talaxian Advisor", trekB21NeelixOracle, 1, 1)
	passPriorityAroundTable(t, g)
	foods := trekB21Tokens(t, g, me.ID, "Food")
	if len(foods) != 1 {
		t.Fatalf("%d Food tokens, want 1", len(foods))
	}
	if n := counterOn(g, neelix, game.CounterPlusOne); n != 0 {
		t.Fatalf("%d counters before any life gain", n)
	}

	// An opponent gaining life does nothing.
	gainLifeNow(t, g, uuid.Nil, opp, 2)
	passPriorityAroundTable(t, g)
	if n := counterOn(g, neelix, game.CounterPlusOne); n != 0 {
		t.Fatalf("an opponent's life gain put %d counters on Neelix", n)
	}

	// Eating the Food: one gain of 3 is one counter.
	life := me.Life
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, foods[0].InstanceID, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("eat the Food: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
	if n := counterOn(g, neelix, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after gaining 3, want 1", n)
	}
}

// --- Photosynthetic Physician --------------------------------------------------------

func TestTrekB21PhysicianMakesAnOfficerFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
		t.Fatalf("the fixture starts in a main phase (%s)", g.Turn.Step)
	}
	id := pushCatalogGraveyardCard(me, "Photosynthetic Physician", "Creature — Plant Doctor", trekB21PhysicianOracle, 5, 5)
	floatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("the refused activation moved the card")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the card was not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	toks := trekB21Tokens(t, g, me.ID, "Officer")
	if len(toks) != 1 {
		t.Fatalf("%d Officers, want 1", len(toks))
	}
	if o := toks[0]; o.Tapped || o.CurrentPower() != 1 || o.CurrentToughness() != 1 || !reflect.DeepEqual(o.Colors, []string{"R"}) {
		t.Errorf("the Officer: tapped %v %d/%d %v, want an untapped 1/1 red", o.Tapped, o.CurrentPower(), o.CurrentToughness(), o.Colors)
	}
}

func TestTrekB21PhysicianHasReachAndNoAbilityOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b27Push(g, me.ID, "Photosynthetic Physician", "Creature — Plant Doctor", trekB21PhysicianOracle, "{4}{G}", 5, 5, "G")
	if !hasEffectiveKeyword(t, g, id, "reach") {
		t.Error("no reach")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated from the battlefield")
	}
}

// --- Rok-Tahk, Budding Scientist -----------------------------------------------------

func TestTrekB21RokTahkMakesMurfAndGivesYourCreaturesGG(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := gaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := gaPut(g, opp.ID, "Bear", "Creature — Bear", "")
	if n := grantedManaRowsOf(g, bear); n != 0 {
		t.Fatalf("the bear has %d granted rows before Rok-Tahk", n)
	}
	rok := castHoldCreature(t, g, "Rok-Tahk, Budding Scientist", "Legendary Creature — Brikar Scientist", trekB21RokTahkOracle, 5, 5)
	passPriorityAroundTable(t, g)

	murfs := trekB21Tokens(t, g, me.ID, "Murf")
	if len(murfs) != 1 {
		t.Fatalf("%d Murf tokens, want 1", len(murfs))
	}
	m := murfs[0]
	if !m.IsLegendary() || !m.HasSubtype("Ooze") || !m.HasSubtype("Slug") || m.CurrentPower() != 0 || m.CurrentToughness() != 1 ||
		!reflect.DeepEqual(m.Colors, []string{"G"}) || !game.HasKeyword(&m, "indestructible") {
		t.Errorf("Murf is %q %d/%d %v, want a legendary 0/1 green Ooze Slug with indestructible", m.TypeLine, m.CurrentPower(), m.CurrentToughness(), m.Colors)
	}

	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's creature has %d granted rows", n)
	}
	if err := tapGrantedMana(t, g, me.ID, bear, "G"); err != nil {
		t.Fatalf("tap the bear: %v", err)
	}
	if n := trekB21GreenInPool(me); n != 2 {
		t.Errorf("pool has %d green, want 2", n)
	}
	// CR 302.6: Rok-Tahk entered this turn and can't tap for it yet.
	if err := tapGrantedMana(t, g, me.ID, rok, "G"); err == nil {
		t.Error("a summoning-sick Rok-Tahk tapped for mana")
	}
}
