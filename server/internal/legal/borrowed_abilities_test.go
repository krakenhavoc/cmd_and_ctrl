package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// borrowed_abilities_test.go — the enumerator half of ADR 0155's seam
// (#1557). A row another object LENT is the borrower's, so it is a move
// exactly when the engine would accept it (#544): every enumerated
// borrowed move dispatches, and a row whose lender has gone is not
// offered.

const (
	legalBorrowLender = "legal-fixture-lender"
	legalBorrowOoze   = "legal-fixture-ooze"
	legalBorrowPing   = "{1}: Lender pings."
)

// stubLegalBorrows layers fixture defs over the real catalog: a lender
// with "{T}: Add {G}" and "{1}: ping", and a borrower with "has all
// activated abilities of all creature cards in all graveyards".
func stubLegalBorrows(t *testing.T) {
	t.Helper()
	graveyardCreatures := func(g *game.Game, _ *game.Card) []*game.Card {
		var out []*game.Card
		for _, p := range g.Seats {
			for i := range p.Graveyard.Cards {
				if p.Graveyard.Cards[i].IsCreature() {
					out = append(out, &p.Graveyard.Cards[i])
				}
			}
		}
		return out
	}
	fixtures := map[string]*game.CardDef{
		legalBorrowLender: {
			ManaAbilities: []game.ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			Activated: []game.ActivatedAbilityShape{{
				Label:  legalBorrowPing,
				Cost:   game.AbilityCost{Mana: "{1}"},
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
		},
		legalBorrowOoze: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Borrows: &game.AbilityBorrow{Lenders: graveyardCreatures, Label: "Has all activated abilities of all creature cards in all graveyards."},
		}}},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if d, ok := fixtures[key]; ok {
			return d
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	t.Cleanup(func() { game.CatalogLookup = prev })
}

func TestBorrowedAbilitiesAreEnumeratedAndDispatch(t *testing.T) {
	stubLegalBorrows(t)
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lender := graveyardCard(active, game.Card{Name: "Lender", TypeLine: "Creature — Elf", OracleID: legalBorrowLender})
	ooze := battlefieldCard(g, active, game.Card{Name: "Ooze", TypeLine: "Creature — Ooze", OracleID: legalBorrowOoze})
	mana(g, active, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	g.BumpLayerVersionForTest()

	moves := legal.EnumerateFor(g, active.ID)
	manaMoves := grantMovesFrom(moves, ooze, legal.KindMana)
	acts := grantMovesFrom(moves, ooze, legal.KindActivate)
	if len(manaMoves) != 1 || len(acts) == 0 {
		t.Fatalf("ooze moves: mana %v, activations %v — want the lent {G} and the lent ping", labels(manaMoves), labels(acts))
	}
	if ref := moveRef(t, manaMoves[0]); ref != "borrow:"+legalBorrowLender+":0:0" {
		t.Errorf("mana move ref = %q", ref)
	}
	for _, a := range acts {
		if ref := moveRef(t, a); !strings.HasPrefix(ref, "borrow:"+legalBorrowLender+":0:") {
			t.Errorf("activation ref = %q", ref)
		}
	}
	// #544: every one of them is accepted.
	dispatchAll(t, g, active.ID, append(manaMoves, acts...))

	// The lender leaves the graveyard: nothing lent is offered.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: active.ID}, game.ZoneRef{Kind: game.ZoneExile}, lender); err != nil {
		t.Fatalf("exile the lender: %v", err)
	}
	after := legal.EnumerateFor(g, active.ID)
	if n := len(grantMovesFrom(after, ooze, legal.KindMana)) + len(grantMovesFrom(after, ooze, legal.KindActivate)); n != 0 {
		t.Errorf("a gone lender's rows are still offered: %d moves", n)
	}
}
