package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b33_cards_test.go — Secrets of Strixhaven, batch 33 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB33PursueOracle    = "4384996f-aeaa-41ac-aa0e-12d72c887731"
	socB33CharmOracle     = "f43e36ea-6bf5-45bd-a4ad-cc507ae45641"
	socB33ProofOracle     = "2cbccc46-bdef-4dfb-90a4-0278c5c8488a"
	socB33LaureateOracle  = "bd9d3a70-faed-4bc6-97d8-1601b36cdb21"
	socB33RabidOracle     = "4db02b34-efec-437c-ab28-aa318b8879ad"
	socB33ArchaicOracle   = "6a68fcaf-b1e8-451d-8e1c-b47db55c83cc"
	socB33RapierOracle    = "cf43518b-bfec-443d-a6c0-f2d848dc5fa8"
	socB33MomentOracle    = "c997afb7-82cb-4579-885e-3628cb1b51e4"
	socB33EmbermareOracle = "95f20067-4471-4bfa-952d-dbbfbb2c2178"
	socB33DebaterOracle   = "e30ccf07-0969-4e81-ba43-c50a2ff9c040"
	socB33BearType        = "Creature — Bear"
)

func TestSocB33EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB33PursueOracle:          "Pursue the Past",
		socB33CharmOracle:           "Quandrix Charm",
		socB33ProofOracle:           "Quandrix, the Proof",
		socB33LaureateOracle:        "Quill-Blade Laureate",
		socB33LaureateOracle + "#1": "Twofold Intent",
		socB33RabidOracle:           "Rabid Attack",
		socB33ArchaicOracle:         "Rancorous Archaic",
		socB33RapierOracle:          "Rapier Wit",
		socB33MomentOracle:          "Rapturous Moment",
		socB33EmbermareOracle:       "Rearing Embermare",
		socB33DebaterOracle:         "Rehearsed Debater",
	}
	for oracle, name := range want {
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

// --- Pursue the Past -------------------------------------------------------------

func TestSocB33PursueGainsTwoThenDiscardsOneAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Pursue the Past", "Sorcery", socB33PursueOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	pitched := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, pitched)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the chosen card was not discarded")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (discard one, draw two)", got)
	}
}

func TestSocB33PursueDecliningTheDiscardDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := lifeOf(g, me.ID)
	castCatalogSpell(t, g, "Pursue the Past", "Sorcery", socB33PursueOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2 even when declining", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand changed from %d to %d after declining", hand, me.Hand.Size())
	}
}

func TestSocB33PursueFlashbackGainsLifeAndExilesTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Pursue the Past", TypeLine: "Sorcery", ManaCost: "{R}{W}",
		OracleID: socB33PursueOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	life := lifeOf(g, me.ID)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "RW")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback", Strict: true}); err == nil {
		t.Fatal("flashback was castable for {R}{W}, want {2}{R}{W}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "RWCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flash it back: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID) - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	if me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell returned to the graveyard instead of being exiled")
	}
}

// --- Quandrix Charm --------------------------------------------------------------

func socB33CharmCounters(t *testing.T, g *game.Game, me *game.Player, pay bool) (spell uuid.UUID) {
	t.Helper()
	toMainForCost(t, g)
	spell = handCardFull(me, "Some Sorcery", "Sorcery", "", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast sorcery: %v", err)
	}
	if pay {
		fillPoolColored(me, "G", 2)
	}
	castModal(t, g, "Quandrix Charm", "Instant", socB33CharmOracle, []int{0}, cardRefs(spell))
	passUntilTaxed(t, g, me.ID)
	answerPayUnless(t, g, me.ID, pay)
	passPriorityAroundTable(t, g)
	return spell
}

func TestSocB33CharmCountersASpellUnlessItsControllerDeclinesToPayTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := socB33CharmCounters(t, g, me, false)
	if !me.Graveyard.Contains(spell) {
		t.Error("declining the {2} should counter the spell")
	}
}

func TestSocB33CharmLetsAPaidSpellResolve(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := socB33CharmCounters(t, g, me, true)
	if !me.Graveyard.Contains(spell) {
		t.Fatal("the sorcery should have resolved into the graveyard")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool holds %d mana, want the {2} spent", me.ManaPool.Total())
	}
}

