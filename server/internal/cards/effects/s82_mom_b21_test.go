package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	momB21ErrantOracle    = "24697554-02fa-48ba-b51d-6bba3f85c293"
	momB21DjeruOracle     = "2c0061ab-4123-4b0c-8286-3207e8f9497d"
	momB21KatildaOracle   = "1990012e-2a92-4e76-917a-f90fb7a4ed66"
	momB21ZimoneOracle    = "5bc66b18-22ac-4138-b527-fa711116e298"
	momB21HidetsuguOracle = "18591d9e-c9f5-4ae8-aee1-8580bc8fa600"
	momB21RashmiOracle    = "bdef246c-cf66-40a1-aae4-a591846e73ca"
	momB21ElendaOracle    = "4b9f463f-3786-4374-bf77-739ae162b3ba"
	momB21IngaOracle      = "3431ee42-c34a-4972-9488-41b83ba5f943"
	momB21OmnathOracle    = "26445dc7-8363-4205-aee8-1cafeb4ba4c0"
)

func TestMomB21Catalogued(t *testing.T) {
	for oracle, name := range map[string]string{
		momB21ErrantOracle:    "Errant and Giada",
		momB21DjeruOracle:     "Djeru and Hazoret",
		momB21KatildaOracle:   "Katilda and Lier",
		momB21ZimoneOracle:    "Zimone and Dina",
		momB21HidetsuguOracle: "Hidetsugu and Kairi",
		momB21RashmiOracle:    "Rashmi and Ragavan",
		momB21ElendaOracle:    "Elenda and Azor",
		momB21IngaOracle:      "Inga and Esika",
		momB21OmnathOracle:    "Omnath, Locus of All",
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

func TestMomB21ErrantCastsFlashOrFlyingFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	b12Push(g, me.ID, "Errant and Giada", "Legendary Creature — Human Angel", momB21ErrantOracle, 2, 3)
	if !topVisibleTo(g, me.ID, me.ID) || topVisibleTo(g, me.ID, opp.ID) {
		t.Error("the look should be private to the controller")
	}
	plain := seedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Power: 2, Toughness: 2})
	if grantedPermissionOn(g, me.ID, plain, game.ZoneLibrary).Granted() {
		t.Error("a plain creature on top is open")
	}
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a plain creature was cast off the top")
	}
	land := seedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if grantedPermissionOn(g, me.ID, land, game.ZoneLibrary).Granted() {
		t.Error("a land on top is open")
	}
	flier := seedTopOfLibrary(me, game.Card{Name: "Top Bird", TypeLine: "Creature — Bird", ManaCost: "{0}", Keywords: []string{"flying"}, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, flier, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("a flying creature off the top: %v", err)
	}
	flash := seedTopOfLibrary(me, game.Card{Name: "Top Trick", TypeLine: "Instant", ManaCost: "{0}", Keywords: []string{"flash"}})
	if !grantedPermissionOn(g, me.ID, flash, game.ZoneLibrary).Granted() {
		t.Error("a flash spell on top is not open")
	}
	oppTop := seedTopOfLibrary(opp, game.Card{Name: "Opp Bird", TypeLine: "Creature — Bird", ManaCost: "{0}", Keywords: []string{"flying"}})
	if grantedPermissionOn(g, me.ID, oppTop, game.ZoneLibrary).Granted() {
		t.Error("an opponent's library top is open")
	}
	_ = uuid.Nil
}

// --- Djeru and Hazoret ---------------------------------------------------------

func momB21HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return hasKeywordOnBattlefield(t, g, id, kw)
}

func TestMomB21DjeruHasVigilanceAndHasteWithOneOrFewerCardsInHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	djeru := b12Push(g, me.ID, "Djeru and Hazoret", "Legendary Creature — Human God", momB21DjeruOracle, 5, 4)
	me.Hand.Cards = nil
	for n := 0; n <= 3; n++ {
		want := n <= 1
		if got := momB21HasKeyword(t, g, djeru, "vigilance") && momB21HasKeyword(t, g, djeru, "haste"); got != want {
			t.Errorf("%d cards in hand: vigilance and haste = %v, want %v", n, got, want)
		}
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
}

func momB21Legend(p *game.Player, name string) uuid.UUID {
	return pushLibraryCardForTest(p, game.Card{Name: name, TypeLine: "Legendary Creature — Human", ManaCost: "{5}{G}{G}", Power: 3, Toughness: 3})
}

