package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Trek Commander batch 32 (#4974). Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	trekB32KimOracle           = "d9bc0f83-c522-44c9-baa6-fffdeb02a0f5"
	trekB32EzriOracle          = "d4fb650a-94fd-4361-9648-45d6ad64b7a2"
	trekB32GuinanOracle        = "b93915ac-744b-4b59-8f16-ae68d6d898c6"
	trekB32HullbreachersOracle = "14c52244-4d09-49b9-b7c2-17ef1ffa2561"
	trekB32HealerOracle        = "9d18253d-a38e-48a4-80bf-aa8fe5283584"
	trekB32JoinOracle          = "b1a6d794-0e48-4bbe-b3a3-d9957ed1bfbf"
	trekB32BashirOracle        = "bbc8c2b4-9949-416f-a722-b1645eff64d7"
	trekB32KorokOracle         = "2f0bad22-a3f5-4b2b-96bf-ff1473e2262d"
	trekB32MartokOracle        = "d17c3079-372d-4e4f-8214-12a9893088ff"
	trekB32ChamberOracle       = "585511e1-4cca-4e70-963d-56f98ae49c29"
)

func TestTrekB32EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB32KimOracle:           "Ensign Harry Kim, Operations",
		trekB32EzriOracle:          "Ezri Dax, Newly Joined",
		trekB32GuinanOracle:        "Guinan, Ancient Advisor",
		trekB32HullbreachersOracle: "Imperial Hullbreachers",
		trekB32HealerOracle:        "Indelicate Healer",
		trekB32JoinOracle:          "Join the Collective",
		trekB32BashirOracle:        "Julian Bashir, Frontier Doctor",
		trekB32KorokOracle:         "Korok, Liberated Rage",
		trekB32MartokOracle:        "Martok, Hero of the Empire",
		trekB32ChamberOracle:       "Maturation Chamber",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func trekB32Me(g *game.Game) *game.Player { return g.Seats[g.Turn.ActiveSeat] }

func trekB32Opp(g *game.Game, n int) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+n)%len(g.Seats)]
}

func trekB32Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return tmtB1Card(t, g, id).CurrentPower()
}

// --- Ensign Harry Kim, Operations ----------------------------------------

func TestTrekB32KimGrowsWhenABiggerCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	kim := b27Push(g, me.ID, "Ensign Harry Kim, Operations", "Legendary Creature — Human Officer", trekB32KimOracle, "{1}{W}", 1, 2, "W")
	advanceToMain(t, g)

	// Not greater: a 1/1 is as big as Kim, and triggers nothing.
	enterCreature(t, g, me.ID, "Small Bear", 1, 1)
	if triggerOnStack(g, kim) != nil {
		t.Fatal("a creature with equal power must not trigger Kim")
	}
	enterCreature(t, g, me.ID, "Big Bear", 2, 2)
	if triggerOnStack(g, kim) == nil {
		t.Fatal("a creature with greater power should trigger Kim")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, kim, game.CounterPlusOne); got != 1 {
		t.Errorf("Kim has %d +1/+1 counters, want 1", got)
	}
	// An opponent's bigger creature is not one you control.
	enterCreature(t, g, trekB32Opp(g, 1).ID, "Their Giant", 5, 5)
	if triggerOnStack(g, kim) != nil {
		t.Error("an opponent's creature must not trigger Kim")
	}
}

// The "if" is intervening (CR 603.4): Kim grown past the newcomer in
// response gets no counter.
func TestTrekB32KimRechecksThePowerAsItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	kim := b27Push(g, me.ID, "Ensign Harry Kim, Operations", "Legendary Creature — Human Officer", trekB32KimOracle, "{1}{W}", 1, 2, "W")
	advanceToMain(t, g)
	enterCreature(t, g, me.ID, "Bear", 3, 3)
	if triggerOnStack(g, kim) == nil {
		t.Fatal("a 3/3 should trigger Kim")
	}
	addCountersForTest(g, kim, game.CounterPlusOne, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, kim, game.CounterPlusOne); got != 2 {
		t.Errorf("Kim has %d counters, want the 2 put in response and no more", got)
	}
}

