package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// gift_test.go — the enumerator half of ADR 0089 (#1267, CR 702.174):
// a gift card is several casts — the unpromised one, and one promise
// per opponent who could receive it — and every one of them is a move
// the engine accepts.

const (
	oracleDawnsTruce = "37c06f89-db36-4937-9404-2b07cd22e1a6"
	oracleWearDown   = "27905301-333e-4cdd-90cf-188159fcf8e9"
)

type giftCastWire struct {
	GiftOpponent  string `json:"gift_opponent"`
	OptionalCosts []int  `json:"optional_costs"`
	Targets       []struct {
		ID string `json:"id"`
	} `json:"targets"`
}

func giftParams(t *testing.T, m legal.Move) giftCastWire {
	t.Helper()
	var p giftCastWire
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("decode %q: %v", m.Label, err)
	}
	return p
}

// TestGiftIsOfferedOncePerOpponent: the promise names a player, so the
// bot is offered each opponent as its own move, plus the cast that
// promises nothing.
func TestGiftIsOfferedOncePerOpponent(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	truce := handCard(active, game.Card{
		Name: "Dawn's Truce", TypeLine: "Instant", OracleID: oracleDawnsTruce, ManaCost: "{1}{W}",
	})
	lands(g, active, "Plains", "Plains", 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), truce)
	unpromised := 0
	promisedTo := map[string]bool{}
	for _, m := range moves {
		p := giftParams(t, m)
		if p.GiftOpponent == "" {
			if len(p.OptionalCosts) != 0 {
				t.Errorf("%q announces an optional cost with no gift opponent", m.Label)
			}
			unpromised++
			continue
		}
		if len(p.OptionalCosts) != 1 {
			t.Errorf("%q names a gift opponent without the gift cost", m.Label)
		}
		if p.GiftOpponent == active.ID.String() {
			t.Errorf("%q promises the gift to its own caster", m.Label)
		}
		promisedTo[p.GiftOpponent] = true
	}
	if unpromised != 1 {
		t.Errorf("want exactly one unpromised cast, got %d; labels: %v", unpromised, labels(moves))
	}
	if len(promisedTo) != len(g.Seats)-1 {
		t.Errorf("want one promise per opponent (%d), got %d; labels: %v", len(g.Seats)-1, len(promisedTo), labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// TestGiftPromisedMovesUseThePromisedClause: Wear Down's promise widens
// "target artifact or enchantment" to "two target artifacts and/or
// enchantments" (CR 702.174m), so only the promised moves name two.
func TestGiftPromisedMovesUseThePromisedClause(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	wear := handCard(active, game.Card{
		Name: "Wear Down", TypeLine: "Sorcery", OracleID: oracleWearDown, ManaCost: "{1}{G}",
	})
	battlefieldCard(g, opp, game.Card{Name: "Rock", TypeLine: "Artifact"})
	battlefieldCard(g, opp, game.Card{Name: "Aura", TypeLine: "Enchantment"})
	lands(g, active, "Forest", "Forest", 3)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), wear)
	sawPromised, sawUnpromised := false, false
	for _, m := range moves {
		p := giftParams(t, m)
		if p.GiftOpponent == "" {
			sawUnpromised = true
			if len(p.Targets) != 1 {
				t.Errorf("unpromised %q names %d targets, want 1", m.Label, len(p.Targets))
			}
			continue
		}
		sawPromised = true
		if len(p.Targets) != 2 {
			t.Errorf("promised %q names %d targets, want 2", m.Label, len(p.Targets))
		}
	}
	if !sawPromised || !sawUnpromised {
		t.Fatalf("want both lines offered; labels: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
