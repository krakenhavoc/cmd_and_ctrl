package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// life_alt_cost_offer_test.go — #695, the card half. Force of Will
// and Snuff Out are the two printed offers with a life component, and
// the bug was that the view offered them to a player who could not
// pay: the picker listed "Pay 4 life", the player chose it, and
// CastSpell answered ErrInvalidParam.
//
// So each card is asserted twice on the same board — the offer the
// VIEW shows, and what CastSpell does with it — because the whole
// point is that the two agree. The engine-level boundary cases live
// in internal/game/alternative_cost_payable_test.go and the projection
// in internal/protocol; these hold the claim to the real cards.

// visibleHandCard is handCardFull plus the visibility the projection
// needs: a card its owner cannot SEE is redacted, and redaction
// clears the alternative-cost offers along with the mana cost — so a
// fixture without KnownBy would "pass" the hidden-offer assertions
// for the wrong reason.
func visibleHandCard(p *game.Player, name, typeLine, manaCost, oracle string, colors []string) uuid.UUID {
	id := handCardFull(p, name, typeLine, manaCost, oracle, colors)
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{p.ID: true}
		}
	}
	return id
}

// offeredKeys reads the alternative-cost keys the view stamps on a
// card in the seat's own hand.
func offeredKeys(t *testing.T, g *game.Game, seat *game.Player, id uuid.UUID) []string {
	t.Helper()
	v := protocol.ViewOfGameFor(g, seat.ID.String())
	for _, s := range v.Seats {
		if s.ID != seat.ID.String() {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.InstanceID != id.String() {
				continue
			}
			keys := make([]string, 0, len(c.AlternativeCosts))
			for _, offer := range c.AlternativeCosts {
				keys = append(keys, offer.Key)
			}
			return keys
		}
	}
	t.Fatalf("the hand card is missing from its owner's view")
	return nil
}

func hasKey(keys []string, want string) bool {
	for _, k := range keys {
		if k == want {
			return true
		}
	}
	return false
}

// Force of Will's pitch costs 1 life. At 1 life it is on the table
// and the cast goes through; at 0 it is neither offered nor accepted.
func TestForceOfWillPitchOfferedOnlyWhenTheLifeIsThere(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	fow := visibleHandCard(me, "Force of Will", "Instant", "{3}{U}{U}", forceOfWillOracle, []string{"U"})
	pitch := visibleHandCard(me, "Brainstorm", "Instant", "{U}", "test-blue-pitch", []string{"U"})
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}

	// Paying down to exactly zero is legal (CR 119.4), so 1 life is
	// the last life total that still carries the offer.
	me.Life = 1
	if keys := offeredKeys(t, g, me, fow); !hasKey(keys, "pitch") {
		t.Fatalf("at 1 life the pitch offer is missing: %v", keys)
	}

	me.Life = 0
	if keys := offeredKeys(t, g, me, fow); hasKey(keys, "pitch") {
		t.Errorf("at 0 life the pitch offer is still shown: %v", keys)
	}
	if err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}, Targets: target,
	}); err == nil {
		t.Errorf("Force of Will pitched at 0 life was allowed")
	}
	if !me.Hand.Contains(fow) || !me.Hand.Contains(pitch) {
		t.Errorf("a refused pitch moved cards out of hand")
	}

	// Back at 1 life the view offers it and the engine takes it —
	// the same board, the same offer, the same answer.
	me.Life = 1
	if keys := offeredKeys(t, g, me, fow); !hasKey(keys, "pitch") {
		t.Fatalf("at 1 life the pitch offer is missing: %v", keys)
	}
	if err := g.CastSpell(me.ID, fow, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch}, Targets: target,
	}); err != nil {
		t.Fatalf("Force of Will pitched at exactly 1 life: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life after the pitch = %d, want 0", me.Life)
	}
}

// Snuff Out's 4 life, with the Swamp already there so the Condition
// is not what is being measured.
func TestSnuffOutOfferHiddenBelowFourLife(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	snuff := visibleHandCard(me, "Snuff Out", "Instant", "{3}{B}", snuffOutOracle, []string{"B"})
	freeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := freeSpellPermanent(g, them.ID, "Grizzly Bears", "Creature — Bear")
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}

	me.Life = 3
	if keys := offeredKeys(t, g, me, snuff); hasKey(keys, "pay_life") {
		t.Errorf("at 3 life the pay_life offer is still shown: %v", keys)
	}
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: target,
	}); err == nil {
		t.Errorf("Snuff Out for 4 life at 3 life was allowed")
	}

	me.Life = 4
	if keys := offeredKeys(t, g, me, snuff); !hasKey(keys, "pay_life") {
		t.Fatalf("at exactly 4 life the pay_life offer is missing: %v", keys)
	}
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: target,
	}); err != nil {
		t.Fatalf("Snuff Out at exactly 4 life: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life after Snuff Out = %d, want 0 — paying down to zero is legal", me.Life)
	}
}