func TestTrekB32KimDrawsForTwoCountersFromAmongCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	kim := b27Push(g, me.ID, "Ensign Harry Kim, Operations", "Legendary Creature — Human Officer", trekB32KimOracle, "{1}{W}", 1, 2, "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	seedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	advanceToMain(t, g)
	addCountersForTest(g, kim, game.CounterPlusOne, 1)
	addCountersForTest(g, bear, game.CounterPlusOne, 1)

	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, kim, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{kim, bear}, CounterCounts: []int{1, 1}, Strict: true,
	}); err == nil {
		t.Fatal("one mana cannot pay {2}")
	}
	floatForTest(g, me, "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, kim, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{kim, bear}, CounterCounts: []int{1, 1}, Strict: true,
	})
	passPriorityAroundTable(t, g)
	if counterCount(g, kim, game.CounterPlusOne) != 0 || counterCount(g, bear, game.CounterPlusOne) != 0 {
		t.Error("the two counters were not removed")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestTrekB32KimNeedsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	kim := b27Push(g, me.ID, "Ensign Harry Kim, Operations", "Legendary Creature — Human Officer", trekB32KimOracle, "{1}{W}", 1, 2, "W")
	advanceToMain(t, g)
	addCountersForTest(g, kim, game.CounterPlusOne, 1)
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, kim, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{kim}, CounterCounts: []int{2}, Strict: true,
	}); err == nil {
		t.Fatal("one +1/+1 counter cannot pay for two")
	}
}

// --- Ezri Dax, Newly Joined ----------------------------------------------

func TestTrekB32EzriCopiesOnlyACreatureThatDiedThisTurnFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB32Me(g), trekB32Opp(g, 1)
	died := b12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	b27Kill(g, died)
	b27Kill(g, theirs)
	old := graveyardCreature(me, "Old Bear", "Bear", 2, 2)
	if !me.Graveyard.Contains(died) || !opp.Graveyard.Contains(theirs) {
		t.Fatal("the giants should be in their owners' graveyards")
	}

	ezri := castGraveyardCopier(t, g, "Ezri Dax, Newly Joined", "Legendary Creature — Trill Advisor", trekB32EzriOracle, 2, 2)
	prompt := waitForCopyPrompt(t, g)
	if !slices.Contains(prompt.CopyOptions, died) {
		t.Errorf("the creature card put into your graveyard from the battlefield this turn should be offered: %v", prompt.CopyOptions)
	}
	if slices.Contains(prompt.CopyOptions, old) || slices.Contains(prompt.CopyOptions, theirs) {
		t.Errorf("only your own graveyard's creatures that died this turn are offered: %v", prompt.CopyOptions)
	}
	resolveWithCopyChoice(t, g, died)

	got := copyBattlefieldCard(t, g, ezri)
	if got.Name != "Ezri Dax, Newly Joined" || got.CurrentPower() != 3 || got.CurrentToughness() != 3 {
		t.Errorf("Ezri = %q %d/%d, want a 3/3 named Ezri Dax, Newly Joined", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
	if !got.IsLegendary() || !got.HasSubtype("Trill") || !got.HasSubtype("Giant") {
		t.Errorf("the copy should be a legendary Trill Giant: %q", got.TypeLine)
	}
	if !me.Graveyard.Contains(died) {
		t.Error("Ezri copies the card; it stays in the graveyard")
	}
	spec, _ := Lookup(trekB32EzriOracle)
	if !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Error("Ezri has flash")
	}
}

// --- Guinan, Ancient Advisor ---------------------------------------------

func trekB32Guinan(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := trekB32Me(g)
	guinan := b27Push(g, me.ID, "Guinan, Ancient Advisor", "Legendary Creature — El-Aurian Advisor", trekB32GuinanOracle, "{2}{W}", 1, 4, "W")
	seedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	advanceToMain(t, g)
	return guinan
}

func trekB32ActivateGuinan(t *testing.T, g *game.Game, guinan uuid.UUID) {
	t.Helper()
	me := trekB32Me(g)
	if err := g.ActivateCatalogAbility(me.ID, guinan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability costs {1}")
	}
	floatForTest(g, me, "C")
	b16Activate(t, g, me.ID, guinan, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
}

func TestTrekB32GuinanCountersEachNewcomerAndDrawsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	guinan := trekB32Guinan(t, g)
	old := b12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	a := enterCreature(t, g, me.ID, "New Bear", 2, 2)
	b := enterCreature(t, g, me.ID, "Newer Bear", 2, 2)
	theirs := enterCreature(t, g, trekB32Opp(g, 1).ID, "Their Bear", 2, 2)
	hand := me.Hand.Size()
	trekB32ActivateGuinan(t, g, guinan)
	for _, id := range []uuid.UUID{a, b} {
		if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("a creature that entered this turn has %d counters, want 1", got)
		}
	}
	for _, id := range []uuid.UUID{old, guinan, theirs} {
		if got := counterCount(g, id, game.CounterPlusOne); got != 0 {
			t.Errorf("%s has %d counters, want none", tmtB1Card(t, g, id).Name, got)
		}
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a card drawn for two counters", hand, me.Hand.Size())
	}
}

