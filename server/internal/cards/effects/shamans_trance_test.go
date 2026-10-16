package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// shamans_trance_test.go — #1987, ADR 0066 amendment 2026-10-16:
// "Other players can't play lands or cast spells from their graveyards
// this turn. You may play lands and cast spells from other players'
// graveyards this turn as though those cards were in your graveyard."

// tranceTable is seat 0's precombat main on a four-seat table, with a
// Mountain, a Lightning Bolt and a Faithless Looting in seat 1's
// graveyard, a Forest in seat 2's, a Plains in seat 0's own, and
// Shaman's Trance cast by seat 0 and resolved, and an untapped Mountain
// under seat 0.
type tranceTable struct {
	g                                   *game.Game
	me, opp, third                      *game.Player
	oppLand, oppBolt, oppLooting, land2 uuid.UUID
	myLand                              uuid.UUID
}

func tranceGraveCard(p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func newTranceTable(t *testing.T) tranceTable {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	tt := tranceTable{g: g, me: g.Seats[0], opp: g.Seats[1], third: g.Seats[2]}
	tt.oppLand = tranceGraveCard(tt.opp, "Mountain", "Basic Land — Mountain", "", "")
	tt.oppBolt = tranceGraveCard(tt.opp, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}")
	tt.oppLooting = tranceGraveCard(tt.opp, "Faithless Looting", "Sorcery", faithlessLootingOracle, "{R}")
	tt.land2 = tranceGraveCard(tt.third, "Forest", "Basic Land — Forest", "", "")
	tt.myLand = tranceGraveCard(tt.me, "Plains", "Basic Land — Plains", "", "")
	// A graveyard is public (CR 404.2): every seat knows its cards, as
	// a card that got there by being put there would be.
	everyone := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		everyone[p.ID] = true
	}
	for _, p := range g.Seats {
		for i := range p.Graveyard.Cards {
			p.Graveyard.Cards[i].KnownBy = everyone
		}
	}
	castCatalogSpell(t, g, shamansTranceName, "Instant", shamansTranceOracle, nil)
	passPriorityAroundTable(t, g)
	// One untapped Mountain, so the {R} spells are affordable to the
	// enumerator and the view, which price a cast before offering it.
	pushLandFor(g, tt.me.ID, "Mountain", "Basic Land — Mountain")
	return tt
}

func graveyardPermission(g *game.Game, player, card uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		perm = g.CastPermissionForLocked(player, c, game.ZoneGraveyard)
	})
	return perm
}

func TestShamansTranceOpensEveryOtherGraveyardAndNotYourOwn(t *testing.T) {
	tt := newTranceTable(t)
	g := tt.g
	for _, id := range []uuid.UUID{tt.oppLand, tt.oppBolt, tt.oppLooting, tt.land2} {
		if graveyardPermission(g, tt.me.ID, id) == nil {
			t.Errorf("card %s in another player's graveyard is not open to the caster", id)
		}
	}
	// "Other players' graveyards": the caster's own pile is untouched,
	// and no other player gets anything.
	if graveyardPermission(g, tt.me.ID, tt.myLand) != nil {
		t.Error("the Trance opened the caster's own graveyard")
	}
	if graveyardPermission(g, tt.third.ID, tt.oppBolt) != nil {
		t.Error("another player got a permission over seat 1's graveyard")
	}
}

// A land from another player's graveyard is a land play: it enters
// under the caster's control (CR 110.2), its owner is unchanged, and it
// spends the turn's land drop (CR 305.2).
func TestShamansTranceLandFromAnotherGraveyardIsTheTurnsLandPlay(t *testing.T) {
	tt := newTranceTable(t)
	g, me := tt.g, tt.me
	if err := g.CastSpell(me.ID, tt.oppLand, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("play the Mountain from seat 1's graveyard: %v", err)
	}
	var land game.Card
	g.ReadSnapshot(func() { land, _ = g.LookupCardForEffect(tt.oppLand) })
	if !g.Battlefield.Contains(tt.oppLand) || land.Controller != me.ID || land.Owner != tt.opp.ID {
		t.Fatalf("the Mountain is not on the battlefield under the caster with its owner kept: %+v", land)
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("lands played = %d, want 1", got)
	}
	if err := g.CastSpell(me.ID, tt.land2, game.CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, game.ErrLandDropUnavailable) {
		t.Errorf("a second land: err = %v, want ErrLandDropUnavailable", err)
	}
}

