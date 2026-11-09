package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_mp_cards_test.go — Tarkir: Dragonstorm, cards from M to P (#4463).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tdmMarduDevoteeOracle       = "24ee57b5-a505-4264-97ca-8b8dfb12298e"
	tdmMarduMonumentOracle      = "5f662383-1684-4bd3-be05-e696d301e26c"
	tdmMarshalOfTheLostOracle   = "af762373-e159-4a36-83b5-a82b8b16fa41"
	tdmMeticulousArtisanOracle  = "2b2b3b16-389c-4581-b889-b8a2e76f83a2"
	tdmMoltenExhaleOracle       = "82aa1402-4b1c-4343-aba5-dd3271cc1ac5"
	tdmMonasteryMessengerOracle = "325b88e1-ec7f-4d22-887c-a7f18a868bb0"
	tdmNarsetWaymasterOracle    = "1a2a74b8-0689-4959-9348-02cc832194cf"
	tdmNerivOracle              = "66290104-ac65-4eb5-bf00-89b944249e07"
	tdmNightbladeBrigadeOracle  = "36d04190-7ce3-431a-b18e-3a1895ec6fc7"
	tdmOsseousExhaleOracle      = "8d84a706-080c-4fb2-b1b9-3dbb3294fb7b"
	tdmOverwhelmingSurgeOracle  = "bc8a37f1-c36b-4d0f-82da-4558ff712565"
	tdmPiercingExhaleOracle     = "ea96fe35-88ed-432b-a495-08e787d23d0b"
	tdmPoisedPractitionerOracle = "e08debdc-2db5-4562-8a02-4158bfc0db88"
)

// Molten Exhale is the one card with a caveat: its flash cast can't also
// take another alternative cost.
func TestTdmMPMoltenExhaleDeclaresItsCaveat(t *testing.T) {
	spec, ok := Lookup(tdmMoltenExhaleOracle)
	if !ok {
		t.Fatal("Molten Exhale is not registered")
	}
	if spec.Name != "Molten Exhale" || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Molten Exhale: name %q, completeness %s, caveats %v, want Caveats with one caveat",
			spec.Name, spec.Completeness, spec.Caveats)
	}
}

func TestTdmMPEveryOtherCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmMarduDevoteeOracle:       "Mardu Devotee",
		tdmMarduMonumentOracle:      "Mardu Monument",
		tdmMarshalOfTheLostOracle:   "Marshal of the Lost",
		tdmMeticulousArtisanOracle:  "Meticulous Artisan",
		tdmMonasteryMessengerOracle: "Monastery Messenger",
		tdmNarsetWaymasterOracle:    "Narset, Jeskai Waymaster",
		tdmNerivOracle:              "Neriv, Heart of the Storm",
		tdmNightbladeBrigadeOracle:  "Nightblade Brigade",
		tdmOsseousExhaleOracle:      "Osseous Exhale",
		tdmOverwhelmingSurgeOracle:  "Overwhelming Surge",
		tdmPiercingExhaleOracle:     "Piercing Exhale",
		tdmPoisedPractitionerOracle: "Poised Practitioner",
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

// --- fixtures ------------------------------------------------------------------

// tdmMPHandCast puts a card with the given mana cost in `p`'s hand and
// casts it where the game stands, without moving to a main phase.
func tdmMPHandCast(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	id := handCardOf(g, p, name, typeLine, cost, oracle)
	return id, g.CastSpell(p.ID, id, params)
}

// tdmMPDragon puts a Dragon creature on `owner`'s battlefield.
func tdmMPDragon(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Creature(g, owner, "Shivan Dragon", "Creature — Dragon", 5, 5)
}

// tdmMPCastFreeSpell casts a {0} sorcery from `p`'s hand and resolves it.
func tdmMPCastFreeSpell(t *testing.T, g *game.Game, p *game.Player, name string) {
	t.Helper()
	id := mayhemHandCard(g, p, name, "Sorcery", "{0}", "")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

// --- Mardu Devotee -------------------------------------------------------------

func TestTdmMarduDevoteeScriesTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Lower", "Upper")
	castCatalogSpell(t, g, "Mardu Devotee", "Creature — Human Scout", tdmMarduDevoteeOracle, nil)
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry on entry")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(c.ScryCards))
	}
}

func TestTdmMarduDevoteeFiltersOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	devotee := b12Push(g, me.ID, "Mardu Devotee", "Creature — Human Scout", tdmMarduDevoteeOracle, 1, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 2)
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Fatal("the Devotee made {U}; it adds {R}, {W}, or {B}")
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("{1}: add {B}: %v", err)
	}
	if got := mcPoolCount(me, "B"); got != 1 {
		t.Errorf("black in the pool %d, want 1", got)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool %d, want 2 (one {G} paid, one {B} made)", me.ManaPool.Total())
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

// --- Mardu Monument ------------------------------------------------------------

func TestTdmMarduMonumentFindsABasicMountainPlainsOrSwamp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Sacred Peaks", "Land — Mountain"),
		searchTestLand("Mountain", "Basic Land — Mountain"),
	)
	castCatalogSpell(t, g, "Mardu Monument", "Artifact", tdmMarduMonumentOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Monument's entry asked for no search")
	}
	if searchOptionNamed(g, c, "Forest") != uuid.Nil || searchOptionNamed(g, c, "Sacred Peaks") != uuid.Nil {
		t.Error("only a basic Mountain, Plains, or Swamp may be found")
	}
	answerSearchNamed(t, g, me.ID, "Mountain")
	if _, ok := handCardNamed(me, "Mountain"); !ok {
		t.Error("the Mountain is not in hand")
	}
}

func TestTdmMarduMonumentMakesThreeWarriorsWithMenaceAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := pushCatalogPermanent(g, me.ID, "Mardu Monument", "Artifact", tdmMarduMonumentOracle, false)
	advanceToMain(t, g)
	fillPoolColored(me, "R", 3)
	fillPoolColored(me, "W", 1)
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(monument) {
		t.Fatal("the Monument was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	warriors := socB8Named(g, me.ID, "Warrior")
	if len(warriors) != 3 {
		t.Fatalf("%d Warriors, want 3", len(warriors))
	}
	for _, w := range warriors {
		if w.Power != 1 || w.Toughness != 1 || len(w.Colors) != 1 || w.Colors[0] != "R" {
			t.Errorf("Warrior %d/%d %v, want a 1/1 red Warrior", w.Power, w.Toughness, w.Colors)
		}
		for _, kw := range []string{"menace", "haste"} {
			if !hasEffectiveKeyword(t, g, w.InstanceID, kw) {
				t.Errorf("a Warrior has no %s", kw)
			}
		}
	}
}

func TestTdmMarduMonumentActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := pushCatalogPermanent(g, me.ID, "Mardu Monument", "Artifact", tdmMarduMonumentOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	fillPoolColored(me, "R", 3)
	fillPoolColored(me, "W", 1)
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Monument's ability was activated in combat")
	}
}

// --- Marshal of the Lost -------------------------------------------------------

func TestTdmMarshalOfTheLostPumpsByTheNumberOfAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marshal := b12Push(g, me.ID, "Marshal of the Lost", "Creature — Orc Warrior", tdmMarshalOfTheLostOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, marshal, "deathtouch") {
		t.Error("Marshal of the Lost has no deathtouch")
	}
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Home Guard", "Creature — Soldier", 1, 1)
	declareAttack(t, g, opp.ID, marshal, a, b)
	b04WaitForPick(t, g, me.ID)
	// Any creature, attacking or not.
	pickCard(t, g, me.ID, home)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, home), effectiveToughness(t, g, home); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4 (+3/+3 for three attackers)", p, tough)
	}
	if triggersOnStackFrom(g, marshal) != 0 {
		t.Error("the trigger is still on the stack")
	}
}