func TestTrekB32GuinanOneCounterDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	guinan := trekB32Guinan(t, g)
	a := enterCreature(t, g, me.ID, "New Bear", 2, 2)
	hand := me.Hand.Size()
	trekB32ActivateGuinan(t, g, guinan)
	if got := counterCount(g, a, game.CounterPlusOne); got != 1 {
		t.Errorf("the newcomer has %d counters, want 1", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want no draw for one counter", hand, me.Hand.Size())
	}
}

// "If you put two or more counters" counts what was put: Hardened Scales
// makes one creature's counter two.
func TestTrekB32GuinanCountsTheCountersActuallyPut(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	guinan := trekB32Guinan(t, g)
	b12Push(g, me.ID, "Hardened Scales", "Enchantment", hardenedScalesOracle, 0, 0)
	a := enterCreature(t, g, me.ID, "New Bear", 2, 2)
	hand := me.Hand.Size()
	trekB32ActivateGuinan(t, g, guinan)
	if got := counterCount(g, a, game.CounterPlusOne); got != 2 {
		t.Fatalf("the newcomer has %d counters, want 2 through Hardened Scales", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want a draw for the two counters put", hand, me.Hand.Size())
	}
}

// --- Imperial Hullbreachers ----------------------------------------------

func trekB32CastHullbreachers(t *testing.T, g *game.Game, squads int) (uuid.UUID, error) {
	t.Helper()
	me := trekB32Me(g)
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Imperial Hullbreachers", "Creature — Klingon Warrior", trekB32HullbreachersOracle, "{2}{W}", 2, 2)
	mana := "{W}{C}{C}"
	var paid []int
	for i := 0; i < squads; i++ {
		mana += "{C}{C}"
		paid = append(paid, 0)
	}
	floatMana(t, g, me, mana)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: paid})
}