func TestSocB33CharmDestroysAnEnchantmentButNotAnythingElse(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	aura := pushTypedCard(g, opp.ID, "Anthem", "Enchantment", "{1}{W}{W}")
	bear := b12Creature(g, opp.ID, "Bear", socB33BearType, 2, 2)
	advanceToMain(t, g)
	if err := castCatalogSpellErrModes(t, g, []int{1}, cardRefs(bear)); err == nil {
		t.Error("a creature was accepted as the target of \"destroy target enchantment\"")
	}
	castModal(t, g, "Quandrix Charm", "Instant", socB33CharmOracle, []int{1}, cardRefs(aura))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the enchantment is still on the battlefield")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

func castCatalogSpellErrModes(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Quandrix Charm", TypeLine: "Instant", OracleID: socB33CharmOracle,
		Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets})
}

func TestSocB33CharmSetsBasePowerAndToughnessToFiveUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", socB33BearType, 2, 2)
	addCountersForTest(g, bear, game.CounterPlusOne, 1)
	advanceToMain(t, g)
	castModal(t, g, "Quandrix Charm", "Instant", socB33CharmOracle, []int{2}, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := socB14PT(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("P/T = %d/%d, want 6/6 (base 5/5 plus the counter)", p, tough)
	}
	endTurn(t, g)
	if p, tough := socB14PT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("P/T = %d/%d after the turn, want 3/3", p, tough)
	}
}

func TestSocB33CharmCanShrinkABigCreatureToFive(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", socB33BearType, 8, 8)
	advanceToMain(t, g)
	castModal(t, g, "Quandrix Charm", "Instant", socB33CharmOracle, []int{2}, cardRefs(big))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 5 || tough != 5 {
		t.Errorf("P/T = %d/%d, want 5/5 (\"has base power and toughness 5/5\" sets, not adds)", p, tough)
	}
}

// --- Quandrix, the Proof ---------------------------------------------------------

func TestSocB33ProofHasFlyingTrampleAndItsOwnCascade(t *testing.T) {
	spec, ok := Lookup(socB33ProofOracle)
	if !ok {
		t.Fatal("Quandrix, the Proof is not registered")
	}
	want := map[string]bool{"flying": false, "trample": false}
	for _, k := range spec.PrintedKeywords {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing printed keyword %q", k)
		}
	}
	if len(spec.Triggered) != 2 {
		t.Fatalf("%d triggered abilities, want its own cascade and the granted one", len(spec.Triggered))
	}
}

func TestSocB33ProofCascadesWhenItIsCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 3)
	castWithCost(t, g, "Quandrix, the Proof", "Legendary Creature — Elder Dragon", "{4}{G}{U}", socB33ProofOracle)
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("casting Quandrix made %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

func TestSocB33ProofGivesCascadeToInstantsAndSorceriesCastFromYourHandOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)
	pushCatalogPermanent(g, me.ID, "Quandrix, the Proof", "Legendary Creature — Elder Dragon", socB33ProofOracle, false)

	castWithCost(t, g, "Bear Spell", socB33BearType, "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("a creature spell got %d cascade offers, want 0", got)
	}

	castWithCost(t, g, "A Sorcery", "Sorcery", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("a sorcery cast from hand got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	castWithCost(t, g, "An Instant", "Instant", "{3}{G}", "")
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("an instant cast from hand got %d cascade offers, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
}

func TestSocB33ProofDoesNotGiveCascadeToASpellCastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 8)
	pushCatalogPermanent(g, me.ID, "Quandrix, the Proof", "Legendary Creature — Elder Dragon", socB33ProofOracle, false)
	advanceToMain(t, g)
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: "Exiled Instant", TypeLine: "Instant", ManaCost: "{3}{G}",
		Owner: me.ID, Controller: me.ID})
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: me.ID, Zone: game.ZoneExile,
		Duration: game.WhileInZoneDuration(), Source: uuid.New()})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "GGGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 0 {
		t.Errorf("a spell cast from exile got %d cascade offers, want 0", got)
	}
}

// --- Quill-Blade Laureate // Twofold Intent --------------------------------------

