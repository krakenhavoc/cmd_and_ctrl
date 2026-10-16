package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kotose_the_silent_spider_test.go — #3988: Kotose's extraction search
// over an opponent's graveyard, hand and library, and its one-play
// permission over the cards it exiled.

const kotoseOracle = "bfef4daa-b950-46ce-8c08-11f9983741cf"

type kotoseTable struct {
	g       *game.Game
	me, opp *game.Player
	cards   map[string]uuid.UUID
}

// kotoseSetUp seeds the opponent's three zones: Bolts in each (two in the
// graveyard), a Tower in the graveyard and the library, a basic Forest in
// the graveyard, and a card of another name in the hand and the library.
func kotoseSetUp(t *testing.T) kotoseTable {
	t.Helper()
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	cards := map[string]uuid.UUID{}
	put := func(key string, zone *game.Zone, name, typeLine, cost string) {
		id := uuid.New()
		zone.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: opp.ID, Controller: opp.ID})
		cards[key] = id
	}
	put("gy bolt", opp.Graveyard, "Lightning Bolt", "Instant", "{R}")
	put("gy bolt 2", opp.Graveyard, "Lightning Bolt", "Instant", "{R}")
	put("gy tower", opp.Graveyard, "Command Tower", "Land", "")
	put("gy forest", opp.Graveyard, "Forest", "Basic Land — Forest", "")
	put("hand bolt", opp.Hand, "Lightning Bolt", "Instant", "{R}")
	put("hand other", opp.Hand, "Counterspell", "Instant", "{U}{U}")
	put("lib bolt", opp.Library, "Lightning Bolt", "Instant", "{R}")
	put("lib tower", opp.Library, "Command Tower", "Land", "")
	put("lib other", opp.Library, "Shock", "Instant", "{R}")
	return kotoseTable{g: g, me: me, opp: opp, cards: cards}
}

// enter casts Kotose and resolves it, targets `target` with its enter
// trigger and resolves that, leaving the search prompt open. It returns
// Kotose.
func (k kotoseTable) enter(t *testing.T, target uuid.UUID) uuid.UUID {
	t.Helper()
	kotose := tdCastFromHand(t, k.g, game.Card{Name: "Kotose, the Silent Spider", TypeLine: "Legendary Creature — Human Ninja",
		OracleID: kotoseOracle, ManaCost: "{3}{U}{B}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, k.g)
	pickCard(t, k.g, k.me.ID, target)
	passPriorityAroundTable(t, k.g)
	return kotose
}

// searchesSince counts the searches of the opponent's library since
// event `from`.
func (k kotoseTable) searchesSince(from int) int {
	n := 0
	for _, ev := range k.g.Events[from:] {
		if ev.Kind == game.EventSearchLibrary && ev.Target == k.opp.ID && ev.Label == game.SearchShuffledLabel {
			n++
		}
	}
	return n
}

func TestKotoseSearchesTheirGraveyardHandAndLibraryAndExilesThem(t *testing.T) {
	k := kotoseSetUp(t)
	from := len(k.g.Events)
	k.enter(t, k.cards["gy bolt"])

	if !k.g.Exile.Contains(k.cards["gy bolt"]) {
		t.Fatal("the target was not exiled")
	}
	choice := searchChoiceFor(k.g, k.me.ID)
	if choice == nil {
		t.Fatal("no search prompt for Kotose's controller")
	}
	want := map[uuid.UUID]string{k.cards["gy bolt 2"]: "graveyard", k.cards["hand bolt"]: "hand", k.cards["lib bolt"]: "library"}
	if len(choice.SearchCards) != len(want) {
		t.Fatalf("candidates %v, want the three other Bolts", choice.SearchCards)
	}
	for _, id := range choice.SearchCards {
		if _, ok := want[id]; !ok {
			t.Errorf("candidate %s is not a Bolt in their zones", id)
		}
	}
	if len(choice.SearchGraveyardCards) != 1 || choice.SearchGraveyardCards[0] != k.cards["gy bolt 2"] {
		t.Errorf("graveyard candidates %v", choice.SearchGraveyardCards)
	}
	if len(choice.SearchHandCards) != 1 || choice.SearchHandCards[0] != k.cards["hand bolt"] {
		t.Errorf("hand candidates %v", choice.SearchHandCards)
	}
	if choice.SearchMax != 3 || k.g.SearchFloorLocked(choice) != 0 {
		t.Errorf("any number: max %d, floor %d; want 3 and 0", choice.SearchMax, k.g.SearchFloorLocked(choice))
	}
	// Searching a hand is looking at all of it (CR 701.23).
	if c, _ := k.g.LookupCardForEffect(k.cards["hand other"]); !c.IsKnownTo(k.me.ID) {
		t.Error("the searcher did not see the rest of their hand")
	}

	answerSearchByID(t, k.g, k.me.ID, k.cards["gy bolt 2"], k.cards["hand bolt"], k.cards["lib bolt"])
	for _, key := range []string{"gy bolt 2", "hand bolt", "lib bolt"} {
		if !k.g.Exile.Contains(k.cards[key]) {
			t.Errorf("%s was not exiled", key)
		}
	}
	if !k.opp.Hand.Contains(k.cards["hand other"]) || !k.opp.Library.Contains(k.cards["lib other"]) {
		t.Error("a card of another name left their zones")
	}
	if n := k.searchesSince(from); n != 1 {
		t.Errorf("%d searches of their library, want 1 (then that player shuffles)", n)
	}
}

// Finding nothing is legal ("any number"), and the library is still
// searched and shuffled: all three zones were searched.
func TestKotoseMayFindNothingAndStillShuffles(t *testing.T) {
	k := kotoseSetUp(t)
	from := len(k.g.Events)
	k.enter(t, k.cards["gy bolt"])
	answerSearchFailToFind(t, k.g, k.me.ID)
	if n := k.searchesSince(from); n != 1 {
		t.Errorf("%d searches, want 1", n)
	}
	if !k.opp.Hand.Contains(k.cards["hand bolt"]) || !k.opp.Graveyard.Contains(k.cards["gy bolt 2"]) {
		t.Error("a card nobody took was exiled")
	}
	perm := exiledPermission(k.g, k.cards["gy bolt"])
	if perm.Player != k.me.ID || !permissionLive(k.g, perm, k.me.ID) {
		t.Errorf("the target alone is still playable: %+v", perm)
	}
}

// "You may play ONE of the exiled cards, and you may spend mana as though
// it were mana of any color to cast it": one cast, paid with blue, spends
// the permission for every other exiled card.
func TestKotosePlaysOneExiledCardWithAnyColor(t *testing.T) {
	k := kotoseSetUp(t)
	k.enter(t, k.cards["gy bolt"])
	answerSearchByID(t, k.g, k.me.ID, k.cards["hand bolt"], k.cards["lib bolt"])

	perm := exiledPermission(k.g, k.cards["hand bolt"])
	if perm.Player != k.me.ID || perm.CastOnly || !perm.AnyColor || perm.CastsLeft != 1 {
		t.Fatalf("%+v: Kotose's controller may play one of them, spending mana as any color", perm)
	}
	for _, key := range []string{"gy bolt", "hand bolt", "lib bolt"} {
		if !permissionLive(k.g, exiledPermission(k.g, k.cards[key]), k.me.ID) {
			t.Errorf("%s is not playable", key)
		}
	}
	fillPoolColored(k.me, "U", 1)
	if err := k.g.CastSpell(k.me.ID, k.cards["lib bolt"], game.CastSpellParams{
		FromZone: "exile", Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: k.opp.ID}},
	}); err != nil {
		t.Fatalf("cast a Bolt from exile with blue mana: %v", err)
	}
	passPriorityAroundTable(t, k.g)
	for _, key := range []string{"gy bolt", "hand bolt"} {
		if p := k.g.CastPermissionOnCardByIDForEffect(k.cards[key]); p != nil && permissionLive(k.g, p, k.me.ID) {
			t.Errorf("%s is still playable after one was played", key)
		}
	}
}