func TestTrekB32HullbreachersSquadMakesFlyingCopiesWithBattleCry(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := trekB32CastHullbreachers(t, g, 2); err != nil {
		t.Fatalf("cast with squad paid twice: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	ids := tmtB1Named(g, "Imperial Hullbreachers")
	if len(ids) != 3 {
		t.Fatalf("%d Hullbreachers, want the card and two token copies", len(ids))
	}
	for _, id := range ids {
		assertKeywords(t, g, id, "flying")
	}
}

func TestTrekB32HullbreachersSquadNeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Imperial Hullbreachers", "Creature — Klingon Warrior", trekB32HullbreachersOracle, "{2}{W}", 2, 2)
	floatMana(t, g, me, "{W}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err == nil {
		t.Fatal("a squad cast was accepted with four mana for a {4}{W} total")
	}
}

func TestTrekB32HullbreachersBattleCryPumpsTheOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB32Me(g), trekB32Opp(g, 1)
	hb := b27Push(g, me.ID, "Imperial Hullbreachers", "Creature — Klingon Warrior", trekB32HullbreachersOracle, "{2}{W}", 2, 2, "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, hb, bear)
	passPriorityAroundTable(t, g)
	if got := trekB32Power(t, g, bear); got != 3 {
		t.Errorf("the other attacker has power %d, want 3", got)
	}
	if got := trekB32Power(t, g, hb); got != 2 {
		t.Errorf("the Hullbreachers have power %d, want 2 (battle cry pumps the others)", got)
	}
}

// --- Indelicate Healer ----------------------------------------------------

func TestTrekB32HealerGivesLifelinkWhenAKlingonOrWarriorEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB32Me(g), trekB32Opp(g, 1)
	healer := b27Push(g, me.ID, "Indelicate Healer", "Creature — Klingon Warrior Doctor", trekB32HealerOracle, "{2}{W}", 3, 3, "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	enterCreature(t, g, me.ID, "Plain Bear", 2, 2)
	if triggerOnStack(g, healer) != nil {
		t.Fatal("a creature that's neither a Klingon nor a Warrior must not trigger the Healer")
	}
	enterCard(t, g, opp.ID, game.Card{Name: "Their Warrior", TypeLine: "Creature — Human Warrior", Power: 2, Toughness: 2})
	if triggerOnStack(g, healer) != nil {
		t.Fatal("an opponent's Warrior must not trigger the Healer")
	}
	enterCard(t, g, me.ID, game.Card{Name: "Warrior", TypeLine: "Creature — Human Warrior", Power: 2, Toughness: 2})
	if triggerOnStack(g, healer) == nil {
		t.Fatal("a Warrior you control should trigger the Healer")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{healer, bear} {
		if !effectiveAbilitiesContain(t, g, id, "lifelink") {
			t.Errorf("%s should have lifelink", tmtB1Card(t, g, id).Name)
		}
	}
	if effectiveAbilitiesContain(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature must not gain lifelink")
	}
}

func TestTrekB32HealerTriggersOnItsOwnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	advanceToMain(t, g)
	healer := enterCard(t, g, me.ID, game.Card{Name: "Indelicate Healer", TypeLine: "Creature — Klingon Warrior Doctor", OracleID: trekB32HealerOracle, Power: 3, Toughness: 3})
	if triggerOnStack(g, healer) == nil {
		t.Fatal("the Healer entering should trigger itself")
	}
	if n := triggersOnStackFrom(g, healer); n != 1 {
		t.Errorf("%d Healer triggers for its own entry, want 1", n)
	}
}

// --- Join the Collective ---------------------------------------------------

func TestTrekB32JoinTheCollectiveMakesABorgPerNonartifactCreatureDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB32Me(g), trekB32Opp(g, 1)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Join the Collective", "Sorcery", trekB32JoinOracle, "{5}{W}{W}", 0, 0)
	floatMana(t, g, me, "{W}{W}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("six mana cannot pay {5}{W}{W}")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	borg := tmtB1Named(g, "Borg")
	if len(borg) != 2 {
		t.Fatalf("%d Borg tokens, want 2 (the golem is an artifact)", len(borg))
	}
	for _, b := range borg {
		c := tmtB1Card(t, g, b)
		if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Borg") || c.Controller != me.ID || c.CurrentPower() != 1 {
			t.Errorf("Borg %+v: want my 1/1 Borg artifact creature token", c)
		}
	}
	for _, name := range []string{"My Bear", "Their Bear", "Their Golem"} {
		if len(tmtB1Named(g, name)) != 0 {
			t.Errorf("%s survived", name)
		}
	}
}

// --- Julian Bashir, Frontier Doctor ---------------------------------------

func TestTrekB32BashirGrowsOnTheSecondCardEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	bashir := b27Push(g, me.ID, "Julian Bashir, Frontier Doctor", "Legendary Creature — Human Doctor", trekB32BashirOracle, "{2}{W}", 1, 1, "W")
	advanceToMain(t, g)
	for i := 0; i < 4; i++ {
		seedLibraryTop(me, "Plains", "Basic Land — Plains", "")
	}
	for g.TurnTallyFor(me.ID).CardsDrawn < 1 {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("draw: %v", err)
		}
	}
	if triggerOnStack(g, bashir) != nil {
		t.Fatal("the first card must not trigger seek out")
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if triggerOnStack(g, bashir) == nil {
		t.Fatal("the second card should trigger seek out")
	}
	passPriorityAroundTable(t, g)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if triggerOnStack(g, bashir) != nil {
		t.Error("the third card must not trigger seek out")
	}
	if got := counterCount(g, bashir, game.CounterPlusOne); got != 1 {
		t.Errorf("Bashir has %d counters, want 1", got)
	}
}

func TestTrekB32BashirReturnsACreatureCardUpToHisPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := trekB32Me(g), trekB32Opp(g, 1)
	bashir := b27Push(g, me.ID, "Julian Bashir, Frontier Doctor", "Legendary Creature — Human Doctor", trekB32BashirOracle, "{2}{W}", 1, 1, "W")
	addCountersForTest(g, bashir, game.CounterPlusOne, 1)
	small := graveyardCardOf(g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	big := graveyardCardOf(g, me, "Hill Giant", "Creature — Giant", "{3}{R}", "")
	spell := graveyardCardOf(g, me, "Divination", "Sorcery", "{1}{U}", "")
	theirs := graveyardCardOf(g, opp, "Their Bears", "Creature — Bear", "{1}{G}", "")

	declareAttack(t, g, opp.ID, bashir)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Bashir's attack trigger")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range p.PickTargetCards {
		offered[id] = true
	}
	if !offered[small] {
		t.Error("a creature card with mana value 2 should be offered at power 2")
	}
	if offered[big] || offered[spell] || offered[theirs] {
		t.Errorf("only your creature cards with mana value 2 or less: %v", p.PickTargetCards)
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(small)...)
	passPriorityAroundTable(t, g)
	if b12ZoneOf(g, small) != game.ZoneBattlefield {
		t.Errorf("the bear is in %s, want the battlefield", b12ZoneOf(g, small))
	}
}

