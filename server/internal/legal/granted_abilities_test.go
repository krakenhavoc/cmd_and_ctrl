package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// granted_abilities_test.go — the enumerator half of ADR 0093's seam.
// An ability another permanent GRANTED is the host's, so it is a move
// exactly when the engine would accept it (#544): every enumerated
// granted move dispatches, a grant that has gone is not offered, and a
// move enumerated before the grant went is refused as stale rather than
// fired on whatever moved into its index.

const (
	legalFixtureRite  = "legal-fixture-rite"
	legalGrantMana    = "legal-fixture/tap-for-green"
	legalGrantPinging = "legal-fixture/ping"
)

// stubLegalGrants layers fixture defs over the real catalog: a grantor
// that gives every creature its controller controls "{T}: Add {G}" and
// "{1}: ping".
func stubLegalGrants(t *testing.T) {
	t.Helper()
	fixtures := map[string]*game.CardDef{
		legalFixtureRite: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			GrantAbilities: []string{legalGrantMana, legalGrantPinging},
		}}},
		game.GrantKey(legalGrantMana): {
			ManaAbilities: []game.ManaAbilityShape{{TapCost: true, Produced: "{G}", Label: "Add {G}"}},
			GrantText:     "{T}: Add {G}.",
		},
		game.GrantKey(legalGrantPinging): {
			Activated: []game.ActivatedAbilityShape{{
				Label:  "{1}: ping",
				Cost:   game.AbilityCost{Mana: "{1}"},
				Effect: func(*game.Game, *game.StackItem) error { return nil },
			}},
			GrantText: "{1}: This creature deals 1 damage to any target.",
		},
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

func grantMovesFrom(moves []legal.Move, source uuid.UUID, kind legal.Kind) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == kind && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

func moveRef(t *testing.T, m legal.Move) string {
	t.Helper()
	var p struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params: %v", err)
	}
	return p.Ref
}

func TestGrantedAbilitiesAreEnumeratedAndDispatch(t *testing.T) {
	stubLegalGrants(t)
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bear := battlefieldCard(g, active, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	rite := battlefieldCard(g, active, game.Card{Name: "Rite", TypeLine: "Enchantment", OracleID: legalFixtureRite})
	mana(g, active, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	g.BumpLayerVersionForTest()

	moves := legal.EnumerateFor(g, active.ID)
	manaMoves := grantMovesFrom(moves, bear, legal.KindMana)
	acts := grantMovesFrom(moves, bear, legal.KindActivate)
	if len(manaMoves) != 1 || len(acts) == 0 {
		t.Fatalf("bear moves: mana %v, activations %v — want the granted {G} and the granted ping", labels(manaMoves), labels(acts))
	}
	if ref := moveRef(t, manaMoves[0]); ref != game.GrantedAbilityRef(legalGrantMana, 0, 0) {
		t.Errorf("mana move ref = %q", ref)
	}
	for _, a := range acts {
		if ref := moveRef(t, a); ref != game.GrantedAbilityRef(legalGrantPinging, 0, 0) {
			t.Errorf("activation ref = %q", ref)
		}
	}
	// #544: every one of them is accepted.
	dispatchAll(t, g, active.ID, append(manaMoves, acts...))

	// The grantor leaves: the grant is gone, nothing from the bear is
	// offered, and the move enumerated a moment ago is refused as
	// stale — before anything is paid — rather than fired elsewhere.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: active.ID}, rite); err != nil {
		t.Fatalf("move the grantor: %v", err)
	}
	after := legal.EnumerateFor(g, active.ID)
	if n := len(grantMovesFrom(after, bear, legal.KindMana)) + len(grantMovesFrom(after, bear, legal.KindActivate)); n != 0 {
		t.Errorf("a removed grant is still offered: %d moves", n)
	}
	stale := actions.Action{Type: actions.Type(manaMoves[0].Type), Player: active.ID, Caller: active.ID, Params: manaMoves[0].Params}
	if err := actions.Dispatch(g.Clone(), stale); !errors.Is(err, game.ErrStaleAbilityRef) {
		t.Errorf("stale mana move: err = %v, want ErrStaleAbilityRef", err)
	}
	staleAct := actions.Action{Type: actions.Type(acts[0].Type), Player: active.ID, Caller: active.ID, Params: acts[0].Params}
	if err := actions.Dispatch(g.Clone(), staleAct); !errors.Is(err, game.ErrStaleAbilityRef) {
		t.Errorf("stale activation: err = %v, want ErrStaleAbilityRef", err)
	}
}

// Every mana and activation move carries a ref, own abilities included,
// so the bot sends one back for everything.
func TestEveryAbilityMoveCarriesARef(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	elves := battlefieldCard(g, active, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", OracleID: oracleLlanowarElves})
	mana(g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	moves := legal.EnumerateFor(g, active.ID)
	seen := 0
	for _, m := range moves {
		if m.Kind != legal.KindMana && m.Kind != legal.KindActivate {
			continue
		}
		seen++
		if moveRef(t, m) == "" {
			t.Errorf("move %q has no ref", m.Label)
		}
	}
	if seen == 0 {
		t.Fatal("no ability moves enumerated; the fixture is wrong")
	}
	if ms := grantMovesFrom(moves, elves, legal.KindMana); len(ms) != 1 || moveRef(t, ms[0]) != game.OwnAbilityRef(0) {
		t.Errorf("Llanowar Elves' mana move = %v", labels(ms))
	}
}