func TestMomB21DjeruAttackExilesALegendaryCreatureAndCastsItFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	djeru := b12Push(g, me.ID, "Djeru and Hazoret", "Legendary Creature — Human God", momB21DjeruOracle, 5, 4)
	me.Library.Cards = nil
	// pushLibraryCardForTest pushes to the bottom: the first pushed is on top.
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Basic Land — Forest"})
	}
	legend := momB21Legend(me, "Legend A")
	plain := pushLibraryCardForTest(me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Deep", TypeLine: "Basic Land — Forest"})
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(djeru, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no look-and-exile prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != legend || prompt.ChooseMin != 0 {
		t.Fatalf("offered %v (min %d), want only the legendary creature, optional", prompt.ChooseCards, prompt.ChooseMin)
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{legend}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Exile.Contains(legend) {
		t.Fatal("the legendary creature was not exiled")
	}
	if inLibrary(g, me.ID, plain) != true {
		t.Error("the non-legendary card must stay in the library (bottom)")
	}
	if c, err := me.Library.Top(); err == nil && (c.InstanceID == plain || c.InstanceID == legend) {
		t.Error("a looked-at card is still on top")
	}
	if me.Library.Size() != 9 {
		t.Errorf("library %d, want 9", me.Library.Size())
	}
	// An ordinary cast keeps its timing: a creature waits for the second main phase.
	if err := g.CastSpell(me.ID, legend, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a creature was cast free at instant speed")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if err := g.CastSpell(me.ID, legend, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("the exiled legendary creature can't be cast free this turn: %v", err)
	}
}

func TestMomB21DjeruAttackWithNoLegendPutsAllSixOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	djeru := b12Push(g, me.ID, "Djeru and Hazoret", "Legendary Creature — Human God", momB21DjeruOracle, 5, 4)
	me.Library.Cards = nil
	var six []uuid.UUID
	for i := 0; i < 6; i++ {
		six = append(six, pushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Creature — Bear", Power: 1, Toughness: 1}))
	}
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Deep", TypeLine: "Basic Land — Forest"})
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(djeru, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if latestChooseCards(g, me.ID) != nil {
		t.Fatal("asked to pick with no legendary creature among them")
	}
	if me.Library.Size() != 10 {
		t.Fatalf("library %d, want 10", me.Library.Size())
	}
	top, _ := me.Library.Top()
	for _, id := range six {
		if top.InstanceID == id {
			t.Error("one of the six looked-at cards is still on top")
		}
	}
}

// --- Katilda and Lier ------------------------------------------------------------

func TestMomB21KatildaGrantsFlashbackOnAHumanSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b12Push(g, me.ID, "Katilda and Lier", "Legendary Creature — Human", momB21KatildaOracle, 3, 3)
	bolt := seedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-b21-bolt")
	bear := seedGraveyardCard(t, g, "Filler Bear", "Creature — Bear", "test-b21-bear")

	// A non-Human spell: no trigger.
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-Human spell triggered Katilda and Lier")
	}
	passPriorityAroundTable(t, g)
	if grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Fatal("flashback granted without a Human spell")
	}

	// A Human spell: the trigger targets an instant or sorcery card only.
	castCatalogSpell(t, g, "Human Bear", "Creature — Human", "", nil)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("a Human spell did not trigger Katilda and Lier")
	}
	for _, id := range pick.PickTargetCards {
		if id == bear {
			t.Error("a creature card in the graveyard is offered as a target")
		}
	}
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || !perm.ExileOnResolution || perm.Cost != "" {
		t.Fatalf("flashback grant = %+v", perm)
	}
	if grantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard).Granted() {
		t.Error("the creature card gained flashback")
	}
}

// --- Zimone and Dina ----------------------------------------------------------------

func TestMomB21ZimoneDrainsOnTheSecondDrawEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	b12Push(g, me.ID, "Zimone and Dina", "Legendary Creature — Human Dryad", momB21ZimoneOracle, 3, 4)
	if n := g.TurnTallyFor(me.ID).CardsDrawn; n != 1 {
		t.Fatalf("draw tally %d at main phase, want 1 (the turn's draw)", n)
	}
	life, oppLife := me.Life, opp.Life
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != life+2 {
		t.Fatalf("life: opp %d→%d, me %d→%d, want -2 and +2", oppLife, opp.Life, life, me.Life)
	}
	// The third card does not trigger it.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the third draw triggered Zimone and Dina")
	}
}

