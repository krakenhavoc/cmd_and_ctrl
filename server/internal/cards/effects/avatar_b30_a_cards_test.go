package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b30_a_cards_test.go — Avatar: The Last Airbender Eternal batch 30,
// group a (#5075).

const (
	avB30aLoAndLiOracle   = "b5018037-cf4e-46d2-86c3-fb8eec6e44d1"
	avB30aRuthlessOracle  = "100771fd-07db-47be-ada3-935dc918fa33"
	avB30aScarringOracle  = "63957ec0-89e6-4be2-8740-846ca56ee48c"
	avB30aDeerDogOracle   = "76c7a339-015b-4074-8e10-14d52b784f9b"
	avB30aOmashuOracle    = "ea58fb10-3d70-4c25-b317-615dfb20a48d"
	avB30aFranticOracle   = "a866cc05-69b0-45f2-8829-b6ef08193ed1"
	avB30aRecruitOracle   = "33e66662-aea6-480c-b69b-1907d987b69e"
	avB30aLongshotOracle  = "f93c3468-5b1d-4fa7-a281-1009ea0f2062"
	avB30aLostOracle      = "69ac3757-66f8-47d4-af7b-dda66c96c254"
	avB30aMokuOracle      = "d6e1b1ff-e4e5-4814-9a20-2d527014463f"
	avB30aScarringFlashAC = "flash-if-you-control-an-attacking-legendary-creature"
)

func TestAvatarB30aCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		avB30aLoAndLiOracle:  "Lo and Li, Royal Advisors",
		avB30aRuthlessOracle: "Ruthless Waterbender",
		avB30aScarringOracle: "Scarring Memories",
		avB30aDeerDogOracle:  "Deer-Dog",
		avB30aOmashuOracle:   "Founding of Omashu",
		avB30aFranticOracle:  "Frantic Confrontation",
		avB30aRecruitOracle:  "Freedom Fighter Recruit",
		avB30aLongshotOracle: "Longshot, Rebel Bowman",
		avB30aLostOracle:     "Lost in Memories",
		avB30aMokuOracle:     "Moku, Meandering Drummer",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		want, caveats := CompletenessFull, 0
		if oracle == avB30aScarringOracle {
			// #5207: its flash is an alternative-cost offer.
			want, caveats = CompletenessCaveats, 1
		}
		if spec.Name != name || spec.Completeness != want || len(spec.Caveats) != caveats {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func avB30aOpp(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }

// --- Lo and Li, Royal Advisors ---------------------------------------------------

func TestAvatarB30LoAndLiCountsEveryCardAnOpponentDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	loli := b12Push(g, me.ID, "Lo and Li, Royal Advisors", "Legendary Creature — Human Advisor", avB30aLoAndLiOracle, 3, 3)
	advisor := b12Creature(g, me.ID, "Court Advisor", "Creature — Human Advisor", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Advisor", "Creature — Human Advisor", 1, 1)
	var err error
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(opp.ID, 2) })
	if err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{loli: 2, advisor: 2, bear: 0, theirs: 0} {
		if got := countersOn(g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s: %d +1/+1 counters, want %d", id, got, want)
		}
	}
	// Your own discard is not an opponent's.
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(me.ID, 1) })
	if err != nil {
		t.Fatalf("discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, loli, game.CounterPlusOne); got != 2 {
		t.Errorf("your own discard added counters: %d, want 2", got)
	}
}

func TestAvatarB30LoAndLiMillsFourAndAMillIsOneTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	loli := b12Push(g, me.ID, "Lo and Li, Royal Advisors", "Legendary Creature — Human Advisor", avB30aLoAndLiOracle, 3, 3)
	theirs := opp.Library.Size()
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, loli, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := theirs - opp.Library.Size(); got != 4 {
		t.Fatalf("the opponent milled %d, want 4", got)
	}
	if got := countersOn(g, loli, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after one mill of four cards, want 1", got)
	}
	// Milling yourself is not an opponent milling.
	mine := me.Library.Size()
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, loli, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := mine - me.Library.Size(); got != 4 {
		t.Fatalf("you milled %d, want 4", got)
	}
	if got := countersOn(g, loli, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after milling yourself, want still 1", got)
	}
}

func TestAvatarB30LoAndLiAbilityRefusesAShortPool(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	loli := b12Push(g, me.ID, "Lo and Li, Royal Advisors", "Legendary Creature — Human Advisor", avB30aLoAndLiOracle, 3, 3)
	fillPoolColored(me, "R", 3)
	if err := g.ActivateCatalogAbility(me.ID, loli, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("{2}{U/B} was paid with three red mana")
	}
}

// --- Ruthless Waterbender --------------------------------------------------------

func TestAvatarB30RuthlessWaterbenderWaterbendsOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wb := b12Push(g, me.ID, "Ruthless Waterbender", "Creature — Human Soldier Ally", avB30aRuthlessOracle, 1, 3)
	tappers := pushTapCostSoldiers(g, me.ID, 2)
	if err := avB4Activate(g, me.ID, wb, 0, tappers); err != nil {
		t.Fatalf("waterbend {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, wb); p != 2 || th != 4 {
		t.Errorf("Ruthless Waterbender is %d/%d, want 2/4", p, th)
	}
}

