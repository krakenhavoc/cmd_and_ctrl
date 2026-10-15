package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3801: Ziatora's Envoy and Rashmi, Eternities Crafter. "Look at (or
// reveal) the top card of your library. You may play a land / cast a spell
// from the top of your library without paying its mana cost. If you
// don't, put that card into your hand."

const (
	ziatorasEnvoyOracle = "93fe1d5b-3ab3-431a-9d78-175550ba9d43"
	rashmiOracle        = "01a79429-7d57-4a51-9815-a20905ed4707"
)

// envoyTable has the Envoy ready to connect for 5 and a chosen card on top
// of the controller's library. attack() runs combat to the damage step and
// settles the trigger, leaving the table at the prompt (if there is one).
type envoyTable struct {
	g       *game.Game
	me, opp *game.Player
	envoy   uuid.UUID
}

func newEnvoyTable(t *testing.T) *envoyTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	envoy := pushDiesCreatureForTest(g, me.ID, "Ziatora's Envoy", ziatorasEnvoyOracle,
		"Creature — Lizard Warrior", 5, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return &envoyTable{g: g, me: me, opp: opp, envoy: envoy}
}

// top puts a card on top of the controller's library.
func (tb *envoyTable) top(name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	tb.g.WithWriteLock(func() {
		tb.me.Library.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: tb.me.ID, Controller: tb.me.ID,
		})
	})
	return id
}

func (tb *envoyTable) attack(t *testing.T) {
	t.Helper()
	attackWith(t, tb.g, tb.opp.ID, tb.envoy)
	passPriorityAroundTable(t, tb.g)
}

func (tb *envoyTable) mayCast() *game.PendingChoice {
	return latestChoiceOfKindFor(tb.g, game.PendingChoiceMayCast, tb.me.ID)
}

func (tb *envoyTable) answer(t *testing.T, apply bool) {
	t.Helper()
	c := tb.mayCast()
	if c == nil {
		t.Fatal("no may_cast prompt is waiting")
	}
	if err := tb.g.ResolveMayCast(c.ID, tb.me.ID, apply); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
}

func (tb *envoyTable) inLibraryTop(id uuid.UUID) bool {
	n := tb.me.Library.Size()
	return n > 0 && tb.me.Library.Cards[n-1].InstanceID == id
}

func TestZiatorasEnvoyIsOnTheCatalogWithTrample(t *testing.T) {
	tb := newEnvoyTable(t)
	if !hasEffectiveKeyword(t, tb.g, tb.envoy, "trample") {
		t.Error("Ziatora's Envoy has trample")
	}
}

// The land path: the top land is offered as a play, accepting puts it on the
// battlefield as this turn's land drop.
func TestZiatorasEnvoyPlaysTheTopLand(t *testing.T) {
	tb := newEnvoyTable(t)
	land := tb.top("Forest", "Basic Land — Forest", "", "")
	before := tb.me.Library.Size()
	tb.attack(t)
	c := tb.mayCast()
	if c == nil {
		t.Fatal("the land was not offered")
	}
	if c.MayCastKeyword != game.MayCastKeywordPlayOrHand || c.MayCastCard != land {
		t.Fatalf("offer %q over %s, want %q over the land", c.MayCastKeyword, c.MayCastCard, game.MayCastKeywordPlayOrHand)
	}
	tb.answer(t, true)
	if !tb.g.Battlefield.Contains(land) {
		t.Fatal("the land was not played")
	}
	if tb.me.Hand.Contains(land) {
		t.Error("a played land also went to the hand")
	}
	if tb.me.Library.Size() != before-1 {
		t.Errorf("library %d, want %d", tb.me.Library.Size(), before-1)
	}
	if got := tb.g.LandsPlayedThisTurnFor(tb.me.ID); got != 1 {
		t.Errorf("lands played this turn = %d, want 1 (it is a land play, CR 305.2a)", got)
	}
}

// "If you don't": declining the land puts it into the hand.
func TestZiatorasEnvoyDecliningTheLandPutsItInHand(t *testing.T) {
	tb := newEnvoyTable(t)
	land := tb.top("Forest", "Basic Land — Forest", "", "")
	tb.attack(t)
	tb.answer(t, false)
	if !tb.me.Hand.Contains(land) || tb.g.Battlefield.Contains(land) {
		t.Fatal("a declined land is in the hand, not on the battlefield")
	}
	if tb.me.Library.Contains(land) {
		t.Error("the card is still in the library")
	}
}

// A land the rules would not let the player play (the land drop is spent)
// is not offered: "if you don't" is already true.
func TestZiatorasEnvoyNoLandDropLeftPutsTheLandInHand(t *testing.T) {
	tb := newEnvoyTable(t)
	land := tb.top("Forest", "Basic Land — Forest", "", "")
	tb.g.WithWriteLock(func() { tb.g.LandsPlayedThisTurn = map[uuid.UUID]int{tb.me.ID: 1} })
	tb.attack(t)
	if tb.mayCast() != nil {
		t.Fatal("a land with no land drop left was offered")
	}
	if !tb.me.Hand.Contains(land) {
		t.Fatal("the unplayable land did not go to the hand")
	}
}

