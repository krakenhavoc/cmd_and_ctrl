package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// grist_the_hunger_tide_test.go — Grist, the Hunger Tide (#5427): the
// "isn't on the battlefield" sentence in every zone (Spec.OffBattlefield,
// game/off_battlefield_self.go), and the three loyalty abilities.

const (
	gristOracle   = "0efb0d7e-dea0-4817-a243-15066e9ef333"
	gristTypeLine = "Legendary Planeswalker — Grist"
)

// gristCard is a Grist card for `owner`, as the importer stamps it: no
// printed power or toughness.
func gristCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Grist, the Hunger Tide", TypeLine: gristTypeLine,
		ManaCost: "{1}{B}{G}", OracleID: gristOracle, Colors: []string{"B", "G"},
		Owner: owner, Controller: owner,
	}
}

// gristTable seeds Grist on seat 0's battlefield with `loyalty`, in a
// main phase, over an EMPTY library, so a test stacks the cards the +1
// mills.
func gristTable(t *testing.T, loyalty int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	me.Library.Cards = nil
	grist := pushBattlefieldCardWithTimestamp(g, gristCard(me.ID))
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(grist, game.CounterLoyalty, loyalty) })
	return g, me, grist
}

func gristLookup(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	// ReadSnapshot: the reader's freshness contract (a stale layer pass
	// is recomputed first), as the view reads a card.
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s is nowhere", id)
	}
	return c
}

// requireGristCreatureCard asserts the off-battlefield reading: a 1/1
// Insect creature in addition to a legendary Grist planeswalker.
func requireGristCreatureCard(t *testing.T, where string, c game.Card) {
	t.Helper()
	if !c.IsCreature() || !c.IsPlaneswalker() {
		t.Errorf("%s: creature %v, planeswalker %v; want both", where, c.IsCreature(), c.IsPlaneswalker())
	}
	if !c.HasSubtype("Insect") || !c.HasSubtype("Grist") {
		t.Errorf("%s: Insect %v, Grist %v; want both", where, c.HasSubtype("Insect"), c.HasSubtype("Grist"))
	}
	if !c.HasSupertype("legendary") {
		t.Errorf("%s: not legendary", where)
	}
	if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 1 || tough != 1 {
		t.Errorf("%s: %d/%d, want 1/1", where, p, tough)
	}
	if c.HasSubtype("Elf") {
		t.Errorf("%s: an Elf with no effect making it one", where)
	}
	eff := c.Effective()
	if len(eff.Types) != 2 || len(eff.Subtypes) != 2 {
		t.Errorf("%s: types %v subtypes %v, want Planeswalker Creature — Grist Insect", where, eff.Types, eff.Subtypes)
	}
}

func TestGristIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gristOracle)
	if !ok {
		t.Fatal("Grist, the Hunger Tide is not registered")
	}
	if spec.Name != "Grist, the Hunger Tide" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// CR 113.6c: in every zone but the battlefield Grist is a 1/1 Insect
// creature card as well as a planeswalker card.
func TestGristIsACreatureCardInEveryZoneOffTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zones := map[string]*game.Zone{
		"hand": me.Hand, "library": me.Library, "graveyard": me.Graveyard,
		"command zone": me.Command, "exile": g.Exile,
	}
	for where, z := range zones {
		c := gristCard(me.ID)
		z.PushTop(c)
		requireGristCreatureCard(t, where, gristLookup(t, g, c.InstanceID))
	}
}

// On the battlefield the sentence is off: a planeswalker with loyalty and
// no power, toughness or creature type, before and after the layer pass.
func TestGristIsOnlyAPlaneswalkerOnTheBattlefield(t *testing.T) {
	g, _, grist := gristTable(t, 3)
	for _, c := range []game.Card{gristLookup(t, g, grist), layeredCard(t, g, grist)} {
		if c.IsCreature() || c.HasSubtype("Insect") {
			t.Errorf("on the battlefield: creature %v, Insect %v; want neither", c.IsCreature(), c.HasSubtype("Insect"))
		}
		if !c.IsPlaneswalker() || !c.HasSubtype("Grist") {
			t.Error("on the battlefield Grist is a Grist planeswalker")
		}
	}
}