func TestAvatarB30RuthlessWaterbenderRefusesOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := avB30aOpp(g)
	wb := b12Push(g, opp.ID, "Ruthless Waterbender", "Creature — Human Soldier Ally", avB30aRuthlessOracle, 1, 3)
	tappers := pushTapCostSoldiers(g, opp.ID, 2)
	if err := avB4Activate(g, opp.ID, wb, 0, tappers); err == nil {
		t.Fatal("the ability was activated during another player's turn")
	}
}

// --- Scarring Memories -----------------------------------------------------------

func TestAvatarB30ScarringMemoriesSacrificesThenDiscardsThenLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	opp := avB30aOpp(g)
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	life, hand := opp.Life, opp.Hand.Size()
	castCatalogSpell(t, g, "Scarring Memories", "Sorcery — Lesson", avB30aScarringOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to sacrifice a creature")
	}
	if discardChoiceFor(g, opp.ID) != nil || opp.Life != life {
		t.Fatal("the discard or the life loss came before the sacrifice was chosen")
	}
	answerSacrifice(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("the chosen creature was not the one sacrificed")
	}
	if opp.Life != life {
		t.Fatal("the life loss came before the discard")
	}
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-1)
	}
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestAvatarB30ScarringMemoriesRefusesToTargetYou(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := castCatalogSpellErr(t, g, "Scarring Memories", "Sorcery — Lesson", avB30aScarringOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Scarring Memories targeted its caster")
	}
}

func TestAvatarB30ScarringMemoriesHasFlashOnlyWithAnAttackingLegend(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	// A legendary creature you control that is not attacking grants no flash.
	b12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	bear := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	fillPoolColored(me, "B", 4)
	if _, err := castFromHandErr(g, me, "Scarring Memories", "Sorcery — Lesson", "{3}{B}", avB30aScarringOracle,
		game.CastSpellParams{Strict: true, Targets: targets, AlternativeCost: avB30aScarringFlashAC}); err == nil {
		t.Fatal("cast with flash while only a nonlegendary creature attacks")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	legend := b12Creature(g, me.ID, "Legendary Bear", "Legendary Creature — Bear", 2, 2)
	targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(legend, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	fillPoolColored(me, "B", 4)
	if _, err := castFromHandErr(g, me, "Scarring Memories", "Sorcery — Lesson", "{3}{B}", avB30aScarringOracle,
		game.CastSpellParams{Strict: true, Targets: targets}); err == nil {
		t.Fatal("the printed cost was cast at instant speed")
	}
	if _, err := castFromHandErr(g, me, "Scarring Memories", "Sorcery — Lesson", "{3}{B}", avB30aScarringOracle,
		game.CastSpellParams{Strict: true, Targets: targets, AlternativeCost: avB30aScarringFlashAC}); err != nil {
		t.Fatalf("cast with flash while a legendary creature attacks: %v", err)
	}
}

// --- Deer-Dog --------------------------------------------------------------------

func TestAvatarB30DeerDogHasFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dog := b12Push(g, me.ID, "Deer-Dog", "Creature — Elk Dog", avB30aDeerDogOracle, 1, 3)
	if !hasEffectiveKeyword(t, g, dog, "first strike") {
		t.Error("Deer-Dog has no first strike")
	}
}

// --- Founding of Omashu ----------------------------------------------------------

func TestAvatarB30FoundingOfOmashuChapters(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	g.WithWriteLock(func() { avB10Fillers(me, 4) })
	saga := castCatalogSpell(t, g, "Founding of Omashu", "Enchantment — Saga", avB30aOmashuOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Ally", me.ID); got != 2 {
		t.Fatalf("%d Allies after chapter I, want 2", got)
	}
	ally := findBattlefieldByName(g, "Ally")
	if p, th := effectivePTOf(t, g, ally); p != 1 || th != 1 {
		t.Errorf("Ally is %d/%d, want 1/1", p, th)
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II offered no discard")
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a discard and a draw", hand, me.Hand.Size())
	}

	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, ally); p != 2 || th != 1 {
		t.Errorf("Ally is %d/%d after chapter III, want 2/1", p, th)
	}
	g.RunStateChecksForTest()
	if onBattlefield(g, saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestAvatarB30FoundingOfOmashuChapterTwoDeclinedDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	g.WithWriteLock(func() { avB10Fillers(me, 4) })
	castCatalogSpell(t, g, "Founding of Omashu", "Enchantment — Saga", avB30aOmashuOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	c := discardChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("chapter II offered no discard")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw without a discard", hand, me.Hand.Size())
	}
}

// --- Frantic Confrontation -------------------------------------------------------

func TestAvatarB30FranticConfrontationPumpsByXWithFirstStrikeAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "R", 4)
	castFromHandForTest(t, g, me, "Frantic Confrontation", "Instant", "{X}{R}", avB30aFranticOracle,
		game.CastSpellParams{Strict: true, XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 5 || th != 2 {
		t.Errorf("Bear is %d/%d, want 5/2", p, th)
	}
	for _, kw := range []string{"first strike", "trample"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("Bear has no %s", kw)
		}
	}
}

