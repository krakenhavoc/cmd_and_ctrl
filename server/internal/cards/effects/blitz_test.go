package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// blitz_test.go — blitz (CR 702.152), #3772, on Sabin, Master Monk.

const sabinOracle = "7234bbf8-b962-4515-bedf-68d111c6137e"

// blitzSabin puts Sabin in `zone` ("hand" or "graveyard") with a filler
// card in hand to discard, floats `pool` and casts him claiming `alt`.
// Returns Sabin's ID, the filler's ID and the cast error.
func blitzSabin(t *testing.T, g *game.Game, me *game.Player, zone, alt, pool string) (uuid.UUID, uuid.UUID, error) {
	t.Helper()
	toMainForCost(t, g)
	sabin := game.Card{
		InstanceID: uuid.New(), Name: "Sabin, Master Monk", TypeLine: "Legendary Creature — Human Noble Monk",
		ManaCost: "{4}{R}", OracleID: sabinOracle, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
		Power: 4, Toughness: 3, KnownBy: map[uuid.UUID]bool{me.ID: true},
	}
	filler := handCardFull(me, "Filler", "Sorcery", "{1}", "", nil)
	params := game.CastSpellParams{Strict: true, AlternativeCost: alt, AltCostIDs: []uuid.UUID{filler}}
	g.WithWriteLock(func() {
		if zone == "graveyard" {
			me.Graveyard.PushTop(sabin)
		} else {
			me.Hand.PushTop(sabin)
		}
	})
	if zone == "graveyard" {
		params.FromZone = "graveyard"
	}
	if alt == "" {
		params.AltCostIDs = nil
	}
	floatForTest(g, me, pool)
	return sabin.InstanceID, filler, g.CastSpell(me.ID, sabin.InstanceID, params)
}

// Blitzed from hand: the discard is paid, Sabin has haste and double
// strike and attacks at once, and at the end step he is sacrificed and
// his controller draws a card.
func TestSabinBlitzHasteThenSacrificedAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id, filler, err := blitzSabin(t, g, me, "hand", "blitz", "CCRR")
	if err != nil {
		t.Fatalf("blitz Sabin: %v", err)
	}
	if !me.Graveyard.Contains(filler) {
		t.Fatal("the discard is paid as the spell is cast")
	}
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, id) {
		t.Fatal("blitzed Sabin should be on the battlefield")
	}
	if !hasEffectiveKeyword(t, g, id, "haste") || !hasEffectiveKeyword(t, g, id, "double strike") {
		t.Fatal("blitzed Sabin has haste and double strike")
	}
	if grantedTriggerCount(g, id) != 1 {
		t.Fatal("blitzed Sabin has \"When this creature dies, draw a card.\"")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("a blitzed creature attacks the turn it was cast: %v", err)
	}
	advanceTo(t, g, game.StepEnd)
	hand := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, id) || !me.Graveyard.Contains(id) {
		t.Fatalf("Sabin is sacrificed at the end step (battlefield=%v graveyard=%v)",
			onBattlefieldByID(g, id), me.Graveyard.Contains(id))
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("Sabin dying draws a card: hand %d want %d", len(me.Hand.Cards), hand+1)
	}
}

// From the graveyard, only through the graveyard-bound blitz offer: not
// for the printed {4}{R}, and not by claiming the hand's offer.
func TestSabinCastsFromGraveyardOnlyUsingBlitz(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if _, _, err := blitzSabin(t, g, me, "graveyard", "", "CCCCR"); err == nil {
		t.Fatal("Sabin cannot be cast from the graveyard for his mana cost")
	}
	g = newCatalogGame(t)
	me = g.Seats[0]
	if _, _, err := blitzSabin(t, g, me, "graveyard", "blitz", "CCRR"); err == nil {
		t.Fatal("the hand's blitz offer cannot be claimed from the graveyard")
	}
	g = newCatalogGame(t)
	me = g.Seats[0]
	id, filler, err := blitzSabin(t, g, me, "graveyard", "blitz-graveyard", "CCRR")
	if err != nil {
		t.Fatalf("blitz Sabin from the graveyard: %v", err)
	}
	if !me.Graveyard.Contains(filler) {
		t.Fatal("the graveyard blitz still discards a card")
	}
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, id) || !hasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("Sabin blitzed from the graveyard enters with haste")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("Sabin is sacrificed back to the graveyard, ready to blitz again")
	}
}

// The discard is the price: with nothing else in hand the offer is
// refused, and nothing is paid.
func TestSabinBlitzNeedsACardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	id := handCardFull(me, "Sabin, Master Monk", "Legendary Creature — Human Noble Monk", "{4}{R}", sabinOracle, []string{"R"})
	floatForTest(g, me, "CCRR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "blitz"}); err == nil {
		t.Fatal("blitz with no card to discard")
	}
	if !me.Hand.Contains(id) {
		t.Error("a refused cast leaves Sabin in hand")
	}
}

// Cast for {4}{R}: no haste, no draw, and he stays.
func TestSabinCastNormallyStays(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id, _, err := blitzSabin(t, g, me, "hand", "", "CCCCR")
	if err != nil {
		t.Fatalf("cast Sabin: %v", err)
	}
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, id, "haste") || grantedTriggerCount(g, id) != 0 {
		t.Fatal("a normal cast gives neither haste nor the draw")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, id) {
		t.Error("a normally cast Sabin stays on the battlefield")
	}
}

