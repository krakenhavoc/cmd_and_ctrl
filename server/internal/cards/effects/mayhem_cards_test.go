package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mayhem_cards_test.go — mayhem (CR 702.187, #4628): the seam against
// the real catalog, then the seven cards. Sadistic Slash and Swarm are
// the seam's probes: an instant and a flash creature, both with a cheap
// mayhem cost that is not the printed one.

const (
	mayhemAbominationOracle = "e0b12c48-7234-40b5-a84c-6901d2b4e5b0"
	mayhemChameleonOracle   = "1ce239f2-79ad-4a30-8e8f-f985045f78dc"
	mayhemGoblinOracle      = "622fbc01-9099-4b6c-b2ce-e41247fc7355"
	mayhemGliderOracle      = "5db9121d-934a-44f1-b178-b389c6595af7"
	mayhemSlashOracle       = "0e471b6d-4ba4-46aa-a802-81e401cbfa56"
	mayhemSwarmOracle       = "10c8cf3b-2530-49e7-b38e-95cdaae8a6e7"
	mayhemSquadronOracle    = "3b7352e6-6508-491c-9935-426a6eebb429"
)

// mayhemHandCard seeds a catalog card into a player's hand.
func mayhemHandCard(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// mayhemDiscard makes `p` discard the card `id` through the engine's
// ordinary discard prompt, so the discard event is the real one.
func mayhemDiscard(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: p.ID, N: 1})
	})
	answerDiscard(t, g, p.ID, id)
	if !p.Graveyard.Contains(id) {
		t.Fatalf("the discarded card is not in the graveyard")
	}
}

func mayhemDiscarded(g *game.Game, id uuid.UUID) bool {
	var got bool
	g.WithWriteLock(func() { got = g.DiscardedThisTurn(id) })
	return got
}

// mayhemSlashGame is a main-phase table with Sadistic Slash discarded
// from seat 0's hand and a 3/3 on seat 1's side to aim it at.
func mayhemSlashGame(t *testing.T) (g *game.Game, me *game.Player, slash, victim uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	toMainForCost(t, g)
	victim = seedCopyableCreature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 3, 3)
	slash = mayhemHandCard(g, me, "Sadistic Slash", "Instant", "{3}{B}", mayhemSlashOracle)
	return g, me, slash, victim
}

func castSlashFromGraveyard(g *game.Game, me *game.Player, slash, victim uuid.UUID, alt string) error {
	return g.CastSpell(me.ID, slash, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: alt,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
}

// The headline: discarded this turn, castable from the graveyard for
// the mayhem cost and for nothing else.
func TestMayhemDiscardedThisTurnCastsFromTheGraveyardForTheMayhemCost(t *testing.T) {
	g, me, slash, victim := mayhemSlashGame(t)
	mayhemDiscard(t, g, me, slash)
	if !mayhemDiscarded(g, slash) {
		t.Fatal("the discard was not recorded")
	}

	// The printed cost is not an option out of the graveyard.
	floatForTest(g, me, "CCCB")
	if err := castSlashFromGraveyard(g, me, slash, victim, ""); !errors.Is(err, game.ErrCastCostRequired) {
		t.Fatalf("printed cost from the graveyard: err = %v, want ErrCastCostRequired", err)
	}
	// One mana short of {1}{B} is refused with the card left alone.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, "B")
	if err := castSlashFromGraveyard(g, me, slash, victim, "mayhem"); err == nil {
		t.Fatal("mayhem {1}{B} paid with one mana")
	}
	if !me.Graveyard.Contains(slash) {
		t.Fatal("a refused cast moved the card")
	}

	floatForTest(g, me, "C")
	if err := castSlashFromGraveyard(g, me, slash, victim, "mayhem"); err != nil {
		t.Fatalf("mayhem cast: %v", err)
	}
	if item := g.StackMeta[slash]; item == nil || item.AltCost != "mayhem" || item.CastFromZone != game.ZoneGraveyard {
		t.Errorf("stack item = %+v, want AltCost mayhem from the graveyard", item)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("-5/-5 left the 3/3 alive")
	}
	// No exile clause: the spell is in the graveyard again, and as the
	// same card it is a new object that was not discarded.
	if !me.Graveyard.Contains(slash) {
		t.Fatal("a mayhem spell is not exiled when it resolves")
	}
	if mayhemDiscarded(g, slash) {
		t.Error("the spell that came back from the stack counts as discarded")
	}
}

// A card put into the graveyard some other way was never discarded.
func TestMayhemMilledCardIsNotCastable(t *testing.T) {
	g, me, slash, victim := mayhemSlashGame(t)
	g.WithWriteLock(func() {
		_, _ = me.Hand.Remove(slash)
		me.Library.PushTop(game.Card{
			InstanceID: slash, Name: "Sadistic Slash", TypeLine: "Instant", ManaCost: "{3}{B}",
			OracleID: mayhemSlashOracle, Owner: me.ID, Controller: me.ID,
		})
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Errorf("mill: %v", err)
		}
	})
	if !me.Graveyard.Contains(slash) {
		t.Fatal("fixture: the card was not milled")
	}
	floatForTest(g, me, "CCCBB")
	for _, alt := range []string{"mayhem", ""} {
		if err := castSlashFromGraveyard(g, me, slash, victim, alt); !errors.Is(err, game.ErrCastZoneNotAllowed) {
			t.Errorf("milled card, alternative cost %q: err = %v, want ErrCastZoneNotAllowed", alt, err)
		}
	}
	if offers := mayhemOffers(g, me, slash); len(offers) != 0 {
		t.Errorf("a milled card is offered %v", offers)
	}
}