// --- Korok, Liberated Rage --------------------------------------------------

func TestTrekB32KorokPumpsByTheOpponentsAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB32Me(g), trekB32Opp(g, 1), trekB32Opp(g, 2)
	korok := b27Push(g, me.ID, "Korok, Liberated Rage", "Legendary Artifact Creature — Klingon Borg", trekB32KorokOracle, "{3}{W}", 2, 2, "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(korok, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	if triggerOnStack(g, korok) == nil {
		t.Fatal("Korok attacking should trigger battlefront")
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{korok, bear, home} {
		c := tmtB1Card(t, g, id)
		if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
			t.Errorf("%s is %d/%d, want 4/4 for two opponents attacked", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
	}
	if got := trekB32Power(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", got)
	}
}

// --- Martok, Hero of the Empire -------------------------------------------

func TestTrekB32MartokCountersUpToXCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2 := trekB32Me(g), trekB32Opp(g, 1), trekB32Opp(g, 2)
	martok := b27Push(g, me.ID, "Martok, Hero of the Empire", "Legendary Creature — Klingon Warrior", trekB32MartokOracle, "{2}{W}", 2, 2, "W")
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp1.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp1.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(c, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Martok's trigger")
	}
	if p.PickTargetMax != 2 {
		t.Errorf("up to %d targets, want 2 for two opponents attacked", p.PickTargetMax)
	}
	for _, id := range p.PickTargetCards {
		if id == theirs {
			t.Error("an opponent's creature must not be offered")
		}
	}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(martok, c)...)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{martok: 1, c: 1, a: 0, b: 0} {
		if got := counterCount(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s has %d counters, want %d", tmtB1Card(t, g, id).Name, got, want)
		}
	}
	if n := triggersOnStackFrom(g, martok); n != 0 {
		t.Errorf("%d Martok triggers left, want one trigger for the declaration", n)
	}
}

// --- Maturation Chamber ----------------------------------------------------

func TestTrekB32ChamberChargesOnArtifactSpellsAndMakesBorg(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	chamber := b27Push(g, me.ID, "Maturation Chamber", "Artifact", trekB32ChamberOracle, "{1}{W}", 0, 0, "W")
	toMainForCost(t, g)

	rock := tmtB11Hand(g, me, "Rock", "Artifact", "", "{1}", 0, 0)
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast an artifact: %v", err)
	}
	if triggerOnStack(g, chamber) == nil {
		t.Fatal("an artifact spell should trigger the Chamber")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	bear := tmtB11Hand(g, me, "Bear", "Creature — Bear", "", "{1}", 2, 2)
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast a creature: %v", err)
	}
	if triggerOnStack(g, chamber) != nil {
		t.Fatal("a nonartifact spell must not trigger the Chamber")
	}
	passPriorityAroundTable(t, g)
	addCountersForTest(g, chamber, game.CounterCharge, 1)
	if got := counterCount(g, chamber, game.CounterCharge); got != 2 {
		t.Fatalf("the Chamber has %d charge counters, want 2", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, chamber, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the ability costs {4}")
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	b16Activate(t, g, me.ID, chamber, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	borg := tmtB1Named(g, "Borg")
	if len(borg) != 2 {
		t.Fatalf("%d Borg tokens, want one per charge counter", len(borg))
	}
	if c := tmtB1Card(t, g, borg[0]); !c.IsToken() || !c.IsArtifact() || c.CurrentPower() != 1 {
		t.Errorf("Borg %+v: want a 1/1 artifact creature token", c)
	}
}

func TestTrekB32ChamberActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := trekB32Me(g)
	chamber := b27Push(g, me.ID, "Maturation Chamber", "Artifact", trekB32ChamberOracle, "{1}{W}", 0, 0, "W")
	addCountersForTest(g, chamber, game.CounterCharge, 1)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, chamber, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Chamber activated outside a main phase")
	}
}