func momB21Lands(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		b12Permanent(g, p.ID, "Forest", "Basic Land — Forest")
	}
}

func TestMomB21ZimoneSacrificeDrawsAndPutsALandThenRepeatsWithEightLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	zim := b12Push(g, me.ID, "Zimone and Dina", "Legendary Creature — Human Dryad", momB21ZimoneOracle, 3, 4)
	momB21Lands(g, me, 7)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	me.Hand.Cards = nil
	me.Library.Cards = nil
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Library Forest", TypeLine: "Basic Land — Forest"})
	}
	b16Activate(t, g, me.ID, zim, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	passPriorityAroundTable(t, g)
	// Draw 1, then the land prompt; seven lands + the put = eight, so repeat.
	if me.Hand.Size() != 1 {
		t.Fatalf("hand %d after the first draw, want 1", me.Hand.Size())
	}
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 {
		t.Fatal("no optional land prompt after the first draw")
	}
	answerChooseCards(t, g, me.ID, prompt.ChooseCards[0])
	if me.Hand.Size() != 1 {
		t.Fatalf("hand %d after the land and the repeat's draw, want 1", me.Hand.Size())
	}
	prompt = latestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("eight lands did not repeat the process")
	}
	answerChooseCards(t, g, me.ID, prompt.ChooseCards[0])
	if got := countControlled(g, me.ID, func(c game.Card) bool { return c.IsLand() }); got != 9 {
		t.Errorf("lands %d, want 9", got)
	}
	if me.Library.Size() != 2 {
		t.Errorf("library %d, want 2 (two draws)", me.Library.Size())
	}
	// The repeat happens once only: no third prompt.
	if latestChooseCards(g, me.ID) != nil {
		t.Error("the process repeated a second time")
	}
}

func TestMomB21ZimoneDoesNotRepeatBelowEightLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	zim := b12Push(g, me.ID, "Zimone and Dina", "Legendary Creature — Human Dryad", momB21ZimoneOracle, 3, 4)
	momB21Lands(g, me, 5)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	me.Hand.Cards = nil
	b16Activate(t, g, me.ID, zim, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	passPriorityAroundTable(t, g)
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil {
		// The drawn card was not a land: filler cards are not lands.
		if me.Hand.Size() != 1 {
			t.Fatalf("hand %d, want exactly one draw", me.Hand.Size())
		}
		return
	}
	answerChooseCards(t, g, me.ID)
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want exactly one draw", me.Hand.Size())
	}
}

func TestMomB21ZimoneCostIsRefusedWithoutAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	zim := b12Push(g, me.ID, "Zimone and Dina", "Legendary Creature — Human Dryad", momB21ZimoneOracle, 3, 4)
	err := g.ActivateCatalogAbility(me.ID, zim, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{zim}})
	if err == nil {
		t.Fatal("Zimone and Dina sacrificed itself to its own ability")
	}
}

// --- Hidetsugu and Kairi --------------------------------------------------------------

func TestMomB21HidetsuguEntersDrawsThreeAndPutsTwoOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Hidetsugu and Kairi", "Legendary Creature — Ogre Demon Dragon", momB21HidetsuguOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand after the draw = %d, want %d", got, before+3)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("put-back prompt = %+v, want exactly two", pick)
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt after the pick")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand after putting two back = %d, want %d", got, before+1)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != first {
		t.Error("the first card of the answer must end on top of the library")
	}
}

func TestMomB21HidetsuguDiesLosesLifeAndMayCastAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	h := b12Push(g, me.ID, "Hidetsugu and Kairi", "Legendary Creature — Ogre Demon Dragon", momB21HidetsuguOracle, 5, 4)
	spell := seedTopOfLibrary(me, game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}"})
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(h) })
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Fatalf("opponent life %d, want %d (its mana value 5)", opp.Life, life-5)
	}
	if !g.Exile.Contains(spell) {
		t.Fatal("the top card was not exiled")
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spell)
	if err := castThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	castInsideTheResolution(t, g, spell)
}

func TestMomB21HidetsuguDiesWithACreatureOnTopOnlyDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	h := b12Push(g, me.ID, "Hidetsugu and Kairi", "Legendary Creature — Ogre Demon Dragon", momB21HidetsuguOracle, 5, 4)
	bear := seedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(h) })
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("opponent life %d, want %d", opp.Life, life-2)
	}
	if !g.Exile.Contains(bear) {
		t.Fatal("the top card was not exiled")
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a creature card was offered a free cast")
	}
	if grantedPermissionOn(g, me.ID, bear, game.ZoneExile).Granted() {
		t.Error("a creature card can be cast from exile")
	}
}

