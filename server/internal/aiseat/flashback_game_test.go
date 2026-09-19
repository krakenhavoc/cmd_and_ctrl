package aiseat_test

import (
	"context"
	"encoding/json"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// flashback_game_test.go is the whole-game half of #673: a bot with a
// flashback card in its GRAVEYARD, playing a real game against the
// real engine and the real enumerator, actually casting it from
// there.
//
// It is here rather than in internal/legal because the unit tests
// next door already prove the move is offered and that the engine
// accepts it. What can only be checked at this level is the whole
// chain: the card is registered with its zone and its offer, the
// enumerator walks the graveyard, the move reaches the policy with a
// price it can read, the policy takes it, and the dispatcher applies
// it — with no rejection anywhere. Before this issue the chain was
// broken at the second link and a bot never flashed anything back in
// its life.

const (
	oracleLingeringSoulsBot = "0b8c3337-04dd-4798-8203-6d8b8cfb936b"
	oracleLoathsomeChimera  = "fe983088-6e0d-4544-90bd-f5ebb95c3418"
)

// castWatcher wraps a policy and records the cast moves it chose,
// with the zone each was cast from.
type castWatcher struct {
	inner aiseat.Policy

	mu       sync.Mutex
	fromZone map[string]int
	offered  int
	// bySource counts graveyard offers per card, so a test can say
	// WHICH card was never offered rather than only that something
	// was missing.
	bySource map[uuid.UUID]int
	// paid counts the alternative-cost keys the bot claimed.
	paid map[string]int
}

func newCastWatcher(inner aiseat.Policy) *castWatcher {
	return &castWatcher{
		inner:    inner,
		fromZone: map[string]int{},
		bySource: map[uuid.UUID]int{},
		paid:     map[string]int{},
	}
}

// offeredFor is how many graveyard casts of one card reached the
// policy across the run.
func (p *castWatcher) offeredFor(src uuid.UUID) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.bySource[src]
}

func (p *castWatcher) Name() string { return p.inner.Name() }

// TargetOrder forwards #687's ordering hook, so the wrapped policy is
// still the TargetOrderer the runner looks for. A wrapper that
// swallowed it would quietly turn the ordering off for every test
// that uses one.
func (p *castWatcher) TargetOrder(in aiseat.Input) legal.TargetOrder {
	o, ok := p.inner.(aiseat.TargetOrderer)
	if !ok {
		return nil
	}
	return o.TargetOrder(in)
}

func (p *castWatcher) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	graveyardOffers := 0
	perSource := map[uuid.UUID]int{}
	for i := range in.Moves {
		if zoneOfCast(in.Moves[i]) == "graveyard" {
			graveyardOffers++
			perSource[in.Moves[i].Source]++
		}
	}
	d, err := p.inner.Decide(ctx, in)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.offered += graveyardOffers
	for id, n := range perSource {
		p.bySource[id] += n
	}
	if err == nil && d.Index >= 0 && d.Index < len(in.Moves) {
		if z := zoneOfCast(in.Moves[d.Index]); z != "" {
			p.fromZone[z]++
			p.paid[altCostOfCast(in.Moves[d.Index])]++
		}
	}
	return d, err
}

// paidKeys is the set of CR 118.9 prices the bot actually claimed,
// with "" for the printed mana cost.
func (p *castWatcher) paidKeys() map[string]int {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make(map[string]int, len(p.paid))
	for k, v := range p.paid {
		out[k] = v
	}
	return out
}

func (p *castWatcher) snapshot() (map[string]int, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make(map[string]int, len(p.fromZone))
	for k, v := range p.fromZone {
		out[k] = v
	}
	return out, p.offered
}

// zoneOfCast reads a cast move's source zone off the wire payload,
// the way a policy would. "" for anything that is not a cast.
func zoneOfCast(m legal.Move) string {
	if m.Kind != legal.KindCast {
		return ""
	}
	var p struct {
		FromZone string `json:"from_zone"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		return ""
	}
	if p.FromZone == "" {
		return "hand"
	}
	return p.FromZone
}

// altCostOfCast reads the CR 118.9 price a cast move claims off the
// wire payload, the way a policy would. "" is the printed mana cost.
func altCostOfCast(m legal.Move) string {
	var p struct {
		AlternativeCost string `json:"alternative_cost"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		return ""
	}
	return p.AlternativeCost
}