func TestTdmMarshalOfTheLostTriggersWhenItStaysHome(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Marshal of the Lost", "Creature — Orc Warrior", tdmMarshalOfTheLostOracle, 3, 3)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, a); p != 3 {
		t.Errorf("the lone attacker has power %d, want 3 (+1/+1 for one attacker)", p)
	}
}

// --- Meticulous Artisan --------------------------------------------------------

func TestTdmMeticulousArtisanMakesATreasureAndHasProwess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	artisan := castCatalogSpell(t, g, "Meticulous Artisan", "Creature — Djinn Artificer", tdmMeticulousArtisanOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if !hasEffectiveKeyword(t, g, artisan, "prowess") {
		t.Error("Meticulous Artisan has no prowess")
	}
}

// --- Molten Exhale -------------------------------------------------------------

func TestTdmMoltenExhaleDealsFourAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	castCatalogSpell(t, g, "Molten Exhale", "Sorcery", tdmMoltenExhaleOracle, cardRefs(wall))
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, wall); got != 4 {
		t.Errorf("the Wall has %d damage, want 4", got)
	}
}

func TestTdmMoltenExhaleFlashesOnlyByBeholdingADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	dragon := tdmMPDragon(g, me.ID)
	advanceTo(t, g, game.StepBeginCombat)
	card := handCardOf(g, me, "Molten Exhale", "Sorcery", "{1}{R}", tdmMoltenExhaleOracle)
	fund := func() {
		g.WithWriteLock(func() { me.ManaPool = nil })
		fillPoolColored(me, "R", 2)
	}
	fund()
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: cardRefs(wall)}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("printed cost in combat: err = %v, want ErrSorcerySpeedRequired", err)
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, AlternativeCost: "flash-behold", AltCostIDs: []uuid.UUID{bear}, Targets: cardRefs(wall),
	}); err == nil {
		t.Fatal("a Bear was beheld as a Dragon")
	}
	if !me.Hand.Contains(card) {
		t.Fatal("the refused cast moved Molten Exhale out of the hand")
	}
	fund()
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, AlternativeCost: "flash-behold", AltCostIDs: []uuid.UUID{dragon}, Targets: cardRefs(wall),
	}); err != nil {
		t.Fatalf("flash cast beholding a Dragon: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left in the pool: %d, want 0 ({1}{R} paid)", len(me.ManaPool))
	}
	if !onBattlefield(g, dragon) || isTapped(g, dragon) {
		t.Error("beholding moved or tapped the Dragon")
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, wall); got != 4 {
		t.Errorf("the Wall has %d damage, want 4", got)
	}
}

func TestTdmMoltenExhaleBeholdsADragonCardFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	advanceTo(t, g, game.StepBeginCombat)
	dragonCard := handCard(me, "Dragon Friend", "Creature — Dragon")
	fillPoolColored(me, "R", 2)
	card, err := tdmMPHandCast(g, me, "Molten Exhale", "Sorcery", "{1}{R}", tdmMoltenExhaleOracle, game.CastSpellParams{
		Strict: true, AlternativeCost: "flash-behold", AltCostIDs: []uuid.UUID{dragonCard}, Targets: cardRefs(wall),
	})
	if err != nil {
		t.Fatalf("flash cast revealing a Dragon card: %v", err)
	}
	if !me.Hand.Contains(dragonCard) {
		t.Fatal("beholding moved the Dragon card out of the hand")
	}
	if got := revealedEvents(g, dragonCard); got != 1 {
		t.Errorf("%d reveal events for the beheld card, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, wall); got != 4 {
		t.Errorf("the Wall has %d damage, want 4", got)
	}
	if !me.Graveyard.Contains(card) {
		t.Error("Molten Exhale is not in the graveyard after resolving")
	}
}

// --- Monastery Messenger -------------------------------------------------------