// The spell path: a spell within the damage dealt is cast from the library
// for free, with its target, and resolves.
func TestZiatorasEnvoyCastsTheTopSpellFromTheLibrary(t *testing.T) {
	tb := newEnvoyTable(t)
	bolt := tb.top("Lightning Bolt", "Instant", "{R}", boltOracleCombat)
	tb.attack(t)
	c := tb.mayCast()
	if c == nil || c.MayCastKeyword != game.MayCastKeywordFreeOrHand || c.MayCastCard != bolt {
		t.Fatalf("the spell was not offered as a free cast: %+v", c)
	}
	tb.answer(t, true)
	if !tb.inLibraryTop(bolt) {
		t.Fatal("the card must stay on top of the library while the free cast is open")
	}
	var perm *game.CastPermission
	tb.g.ReadSnapshot(func() { perm = tb.g.CastPermissionOnCardByIDForEffect(bolt) })
	if perm == nil || perm.Zone != game.ZoneLibrary || perm.Cost != "{0}" || perm.LapseOnPass != game.LapseToHand {
		t.Fatalf("grant %+v, want a free library cast that lapses to the hand", perm)
	}
	if err := tb.g.CastSpell(tb.me.ID, bolt, game.CastSpellParams{
		FromZone: "library",
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: tb.opp.ID}},
	}); err != nil {
		t.Fatalf("the free cast from the library: %v", err)
	}
	if tb.me.Library.Contains(bolt) {
		t.Fatal("the spell did not leave the library")
	}
	passPriorityAroundTable(t, tb.g)
	if tb.opp.Life != 40-5-3 {
		t.Errorf("opponent life %d, want 32 (5 combat + 3 from the free Bolt)", tb.opp.Life)
	}
	if tb.me.Hand.Contains(bolt) {
		t.Error("a cast spell also went to the hand")
	}
}

// Declining the cast puts the card into the hand.
func TestZiatorasEnvoyDecliningTheSpellPutsItInHand(t *testing.T) {
	tb := newEnvoyTable(t)
	bolt := tb.top("Lightning Bolt", "Instant", "{R}", boltOracleCombat)
	tb.attack(t)
	tb.answer(t, false)
	if !tb.me.Hand.Contains(bolt) || tb.me.Library.Contains(bolt) {
		t.Fatal("a declined spell goes to the hand")
	}
}

// Accepting and then passing without casting closes the window with the card
// in the hand, and the library is not left holding a free cast.
func TestZiatorasEnvoyPassingWithoutCastingPutsItInHand(t *testing.T) {
	tb := newEnvoyTable(t)
	bolt := tb.top("Lightning Bolt", "Instant", "{R}", boltOracleCombat)
	tb.attack(t)
	tb.answer(t, true)
	if err := tb.g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !tb.me.Hand.Contains(bolt) || tb.me.Library.Contains(bolt) {
		t.Fatal("an uncast free spell goes to the hand when the window closes")
	}
	var perm *game.CastPermission
	tb.g.ReadSnapshot(func() { perm = tb.g.CastPermissionOnCardByIDForEffect(bolt) })
	if perm != nil {
		t.Error("the grant outlived the window")
	}
}

// A spell above the damage dealt cannot be cast: no question, to the hand.
func TestZiatorasEnvoySpellAboveTheDamageGoesToHand(t *testing.T) {
	tb := newEnvoyTable(t)
	big := tb.top("Big Spell", "Sorcery", "{4}{R}{R}", "")
	tb.attack(t)
	if tb.mayCast() != nil {
		t.Fatal("a mana value 6 spell was offered after 5 damage")
	}
	if !tb.me.Hand.Contains(big) {
		t.Fatal("the too-expensive spell did not go to the hand")
	}
}

// The edge of the cap: mana value equal to the damage is allowed.
func TestZiatorasEnvoySpellAtTheDamageIsOffered(t *testing.T) {
	tb := newEnvoyTable(t)
	five := tb.top("Five Spell", "Sorcery", "{3}{R}{R}", "")
	tb.attack(t)
	c := tb.mayCast()
	if c == nil || c.MayCastCard != five {
		t.Fatalf("a mana value 5 spell was not offered after 5 damage: %+v", c)
	}
}

// Looking is private (CR 701.20): only the controller learns the card.
func TestZiatorasEnvoyTheLookIsPrivate(t *testing.T) {
	tb := newEnvoyTable(t)
	land := tb.top("Forest", "Basic Land — Forest", "", "")
	tb.attack(t)
	c := tb.mayCast()
	if c == nil || !c.PrivateText() {
		t.Fatal("the offer's words name a card only the controller has seen, so it must be private")
	}
	var known map[uuid.UUID]bool
	tb.g.ReadSnapshot(func() {
		card, _ := tb.g.LookupCardForEffect(land)
		known = card.KnownBy
	})
	if !known[tb.me.ID] || known[tb.opp.ID] {
		t.Errorf("knowers %v: want the controller only", known)
	}
}

