package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// null_summoner_test.go — #3516: the revealed-hand exile records "exiled
// with this" (CR 607.2a), and a threshold-gated standing permission
// lets you cast the card for as long as the Summoner stays.

const nullSummonerOracle = "56c412be-daab-4461-931f-1e2ea8ecd75d"

// nullSummonerTable casts Null Summoner on seat 0's turn against seat 1,
// answers the target and the pick, and returns the exiled card. The
// victim's hand is a Forest, a Lightning Bolt and a Big Bear that costs
// {3}{R}{R}.
type nullSummonerGame struct {
	g        *game.Game
	me, opp  *game.Player
	summoner uuid.UUID
	bear     uuid.UUID
	bolt     uuid.UUID
}

func nullSummonerTable(t *testing.T) nullSummonerGame {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := revealHand(opp, rhForest(), rhBolt(),
		game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3})
	n := nullSummonerGame{g: g, me: me, opp: opp, bolt: ids[1], bear: ids[2]}

	n.summoner = castCatalogSpell(t, g, "Null Summoner", "Creature — Human Warlock", nullSummonerOracle, nil)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids[1:]) || pick.PickDestination != game.PickExile {
		t.Fatalf("pick = options %v destination %q; want the two nonland cards, exile", pick.DiscardOptions, pick.PickDestination)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{n.bear}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	if !g.Exile.Contains(n.bear) {
		t.Fatal("the chosen card was not exiled")
	}
	// Resolving the trigger walked the turn on; the cast needs a main phase.
	g.Turn.Step = game.StepPrecombatMain
	// Five blue mana in the pool: the enumerator and the view only offer
	// a cast the pool can pay for, and blue pays {3}{R}{R} only because
	// the permission says any type of mana can be spent.
	floatForTest(g, me, "UUUUU")
	return n
}

func (n nullSummonerGame) fillGraveyard(count int) {
	for i := 0; i < count; i++ {
		pushGraveyardTyped(n.me, "Filler", "Sorcery")
	}
}

func (n nullSummonerGame) castBear() error {
	return n.g.CastSpell(n.me.ID, n.bear, game.CastSpellParams{FromZone: "exile", Strict: true})
}

// The exile is linked to the Summoner as the object it is now, the card
// the Summoner did not take stays in the hand unlinked, and a Forest in
// the hand was never a candidate.
func TestNullSummonerLinksTheCardItExiles(t *testing.T) {
	n := nullSummonerTable(t)
	var want game.PermissionCardRef
	n.g.ReadSnapshot(func() {
		c, ok := n.g.LookupCardForEffect(n.summoner)
		if !ok {
			t.Fatal("Null Summoner is not on the battlefield")
		}
		want = game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
	})
	if got := cardInExile(t, n.g, n.bear).ExiledWith; got != want {
		t.Fatalf("ExiledWith = %+v, want the Summoner object %+v", got, want)
	}
	if !n.opp.Hand.Contains(n.bolt) {
		t.Error("the card the Summoner did not choose left the hand")
	}
}

// Threshold: with six cards in your graveyard the exiled card cannot be
// cast, by the engine, the enumerator or the view; with seven it can,
// and the permission is dormant again the moment one leaves.
func TestNullSummonerThresholdGatesTheCast(t *testing.T) {
	n := nullSummonerTable(t)
	n.fillGraveyard(6)
	if enumeratedFor(n.g, n.me.ID, n.bear) || viewCastable(t, n.g, n.me.ID, n.bear) {
		t.Error("the exiled card is offered with six cards in the graveyard")
	}
	if err := n.castBear(); err == nil {
		t.Fatal("cast the exiled card without threshold")
	}

	extra := pushGraveyardTyped(n.me, "Seventh", "Sorcery")
	if !enumeratedFor(n.g, n.me.ID, n.bear) || !viewCastable(t, n.g, n.me.ID, n.bear) {
		t.Fatal("with seven cards the enumerator and the view must offer the exiled card")
	}
	if err := n.g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: n.me.ID}, game.ZoneRef{Kind: game.ZoneExile}, extra); err != nil {
		t.Fatalf("remove a card from the graveyard: %v", err)
	}
	if enumeratedFor(n.g, n.me.ID, n.bear) || viewCastable(t, n.g, n.me.ID, n.bear) {
		t.Error("the permission stayed live after the graveyard fell back to six")
	}
}

// Mana of any type: {3}{R}{R} paid with five blue mana, and the cast
// card is the opponent's card that resolves under your control.
func TestNullSummonerCastsTheExiledCardWithAnyTypeOfMana(t *testing.T) {
	n := nullSummonerTable(t)
	n.fillGraveyard(7)
	if err := n.castBear(); err != nil {
		t.Fatalf("cast the exiled card with blue mana: %v", err)
	}
	passPriorityAroundTable(t, n.g)
	got, ok := battlefieldCard(n.g, n.bear)
	if !ok {
		t.Fatal("the exiled creature did not resolve onto the battlefield")
	}
	if got.Controller != n.me.ID {
		t.Error("the cast creature should be under the caster's control")
	}
	if got.ExiledWith.ID != uuid.Nil {
		t.Errorf("the link followed the card out of exile: %+v", got.ExiledWith)
	}
}