func TestMomB21HidetsuguDiesDeclinedCastLeavesTheCardInExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	h := b12Push(g, me.ID, "Hidetsugu and Kairi", "Legendary Creature — Ogre Demon Dragon", momB21HidetsuguOracle, 5, 4)
	spell := seedTopOfLibrary(me, game.Card{Name: "Cheap Instant", TypeLine: "Instant", ManaCost: "{R}"})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(h) })
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, me.ID)
	if !g.Exile.Contains(spell) {
		t.Error("the declined card left exile")
	}
}

// --- Rashmi and Ragavan --------------------------------------------------------------

// momB21RashmiTurn seeds Rashmi and the opponent's top card, casts the turn's
// first spell and answers the trigger's target, leaving the trigger
// resolved up to any prompt it raises.
func momB21RashmiTurn(t *testing.T, artifacts int, stolen game.Card) (g *game.Game, me, opp *game.Player, top uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	b12Push(g, me.ID, "Rashmi and Ragavan", "Legendary Creature — Elf Monkey", momB21RashmiOracle, 2, 4)
	for i := 0; i < artifacts; i++ {
		b12Permanent(g, me.ID, "Trinket", "Artifact")
	}
	top = seedTopOfLibrary(opp, stolen)
	castCatalogSpell(t, g, "First Spell", "Sorcery", "", nil)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	return g, me, opp, top
}

func TestMomB21RashmiExilesMakesATreasureAndOffersAFreeCastBelowTheArtifactCount(t *testing.T) {
	// One Treasure plus two Trinkets: three artifacts, so mana value 2 is free.
	g, me, _, top := momB21RashmiTurn(t, 2, game.Card{Name: "Stolen Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	if !g.Exile.Contains(top) {
		t.Fatal("the opponent's top card was not exiled")
	}
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("Treasures %d, want 1", n)
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, top)
	if err := castThroughResolutionCast(t, g, me.ID, top, game.CastSpellParams{}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	castInsideTheResolution(t, g, top)
}

func TestMomB21RashmiDeclinedFreeCastFallsBackToAPaidCastThisTurn(t *testing.T) {
	g, me, _, top := momB21RashmiTurn(t, 2, game.Card{Name: "Stolen Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	declineResolutionCast(t, g, me.ID)
	perm := grantedPermissionOn(g, me.ID, top, game.ZoneExile)
	if !perm.Granted() || perm.Cost == "{0}" {
		t.Fatalf("after declining, the exiled card's permission = %+v, want a paid cast this turn", perm)
	}
}

func TestMomB21RashmiNotBelowTheArtifactCountIsAPaidCastOnly(t *testing.T) {
	// Only the Treasure: one artifact, so mana value 2 is not free.
	g, me, _, top := momB21RashmiTurn(t, 0, game.Card{Name: "Stolen Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a free cast was offered for a card that is not cheaper than the artifact count")
	}
	perm := grantedPermissionOn(g, me.ID, top, game.ZoneExile)
	if !perm.Granted() {
		t.Fatal("the exiled card may not be cast this turn")
	}
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast the exiled card without paying its cost")
	}
	b06AddMana(me, "R", "R")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("paid cast of the exiled card: %v", err)
	}
}

func TestMomB21RashmiAFreeCastNeedsASpell(t *testing.T) {
	// A land with mana value 0 is below the count of one, but is not a spell.
	g, me, _, top := momB21RashmiTurn(t, 0, game.Card{Name: "Stolen Forest", TypeLine: "Basic Land — Forest"})
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a land was offered a free cast")
	}
	if !g.Exile.Contains(top) {
		t.Error("the land was not exiled")
	}
}

func TestMomB21RashmiOnlyTheFirstSpellOnYourOwnTurnTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	b12Push(g, me.ID, "Rashmi and Ragavan", "Legendary Creature — Elf Monkey", momB21RashmiOracle, 2, 4)
	seedTopOfLibrary(opp, game.Card{Name: "Stolen", TypeLine: "Sorcery", ManaCost: "{R}"})
	castCatalogSpell(t, g, "First Spell", "Sorcery", "", nil)
	pickPlayerTarget(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Second Spell", "Sorcery", "", nil)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the second spell triggered Rashmi and Ragavan")
	}
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures %d, want 1", n)
	}
}