// An empty library has nothing to look at.
func TestZiatorasEnvoyEmptyLibraryDoesNothing(t *testing.T) {
	tb := newEnvoyTable(t)
	tb.g.WithWriteLock(func() { tb.me.Library.Cards = nil })
	handBefore := tb.me.Hand.Size()
	tb.attack(t)
	if tb.mayCast() != nil || tb.me.Hand.Size() != handBefore {
		t.Fatal("an empty library offered or moved something")
	}
}

// Rashmi: a revealed spell of lesser mana value than the first spell cast is
// offered; one of equal value, a land, or a second spell is not.
type rashmiTable struct {
	g       *game.Game
	me, opp *game.Player
}

func newRashmiTable(t *testing.T) *rashmiTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushDiesCreatureForTest(g, me.ID, "Rashmi, Eternities Crafter", rashmiOracle,
		"Legendary Creature — Elf Druid", 2, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return &rashmiTable{g: g, me: me, opp: opp}
}

func (rt *rashmiTable) top(name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	rt.g.WithWriteLock(func() {
		rt.me.Library.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: rt.me.ID, Controller: rt.me.ID,
		})
	})
	return id
}

// castFromHand casts a free-valued spell of the given cost as the first spell.
func (rt *rashmiTable) castFirst(t *testing.T, cost string) {
	t.Helper()
	id := handCardFull(rt.me, "First Spell", "Instant", cost, "", nil)
	rt.g.WithWriteLock(func() { rt.me.ManaPool = game.ManaPool{} })
	if err := rt.g.CastSpell(rt.me.ID, id, game.CastSpellParams{Strict: false}); err != nil {
		t.Fatalf("casting the first spell: %v", err)
	}
	passPriorityAroundTable(t, rt.g)
}

func (rt *rashmiTable) mayCast() *game.PendingChoice {
	return latestChoiceOfKindFor(rt.g, game.PendingChoiceMayCast, rt.me.ID)
}

func TestRashmiOffersARevealedLesserSpellAndRevealsItToTheTable(t *testing.T) {
	rt := newRashmiTable(t)
	bolt := rt.top("Lightning Bolt", "Instant", "{R}", boltOracleCombat)
	rt.castFirst(t, "{2}")
	c := rt.mayCast()
	if c == nil || c.MayCastCard != bolt || c.MayCastKeyword != game.MayCastKeywordFreeOrHand {
		t.Fatalf("a mana value 1 card was not offered under a mana value 2 spell: %+v", c)
	}
	if c.PrivateText() {
		t.Error("a revealed card is not private")
	}
	var known map[uuid.UUID]bool
	rt.g.ReadSnapshot(func() {
		card, _ := rt.g.LookupCardForEffect(bolt)
		known = card.KnownBy
	})
	if !known[rt.opp.ID] {
		t.Error("the opponent did not see the revealed card")
	}
	var perm *game.CastPermission
	if err := rt.g.ResolveMayCast(c.ID, rt.me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	rt.g.ReadSnapshot(func() { perm = rt.g.CastPermissionOnCardByIDForEffect(bolt) })
	if perm == nil {
		t.Fatal("accepting did not open the free cast")
	}
}

func TestRashmiEqualManaValueGoesToHand(t *testing.T) {
	rt := newRashmiTable(t)
	two := rt.top("Two Spell", "Sorcery", "{1}{R}", "")
	rt.castFirst(t, "{2}")
	if rt.mayCast() != nil {
		t.Fatal("an equal mana value is not lesser")
	}
	if !rt.me.Hand.Contains(two) {
		t.Fatal("the not-lesser card goes to the hand")
	}
}

func TestRashmiRevealedLandGoesToHand(t *testing.T) {
	rt := newRashmiTable(t)
	land := rt.top("Forest", "Basic Land — Forest", "", "")
	rt.castFirst(t, "{2}")
	if rt.mayCast() != nil {
		t.Fatal("a land is not a spell and must not be offered")
	}
	if !rt.me.Hand.Contains(land) {
		t.Fatal("the land goes to the hand")
	}
}

func TestRashmiOnlyTriggersOnTheFirstSpellEachTurn(t *testing.T) {
	rt := newRashmiTable(t)
	rt.top("Lightning Bolt", "Instant", "{R}", boltOracleCombat)
	rt.castFirst(t, "{3}")
	c := rt.mayCast()
	if c == nil {
		t.Fatal("the first spell did not trigger Rashmi")
	}
	if err := rt.g.ResolveMayCast(c.ID, rt.me.ID, false); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	second := rt.top("Another Bolt", "Instant", "{R}", boltOracleCombat)
	rt.castFirst(t, "{3}")
	if rt.mayCast() != nil || rt.me.Hand.Contains(second) || !rt.me.Library.Contains(second) {
		t.Fatal("the second spell this turn triggered Rashmi")
	}
}
