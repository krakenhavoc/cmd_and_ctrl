package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestZiatorasEnvoyIsOnTheCatalogWithTrample(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	if !effectstest.HasEffectiveKeyword(t, tb.G, tb.Envoy, "trample") {
		t.Error("Ziatora's Envoy has trample")
	}
}

// The land path: the top land is offered as a play, accepting puts it on the
// battlefield as this turn's land drop.
func TestZiatorasEnvoyPlaysTheTopLand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	land := tb.Top("Forest", "Basic Land — Forest", "", "")
	before := tb.Me.Library.Size()
	tb.Attack(t)
	c := tb.MayCast()
	if c == nil {
		t.Fatal("the land was not offered")
	}
	if c.MayCastKeyword != game.MayCastKeywordPlayOrHand || c.MayCastCard != land {
		t.Fatalf("offer %q over %s, want %q over the land", c.MayCastKeyword, c.MayCastCard, game.MayCastKeywordPlayOrHand)
	}
	tb.Answer(t, true)
	if !tb.G.Battlefield.Contains(land) {
		t.Fatal("the land was not played")
	}
	if tb.Me.Hand.Contains(land) {
		t.Error("a played land also went to the hand")
	}
	if tb.Me.Library.Size() != before-1 {
		t.Errorf("library %d, want %d", tb.Me.Library.Size(), before-1)
	}
	if got := tb.G.LandsPlayedThisTurnFor(tb.Me.ID); got != 1 {
		t.Errorf("lands played this turn = %d, want 1 (it is a land play, CR 305.2a)", got)
	}
}

// "If you don't": declining the land puts it into the hand.
func TestZiatorasEnvoyDecliningTheLandPutsItInHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	land := tb.Top("Forest", "Basic Land — Forest", "", "")
	tb.Attack(t)
	tb.Answer(t, false)
	if !tb.Me.Hand.Contains(land) || tb.G.Battlefield.Contains(land) {
		t.Fatal("a declined land is in the hand, not on the battlefield")
	}
	if tb.Me.Library.Contains(land) {
		t.Error("the card is still in the library")
	}
}

// A land the rules would not let the player play (the land drop is spent)
// is not offered: "if you don't" is already true.
func TestZiatorasEnvoyNoLandDropLeftPutsTheLandInHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	land := tb.Top("Forest", "Basic Land — Forest", "", "")
	tb.G.WithWriteLock(func() { tb.G.LandsPlayedThisTurn = map[uuid.UUID]int{tb.Me.ID: 1} })
	tb.Attack(t)
	if tb.MayCast() != nil {
		t.Fatal("a land with no land drop left was offered")
	}
	if !tb.Me.Hand.Contains(land) {
		t.Fatal("the unplayable land did not go to the hand")
	}
}

// The spell path: a spell within the damage dealt is cast from the library
// for free, with its target, and resolves.
func TestZiatorasEnvoyCastsTheTopSpellFromTheLibrary(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	bolt := tb.Top("Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	tb.Attack(t)
	c := tb.MayCast()
	if c == nil || c.Kind != game.PendingChoiceResolutionCast || c.MayCastKeyword != game.MayCastKeywordFreeOrHand || effectstest.KeywordCastCard(c) != bolt {
		t.Fatalf("the spell was not offered as a free cast: %+v", c)
	}
	if !tb.G.ResolutionPaused() {
		t.Fatal("the offer is not holding the trigger's resolution open")
	}
	if !tb.InLibraryTop(bolt) {
		t.Fatal("the card must stay on top of the library while the free cast is open")
	}
	var perm *game.CastPermission
	tb.G.ReadSnapshot(func() { perm = tb.G.CastPermissionOnCardByIDForEffect(bolt) })
	if perm == nil || perm.Zone != game.ZoneLibrary || perm.Cost != "{0}" || perm.LapseOnPass != game.LapseToHand {
		t.Fatalf("grant %+v, want a free library cast that lapses to the hand", perm)
	}
	if err := tb.G.CastSpell(tb.Me.ID, bolt, game.CastSpellParams{
		FromZone: "library",
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: tb.Opp.ID}},
	}); err != nil {
		t.Fatalf("the free cast from the library: %v", err)
	}
	if tb.Me.Library.Contains(bolt) {
		t.Fatal("the spell did not leave the library")
	}
	if tb.G.ResolutionPaused() || tb.MayCast() != nil {
		t.Fatal("the cast did not answer the offer and finish the trigger's resolution")
	}
	effectstest.PassPriorityAroundTable(t, tb.G)
	if tb.Opp.Life != 40-5-3 {
		t.Errorf("opponent life %d, want 32 (5 combat + 3 from the free Bolt)", tb.Opp.Life)
	}
	if tb.Me.Hand.Contains(bolt) {
		t.Error("a cast spell also went to the hand")
	}
}