func TestSocB33LaureateEntersPreparedWithDoubleStrikeAndTwofoldIntentPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	c := sos1Prep(me.ID, socB33LaureateOracle, []string{"double strike"}, "Quill-Blade Laureate", "Creature — Human Cleric", "{1}{W}", 1, 1,
		"Twofold Intent", "Sorcery", "{1}{W}")
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Quill-Blade Laureate did not enter prepared")
	}
	if !hasEffectiveKeyword(t, g, c.InstanceID, "double strike") {
		t.Error("Quill-Blade Laureate has no double strike")
	}
	bear := b12Creature(g, me.ID, "Bear", socB33BearType, 2, 2)
	fraBCastCopy(t, g, me, "Twofold Intent", game.TargetRef{Kind: game.TargetCard, ID: bear})
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("target P/T = %d/%d, want 3/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Error("the target has no double strike")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
	endTurn(t, g)
	if p := effectivePower(t, g, bear); p != 2 || hasEffectiveKeyword(t, g, bear, "double strike") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

// --- Rabid Attack ----------------------------------------------------------------

func TestSocB33RabidAttackPumpsEachTargetAndTheyDrawWhenTheyDie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", socB33BearType, 2, 2)
	b := b12Creature(g, me.ID, "B", socB33BearType, 2, 2)
	c := b12Creature(g, me.ID, "C", socB33BearType, 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Rabid Attack", "Instant", socB33RabidOracle, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p := effectivePower(t, g, id); p != 3 {
			t.Errorf("a target has power %d, want 3", p)
		}
	}
	if p := effectivePower(t, g, c); p != 2 {
		t.Errorf("the untargeted creature has power %d, want 2", p)
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(a))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Fatal("the bolted creature survived")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		// castCatalogSpell puts the Bolt in hand and casts it (net 0); the dies trigger drew one.
		t.Errorf("hand changed by %d, want +1 (the dies trigger drew a card)", got)
	}
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, cardRefs(c))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d after a creature with no grant died, want it still +1", got)
	}
}

func TestSocB33RabidAttackEndsAtTheEndOfTheTurnAndAcceptsNoTargets(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "A", socB33BearType, 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Rabid Attack", "Instant", socB33RabidOracle, nil)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, a); p != 2 {
		t.Errorf("power %d after a spell with no targets, want 2", p)
	}
	castCatalogSpell(t, g, "Rabid Attack", "Instant", socB33RabidOracle, cardRefs(a))
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	if p := effectivePower(t, g, a); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
	hand := me.Hand.Size()
	opp := g.Seats[g.Turn.ActiveSeat]
	bolt := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle,
		Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: cardRefs(a)}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Fatal("the creature survived the Bolt")
	}
	if me.Hand.Size() != hand {
		t.Errorf("the dies trigger outlasted the turn: hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestSocB33RabidAttackCannotTargetAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", socB33BearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Rabid Attack", "Instant", socB33RabidOracle, cardRefs(theirs)); err == nil {
		t.Error("an opponent's creature was accepted for \"creatures you control\"")
	}
}

// --- Rancorous Archaic -----------------------------------------------------------

func socB33CastArchaic(t *testing.T, g *game.Game, me *game.Player, mana string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rancorous Archaic", TypeLine: "Creature — Avatar", ManaCost: "{5}",
			OracleID: socB33ArchaicOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.ManaPool.EmptyPool()
	})
	floatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Rancorous Archaic with %s: %v", mana, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestSocB33ArchaicHasReachAndTrample(t *testing.T) {
	spec, ok := Lookup(socB33ArchaicOracle)
	if !ok {
		t.Fatal("Rancorous Archaic is not registered")
	}
	want := map[string]bool{"reach": false, "trample": false}
	for _, k := range spec.PrintedKeywords {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing printed keyword %q", k)
		}
	}
}

func TestSocB33ArchaicEntersWithACounterPerColorSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	five := socB33CastArchaic(t, g, me, "WUBRG")
	if got := counterCount(g, five, game.CounterPlusOne); got != 5 {
		t.Errorf("five colours spent: %d counters, want 5", got)
	}
	if p, tough := socB14PT(t, g, five); p != 7 || tough != 7 {
		t.Errorf("P/T = %d/%d, want 7/7 (2/2 with five counters)", p, tough)
	}
	two := socB33CastArchaic(t, g, me, "WWWUU")
	if got := counterCount(g, two, game.CounterPlusOne); got != 2 {
		t.Errorf("white and blue spent: %d counters, want 2 (colours, not mana)", got)
	}
	none := socB33CastArchaic(t, g, me, "CCCCC")
	if got := counterCount(g, none, game.CounterPlusOne); got != 0 {
		t.Errorf("colourless mana spent: %d counters, want 0", got)
	}
}