// "This turn" ends with the turn: the record is flushed, and the card
// is a graveyard card like any other the next turn.
func TestMayhemDiscardedLastTurnIsNotCastable(t *testing.T) {
	g, me, slash, victim := mayhemSlashGame(t)
	mayhemDiscard(t, g, me, slash)
	advanceToMainOf(t, g, 1)
	if mayhemDiscarded(g, slash) {
		t.Error("the record outlived the turn")
	}
	g.WithWriteLock(func() {
		if len(g.TurnTally.Resolved) != 0 {
			t.Errorf("turn tally still holds %v", g.TurnTally.Resolved)
		}
	})
	floatForTest(g, me, "CCCBB")
	if err := castSlashFromGraveyard(g, me, slash, victim, "mayhem"); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("discarded last turn: err = %v, want ErrCastZoneNotAllowed", err)
	}
}

// CR 400.7: discarded, left the graveyard, came back by another route
// — a new object that was not discarded.
func TestMayhemCardThatLeftTheGraveyardAndReturnedIsNew(t *testing.T) {
	g, me, slash, victim := mayhemSlashGame(t)
	mayhemDiscard(t, g, me, slash)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(slash, game.ZoneHand); err != nil {
			t.Errorf("return to hand: %v", err)
		}
		if err := g.TuckToLibraryForEffect(slash, false); err != nil {
			t.Errorf("tuck: %v", err)
		}
		if err := g.MillNForEffect(me.ID, 1); err != nil {
			t.Errorf("mill: %v", err)
		}
	})
	if !me.Graveyard.Contains(slash) {
		t.Fatal("fixture: the card is not back in the graveyard")
	}
	if mayhemDiscarded(g, slash) {
		t.Fatal("the record followed the card through two zone changes")
	}
	floatForTest(g, me, "CCCBB")
	if err := castSlashFromGraveyard(g, me, slash, victim, "mayhem"); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("returned card: err = %v, want ErrCastZoneNotAllowed", err)
	}
}

// Discarded again later in the same turn, it is castable again: a
// mayhem spell that went back to the graveyard is just a card.
func TestMayhemTheSameCardDiscardedAgainIsCastableAgain(t *testing.T) {
	g, me, slash, victim := mayhemSlashGame(t)
	mayhemDiscard(t, g, me, slash)
	floatForTest(g, me, "CB")
	if err := castSlashFromGraveyard(g, me, slash, victim, "mayhem"); err != nil {
		t.Fatalf("first mayhem: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.ReturnFromGraveyardForEffect(slash, game.ZoneHand); err != nil {
			t.Errorf("return to hand: %v", err)
		}
	})
	mayhemDiscard(t, g, me, slash)
	victim2 := seedCopyableCreature(g, g.Seats[1].ID, "Their Other Bear", "Creature — Bear", 2, 2)
	floatForTest(g, me, "CB")
	if err := castSlashFromGraveyard(g, me, slash, victim2, "mayhem"); err != nil {
		t.Fatalf("second mayhem after a second discard: %v", err)
	}
}

// The offer list the view's stamp and the bot's enumerator read.
func mayhemOffers(g *game.Game, me *game.Player, id uuid.UUID) []string {
	var keys []string
	g.WithWriteLock(func() {
		var card game.Card
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == id {
				card = c
			}
		}
		for _, o := range g.CastOffersForLocked(me.ID, card, game.ZoneGraveyard, nil) {
			if o == nil {
				keys = append(keys, "")
			} else {
				keys = append(keys, o.Key)
			}
		}
	})
	return keys
}

func TestMayhemOnlyOffersTheMayhemPriceForADiscardedCard(t *testing.T) {
	g, me, slash, _ := mayhemSlashGame(t)
	mayhemDiscard(t, g, me, slash)
	floatForTest(g, me, "CB")
	offers := mayhemOffers(g, me, slash)
	if len(offers) != 1 || offers[0] != "mayhem" {
		t.Fatalf("offers = %q, want exactly mayhem", offers)
	}
}