// newGraveyardRoom seats the bot behind a deck of nothing but
// Forests, with an ordinary deck opposite.
//
// The land-only deck is the fixture's whole trick, and it is there to
// make the assertion mean something. On an ordinary deck the bot is
// spoilt for choice — a 4/4 Ogre out of hand outranks most things —
// so "the bot did not cast out of its graveyard" would be a statement
// about the SCORER rather than about whether the move exists at all.
// With nothing else to do, the graveyard is the only place a play can
// come from.
//
// The commander is red and the lands are green, so the one other
// nonland cast on offer is unaffordable all game.
func newGraveyardRoom(t *testing.T, seed uint64) *ws.Room {
	t.Helper()
	g := game.NewGame()
	forests := func(owner uuid.UUID) []game.Card {
		cmdr := game.NewCommander("Commander Bear", owner)
		cmdr.TypeLine = "Legendary Creature — Bear"
		cmdr.ManaCost = "{2}{R}"
		cmdr.Power, cmdr.Toughness = 3, 3
		deck := []game.Card{cmdr}
		for i := 0; i < 40; i++ {
			c := game.NewCard("Forest", owner)
			c.TypeLine = "Basic Land — Forest"
			deck = append(deck, c)
		}
		return deck
	}
	if _, err := g.AddPlayer("Bot0", forests(uuid.Nil)); err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if _, err := g.AddPlayer("Bot1", battleDeck(uuid.Nil)); err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	if err := g.Start(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return ws.NewRoom(g, testLogger(), "")
}

// seedGraveyard drops a card into the bot's graveyard, known to the
// table the way a graveyard card always is.
func seedGraveyard(g *game.Game, owner *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner.ID, owner.ID
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		c.KnownBy[p.ID] = true
	}
	owner.Graveyard.PushTop(c)
	return c.InstanceID
}

// seedBattlefield puts an untapped, already-settled permanent in play.
func seedBattlefield(g *game.Game, owner *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner.ID, owner.ID
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		c.KnownBy[p.ID] = true
	}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// The test. A flashback sorcery and an escape creature in the
// graveyard, mana for both, and an opponent that does nothing.
//
// Three claims, in the order the chain runs: both offers reach the
// policy (the enumerator walked the graveyard), the bot takes them
// (the scorer can price a cast whose card is not in its hand), and
// each one claimed its own CR 118.9 key on the wire — a graveyard
// cast that claimed nothing would have been refused by CastSpell, so
// the last assertion is what proves the price travelled end to end.
//
// Before #673 the chain broke at the first link and none of it ran.
func TestBotCastsOutOfItsGraveyardInARun(t *testing.T) {
	requireGameTests(t)

	room := newGraveyardRoom(t, 6730)
	g := room.Game
	bot := g.Seats[0]

	// The escape fuel goes in FIRST: the enumerator's single payment
	// takes the oldest cards in the graveyard (maxEnumeratedCostPayments),
	// and a fixture that let it eat its own test subjects would be
	// asserting about whichever one survived.
	for i := 0; i < 3; i++ {
		seedGraveyard(g, bot, game.Card{Name: "Fuel", TypeLine: "Instant", ManaCost: "{1}"})
	}
	souls := seedGraveyard(g, bot, game.Card{
		Name: "Lingering Souls", TypeLine: "Sorcery", ManaCost: "{2}{W}",
		OracleID: oracleLingeringSoulsBot,
	})
	chimera := seedGraveyard(g, bot, game.Card{
		Name: "Loathsome Chimera", TypeLine: "Creature — Chimera", ManaCost: "{3}{G}",
		Power: 3, Toughness: 3, OracleID: oracleLoathsomeChimera,
	})
	// Five Forests for the escape's {4}{G} and two Swamps for the
	// flashback's {1}{B} — both offers have to be AFFORDABLE, or
	// "never offered" would mean "could not pay for it" and the
	// assertion below would be about the mana gate.
	for i := 0; i < 5; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	for i := 0; i < 2; i++ {
		seedBattlefield(g, bot, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	}

	watch := newCastWatcher(heuristic.New())
	res := playGameIn(t, room, 6730, []aiseat.Policy{watch, passPolicy{}}, 6, 120*time.Second)
	zones, offered := watch.snapshot()

	for name, id := range map[string]uuid.UUID{"Lingering Souls": souls, "Loathsome Chimera": chimera} {
		if watch.offeredFor(id) == 0 {
			t.Errorf("%s was never offered out of the graveyard in %d turns — "+
				"the enumerator did not walk the zone", name, res.turns)
		}
	}
	if offered == 0 {
		t.Fatalf("no graveyard cast was offered at all in %d turns; the rest of this test proves nothing",
			res.turns)
	}
	if zones["graveyard"] == 0 {
		t.Errorf("the bot was offered %d graveyard casts across %d turns and took none (chose: %v)",
			offered, res.turns, zones)
	}
	// And the prices it claimed were the cards' own, not the printed
	// cost: a graveyard cast that claimed nothing would have been
	// refused by CastSpell, so this is the assertion that the move
	// carried its CR 118.9 key all the way through.
	paid := watch.paidKeys()
	if paid["flashback"] == 0 {
		t.Errorf("the bot never flashed anything back; prices claimed: %v", paid)
	}
	if paid["escape"] == 0 {
		t.Errorf("the bot never escaped anything; prices claimed: %v", paid)
	}
	// And it really resolved.
	var chimeras int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == bot.ID && c.Name == "Loathsome Chimera" {
				chimeras++
			}
		}
	})
	if chimeras == 0 {
		t.Errorf("no escaped Chimera on the board after %d turns", res.turns)
	}
	t.Logf("%d graveyard casts offered, taken from %v, %d Chimeras, %d turns", offered, zones, chimeras, res.turns)
	assertNoEnumeratorBugs(t, res)
}
