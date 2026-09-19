package legal_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// non_hand_cast_test.go — #673. The enumerator offers a cast from
// EVERY zone the engine would accept one from, at EVERY price it
// would charge: the printed cost, and each alternative cost the card
// or a permission makes claimable from that zone.
//
// Every test here ends in dispatchAll, which is the contract: each
// move is applied to a fresh clone of the board it was enumerated on
// and the dispatcher must accept it. A move that names the wrong
// price, the wrong zone, the wrong face or an unpayable card
// component fails there rather than in an assertion about labels.

// Catalog oracle IDs for the cost families. One real card per family
// rather than a synthetic catalog, because the thing under test is
// that the enumerator agrees with the engine about cards as they are
// actually registered.
const (
	oracleFaithlessLooting  = "3d6fa57a-aa53-4b5c-b8af-a7612c823117" // flashback, graveyard
	oracleLingeringSouls    = "0b8c3337-04dd-4798-8203-6d8b8cfb936b" // flashback, no targets
	oracleVoraciousTyphon   = "f62784c0-9c67-4a05-a09b-aabf03a9390f" // escape, exile four
	oracleCyclonicRift      = "d75b9c82-1b49-4c3e-a1b5-aeef57d6644b" // overload, clears targets
	oracleForceOfWill       = "956381ba-6d37-4a8a-846c-bad79222dbee" // pitch: 1 life + a blue card
	oracleSnuffOut          = "324824cb-f938-401c-b9b5-d8908b431ef0" // pay 4 life, Swamp condition
	oracleDaze              = "70486bee-6ee7-41ea-b834-8caf4699302b" // return an Island
	oracleSolitude          = "dcb9c2a7-ae54-4ddc-a567-640bf4bf4366" // evoke: exile a white card
	oracleWeftstalkerArdent = "926d52a5-4db1-46ce-9567-17c28bf56ae7" // warp
)

// lands seeds n untapped basics of one subtype, so a scenario can
// fund a coloured cost. `mana` only ever produces Islands.
func basicLands(g *game.Game, p *game.Player, n int, subtype string) {
	for i := 0; i < n; i++ {
		battlefieldCard(g, p, basic(fmt.Sprintf("%s %d", subtype, i), subtype))
	}
}

// castSpellForTest puts a spell on the stack so a counterspell in the
// scenario has something to target, and hands priority back to the
// seat under test. The engine's own cast path, not a hand-built
// StackItem: a probe that faked the stack would not be testing what
// the enumerator sees.
func castSpellForTest(t *testing.T, g *game.Game, caster, cardID, target uuid.UUID) {
	t.Helper()
	params := game.CastSpellParams{Strict: true, AutoTap: true}
	if spec := game.TargetSpecFor(game.CatalogKey(*findCardForTest(t, g, cardID))); spec != nil {
		params.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: target}}
	}
	if err := g.CastSpell(caster, cardID, params); err != nil {
		t.Fatalf("seed the stack: %v", err)
	}
}

// findCardForTest is the scenario's card lookup by instance ID.
func findCardForTest(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	var out *game.Card
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			for _, z := range []*game.Zone{p.Hand, p.Graveyard, p.Command, p.Library} {
				if z == nil {
					continue
				}
				for i := range z.Cards {
					if z.Cards[i].InstanceID == id {
						c := z.Cards[i]
						out = &c
						return
					}
				}
			}
		}
		for _, z := range []*game.Zone{g.Battlefield, g.Exile, g.Stack} {
			if z == nil {
				continue
			}
			for i := range z.Cards {
				if z.Cards[i].InstanceID == id {
					c := z.Cards[i]
					out = &c
					return
				}
			}
		}
	})
	if out == nil {
		t.Fatalf("card %s is in no zone", id)
	}
	return out
}

// castParamsOf decodes the cast payloads one source produced. The
// whole payload, because #673 is about fields the older helpers
// (castsOf) do not look at: the zone, the claimed cost and the cards
// paid to it.
type castPayload struct {
	FromZone        string   `json:"from_zone"`
	AlternativeCost string   `json:"alternative_cost"`
	AltCostIDs      []string `json:"alt_cost_ids"`
	Targets         []struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	} `json:"targets"`
	Face int `json:"face"`
}