// Cast from hand it is a creature spell: "whenever you cast a creature
// spell" sees it. Then it resolves as a planeswalker with 3 loyalty, and "whenever
// another creature enters" does not see a creature enter.
func TestGristIsACreatureSpellAndEntersAsAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Beast Whisperer", "Creature — Elf Druid", beastWhispererOracle, false)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", b04SoulWardenOracle, false)
	advanceToMain(t, g)
	c := gristCard(me.ID)
	me.Hand.PushTop(c)
	hand, life := me.Hand.Size(), me.Life

	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grist: %v", err)
	}
	requireGristCreatureCard(t, "stack", gristLookup(t, g, c.InstanceID))
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand-1+1 {
		t.Errorf("hand %d → %d: Beast Whisperer draws for a creature spell", hand, got)
	}
	perm, ok := battlefieldCard(g, c.InstanceID)
	if !ok {
		t.Fatal("Grist did not resolve onto the battlefield")
	}
	if perm.IsCreature() || !perm.IsPlaneswalker() {
		t.Errorf("resolved: creature %v, planeswalker %v", perm.IsCreature(), perm.IsPlaneswalker())
	}
	if got := countersOn(g, c.InstanceID, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
	if me.Life != life {
		t.Errorf("life %d → %d: Soul Warden saw a creature enter, and Grist is not one on the battlefield", life, me.Life)
	}
}

// "Counter target creature spell" may target Grist on the stack, and a
// noncreature planeswalker spell is still refused.
func TestGristOnTheStackIsCounteredAsACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	other := game.Card{
		InstanceID: uuid.New(), Name: "Plain Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID,
	}
	me.Hand.PushTop(other)
	if err := g.CastSpell(me.ID, other.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the plain planeswalker: %v", err)
	}
	scatter := handCardFull(opp, "Essence Scatter", "Instant", "{1}{U}", b25EssenceScatterOracle, nil)
	if err := g.CastSpell(opp.ID, scatter, game.CastSpellParams{Targets: b16TargetCard(other.InstanceID)}); err == nil {
		t.Fatal("a planeswalker spell that is not a creature spell is refused as a target")
	}
	passPriorityAroundTable(t, g)

	c := gristCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grist: %v", err)
	}
	if err := g.CastSpell(opp.ID, scatter, game.CastSpellParams{Targets: b16TargetCard(c.InstanceID)}); err != nil {
		t.Fatalf("Essence Scatter on Grist: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c.InstanceID) || !me.Graveyard.Contains(c.InstanceID) {
		t.Error("Grist was not countered")
	}
	requireGristCreatureCard(t, "graveyard after the counter", gristLookup(t, g, c.InstanceID))
}

// --- +1 ------------------------------------------------------------------

// gristStack puts cards on top of the library, the first argument on top.
func gristStack(me *game.Player, cards ...game.Card) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		if c.InstanceID == uuid.Nil {
			c.InstanceID = uuid.New()
		}
		c.Owner, c.Controller = me.ID, me.ID
		ids[i] = c.InstanceID
		me.Library.PushTop(c)
	}
	return ids
}

func gristInsectTokens(t *testing.T, g *game.Game, controller uuid.UUID) int {
	t.Helper()
	n := 0
	for _, c := range g.Battlefield.Cards {
		if !IsToken(c) || c.Name != "Insect" || c.Controller != controller {
			continue
		}
		n++
		l := layeredCard(t, g, c.InstanceID)
		cols := l.EffectiveColors()
		if l.CurrentPower() != 1 || l.CurrentToughness() != 1 || !l.HasSubtype("Insect") ||
			len(cols) != 2 || cols[0] != "B" || cols[1] != "G" || len(l.Effective().Abilities) != 0 {
			t.Errorf("token is %d/%d %v %v, want a 1/1 black and green Insect with no abilities",
				l.CurrentPower(), l.CurrentToughness(), cols, l.Effective().Abilities)
		}
	}
	return n
}