// --- Elenda and Azor -------------------------------------------------------------------

func TestMomB21ElendaAttackPaysXWUBToDrawX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	elenda := b12Push(g, me.ID, "Elenda and Azor", "Legendary Creature — Vampire Knight Sphinx", momB21ElendaOracle, 6, 6)
	before := me.Hand.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(elenda, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	b06AddMana(me, "W", "U", "B", "C", "C", "C")
	passPriorityAroundTable(t, g)
	pa := answerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Errorf("X prompt = %+v, want 0..3", pa)
	}
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
	if left := me.ManaPool.Total(); left != 1 {
		t.Errorf("%d mana left, want 1 (paid {2}{W}{U}{B})", left)
	}
}

func TestMomB21ElendaEndStepPaysLifeForAVampireKnightPerCardDrawn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b12Push(g, me.ID, "Elenda and Azor", "Legendary Creature — Vampire Knight Sphinx", momB21ElendaOracle, 6, 6)
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(me.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	life := me.Life
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
	n, _ := countTokensNamed(g, me.ID, "Vampire Knight")
	if n != drawn {
		t.Errorf("Vampire Knights %d, want %d (cards drawn this turn)", n, drawn)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Vampire Knight" && (!game.HasKeyword(&c, "lifelink") || c.Power != 1 || c.Toughness != 1) {
			t.Errorf("token %+v is not a 1/1 with lifelink", c)
		}
	}
}

func TestMomB21ElendaEndStepDeclinedPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b12Push(g, me.ID, "Elenda and Azor", "Legendary Creature — Vampire Knight Sphinx", momB21ElendaOracle, 6, 6)
	life := me.Life
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
	if n, _ := countTokensNamed(g, me.ID, "Vampire Knight"); n != 0 {
		t.Errorf("Vampire Knights %d, want 0", n)
	}
}

func TestMomB21ElendaTriggersOnAnOpponentsEndStepToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	oppSeat := 1
	toMain(t, g)
	b12Push(g, me.ID, "Elenda and Azor", "Legendary Creature — Vampire Knight Sphinx", momB21ElendaOracle, 6, 6)
	// #5793: Elenda's offer at our own end step holds the table until it
	// is answered, so the walk to the opponent's end step declines it.
	for i := 0; i < 300 && (g.Turn.Step != game.StepEnd || g.Turn.ActiveSeat != oppSeat); i++ {
		if hasPayUnlessFor(g, me.ID) {
			answerPayUnless(t, g, me.ID, false)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep toward end step of seat %d: %v", oppSeat, err)
		}
	}
	if g.Turn.Step != game.StepEnd || g.Turn.ActiveSeat != oppSeat {
		t.Fatalf("never reached the end step of seat %d", oppSeat)
	}
	for i := 0; i < 2; i++ {
		_ = g.DrawCard(me.ID)
	}
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no life-payment offer at the opponent's end step")
	}
}

// --- Inga and Esika ------------------------------------------------------------------------

func TestMomB21IngaGrantsVigilanceAndRestrictedManaToEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	inga := b12Push(g, me.ID, "Inga and Esika", "Legendary Creature — Human God", momB21IngaOracle, 4, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{inga, bear} {
		if !momB21HasKeyword(t, g, id, "vigilance") {
			t.Errorf("%s has no vigilance", id)
		}
	}
	if momB21HasKeyword(t, g, theirs, "vigilance") {
		t.Error("an opponent's creature has vigilance")
	}
	// The bear taps for one mana of any color, restricted to creature spells.
	if err := tapGrantedMana(t, g, me.ID, bear, "G"); err != nil {
		t.Fatalf("tap: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool %d after tapping the Bear, want 1", me.ManaPool.Total())
	}
	if err := g.CastSpell(me.ID, momB21Hand(me, game.Card{Name: "Instant", TypeLine: "Instant", ManaCost: "{W}"}), game.CastSpellParams{Strict: true}); err == nil {
		t.Error("creature-only mana paid for an instant")
	}
}

func momB21Hand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