// sabinOffersOnTheWire is the alternative-cost keys the view stamps on
// Sabin for his controller, in whichever zone he sits, and whether each
// offer carries its discard picker.
func sabinOffersOnTheWire(g *game.Game, me *game.Player, id uuid.UUID) map[string]bool {
	out := map[string]bool{}
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for _, zone := range [][]protocol.CardView{s.Hand.Cards, s.Graveyard.Cards} {
			for _, c := range zone {
				if c.InstanceID != id.String() {
					continue
				}
				for _, o := range c.AlternativeCosts {
					out[o.Key] = o.Discards && o.PayOptions != nil
				}
			}
		}
	}
	return out
}

// The client's picker needs nothing new: the view stamps the hand offer
// on Sabin in hand and the graveyard offer on Sabin in the graveyard,
// each with its discard picker.
func TestSabinBlitzOffersOnTheWire(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	handCardFull(me, "Filler", "Sorcery", "{1}", "", nil)
	inHand := handCardFull(me, "Sabin, Master Monk", "Legendary Creature — Human Noble Monk", "{4}{R}", sabinOracle, []string{"R"})
	inGrave := graveyardCardOf(g, me, "Sabin, Master Monk", "Legendary Creature — Human Noble Monk", "{4}{R}", sabinOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			me.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{me.ID: true}
		}
	})
	if got := sabinOffersOnTheWire(g, me, inHand); len(got) != 1 || !got["blitz"] {
		t.Errorf("in hand: offers %v, want blitz with a discard picker", got)
	}
	if got := sabinOffersOnTheWire(g, me, inGrave); len(got) != 1 || !got["blitz-graveyard"] {
		t.Errorf("in the graveyard: offers %v, want blitz-graveyard with a discard picker", got)
	}
}

// A blitzed creature that dies before the end step still draws, and the
// end-step sacrifice then finds nothing.
func TestBlitzedCreatureDyingEarlyDrawsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id, _, err := blitzSabin(t, g, me, "hand", "blitz", "CCRR")
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	hand := len(me.Hand.Cards)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}, game.DestroyOptions{}) })
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || len(me.Hand.Cards) != hand+1 {
		t.Fatalf("a destroyed blitzed creature draws (graveyard=%v hand %d want %d)",
			me.Graveyard.Contains(id), len(me.Hand.Cards), hand+1)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 || !me.Graveyard.Contains(id) {
		t.Error("the end-step sacrifice finds nothing and draws nothing more")
	}
}

// A blitzed creature an opponent has taken is not sacrificed: the
// delayed trigger is the caster's, and a player can sacrifice only a
// permanent they control (CR 701.17a). It keeps its haste and its draw,
// which is now the thief's.
func TestBlitzedCreatureStolenIsNotSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id, _, err := blitzSabin(t, g, me, "hand", "blitz", "CCRR")
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, id, opp.ID, game.IndefiniteDuration(), "test: steal") })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, id) {
		t.Fatal("the caster cannot sacrifice a creature they no longer control")
	}
	if !hasEffectiveKeyword(t, g, id, "haste") || grantedTriggerCount(g, id) != 1 {
		t.Error("the stolen creature keeps its blitz riders")
	}

	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	id, _, err = blitzSabin(t, g, me, "hand", "blitz", "CCRR")
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, id, opp.ID, game.IndefiniteDuration(), "test: steal") })
	hand := len(opp.Hand.Cards)
	b18Kill(t, g, id)
	if len(opp.Hand.Cards) != hand+1 {
		t.Errorf("the thief draws when it dies: hand %d want %d", len(opp.Hand.Cards), hand+1)
	}
}

// A blitzed creature that died and came back under the same instance ID
// is a new object (CR 400.7): the end-step sacrifice does not touch it.
func TestBlitzedCreatureThatDiedAndCameBackIsNotSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id, _, err := blitzSabin(t, g, me, "hand", "blitz", "CCRR")
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	b18Kill(t, g, id)
	var back uuid.UUID
	var retErr error
	g.WithWriteLock(func() { back, retErr = g.ReturnFromGraveyardWithCountersForEffect(id, uuid.Nil, false, nil) })
	if retErr != nil || back != id {
		t.Fatalf("setup: return under the same ID (got %s, %v)", back, retErr)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, id) {
		t.Error("the returned creature is not the object that was blitzed, and stays")
	}
}

// A blitzed creature flickered before the end step is a new object
// (CR 400.7): it has no haste, no draw, and is not sacrificed.
func TestBlitzedCreatureFlickeredIsNotSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id, _, err := blitzSabin(t, g, me, "hand", "blitz", "CCRR")
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	var flickerErr error
	g.WithWriteLock(func() {
		flickerErr = (Flicker{Target: id}).Apply(ctxFor(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()}))
	})
	if flickerErr != nil {
		t.Fatalf("flicker: %v", flickerErr)
	}
	passPriorityAroundTable(t, g)
	newID := id
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Sabin, Master Monk" {
			newID = c.InstanceID
		}
	}
	if hasEffectiveKeyword(t, g, newID, "haste") || grantedTriggerCount(g, newID) != 0 {
		t.Fatal("the returned creature is a new object, not blitzed")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !onBattlefieldByID(g, newID) {
		t.Error("a flickered blitzed creature is not sacrificed")
	}
}