func TestAvatarB30FranticConfrontationRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	advanceToMain(t, g)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	fillPoolColored(me, "R", 2)
	if _, err := castFromHandErr(g, me, "Frantic Confrontation", "Instant", "{X}{R}", avB30aFranticOracle,
		game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Fatal("Frantic Confrontation targeted an opponent's creature")
	}
}

// --- Freedom Fighter Recruit -----------------------------------------------------

func TestAvatarB30FreedomFighterRecruitCountsYourCreaturesInEveryZone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	recruit := b12Push(g, me.ID, "Freedom Fighter Recruit", "Creature — Human Rebel Ally", avB30aRecruitOracle, 0, 2)
	if p, _ := effectivePTOf(t, g, recruit); p != 1 {
		t.Fatalf("alone: power %d, want 1", p)
	}
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, th := effectivePTOf(t, g, recruit); p != 3 || th != 2 {
		t.Errorf("Recruit is %d/%d, want 3/2", p, th)
	}
	// The ability functions in a hand too (the ruling).
	inHand := game.Card{InstanceID: uuid.New(), Name: "Freedom Fighter Recruit", TypeLine: "Creature — Human Rebel Ally",
		OracleID: avB30aRecruitOracle, Toughness: 2, Owner: me.ID, Controller: me.ID}
	var p int
	g.WithWriteLock(func() { p = g.PowerAnywhereForEffect(inHand) })
	if p != 3 {
		t.Errorf("in hand: power %d, want 3", p)
	}
}

// --- Longshot, Rebel Bowman ------------------------------------------------------

func TestAvatarB30LongshotDiscountsNoncreatureSpellsAndPingsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	longshot := b12Push(g, me.ID, "Longshot, Rebel Bowman", "Legendary Creature — Human Rebel Ally", avB30aLongshotOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, longshot, "reach") {
		t.Error("Longshot has no reach")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	fillPoolColored(me, "R", 1)
	castFromHandForTest(t, g, me, "Two Drop", "Sorcery", "{1}{R}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 2
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s: life %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestAvatarB30LongshotIgnoresCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Longshot, Rebel Bowman", "Legendary Creature — Human Rebel Ally", avB30aLongshotOracle, 3, 3)
	fillPoolColored(me, "R", 1)
	if _, err := castFromHandErr(g, me, "Two Drop Bear", "Creature — Bear", "{1}{R}", "", game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a creature spell was discounted")
	}
	life := opp.Life
	fillPoolColored(me, "R", 1)
	castFromHandForTest(t, g, me, "Two Drop Bear", "Creature — Bear", "{1}{R}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("a creature spell dealt damage: life %d, want %d", opp.Life, life)
	}
}

// --- Lost in Memories ------------------------------------------------------------

func TestAvatarB30LostInMemoriesPumpsAndGrantsFlashbackOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	bolt := seedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-lost-in-memories-bolt")
	bear := auraBear(g, me.ID)
	castCatalogSpell(t, g, "Lost in Memories", auraTypeLine, avB30aLostOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 3 || th != 3 {
		t.Fatalf("the enchanted creature is %d/%d, want 3/3", p, th)
	}
	if grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard).Granted() {
		t.Fatal("flashback before any combat damage")
	}
	trekB18AAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || perm.Cost != "" || !perm.ExileOnResolution {
		t.Fatalf("granted permission %+v, want flashback at its mana cost", perm)
	}
}

func TestAvatarB30LostInMemoriesRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := avB30aOpp(g)
	theirs := auraBear(g, opp.ID)
	if err := castCatalogSpellErr(t, g, "Lost in Memories", auraTypeLine, avB30aLostOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("Lost in Memories enchanted an opponent's creature")
	}
}

// --- Moku, Meandering Drummer ----------------------------------------------------

func TestAvatarB30MokuPaidPumpsItselfAndHastesYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], avB30aOpp(g)
	advanceToMain(t, g)
	moku := b12Push(g, me.ID, "Moku, Meandering Drummer", "Legendary Creature — Human Bard Ally", avB30aMokuOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	fillPoolColored(me, "R", 1)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, moku); p != 4 || th != 3 {
		t.Errorf("Moku is %d/%d, want 4/3", p, th)
	}
	if p, th := effectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("Bear is %d/%d, want 2/2 (only Moku is pumped)", p, th)
	}
	for _, id := range []uuid.UUID{moku, bear} {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", id)
		}
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature gained haste")
	}
}

func TestAvatarB30MokuDeclinedDoesNothingAndCreatureSpellsDontTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	moku := b12Push(g, me.ID, "Moku, Meandering Drummer", "Legendary Creature — Human Bard Ally", avB30aMokuOracle, 2, 2)
	castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a creature spell triggered Moku")
	}
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, moku); p != 2 || th != 2 {
		t.Errorf("Moku is %d/%d after declining, want 2/2", p, th)
	}
	if hasEffectiveKeyword(t, g, moku, "haste") {
		t.Error("Moku gained haste without the payment")
	}
}