// A spell cast from another player's graveyard is the caster's spell,
// and goes back to its OWNER's graveyard when it resolves (CR 400.3).
func TestShamansTranceSpellFromAnotherGraveyardReturnsToItsOwner(t *testing.T) {
	tt := newTranceTable(t)
	g, me, opp := tt.g, tt.me, tt.opp
	life := tt.third.Life
	if err := g.CastSpell(me.ID, tt.oppBolt, game.CastSpellParams{
		FromZone: "graveyard",
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: tt.third.ID}},
	}); err != nil {
		t.Fatalf("cast the Bolt from seat 1's graveyard: %v", err)
	}
	if item := g.StackMeta[tt.oppBolt]; item == nil || item.Controller != me.ID {
		t.Fatalf("the Bolt on the stack is not the caster's: %+v", item)
	}
	passPriorityAroundTable(t, g)
	if tt.third.Life != life-3 {
		t.Errorf("seat 2 life = %d, want %d", tt.third.Life, life-3)
	}
	if !opp.Graveyard.Contains(tt.oppBolt) || me.Graveyard.Contains(tt.oppBolt) {
		t.Error("the resolved Bolt did not go back to its owner's graveyard")
	}
}

// "As though those cards were in your graveyard": a flashback card in
// another player's graveyard may be flashed back (the 2004-10-04
// ruling) or cast for its mana cost, and the caster picks.
func TestShamansTranceFlashbackCardOffersBothPrices(t *testing.T) {
	tt := newTranceTable(t)
	g, me := tt.g, tt.me
	var offers []string
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(tt.oppLooting)
		perm := g.CastPermissionForLocked(me.ID, c, game.ZoneGraveyard)
		for _, o := range g.CastOffersForLocked(me.ID, c, game.ZoneGraveyard, perm) {
			if o == nil {
				offers = append(offers, "printed")
				continue
			}
			offers = append(offers, o.Key)
		}
	})
	if strings.Join(offers, ",") != "printed,flashback" {
		t.Fatalf("offers = %v, want [printed flashback]", offers)
	}
	if err := g.CastSpell(me.ID, tt.oppLooting, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flash back seat 1's Faithless Looting: %v", err)
	}
	if item := g.StackMeta[tt.oppLooting]; item == nil || item.AltCost != "flashback" || item.Controller != me.ID {
		t.Errorf("the flashback cast is not the caster's flashback: %+v", item)
	}
}

// The ban: every other player is refused a land or a spell out of their
// OWN graveyard, flashback included, and nothing else.
func TestShamansTranceBansOtherPlayersOwnGraveyards(t *testing.T) {
	tt := newTranceTable(t)
	g, me, opp := tt.g, tt.me, tt.opp
	var cast, land, hand, mine error
	g.ReadSnapshot(func() {
		looting, _ := g.LookupCardForEffect(tt.oppLooting)
		cast = g.CastGateLocked(opp.ID, looting, game.ZoneGraveyard, game.CastSpellParams{AlternativeCost: "flashback"})
		hand = g.CastGateLocked(opp.ID, looting, game.ZoneHand, game.CastSpellParams{})
		forest, _ := g.LookupCardForEffect(tt.land2)
		land = g.LandPlayGateLocked(tt.third.ID, forest, game.ZoneGraveyard)
		plains, _ := g.LookupCardForEffect(tt.myLand)
		mine = g.LandPlayGateLocked(me.ID, plains, game.ZoneGraveyard)
	})
	var cant *game.CantCastError
	if !errors.As(cast, &cant) || !strings.Contains(cant.Reason, shamansTranceName) {
		t.Errorf("seat 1 flashing back from their own graveyard: %v, want the Trance's ban", cast)
	}
	if !errors.Is(land, game.ErrCantPlayLand) {
		t.Errorf("seat 2 playing a land from their own graveyard: %v, want ErrCantPlayLand", land)
	}
	if hand != nil {
		t.Errorf("the ban reached a hand: %v", hand)
	}
	if mine != nil {
		t.Errorf("the ban reached the caster: %v", mine)
	}
}