// The permission is a standing one: it ends when the Summoner leaves,
// and a Summoner that comes back is a new object with no claim on the
// card its earlier self exiled (CR 400.7).
func TestNullSummonerPermissionEndsWithTheSummoner(t *testing.T) {
	n := nullSummonerTable(t)
	n.fillGraveyard(7)
	if !enumeratedFor(n.g, n.me.ID, n.bear) {
		t.Fatal("non-vacuity: threshold with the Summoner in play offers the card")
	}
	n.g.WithWriteLock(func() {
		if err := n.g.DestroyPermanentForEffect(n.summoner); err != nil {
			t.Fatalf("destroy the Summoner: %v", err)
		}
	})
	if enumeratedFor(n.g, n.me.ID, n.bear) || viewCastable(t, n.g, n.me.ID, n.bear) {
		t.Error("the card is still offered after the Summoner left")
	}

	// A second Summoner is a different object; the card is exiled with
	// the first.
	pushBattlefieldCardWithTimestamp(n.g, game.Card{
		InstanceID: uuid.New(), Name: "Null Summoner", TypeLine: "Creature — Human Warlock",
		OracleID: nullSummonerOracle, Power: 4, Toughness: 2, Owner: n.me.ID, Controller: n.me.ID,
	})
	if enumeratedFor(n.g, n.me.ID, n.bear) {
		t.Error("another Summoner inherited the card the first one exiled")
	}
}

// A Summoner that is blinked comes back as a new object (CR 400.7), so
// the card its earlier object exiled is stranded.
func TestNullSummonerFlickeredIsANewObject(t *testing.T) {
	n := nullSummonerTable(t)
	n.fillGraveyard(7)
	if !enumeratedFor(n.g, n.me.ID, n.bear) {
		t.Fatal("non-vacuity: threshold with the Summoner in play offers the card")
	}
	n.g.WithWriteLock(func() {
		item := &game.StackItem{Controller: n.me.ID}
		if err := (Flicker{Target: n.summoner}).Apply(ctxFor(n.g, item)); err != nil {
			t.Fatalf("flicker the Summoner: %v", err)
		}
	})
	passPriorityAroundTable(t, n.g)
	n.g.Turn.Step = game.StepPrecombatMain
	if enumeratedFor(n.g, n.me.ID, n.bear) || viewCastable(t, n.g, n.me.ID, n.bear) {
		t.Error("a blinked Summoner kept its claim on the card its earlier object exiled")
	}
}

// A card exiled from the same hand by anything else is not "the exiled
// card": the standing permission names the link, not the zone.
func TestNullSummonerIgnoresACardNothingLinkedToIt(t *testing.T) {
	n := nullSummonerTable(t)
	n.fillGraveyard(7)
	elsewhere := stashed(n.g, n.opp.ID, "Plain Exile", "Creature — Bear", "{1}{G}", 0)
	if enumeratedFor(n.g, n.me.ID, elsewhere) || viewCastable(t, n.g, n.me.ID, elsewhere) {
		t.Error("a card that was not exiled with Null Summoner is offered")
	}
	if err := n.g.CastSpell(n.me.ID, elsewhere, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a card not exiled with Null Summoner was cast through it")
	}
}

// "If you cast it": a Summoner that enters without being cast takes
// nothing, and raises no prompt (CR 603.4).
func TestNullSummonerNeedsToHaveBeenCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealHand(opp, rhBolt())
	card := game.Card{InstanceID: uuid.New(), Name: "Null Summoner", TypeLine: "Creature — Human Warlock",
		OracleID: nullSummonerOracle, Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(card)
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(card.InstanceID, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put onto the battlefield: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an uncast Null Summoner asked for a target")
	}
	noVariant(t, g)
}

// The link rides a prompt that is answered later: a restore point taken
// while the pick is open still links the card it exiles.
func TestNullSummonerLinkSurvivesARestoreWhileThePickIsOpen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := revealHand(opp, rhBolt())
	summoner := castCatalogSpell(t, g, "Null Summoner", "Creature — Human Warlock", nullSummonerOracle, nil)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if pick := openVariant(t, g); pick.PickExiledWith.ID != summoner {
		t.Fatalf("the open pick links %+v, want the Summoner %s", pick.PickExiledWith, summoner)
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	pick := openVariant(t, restored)
	if pick.PickExiledWith.ID != summoner {
		t.Fatalf("the restored pick links %+v, want the Summoner %s", pick.PickExiledWith, summoner)
	}
	if err := restored.ResolvePendingChoice(pick.ID, me.ID, ids); err != nil {
		t.Fatalf("resolve the restored pick: %v", err)
	}
	if got := cardInExile(t, restored, ids[0]).ExiledWith.ID; got != summoner {
		t.Errorf("the restored pick exiled the card linked to %s, want %s", got, summoner)
	}
}

// A revealed-hand exile that does not ask for the link is unchanged:
// Appetite for Brains' card is exiled with nothing.
func TestAnUnlinkedRevealedHandExileStaysUnlinked(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	ids := revealHand(victim, rvBig())
	castAt(t, g, "Appetite for Brains", "Sorcery", appetiteForBrainsOracle, pbPlayer(victim.ID))
	pick := openVariant(t, g)
	if pick.PickExiledWith.ID != uuid.Nil {
		t.Fatalf("an unlinked pick carries a link: %+v", pick.PickExiledWith)
	}
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got := cardInExile(t, g, ids[0]).ExiledWith; got.ID != uuid.Nil {
		t.Errorf("an unlinked exile is linked to %+v", got)
	}
}