func TestMomB21IngaDrawsWhenThreeManaFromCreaturesPaidACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b12Push(g, me.ID, "Inga and Esika", "Legendary Creature — Human God", momB21IngaOracle, 4, 4)
	var bears []uuid.UUID
	for i := 0; i < 3; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	for _, b := range bears {
		if err := tapGrantedMana(t, g, me.ID, b, "G"); err != nil {
			t.Fatalf("tap: %v", err)
		}
	}
	spell := momB21Hand(me, game.Card{Name: "Three Drop", TypeLine: "Creature — Elf", ManaCost: "{3}", Power: 3, Toughness: 3})
	before := me.Hand.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with three creature mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The spell left the hand (-1); the trigger drew one card (+1).
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand %d, want %d (a card drawn)", got, before)
	}
}

func TestMomB21IngaDoesNotDrawForTwoManaFromCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b12Push(g, me.ID, "Inga and Esika", "Legendary Creature — Human God", momB21IngaOracle, 4, 4)
	var bears []uuid.UUID
	for i := 0; i < 2; i++ {
		bears = append(bears, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	}
	for _, b := range bears {
		if err := tapGrantedMana(t, g, me.ID, b, "G"); err != nil {
			t.Fatalf("tap: %v", err)
		}
	}
	b06AddMana(me, "G")
	spell := momB21Hand(me, game.Card{Name: "Three Drop", TypeLine: "Creature — Elf", ManaCost: "{3}", Power: 3, Toughness: 3})
	before := me.Hand.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before-1 {
		t.Errorf("hand %d, want %d (no draw: only two mana from creatures)", got, before-1)
	}
}

// --- Omnath, Locus of All --------------------------------------------------------------------

func TestMomB21OmnathTurnsLostManaBlackInsteadAtEveryBoundary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b12Push(g, me.ID, "Omnath, Locus of All", "Legendary Creature — Phyrexian Elemental", momB21OmnathOracle, 4, 4)
	b06AddMana(me, "G", "R", "C")
	advanceTo(t, g, game.StepBeginCombat)
	if got := me.ManaPool.Total(); got != 3 {
		t.Fatalf("pool %d after the phase boundary, want 3 (kept as black)", got)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "B" {
			t.Errorf("mana %q is not black", tok.Color)
		}
	}
}

func momB21OmnathMain(t *testing.T, top game.Card) (g *game.Game, me *game.Player, id uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Omnath, Locus of All", "Legendary Creature — Phyrexian Elemental", momB21OmnathOracle, 4, 4)
	id = seedTopOfLibrary(me, top)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	return g, me, id
}

func TestMomB21OmnathPutsAnIneligibleTopCardIntoHandWithoutAsking(t *testing.T) {
	g, me, id := momB21OmnathMain(t, game.Card{Name: "Two Symbols", TypeLine: "Creature — Bear", ManaCost: "{4}{G}{G}", Power: 2, Toughness: 2})
	if latestChooseCards(g, me.ID) != nil {
		t.Fatal("asked to reveal a card with two colored symbols")
	}
	if !inHand(me, id) {
		t.Fatal("the top card did not go to hand")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d, want 0", me.ManaPool.Total())
	}
}

func TestMomB21OmnathRevealsAThreeSymbolCardForThreeManaOfItsColors(t *testing.T) {
	g, me, id := momB21OmnathMain(t, game.Card{Name: "Three Symbols", TypeLine: "Creature — Bear", ManaCost: "{R}{R}{G}", Colors: []string{"R", "G"}, Power: 2, Toughness: 2})
	prompt := latestChooseCards(g, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Fatalf("reveal prompt = %+v, want an optional pick of the one card", prompt)
	}
	answerChooseCards(t, g, me.ID, id)
	if !inHand(me, id) {
		t.Fatal("the revealed card did not go to hand")
	}
	// The three mana are red or green (a pick per mana); answer any open prompts.
	for _, ch := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if ch != nil && ch.Kind == game.PendingChoiceMana && ch.Chooser == me.ID {
			if err := g.ResolveManaChoice(ch.ID, me.ID, "R"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
		}
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool %d, want 3", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "R" && tok.Color != "G" {
			t.Errorf("mana %q is not one of the card's colors", tok.Color)
		}
	}
}

func TestMomB21OmnathNotRevealingStillPutsTheCardIntoHand(t *testing.T) {
	g, me, id := momB21OmnathMain(t, game.Card{Name: "Three Symbols", TypeLine: "Creature — Bear", ManaCost: "{W/U}{W/U}{B/P}", Colors: []string{"W", "U", "B"}, Power: 2, Toughness: 2})
	answerChooseCards(t, g, me.ID)
	if !inHand(me, id) {
		t.Fatal("the unrevealed card did not go to hand")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d, want 0", me.ManaPool.Total())
	}
}