func TestTdmMonasteryMessengerPutsANoncreatureNonlandCardOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	land := seedGraveyardCard(t, g, "Old Land", "Land", "")
	spell := seedGraveyardCard(t, g, "Old Spell", "Instant", "")
	bird := castCatalogSpell(t, g, "Monastery Messenger", "Creature — Bird Scout", tdmMonasteryMessengerOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	for name, id := range map[string]uuid.UUID{"a creature card": creature, "a land card": land} {
		if err := g.ResolvePickTargets(latestPickTarget(g, me.ID).ID, me.ID, cardRefs(id)); err == nil {
			t.Fatalf("the Messenger targeted %s", name)
		}
	}
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != spell {
		t.Fatalf("library top %q (%v), want the instant", top.Name, err)
	}
	if !me.Graveyard.Contains(creature) || !me.Graveyard.Contains(land) {
		t.Error("an untargeted card left the graveyard")
	}
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, bird, kw) {
			t.Errorf("Monastery Messenger has no %s", kw)
		}
	}
}

func TestTdmMonasteryMessengerMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Monastery Messenger", "Creature — Bird Scout", tdmMonasteryMessengerOracle, nil)
	passPriorityAroundTable(t, g)
	if c := latestPickTarget(g, me.ID); c != nil {
		pickTargetRefs(t, g, me.ID, nil)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Errorf("library %d -> %d with nothing to put back", lib, me.Library.Size())
	}
}

// --- Narset, Jeskai Waymaster --------------------------------------------------

func TestTdmNarsetDiscardsHandAndDrawsPerSpellCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three", "Four")
	pushCatalogPermanent(g, me.ID, "Narset, Jeskai Waymaster", "Legendary Creature — Human Monk", tdmNarsetWaymasterOracle, false)
	toMainForCost(t, g)
	tdmMPCastFreeSpell(t, g, me, "First")
	tdmMPCastFreeSpell(t, g, me, "Second")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	kept := []uuid.UUID{handCard(me, "Held A", "Sorcery"), handCard(me, "Held B", "Sorcery"), handCard(me, "Held C", "Sorcery")}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	for _, id := range kept {
		if !me.Graveyard.Contains(id) {
			t.Error("a card in hand was not discarded")
		}
	}
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand %d, want 2 (one card per spell cast this turn)", got)
	}
}

func TestTdmNarsetDrawsFromAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two")
	pushCatalogPermanent(g, me.ID, "Narset, Jeskai Waymaster", "Legendary Creature — Human Monk", tdmNarsetWaymasterOracle, false)
	toMainForCost(t, g)
	tdmMPCastFreeSpell(t, g, me, "Only")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand %d, want 1: an empty hand may be discarded, and the draw still happens", got)
	}
}

func TestTdmNarsetDeclinedKeepsTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two")
	pushCatalogPermanent(g, me.ID, "Narset, Jeskai Waymaster", "Legendary Creature — Human Monk", tdmNarsetWaymasterOracle, false)
	toMainForCost(t, g)
	tdmMPCastFreeSpell(t, g, me, "Only")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	held := handCard(me, "Held", "Sorcery")
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(held) || me.Hand.Size() != 1 {
		t.Errorf("hand %d, want only the held card", me.Hand.Size())
	}
}

// --- Neriv, Heart of the Storm -------------------------------------------------

// tdmMPDamage deals `n` damage from `source` to the player `to` through an
// effect `by` controls.
func tdmMPDamage(t *testing.T, g *game.Game, by, source, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: by, Owner: by})
		if err := (DealDamage{Source: source, Target: to, Amount: n}).Apply(ctx); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
}