// A card that is not an Insect: one token, one card milled, and only the
// cost's counter.
func TestGristPlusOneMakesAnInsectAndMillsACard(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	ids := gristStack(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"}, game.Card{Name: "Under", TypeLine: "Sorcery"})
	b16Activate(t, g, me.ID, grist, 0, game.ActivateAbilityParams{})
	if got := gristInsectTokens(t, g, me.ID); got != 1 {
		t.Errorf("Insect tokens %d, want 1", got)
	}
	if !me.Graveyard.Contains(ids[0]) || !me.Library.Contains(ids[1]) {
		t.Error("exactly the top card is milled")
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
}

// An Insect card milled: a loyalty counter, and the whole process again,
// until a card that is not one. Another Grist is an Insect card in the
// graveyard, by its own first sentence.
func TestGristPlusOneRepeatsWhileItMillsInsectCards(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	second := gristCard(me.ID)
	ids := gristStack(me,
		game.Card{Name: "Beetle", TypeLine: "Creature — Insect"},
		second,
		game.Card{Name: "Hive", TypeLine: "Kindred Enchantment — Insect"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Under", TypeLine: "Creature — Insect"},
	)
	b16Activate(t, g, me.ID, grist, 0, game.ActivateAbilityParams{})
	if got := gristInsectTokens(t, g, me.ID); got != 4 {
		t.Errorf("Insect tokens %d, want 4 (three repeats)", got)
	}
	for i, id := range ids[:4] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %d was not milled", i)
		}
	}
	if !me.Library.Contains(ids[4]) {
		t.Error("the process stopped after the Bear; the card under it stays")
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 3+1+3 {
		t.Errorf("loyalty %d, want 7 (the cost's one and three Insect cards)", got)
	}
}

// An empty library: the token, no card, no repeat.
func TestGristPlusOneOnAnEmptyLibraryMakesOneToken(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	b16Activate(t, g, me.ID, grist, 0, game.ActivateAbilityParams{})
	if got := gristInsectTokens(t, g, me.ID); got != 1 {
		t.Errorf("Insect tokens %d, want 1", got)
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	if me.Eliminated {
		t.Error("milling an empty library is not a draw")
	}
}

// CR 701.17c: under Rest in Peace the milled Insect card is exiled
// instead, and is still the milled card, read in exile.
func TestGristPlusOneReadsAMilledCardRestInPeaceExiled(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	pushCatalogPermanent(g, me.ID, "Rest in Peace", "Enchantment", restInPeaceOracle, false)
	ids := gristStack(me,
		game.Card{Name: "Beetle", TypeLine: "Creature — Insect"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	b16Activate(t, g, me.ID, grist, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) {
		t.Fatal("setup: Rest in Peace exiles the milled cards")
	}
	if got := gristInsectTokens(t, g, me.ID); got != 2 {
		t.Errorf("Insect tokens %d, want 2", got)
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

// Grist destroyed in response: no counter lands, and the process still
// repeats, because the repeat is not about Grist.
func TestGristPlusOneRepeatsAfterGristHasLeft(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	gristStack(me,
		game.Card{Name: "Beetle", TypeLine: "Creature — Insect"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	if err := g.ActivateCatalogAbility(me.ID, grist, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(grist) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(grist) {
		t.Fatal("setup: Grist was destroyed in response")
	}
	if got := gristInsectTokens(t, g, me.ID); got != 2 {
		t.Errorf("Insect tokens %d, want 2", got)
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 0 {
		t.Errorf("the card in the graveyard has %d loyalty counters", got)
	}
}

// --- −2 ------------------------------------------------------------------

// The sacrifice happens as the ability resolves, and only then does the
// reflexive trigger choose its target: a creature or a planeswalker.
func TestGristMinusTwoSacrificesThenDestroysATarget(t *testing.T) {
	for _, typeLine := range []string{"Creature — Rat", "Legendary Planeswalker — Test"} {
		t.Run(typeLine, func(t *testing.T) {
			g, me, grist := gristTable(t, 3)
			opp := g.Seats[1]
			fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 4, 4)
			victim := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Victim", TypeLine: typeLine, Power: 1, Toughness: 2,
				Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{},
			})
			if typeLine != "Creature — Rat" {
				g.WithWriteLock(func() { _ = g.AddCounterForEffect(victim, game.CounterLoyalty, 4) })
			}
			if err := g.ActivateCatalogAbility(me.ID, grist, 1, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if latestPickTarget(g, me.ID) != nil {
				t.Fatal("the −2 has no target of its own")
			}
			passPriorityAroundTable(t, g)
			answerOwnPermanents(t, g, me.ID, fodder)
			b04WaitForPick(t, g, me.ID)
			if g.Battlefield.Contains(fodder) {
				t.Fatal("the sacrifice happens before the reflexive trigger asks for a target")
			}
			pickCard(t, g, me.ID, victim)
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("the target was not destroyed")
			}
			if got := countersOn(g, grist, game.CounterLoyalty); got != 1 {
				t.Errorf("loyalty %d, want 1", got)
			}
		})
	}
}

// "You may": declining sacrifices nothing and raises no trigger.
func TestGristMinusTwoDeclinedDoesNothing(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	opp := g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	if err := g.ActivateCatalogAbility(me.ID, grist, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no \"when you do\"")
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("declining leaves both creatures")
	}
}

// With no creature to sacrifice there is no question and no trigger.
func TestGristMinusTwoWithNoCreatureAsksNothing(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	victim := b12Creature(g, g.Seats[1].ID, "Victim", "Creature — Rat", 1, 2)
	b16Activate(t, g, me.ID, grist, 1, game.ActivateAbilityParams{})
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil || latestPickTarget(g, me.ID) != nil {
		t.Error("nothing to sacrifice: nothing is asked")
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("the opponent's creature is untouched")
	}
}

// --- −5 ------------------------------------------------------------------

// The count is the creature cards in the graveyard as the ability
// resolves, and a Grist card there is one of them.
func TestGristMinusFiveCountsCreatureCardsInYourGraveyard(t *testing.T) {
	g, me, grist := gristTable(t, 6)
	seedGraveyardCreature(me, "Zombie A", "{4}{B}")
	seedGraveyardCreature(me, "Zombie B", "{4}{B}")
	me.Graveyard.PushTop(gristCard(me.ID))
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	seedGraveyardCreature(g.Seats[1], "Theirs", "{4}{B}")
	before := opponentLives(g, me.ID)
	life := me.Life
	b16Activate(t, g, me.ID, grist, 2, game.ActivateAbilityParams{})
	assertOpponentsLost(t, g, me.ID, before, 3)
	if me.Life != life {
		t.Errorf("my life %d → %d", life, me.Life)
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// Loyalty 3 cannot pay −5.
func TestGristMinusFiveIsRefusedAtThreeLoyalty(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	if err := g.ActivateCatalogAbility(me.ID, grist, 2, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("−5 at 3 loyalty is refused")
	}
	if got := countersOn(g, grist, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d after a refused activation", got)
	}
}

// --- the first sentence, through other cards -------------------------------

// Reanimate's "target creature card from a graveyard" takes Grist, which
// returns as a planeswalker with its printed loyalty; a planeswalker card
// without the sentence is refused.
func TestGristInAGraveyardIsALegalReanimateTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plain := game.Card{InstanceID: uuid.New(), Name: "Plain Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{2}", Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(plain)
	c := gristCard(me.ID)
	me.Graveyard.PushTop(c)
	if err := castCatalogSpellErr(t, g, "Reanimate", "Sorcery", reanimateOracle, b16TargetCard(plain.InstanceID)); err == nil {
		t.Fatal("a planeswalker card that is not a creature card is refused")
	}
	life := me.Life
	castCatalogSpell(t, g, "Reanimate", "Sorcery", reanimateOracle, b16TargetCard(c.InstanceID))
	passPriorityAroundTable(t, g)
	perm, ok := battlefieldCard(g, c.InstanceID)
	if !ok {
		t.Fatal("Grist was not returned to the battlefield")
	}
	if perm.IsCreature() || !perm.IsPlaneswalker() {
		t.Errorf("returned: creature %v, planeswalker %v", perm.IsCreature(), perm.IsPlaneswalker())
	}
	if got := countersOn(g, c.InstanceID, game.CounterLoyalty); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want -3 (Grist's mana value)", life, me.Life)
	}
}

// Worldly Tutor's "creature card" search offers Grist in the library.
func TestGristInALibraryIsFoundByACreatureTutor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := gristStack(me, gristCard(me.ID), game.Card{Name: "Plain Walker", TypeLine: "Legendary Planeswalker — Test"})
	castCatalogSpell(t, g, "Worldly Tutor", "Instant", worldlyTutorOracle, nil)
	passPriorityAroundTable(t, g)
	ch := searchChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.SearchCards {
		offered[id] = true
	}
	if !offered[ids[0]] || offered[ids[1]] {
		t.Errorf("offered Grist %v, the plain planeswalker %v; want only Grist", offered[ids[0]], offered[ids[1]])
	}
}

// CR 603.10a: a dies trigger looks back at the permanent, and Grist on
// the battlefield was not a creature, though the card that arrives in the
// graveyard is one.
func TestGristLeavingTheBattlefieldIsNotACreatureDying(t *testing.T) {
	g, me, grist := gristTable(t, 3)
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(grist) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || me.Life != life {
		t.Error("Blood Artist triggered on a planeswalker going to the graveyard")
	}
	requireGristCreatureCard(t, "graveyard", gristLookup(t, g, grist))
}

// CR 613.8a: Maskwood Nexus' "creature cards you own that aren't on the
// battlefield" depends on Grist's own sentence, so Grist in a hand is
// every creature type whichever came first; on the battlefield the Nexus
// reads "creatures you control" and Grist is not one.
func TestGristInHandIsEveryCreatureTypeUnderMaskwoodNexus(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c := gristCard(me.ID)
	me.Hand.PushTop(c)
	pushNexusFor(g, me.ID)
	inHand := gristLookup(t, g, c.InstanceID)
	if !inHand.HasSubtype("Elf") || !inHand.HasSubtype("Insect") || !inHand.IsCreature() {
		t.Errorf("in hand under the Nexus: Elf %v, Insect %v, creature %v", inHand.HasSubtype("Elf"), inHand.HasSubtype("Insect"), inHand.IsCreature())
	}
	if p, tough := inHand.CurrentPower(), inHand.CurrentToughness(); p != 1 || tough != 1 {
		t.Errorf("in hand under the Nexus: %d/%d, want 1/1", p, tough)
	}
	g2, me2, grist := gristTable(t, 3)
	pushNexusFor(g2, me2.ID)
	if perm := layeredCard(t, g2, grist); perm.HasSubtype("Elf") || perm.IsCreature() {
		t.Error("on the battlefield Grist is not a creature, so the Nexus gives it no creature type")
	}
}

// "Creatures your opponents control enter tapped" reads the permanent as
// it will exist on the battlefield (CR 614.12), and there Grist is not a
// creature: it enters untapped, and "whenever a creature an opponent
// controls enters" does not trigger.
func TestGristDoesNotEnterAsACreatureUnderAuthorityOfTheConsuls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Authority of the Consuls", "Enchantment", authorityOfTheConsulsOracle, false)
	advanceToMain(t, g)
	c := gristCard(me.ID)
	me.Hand.PushTop(c)
	life := opp.Life
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Grist: %v", err)
	}
	passPriorityAroundTable(t, g)
	perm, ok := battlefieldCard(g, c.InstanceID)
	if !ok {
		t.Fatal("Grist did not resolve")
	}
	if perm.Tapped {
		t.Error("Grist entered tapped: it is not a creature on the battlefield")
	}
	if opp.Life != life {
		t.Errorf("Authority of the Consuls gained %d life for a planeswalker entering", opp.Life-life)
	}
}