// A land among the exiled cards is PLAYED with the turn's land drop, and
// the play spends the one use as a cast does.
func TestKotoseLandPlaySpendsTheOneUse(t *testing.T) {
	k := kotoseSetUp(t)
	k.enter(t, k.cards["gy tower"])
	answerSearchByID(t, k.g, k.me.ID, k.cards["lib tower"])
	if !k.g.Exile.Contains(k.cards["lib tower"]) {
		t.Fatal("the library Tower was not exiled")
	}
	if err := k.g.CastSpell(k.me.ID, k.cards["gy tower"], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the exiled Tower: %v", err)
	}
	if !k.g.Battlefield.Contains(k.cards["gy tower"]) {
		t.Fatal("the Tower is not on the battlefield")
	}
	if p := k.g.CastPermissionOnCardByIDForEffect(k.cards["lib tower"]); p != nil && permissionLive(k.g, p, k.me.ID) {
		t.Error("the other Tower is still playable after a land play")
	}
	nextTurnOfMine(t, k.g)
	if err := k.g.CastSpell(k.me.ID, k.cards["lib tower"], game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("a second exiled card was played")
	}
}

// "Target card other than a basic land card", and only while Kotose's
// controller controls Kotose.
func TestKotoseRefusesABasicLandAndEndsWithControl(t *testing.T) {
	k := kotoseSetUp(t)
	tdCastFromHand(t, k.g, game.Card{Name: "Kotose, the Silent Spider", TypeLine: "Legendary Creature — Human Ninja",
		OracleID: kotoseOracle, ManaCost: "{3}{U}{B}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	passPriorityAroundTable(t, k.g)
	p := latestPickTarget(k.g, k.me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := k.g.ResolvePickTarget(p.ID, k.me.ID, game.TargetRef{Kind: game.TargetCard, ID: k.cards["gy forest"]}); err == nil {
		t.Error("a basic land card was a legal target")
	}
	var kotose uuid.UUID
	k.g.ReadSnapshot(func() {
		for _, c := range k.g.Battlefield.Cards {
			if c.OracleID == kotoseOracle {
				kotose = c.InstanceID
			}
		}
	})
	pickCard(t, k.g, k.me.ID, k.cards["gy bolt"])
	passPriorityAroundTable(t, k.g)
	answerSearchFailToFind(t, k.g, k.me.ID)
	if !permissionLive(k.g, exiledPermission(k.g, k.cards["gy bolt"]), k.me.ID) {
		t.Fatal("the exiled Bolt is not playable")
	}
	k.g.WithWriteLock(func() { _ = k.g.DestroyPermanentForEffect(kotose) })
	if permissionLive(k.g, exiledPermission(k.g, k.cards["gy bolt"]), k.me.ID) {
		t.Error("playable after Kotose left the battlefield")
	}
}