func TestTdmNerivDoublesOnlyACreatureOfYoursThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neriv := b12Push(g, me.ID, "Neriv, Heart of the Storm", "Legendary Creature — Spirit Dragon", tdmNerivOracle, 4, 5)
	if !hasEffectiveKeyword(t, g, neriv, "flying") {
		t.Error("Neriv has no flying")
	}
	veteran := b12Creature(g, me.ID, "Veteran", "Creature — Soldier", 2, 2)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 black Zombie Druid"), 1)
		_ = g.CreateTokenForEffect(opp.ID, TokenCard("2/2 black Zombie Druid"), 1)
	})
	passPriorityAroundTable(t, g)
	mine, theirs := socB8Named(g, me.ID, "Zombie Druid"), socB8Named(g, opp.ID, "Zombie Druid")
	if len(mine) != 1 || len(theirs) != 1 {
		t.Fatalf("tokens: %d mine, %d theirs, want one each", len(mine), len(theirs))
	}

	life := opp.Life
	tdmMPDamage(t, g, me.ID, mine[0].InstanceID, opp.ID, 2)
	if got := life - opp.Life; got != 4 {
		t.Errorf("a creature of yours that entered this turn dealt %d, want 4", got)
	}
	life = opp.Life
	tdmMPDamage(t, g, me.ID, veteran, opp.ID, 2)
	if got := life - opp.Life; got != 2 {
		t.Errorf("a creature that did not enter this turn dealt %d, want 2", got)
	}
	myLife := me.Life
	tdmMPDamage(t, g, opp.ID, theirs[0].InstanceID, me.ID, 2)
	if got := myLife - me.Life; got != 2 {
		t.Errorf("an opponent's creature that entered this turn dealt %d, want 2", got)
	}
}

// --- Nightblade Brigade --------------------------------------------------------

func TestTdmNightbladeBrigadeSurveilsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	brigade := castCatalogSpell(t, g, "Nightblade Brigade", "Creature — Goblin Soldier", tdmNightbladeBrigadeOracle, nil)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil on entry")
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	if !hasEffectiveKeyword(t, g, brigade, "deathtouch") {
		t.Error("Nightblade Brigade has no deathtouch")
	}
}

func TestTdmNightbladeBrigadeMobilizesOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	brigade := b12Push(g, me.ID, "Nightblade Brigade", "Creature — Goblin Soldier", tdmNightbladeBrigadeOracle, 1, 3)
	declareAttack(t, g, opp.ID, brigade)
	passPriorityAroundTable(t, g)
	tdmPlaceWarriors(t, g, me.ID)
	warriors := socB8Named(g, me.ID, "Warrior")
	if len(warriors) != 1 {
		t.Fatalf("mobilize 1 made %d Warriors, want 1", len(warriors))
	}
	if w := warriors[0]; !w.Tapped || w.AttackingTarget == uuid.Nil {
		t.Errorf("Warrior tapped=%v attacking=%v, want tapped and attacking", w.Tapped, w.AttackingTarget != uuid.Nil)
	}
}

// --- Osseous Exhale ------------------------------------------------------------

// tdmMPAttackingBigCreature has seat 0 attack seat 1 with a 2/7 and
// returns it, with the defenders done blocking (nobody blocks) and seat 0
// holding priority.
func tdmMPAttackingBigCreature(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	big := b12Creature(g, g.Seats[0].ID, "Big Attacker", "Creature — Giant", 2, 7)
	declareAttack(t, g, g.Seats[1].ID, big)
	finishBlockDeclarations(t, g)
	return big
}

func TestTdmOsseousExhaleBeheldDealsFiveAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := tdmMPDragon(g, me.ID)
	big := tdmMPAttackingBigCreature(t, g)
	life := me.Life
	if _, err := tdmMPHandCast(g, me, "Osseous Exhale", "Instant", "", tdmOsseousExhaleOracle, game.CastSpellParams{
		OptionalCosts: []int{beholdOfferIdx}, RevealIDs: []uuid.UUID{dragon}, Targets: cardRefs(big),
	}); err != nil {
		t.Fatalf("cast beholding a Dragon: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, big); got != 5 {
		t.Errorf("the attacker has %d damage, want 5", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want 2 gained for the beheld Dragon", life, me.Life)
	}
}

func TestTdmOsseousExhaleWithoutBeholdGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := tdmMPAttackingBigCreature(t, g)
	life := me.Life
	if _, err := tdmMPHandCast(g, me, "Osseous Exhale", "Instant", "", tdmOsseousExhaleOracle, game.CastSpellParams{
		Targets: cardRefs(big),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, big); got != 5 {
		t.Errorf("the attacker has %d damage, want 5", got)
	}
	if me.Life != life {
		t.Errorf("life %d -> %d, want no change without a behold", life, me.Life)
	}
}

func TestTdmOsseousExhaleRefusesACreatureOutOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Osseous Exhale", "Instant", tdmOsseousExhaleOracle, cardRefs(idle)); err == nil {
		t.Error("Osseous Exhale targeted a creature that is neither attacking nor blocking")
	}
}