// Both halves end with the turn.
func TestShamansTranceEndsWithTheTurn(t *testing.T) {
	tt := newTranceTable(t)
	g := tt.g
	passTurnsTo(t, g, 1)
	advanceToMain(t, g)
	if graveyardPermission(g, tt.me.ID, tt.oppBolt) != nil {
		t.Error("the permission outlived the turn")
	}
	var cast error
	g.ReadSnapshot(func() {
		looting, _ := g.LookupCardForEffect(tt.oppLooting)
		cast = g.CastGateLocked(tt.opp.ID, looting, game.ZoneGraveyard, game.CastSpellParams{AlternativeCost: "flashback"})
	})
	if cast != nil {
		t.Errorf("the ban outlived the turn: %v", cast)
	}
}

// The bot is offered the other graveyards, and not its own.
func TestShamansTranceEnumeratorOffersTheOtherGraveyards(t *testing.T) {
	tt := newTranceTable(t)
	g, me := tt.g, tt.me
	if !enumeratedFor(g, me.ID, tt.oppLand) || !enumeratedFor(g, me.ID, tt.oppBolt) || !enumeratedFor(g, me.ID, tt.land2) {
		t.Error("the enumerator does not offer the cards in the other graveyards")
	}
	if enumeratedFor(g, me.ID, tt.myLand) {
		t.Error("the enumerator offers a land from the caster's own graveyard")
	}
}

// The client's graveyard browser reads `castable_here`, the viewer's
// own answer: set on another player's graveyard card for the caster,
// for a land and a spell, and for nobody else.
func TestShamansTranceViewStampsTheCastersPlays(t *testing.T) {
	tt := newTranceTable(t)
	g := tt.g
	stamped := func(viewer, owner, card uuid.UUID) bool {
		v := protocol.ViewOfGameFor(g, viewer.String())
		for _, s := range v.Seats {
			if s.ID != owner.String() {
				continue
			}
			for _, c := range s.Graveyard.Cards {
				if c.InstanceID == card.String() {
					return c.CastableHere
				}
			}
		}
		t.Fatalf("card %s is not in seat %s's graveyard view", card, owner)
		return false
	}
	if !stamped(tt.me.ID, tt.opp.ID, tt.oppLand) || !stamped(tt.me.ID, tt.opp.ID, tt.oppBolt) {
		t.Error("the caster's view does not offer seat 1's land and Bolt")
	}
	if !stamped(tt.me.ID, tt.third.ID, tt.land2) {
		t.Error("the caster's view does not offer seat 2's land")
	}
	if stamped(tt.opp.ID, tt.opp.ID, tt.oppLooting) {
		t.Error("seat 1 is offered a flashback out of their banned graveyard")
	}
	if stamped(tt.third.ID, tt.opp.ID, tt.oppBolt) {
		t.Error("a bystander is offered seat 1's Bolt")
	}
	if stamped(tt.me.ID, tt.me.ID, tt.myLand) {
		t.Error("the caster is offered a land from their own graveyard")
	}
	// The banned owner's flashback card says why it is refused.
	for _, s := range protocol.ViewOfGameFor(g, tt.opp.ID.String()).Seats {
		if s.ID != tt.opp.ID.String() {
			continue
		}
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == tt.oppLooting.String() && !strings.Contains(c.CantCast, shamansTranceName) {
				t.Errorf("seat 1's Faithless Looting cant_cast = %q, want the Trance's ban", c.CantCast)
			}
		}
	}
}

func TestShamansTranceSurvivesARestore(t *testing.T) {
	tt := newTranceTable(t)
	g := restoreRoundTrip(t, tt.g, true)
	me, opp := g.Seats[0], g.Seats[1]
	if graveyardPermission(g, me.ID, tt.oppBolt) == nil {
		t.Error("after a restore the permission is gone")
	}
	var cast error
	g.ReadSnapshot(func() {
		looting, _ := g.LookupCardForEffect(tt.oppLooting)
		cast = g.CastGateLocked(opp.ID, looting, game.ZoneGraveyard, game.CastSpellParams{AlternativeCost: "flashback"})
	})
	if cast == nil {
		t.Error("after a restore the ban is gone")
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == tt.myLand {
			t.Error("after a restore the caster's own graveyard is open")
		}
	}
}