// Declining the cast puts the card into the hand.
func TestZiatorasEnvoyDecliningTheSpellPutsItInHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	bolt := tb.Top("Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	tb.Attack(t)
	tb.Answer(t, false)
	if !tb.Me.Hand.Contains(bolt) || tb.Me.Library.Contains(bolt) {
		t.Fatal("a declined spell goes to the hand")
	}
}

// Nobody passes while the offer is open (it is mid-resolution), and a
// table paused on it is a restore point: the restored decline puts the
// card into the hand, and the library is not left holding a free cast.
func TestZiatorasEnvoyOfferSurvivesARestoreMidPrompt(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	bolt := tb.Top("Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	tb.Attack(t)
	if err := tb.G.PassPriority(); err == nil {
		t.Fatal("a pass was taken while the free cast is open mid-resolution")
	}
	restored := effectstest.RestoreRoundTrip(t, tb.G, true)
	rme := restored.PlayerByIDForEffect(tb.Me.ID)
	rc := effectstest.KeywordCastPromptFor(restored, rme.ID)
	if rc == nil || effectstest.KeywordCastCard(rc) != bolt || !restored.ResolutionPaused() {
		t.Fatalf("the restored table lost the offer: %+v", rc)
	}
	if err := restored.ResolvePendingChoice(rc.ID, rme.ID, nil); err != nil {
		t.Fatalf("the restored decline: %v", err)
	}
	if !rme.Hand.Contains(bolt) || rme.Library.Contains(bolt) {
		t.Fatal("the restored decline did not put the spell into the hand")
	}
	var perm *game.CastPermission
	restored.ReadSnapshot(func() { perm = restored.CastPermissionOnCardByIDForEffect(bolt) })
	if perm != nil {
		t.Error("the grant outlived the decline")
	}
}

// A spell above the damage dealt cannot be cast: no question, to the hand.
func TestZiatorasEnvoySpellAboveTheDamageGoesToHand(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	big := tb.Top("Big Spell", "Sorcery", "{4}{R}{R}", "")
	tb.Attack(t)
	if tb.MayCast() != nil {
		t.Fatal("a mana value 6 spell was offered after 5 damage")
	}
	if !tb.Me.Hand.Contains(big) {
		t.Fatal("the too-expensive spell did not go to the hand")
	}
}

// The edge of the cap: mana value equal to the damage is allowed.
func TestZiatorasEnvoySpellAtTheDamageIsOffered(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	five := tb.Top("Five Spell", "Sorcery", "{3}{R}{R}", "")
	tb.Attack(t)
	c := tb.MayCast()
	if c == nil || effectstest.KeywordCastCard(c) != five {
		t.Fatalf("a mana value 5 spell was not offered after 5 damage: %+v", c)
	}
}

// Looking is private (CR 701.20): only the controller learns the card.
func TestZiatorasEnvoyTheLookIsPrivate(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	land := tb.Top("Forest", "Basic Land — Forest", "", "")
	tb.Attack(t)
	c := tb.MayCast()
	if c == nil || !c.PrivateText() {
		t.Fatal("the offer's words name a card only the controller has seen, so it must be private")
	}
	var known map[uuid.UUID]bool
	tb.G.ReadSnapshot(func() {
		card, _ := tb.G.LookupCardForEffect(land)
		known = card.KnownBy
	})
	if !known[tb.Me.ID] || known[tb.Opp.ID] {
		t.Errorf("knowers %v: want the controller only", known)
	}
}

// An empty library has nothing to look at.
func TestZiatorasEnvoyEmptyLibraryDoesNothing(t *testing.T) {
	tb := effectstest.NewEnvoyTable(t)
	tb.G.WithWriteLock(func() { tb.Me.Library.Cards = nil })
	handBefore := tb.Me.Hand.Size()
	tb.Attack(t)
	if tb.MayCast() != nil || tb.Me.Hand.Size() != handBefore {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushDiesCreatureForTest(g, me.ID, "Rashmi, Eternities Crafter", effectstest.RashmiOracle,
		"Legendary Creature — Elf Druid", 2, 3)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return &rashmiTable{g: g, me: me, opp: opp}
}

func (rt *rashmiTable) Top(name, typeLine, cost, oracle string) uuid.UUID {
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
	id := effectstest.HandCardFull(rt.me, "First Spell", "Instant", cost, "", nil)
	rt.g.WithWriteLock(func() { rt.me.ManaPool = game.ManaPool{} })
	if err := rt.g.CastSpell(rt.me.ID, id, game.CastSpellParams{Strict: false}); err != nil {
		t.Fatalf("casting the first spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, rt.g)
}

// mayCast is the revealed spell's offer, a resolution_cast (#4250).
func (rt *rashmiTable) MayCast() *game.PendingChoice {
	return effectstest.KeywordCastPromptFor(rt.g, rt.me.ID)
}

func TestRashmiOffersARevealedLesserSpellAndRevealsItToTheTable(t *testing.T) {
	rt := newRashmiTable(t)
	bolt := rt.Top("Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	rt.castFirst(t, "{2}")
	c := rt.MayCast()
	if c == nil || effectstest.KeywordCastCard(c) != bolt || c.MayCastKeyword != game.MayCastKeywordFreeOrHand {
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
	rt.g.ReadSnapshot(func() { perm = rt.g.CastPermissionOnCardByIDForEffect(bolt) })
	if perm == nil {
		t.Fatal("the offer did not open the free cast")
	}
}

func TestRashmiEqualManaValueGoesToHand(t *testing.T) {
	rt := newRashmiTable(t)
	two := rt.Top("Two Spell", "Sorcery", "{1}{R}", "")
	rt.castFirst(t, "{2}")
	if rt.MayCast() != nil {
		t.Fatal("an equal mana value is not lesser")
	}
	if !rt.me.Hand.Contains(two) {
		t.Fatal("the not-lesser card goes to the hand")
	}
}

func TestRashmiRevealedLandGoesToHand(t *testing.T) {
	rt := newRashmiTable(t)
	land := rt.Top("Forest", "Basic Land — Forest", "", "")
	rt.castFirst(t, "{2}")
	if rt.MayCast() != nil {
		t.Fatal("a land is not a spell and must not be offered")
	}
	if !rt.me.Hand.Contains(land) {
		t.Fatal("the land goes to the hand")
	}
}

func TestRashmiOnlyTriggersOnTheFirstSpellEachTurn(t *testing.T) {
	rt := newRashmiTable(t)
	rt.Top("Lightning Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	rt.castFirst(t, "{3}")
	c := rt.MayCast()
	if c == nil {
		t.Fatal("the first spell did not trigger Rashmi")
	}
	if err := rt.g.ResolvePendingChoice(c.ID, rt.me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	second := rt.Top("Another Bolt", "Instant", "{R}", effectstest.BoltOracleCombat)
	rt.castFirst(t, "{3}")
	if rt.MayCast() != nil || rt.me.Hand.Contains(second) || !rt.me.Library.Contains(second) {
		t.Fatal("the second spell this turn triggered Rashmi")
	}
}