// --- Overwhelming Surge --------------------------------------------------------

func TestTdmOverwhelmingSurgeBothModes(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	castModal(t, g, "Overwhelming Surge", "Instant", tdmOverwhelmingSurgeOracle, []int{0, 1},
		[]game.TargetRef{modeRef(game.TargetCard, wall, 0, 0), modeRef(game.TargetCard, relic, 1, 0)})
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, wall); got != 3 {
		t.Errorf("the Wall has %d damage, want 3", got)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not destroyed")
	}
}

func TestTdmOverwhelmingSurgeDestroysOnlyANoncreatureArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 3, 5)
	if _, err := castModalErr(t, g, "Overwhelming Surge", "Instant", tdmOverwhelmingSurgeOracle, []int{1}, cardRefs(golem)); err == nil {
		t.Fatal("the destroy mode targeted an artifact creature")
	}
	castModal(t, g, "Overwhelming Surge", "Instant", tdmOverwhelmingSurgeOracle, []int{0}, cardRefs(golem))
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, golem); got != 3 {
		t.Errorf("the Golem has %d damage, want 3", got)
	}
}

// --- Piercing Exhale -----------------------------------------------------------

func TestTdmPiercingExhaleBitesAndSurveilsWhenBeheld(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Wall", 0, 6)
	dragonCard := handCard(me, "Dragon Friend", "Creature — Dragon")
	if _, err := castWithTapParams(t, g, "Piercing Exhale", "Instant", "{1}{G}", tdmPiercingExhaleOracle, game.CastSpellParams{
		OptionalCosts: []int{beholdOfferIdx}, RevealIDs: []uuid.UUID{dragonCard}, Targets: slotted(mine, foe),
	}); err != nil {
		t.Fatalf("cast beholding a Dragon card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 3 {
		t.Errorf("the foe has %d damage, want 3 (the biter's power)", got)
	}
	if got := pr6Marked(g, mine); got != 0 {
		t.Errorf("the biter took %d damage; the bite is one-sided", got)
	}
	if c := surveilChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Errorf("surveil prompt %+v, want surveil 2 for the beheld Dragon", c)
	}
}

func TestTdmPiercingExhaleWithoutBeholdDoesNotSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Wall", 0, 6)
	if _, err := castWithTapParams(t, g, "Piercing Exhale", "Instant", "{1}{G}", tdmPiercingExhaleOracle, game.CastSpellParams{
		Targets: slotted(mine, foe),
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := pr6Marked(g, foe); got != 3 {
		t.Errorf("the foe has %d damage, want 3", got)
	}
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("surveiled without beholding a Dragon")
	}
}

func TestTdmPiercingExhaleTheBiterMustBeYours(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Wall", 0, 6)
	if err := castCatalogSpellErr(t, g, "Piercing Exhale", "Instant", tdmPiercingExhaleOracle, slotted(theirs, foe)); err == nil {
		t.Error("an opponent's creature was named as the biter")
	}
}

// --- Poised Practitioner -------------------------------------------------------

func TestTdmPoisedPractitionerGrowsAndScriesOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two")
	monk := b12Push(g, me.ID, "Poised Practitioner", "Creature — Human Monk", tdmPoisedPractitionerOracle, 2, 3)
	toMainForCost(t, g)
	tdmMPCastFreeSpell(t, g, me, "First")
	if got := countersOn(g, monk, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters after the first spell, want 0", got)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("scried on the first spell")
	}
	tdmMPCastFreeSpell(t, g, me, "Second")
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry on the second spell")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, monk, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after the second spell, want 1", got)
	}
	tdmMPCastFreeSpell(t, g, me, "Third")
	if got := countersOn(g, monk, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after the third spell, want still 1", got)
	}
}