func castPayloadsOf(t *testing.T, moves []legal.Move, src uuid.UUID) []castPayload {
	t.Helper()
	var out []castPayload
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p castPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, p)
	}
	return out
}

// offerKeys is the set of alternative-cost keys a source was offered,
// with "" standing for the printed mana cost.
func offerKeys(t *testing.T, moves []legal.Move, src uuid.UUID) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, p := range castPayloadsOf(t, moves, src) {
		out[p.AlternativeCost]++
	}
	return out
}

// --- flashback (CR 702.34) ---------------------------------------

// The headline of #673: a bot with Faithless Looting in its graveyard
// and three mana is offered the flashback, and is NOT offered the
// printed {1}{R} — a zone the card prices must be paid for.
func TestEnumeratorOffersFlashbackFromTheGraveyard(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Mountain")

	looting := graveyardCard(seat, game.Card{
		Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: oracleFaithlessLooting,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := offerKeys(t, moves, looting)
	if got["flashback"] != 1 {
		t.Errorf("want exactly one flashback cast, offers = %v: %v", got, labels(moves))
	}
	if got[""] != 0 {
		t.Errorf("the printed cost was offered from the graveyard (%v) — CR 601.2b, "+
			"a zone the card prices must be paid for", got)
	}
	if !hasLabel(moves, "Cast Faithless Looting from graveyard (Flashback {2}{R})") {
		t.Errorf("the flashback label does not name the price: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// One mana short of the flashback cost is no move at all, rather than
// a move at the printed price.
func TestEnumeratorOffersNoFlashbackItCannotPayFor(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Mountain")

	looting := graveyardCard(seat, game.Card{
		Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: oracleFaithlessLooting,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	if n := len(castMovesFor(moves, looting)); n != 0 {
		t.Errorf("offered %d casts of an unaffordable flashback: %v", n, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A flashback card with no target clause still reaches the
// enumerator, and the graveyard walk needs no grant to find it —
// AnyCastPermissionsForEffect is false on this board.
func TestFlashbackNeedsNoGrantedPermission(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Swamp")

	souls := graveyardCard(seat, game.Card{
		Name: "Lingering Souls", TypeLine: "Sorcery", ManaCost: "{2}{W}",
		OracleID: oracleLingeringSouls,
	})
	if g.AnyCastPermissionsForEffect() {
		t.Fatal("fixture grants a permission; the point of this test is that none is needed")
	}

	moves := legal.EnumerateFor(g, seat.ID)
	if got := offerKeys(t, moves, souls); got["flashback"] != 1 {
		t.Errorf("want one flashback cast with nothing granted, offers = %v: %v", got, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- escape (CR 702.138), the card component ----------------------

// Escape is the combination search #1004 left open at
// legal/cast.go:169: a price with a CARD half. The move names the
// exact four cards it exiles, and the engine accepts them.
func TestEnumeratorOffersEscapeAndNamesTheExiledCards(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 7, "Forest")

	typhon := graveyardCard(seat, game.Card{
		Name: "Voracious Typhon", TypeLine: "Creature — Hydra", ManaCost: "{4}{G}",
		Power: 4, Toughness: 4, OracleID: oracleVoraciousTyphon,
	})
	fuel := make([]uuid.UUID, 0, 4)
	for i := 0; i < 4; i++ {
		fuel = append(fuel, graveyardCard(seat, game.Card{
			Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant", ManaCost: "{1}",
		}))
	}

	moves := legal.EnumerateFor(g, seat.ID)
	pays := castPayloadsOf(t, moves, typhon)
	var escapes []castPayload
	for _, p := range pays {
		if p.AlternativeCost == "escape" {
			escapes = append(escapes, p)
		}
	}
	if len(escapes) != 1 {
		t.Fatalf("want exactly one escape cast (maxEnumeratedCostPayments), got %d of %v: %v",
			len(escapes), pays, labels(moves))
	}
	if len(escapes[0].AltCostIDs) != 4 {
		t.Fatalf("escape names %d cards, the cost is four: %+v", len(escapes[0].AltCostIDs), escapes[0])
	}
	isFuel := map[string]bool{}
	for _, f := range fuel {
		isFuel[f.String()] = true
	}
	named := map[string]bool{}
	for _, id := range escapes[0].AltCostIDs {
		named[id] = true
		if id == typhon.String() {
			t.Error("escape named the spell itself — CR 601.2a put it on the stack, 'other' means other")
		}
		if !isFuel[id] {
			t.Errorf("escape named %s, which is not one of the four cards in the graveyard", id)
		}
	}
	if len(named) != 4 {
		t.Errorf("escape named a card twice: %v", escapes[0].AltCostIDs)
	}
	if !hasLabel(moves, "Cast Voracious Typhon from graveyard (Escape—{5}{G}{G}") {
		t.Errorf("the escape label does not name the price: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// One card short of escape's four is no escape move: the offer is
// filtered out where the rule lives, not discovered by a rejection.
func TestEscapeIsNotOfferedWithoutEnoughFuel(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 7, "Forest")

	typhon := graveyardCard(seat, game.Card{
		Name: "Voracious Typhon", TypeLine: "Creature — Hydra", ManaCost: "{4}{G}",
		Power: 4, Toughness: 4, OracleID: oracleVoraciousTyphon,
	})
	for i := 0; i < 3; i++ {
		graveyardCard(seat, game.Card{
			Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant", ManaCost: "{1}",
		})
	}

	moves := legal.EnumerateFor(g, seat.ID)
	if n := len(castMovesFor(moves, typhon)); n != 0 {
		t.Errorf("offered %d escape casts with three cards of fuel for a four-card cost: %v",
			n, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- overload (CR 702.96) ----------------------------------------

// Overload is the offer that REWRITES the announcement: an overloaded
// Cyclonic Rift has no targets at all, and an enumerator that carried
// the printed clause into it would have every move refused.
func TestEnumeratorOffersOverloadWithNoTargets(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 7)
	battlefieldCard(g, other, creature("Bear", "{1}{G}", 2, 2))

	rift := handCard(seat, game.Card{
		Name: "Cyclonic Rift", TypeLine: "Instant", ManaCost: "{1}{U}",
		OracleID: oracleCyclonicRift,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := offerKeys(t, moves, rift)
	if got[""] == 0 {
		t.Errorf("the printed targeted cast disappeared: %v", labels(moves))
	}
	if got["overload"] != 1 {
		t.Errorf("want one overload cast, offers = %v: %v", got, labels(moves))
	}
	for _, p := range castPayloadsOf(t, moves, rift) {
		if p.AlternativeCost == "overload" && len(p.Targets) != 0 {
			t.Errorf("an overloaded cast announced %d targets — the clause is deleted", len(p.Targets))
		}
		if p.AlternativeCost == "" && len(p.Targets) != 1 {
			t.Errorf("the printed cast announced %d targets, want one", len(p.Targets))
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- pitch: life AND a card (Force of Will) -----------------------

// Both halves of the pitch are costs. The move claims the offer,
// names the card it exiles, and carries the life on Move.Cost so a
// policy reading only the payload does not price it as free.
func TestEnumeratorOffersAPitchCostWithItsLifeAndItsCard(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	clearHand(other)
	advanceTo(t, g, game.StepPrecombatMain)

	fow := handCard(seat, game.Card{
		Name: "Force of Will", TypeLine: "Instant", ManaCost: "{3}{U}{U}",
		Colors: []string{"U"}, OracleID: oracleForceOfWill,
	})
	pitch := handCard(seat, game.Card{
		Name: "Blue Card", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"},
	})
	// Something to counter. Force of Will targets a spell, so without
	// one on the stack there is no legal announcement at any price.
	boltID := handCard(other, bolt())
	basicLands(g, other, 1, "Mountain")
	castSpellForTest(t, g, other.ID, boltID, seat.ID)

	moves := legal.EnumerateFor(g, seat.ID)
	pays := castPayloadsOf(t, moves, fow)
	var pitched *castPayload
	for i := range pays {
		if pays[i].AlternativeCost == "pitch" {
			pitched = &pays[i]
		}
	}
	if pitched == nil {
		t.Fatalf("no pitch cast of Force of Will: %+v / %v", pays, labels(moves))
	}
	if len(pitched.AltCostIDs) != 1 || pitched.AltCostIDs[0] != pitch.String() {
		t.Errorf("pitch named %v, want the one other blue card %s", pitched.AltCostIDs, pitch)
	}
	for _, m := range moves {
		if m.Source != fow || m.Kind != legal.KindCast {
			continue
		}
		if !strings.Contains(m.Label, "Pay 1 life") {
			continue
		}
		if m.Cost == nil || m.Cost.Life != 1 {
			t.Errorf("the pitch move prices its life as %+v, want 1 (CR 119.4)", m.Cost)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Force of Will cannot pitch itself (CR 601.2a), so a hand holding
// only the Force is not offered the pitch.
func TestPitchIsNotOfferedWithNothingToPitch(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	clearHand(other)
	advanceTo(t, g, game.StepPrecombatMain)

	fow := handCard(seat, game.Card{
		Name: "Force of Will", TypeLine: "Instant", ManaCost: "{3}{U}{U}",
		Colors: []string{"U"}, OracleID: oracleForceOfWill,
	})
	boltID := handCard(other, bolt())
	basicLands(g, other, 1, "Mountain")
	castSpellForTest(t, g, other.ID, boltID, seat.ID)

	moves := legal.EnumerateFor(g, seat.ID)
	if got := offerKeys(t, moves, fow); got["pitch"] != 0 {
		t.Errorf("offered a pitch with nothing to pitch: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- pay life (Snuff Out): the condition, and the bot's own floor --

func TestPayLifeOfferFollowsItsConditionAndTheBotsLifeFloor(t *testing.T) {
	for _, tc := range []struct {
		name  string
		swamp bool
		life  int
		want  int
	}{
		{"no Swamp, no offer", false, 40, 0},
		{"Swamp and life to spare", true, 40, 1},
		{"exactly the price is a losing line the bot declines", true, 4, 0},
		{"below the price is not payable at all", true, 3, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			seat := g.Seats[g.Turn.ActiveSeat]
			other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			clearHand(seat)
			advanceTo(t, g, game.StepPrecombatMain)
			seat.Life = tc.life
			if tc.swamp {
				basicLands(g, seat, 1, "Swamp")
			}
			battlefieldCard(g, other, creature("Bear", "{1}{G}", 2, 2))

			snuff := handCard(seat, game.Card{
				Name: "Snuff Out", TypeLine: "Instant", ManaCost: "{3}{B}",
				OracleID: oracleSnuffOut,
			})

			moves := legal.EnumerateFor(g, seat.ID)
			if got := offerKeys(t, moves, snuff)["pay_life"]; got != tc.want {
				t.Errorf("pay_life offers = %d, want %d: %v", got, tc.want, labels(moves))
			}
			dispatchAll(t, g, seat.ID, moves)
		})
	}
}

// --- return a permanent (Daze) ------------------------------------

func TestEnumeratorOffersDazesBounceAndNamesTheIsland(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	clearHand(other)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 1)

	daze := handCard(seat, game.Card{
		Name: "Daze", TypeLine: "Instant", ManaCost: "{1}{U}", OracleID: oracleDaze,
	})
	boltID := handCard(other, bolt())
	basicLands(g, other, 1, "Mountain")
	castSpellForTest(t, g, other.ID, boltID, seat.ID)

	moves := legal.EnumerateFor(g, seat.ID)
	var returned *castPayload
	for _, p := range castPayloadsOf(t, moves, daze) {
		if p.AlternativeCost == "return" {
			q := p
			returned = &q
		}
	}
	if returned == nil {
		t.Fatalf("no Daze bounce cast offered: %v", labels(moves))
	}
	if len(returned.AltCostIDs) != 1 {
		t.Errorf("the bounce named %v, want one Island", returned.AltCostIDs)
	}
	if !hasLabel(moves, "Cast Daze (Return an Island you control to its owner's hand, returning Island 0)") {
		t.Errorf("the bounce label does not name what it returns: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- warp (CR 702.185) and evoke ----------------------------------

func TestEnumeratorOffersWarpFromHand(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 1, "Mountain")

	ardent := handCard(seat, game.Card{
		Name: "Weftstalker Ardent", TypeLine: "Creature — Spirit", ManaCost: "{3}{R}",
		Power: 3, Toughness: 3, OracleID: oracleWeftstalkerArdent,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := offerKeys(t, moves, ardent)
	if got["warp"] != 1 {
		t.Errorf("want one warp cast off a single Mountain, offers = %v: %v", got, labels(moves))
	}
	if got[""] != 0 {
		t.Errorf("the {3}{R} printed cast was offered off one land: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

func TestEnumeratorOffersEvokesCardPitch(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)

	solitude := handCard(seat, game.Card{
		Name: "Solitude", TypeLine: "Creature — Elemental Incarnation", ManaCost: "{3}{W}{W}",
		Power: 3, Toughness: 2, Colors: []string{"W"}, OracleID: oracleSolitude,
	})
	handCard(seat, game.Card{
		Name: "White Card", TypeLine: "Instant", ManaCost: "{W}", Colors: []string{"W"},
	})

	moves := legal.EnumerateFor(g, seat.ID)
	if got := offerKeys(t, moves, solitude); got["evoke"] != 1 {
		t.Errorf("want one evoke cast, offers = %v: %v", got, labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- granted permissions with a card component --------------------

// The case #1004 called out at legal/cast.go:169 and left open: a
// GRANTED escape (Underworld Breach's shape). The grant prices the
// zone with a card component, and the enumerator now searches it.
func TestEnumeratorOffersAGrantedEscapeWithItsCardComponent(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Mountain")

	spell := graveyardCard(seat, game.Card{
		Name: "Breached Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}", Layout: "normal",
	})
	for i := 0; i < 3; i++ {
		graveyardCard(seat, game.Card{Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant"})
	}
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{
			Player: seat.ID, Zone: game.ZoneGraveyard,
			AltCostKey: "escape", Label: "Escape—{1}{R}, Exile three other cards",
			ExileOtherFromGraveyard: 3, ExileOnResolution: true,
		})
	})

	moves := legal.EnumerateFor(g, seat.ID)
	pays := castPayloadsOf(t, moves, spell)
	if len(pays) != 1 || pays[0].AlternativeCost != "escape" {
		t.Fatalf("want one granted escape cast, got %+v: %v", pays, labels(moves))
	}
	if len(pays[0].AltCostIDs) != 3 {
		t.Errorf("the granted escape named %d cards, the grant asks for three", len(pays[0].AltCostIDs))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- multikicker, and the offer × optional-cost product -----------

// Kicker's enumerator half is TestKickedAndUnkickedAreBothEnumerated
// (cast_gate_test.go). This is its repeatable sibling: multikicker is
// capped at maxEnumeratedRepeats payments, and the cap composes with
// the alternative-cost loop rather than multiplying past it.
func TestMultikickerIsOfferedUpToItsCap(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 12, "Forest")

	wolfbriar := handCard(seat, game.Card{
		Name: "Wolfbriar Elemental", TypeLine: "Creature — Elemental", ManaCost: "{2}{G}{G}",
		Power: 4, Toughness: 4, OracleID: "2f8872fe-84dc-4cda-a253-e7503a5c96a3",
	})

	moves := castMovesFor(legal.EnumerateFor(g, seat.ID), wolfbriar)
	kicks := map[string]bool{}
	for _, m := range moves {
		kicks[m.Label] = true
	}
	// Unkicked, ×1, ×2, ×3 — and nothing above the cap, which is a
	// policy number rather than the card's printed 20.
	if len(kicks) != 4 {
		t.Errorf("want 4 announcements (unkicked + up to maxEnumeratedRepeats), got %d: %v",
			len(kicks), labels(moves))
	}
	if !hasLabel(moves, "Cast Wolfbriar Elemental (Multikicker {G} ×3)") {
		t.Errorf("the triple kick is missing: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- foretell (CR 702.143) and suspend (CR 702.62) ----------------

// A foretold card is an exile cast under a grant that PRICES the
// zone, so the enumerator offers the foretell cost and never the
// printed one. The permission is built exactly as foretell.go:111
// builds it — the enumerator level is the level under test, and
// running the special action and three turns of upkeep here would be
// testing game/foretell.go instead.
func TestEnumeratorOffersAForetoldCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	mana(g, seat, 2)

	behold := exileCardWithGrant(g, seat, game.Card{
		Name: "Behold the Multiverse", TypeLine: "Instant", ManaCost: "{3}{U}",
		OracleID: "d7d2f701-77df-4169-bf98-0d51d6886e9b",
	}, game.CastPermission{
		Player:     seat.ID,
		Zone:       game.ZoneExile,
		AltCostKey: game.AltCostKeyForetell,
		Cost:       "{1}{U}",
		Duration:   game.WhileInZoneDuration(),
		CastOnly:   true,
		Label:      "Foretell",
	})

	moves := legal.EnumerateFor(g, seat.ID)
	got := offerKeys(t, moves, behold)
	if got[game.AltCostKeyForetell] != 1 {
		t.Errorf("want one foretold cast off two lands, offers = %v: %v", got, labels(moves))
	}
	if got[""] != 0 {
		t.Errorf("the printed {3}{U} was offered for a foretold card: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, moves)
}

// A suspended card's last time counter grants a FREE cast: a flat
// "{0}" price and no claimable offer, so the move claims nothing and
// the printed-cost entry is the right one. The shape is
// suspend.go:221's.
//
// Enumerated in the END STEP on purpose. Rift Bolt is a SORCERY, so
// the only thing that makes the cast legal there is the grant's
// TimingFlash (CR 608.2g — the cast happens as the suspend trigger
// resolves, and the trigger fires in an upkeep). An enumerator that
// read the card's timing instead of the permission's would offer
// nothing here.
func TestEnumeratorOffersASuspendedFreeCast(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepEnd)

	riftBolt := exileCardWithGrant(g, seat, game.Card{
		Name: "Rift Bolt", TypeLine: "Sorcery", ManaCost: "{2}{R}",
		OracleID: "2b8afa9f-4236-4c02-a8d5-3c145caecfd6",
	}, game.CastPermission{
		Player:      seat.ID,
		Zone:        game.ZoneExile,
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		GrantsHaste: true,
		Duration:    g.UntilEndOfTurnDuration(),
		Label:       game.SuspendFreeCastLabel,
	})

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castMovesFor(moves, riftBolt)
	if len(casts) == 0 {
		t.Fatalf("no suspended free cast offered in the end step with no mana: %v", labels(moves))
	}
	for _, p := range castPayloadsOf(t, moves, riftBolt) {
		if p.AlternativeCost != "" {
			t.Errorf("the free cast claimed %q; a flat grant price is not a claimable offer", p.AlternativeCost)
		}
		if p.FromZone != "exile" {
			t.Errorf("the free cast came from %q, want exile", p.FromZone)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// --- adventures (CR 715) and madness (CR 702.35) ------------------
//
// Not here, and deliberately: #719 and #657 each shipped their own
// enumerator tests — adventure_moves_test.go and
// madness_cast_moves_test.go — and both were written against the
// walk this file replaced. They still pass, which is the claim worth
// making: an adventure's two castable faces and a madness grant's
// flash-timed exile cast needed NO new enumerator code, because the
// new walk asks the same two engine functions (Card.CastableFaces
// narrowed by CastPermission.Faces, and CastOffersForLocked) the old
// one did. Adding a third copy of those scenarios here would be
// duplicating them rather than covering anything.

// --- the converse of the soundness contract -----------------------

// dispatchAll proves every enumerated move is accepted. This is the
// other direction, for a bounded fixture: every (card, zone, cost
// choice) the ENGINE accepts is one the enumerator offered.
//
// Bounded deliberately, and the bound is stated rather than implied:
// the probe walks the cost dimension exhaustively (each zone the card
// could be in × the printed cost × every alternative cost the catalog
// declares) and takes the engine's own answer for the target and
// payment dimensions, because re-deriving those in the test would be
// re-deriving the enumerator. The gap #673 closed was whole ZONES and
// whole COST FAMILIES missing, which is exactly what this covers.
func TestEveryCastTheEngineAcceptsIsEnumerated(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 4, "Mountain")
	basicLands(g, seat, 4, "Island")
	battlefieldCard(g, other, creature("Bear", "{1}{G}", 2, 2))

	type probe struct {
		id   uuid.UUID
		zone string
		kind game.ZoneKind
	}
	var probes []probe

	probes = append(probes, probe{handCard(seat, game.Card{
		Name: "Cyclonic Rift", TypeLine: "Instant", ManaCost: "{1}{U}",
		OracleID: oracleCyclonicRift,
	}), "hand", game.ZoneHand})
	probes = append(probes, probe{handCard(seat, bolt()), "hand", game.ZoneHand})
	probes = append(probes, probe{graveyardCard(seat, game.Card{
		Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: oracleFaithlessLooting,
	}), "graveyard", game.ZoneGraveyard})
	probes = append(probes, probe{graveyardCard(seat, game.Card{
		Name: "Lingering Souls", TypeLine: "Sorcery", ManaCost: "{2}{W}",
		OracleID: oracleLingeringSouls,
	}), "graveyard", game.ZoneGraveyard})

	moves := legal.EnumerateFor(g, seat.ID)
	offered := map[string]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindCast {
			continue
		}
		var p castPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		offered[m.Source.String()+"|"+p.FromZone+"|"+p.AlternativeCost] = true
	}

	for _, pr := range probes {
		card := findCardForTest(t, g, pr.id)
		keys := []string{""}
		for _, ac := range game.AlternativeCostsOfferedFromZone(game.CatalogKey(*card), pr.kind) {
			keys = append(keys, ac.Key)
		}
		for _, key := range keys {
			params := buildCastProbe(t, g, seat.ID, *card, pr.zone, pr.kind, key)
			clone := g.Clone()
			if err := clone.CastSpell(seat.ID, pr.id, params); err != nil {
				continue // the engine refuses it; the enumerator is right not to offer it
			}
			want := pr.id.String() + "|" + pr.zone + "|" + key
			if !offered[want] {
				t.Errorf("the engine accepts %s from %s for %q and the enumerator never offered it",
					card.Name, pr.zone, key)
			}
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// buildCastProbe assembles the announcement the converse test sends:
// the engine's own legal targets and its own payment candidates, so
// the probe tests the enumerator's COVERAGE rather than its taste.
func buildCastProbe(t *testing.T, g *game.Game, seat uuid.UUID, card game.Card, zone string, kind game.ZoneKind, key string) game.CastSpellParams {
	t.Helper()
	params := game.CastSpellParams{FromZone: zone, AlternativeCost: key, Strict: true, AutoTap: true}
	var offer *game.AlternativeCost
	if key != "" {
		for _, ac := range game.AlternativeCostsOfferedFromZone(game.CatalogKey(card), kind) {
			if ac.Key == key {
				one := ac
				offer = &one
			}
		}
	}
	g.ReadSnapshot(func() {
		if spec := game.TargetSpecUnderAlternativeCost(game.TargetSpecFor(game.CatalogKey(card)), offer); spec != nil && spec.Min > 0 {
			lt := g.LegalTargetsForEffect(game.SourceObject(seat, &card), spec)
			for i := 0; i < spec.Min && i < len(lt.Cards); i++ {
				params.Targets = append(params.Targets, game.TargetRef{Kind: game.TargetCard, ID: lt.Cards[i]})
			}
			for i := len(params.Targets); i < spec.Min && i-len(params.Targets) < len(lt.Players); i++ {
				params.Targets = append(params.Targets, game.TargetRef{Kind: game.TargetPlayer, ID: lt.Players[i]})
			}
		}
		if n := offer.CardPaymentCount(); n > 0 {
			pool := g.AltCostCandidatesLocked(seat, card.InstanceID, offer)
			if len(pool) >= n {
				params.AltCostIDs = append(params.AltCostIDs, pool[:n]...)
			}
		}
	})
	return params
}