// --- Rapier Wit ------------------------------------------------------------------

func TestSocB33RapierWitOnYourTurnTapsStunsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", socB33BearType, 2, 2)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Rapier Wit", "Instant", socB33RapierOracle, cardRefs(theirs))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, theirs) {
		t.Error("the target is untapped")
	}
	if n := countersOn(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d on your own turn, want 1", n)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestSocB33RapierWitOnAnotherPlayersTurnOnlyTapsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", socB33BearType, 2, 2)
	batch01AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rapier Wit", TypeLine: "Instant", OracleID: socB33RapierOracle,
		Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(theirs)}); err != nil {
		t.Fatalf("cast Rapier Wit: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedOnBattlefield(t, g, theirs) {
		t.Error("the target is untapped")
	}
	if n := countersOn(g, theirs, game.CounterStun); n != 0 {
		t.Errorf("stun counters %d on another player's turn, want 0", n)
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand changed by %d, want 0 (cast one, drew one)", got)
	}
}

func TestSocB33RapierWitStunsAnAlreadyTappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", socB33BearType, 2, 2)
	g.WithWriteLock(func() { battlefieldCardPtr(t, g, theirs).Tapped = true })
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Rapier Wit", "Instant", socB33RapierOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if n := countersOn(g, theirs, game.CounterStun); n != 1 {
		t.Errorf("stun counters %d on an already tapped creature, want 1", n)
	}
}

// --- Rapturous Moment ------------------------------------------------------------

func TestSocB33RapturousMomentDrawsThreeDiscardsTwoAndAddsFiveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	castCatalogSpell(t, g, "Rapturous Moment", "Sorcery", socB33MomentOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d cards before the discard, want 3", got)
	}
	if got := discardOwed(g, me.ID); got != 2 {
		t.Fatalf("the discard prompt asks for %d cards, want 2", got)
	}
	pitched := []uuid.UUID{me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID}
	answerDiscard(t, g, me.ID, pitched...)
	passPriorityAroundTable(t, g)
	for _, id := range pitched {
		if !me.Graveyard.Contains(id) {
			t.Error("a chosen card was not discarded")
		}
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (three drawn, two discarded)", got)
	}
	if got := me.ManaPool.Total(); got != 5 {
		t.Errorf("pool holds %d mana, want 5", got)
	}
	blue, red := me.ManaPool.CountColor("U"), me.ManaPool.CountColor("R")
	if blue != 2 || red != 3 {
		t.Errorf("pool holds %d blue and %d red, want {U}{U}{R}{R}{R}", blue, red)
	}
}

// --- Rearing Embermare -----------------------------------------------------------

func TestSocB33EmbermareHasReachAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Rearing Embermare", "Creature — Horse Beast", socB33EmbermareOracle, 4, 5)
	advanceToMain(t, g)
	for _, kw := range []string{"reach", "haste"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Rearing Embermare has no %s", kw)
		}
	}
}

// --- Rehearsed Debater -----------------------------------------------------------

func TestSocB33DebaterHasVigilanceAndReparteeGivesPlusOnePlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	debater := b12Push(g, me.ID, "Rehearsed Debater", "Creature — Djinn Bard", socB33DebaterOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", socB33BearType, 2, 2)
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, debater, "vigilance") {
		t.Error("the Debater has no vigilance")
	}
	castCatalogSpell(t, g, "Interjection", "Instant", "47bbadcf-11c6-4a4d-9575-1eb8284a174c", cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, debater), effectiveToughness(t, g, debater); p != 4 || tough != 4 {
		t.Errorf("Debater P/T = %d/%d, want 4/4", p, tough)
	}
	endTurn(t, g)
	if p, tough := effectivePower(t, g, debater), effectiveToughness(t, g, debater); p != 3 || tough != 3 {
		t.Errorf("the pump outlasted the turn: %d/%d", p, tough)
	}
}

func TestSocB33DebaterReparteeIgnoresASpellWithNoCreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	debater := b12Push(g, me.ID, "Rehearsed Debater", "Creature — Djinn Bard", socB33DebaterOracle, 3, 3)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Pursue the Past", "Sorcery", socB33PursueOracle, nil)
	if n := triggersOnStackFrom(g, debater); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell with no creature target", n)
	}
}
