package legal_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// legal_actions_agreement_test.go is the server half of ADR 0105's
// contract fixture (§9, #1789), built the way agreement_test.go builds
// the timing one.
//
// Each scenario declares by hand which cards are READY — have a legal
// move in the viewer's own `legal_actions` digest — and what the
// digest must say about them: which kinds of move, which ability rows,
// which zones and faces, whom they may attack or block. This file
// asserts the digest in the real filtered frame matches that
// declaration. The generated testdata/legal_actions_agreement.json
// carries the same frames to the client, whose legalActions.ts lookups
// (ADR 0105 sub-PR 2) assert against the same declarations. Neither
// side is the other's oracle.
//
// Regenerate with:
//
//	go test ./internal/legal/ -run TestLegalActionsAgreement -update

const (
	legalActionsFixturePath = "testdata/legal_actions_agreement.json"
	oracleViviOrnitier      = "0452be2a-e97a-4269-a9e3-b616265ecb2e"
)

// readyExpectation is one card and what the viewer's digest must say
// about it. With Ready false the card must have no entry at all, and
// the lists are empty. With Ready true every list is compared as a
// set, and an omitted list means the entry's must be empty too.
type readyExpectation struct {
	InstanceID    string   `json:"instance_id"`
	Name          string   `json:"name"`
	Ready         bool     `json:"ready"`
	Kinds         []string `json:"kinds,omitempty"`
	Abilities     []string `json:"abilities,omitempty"`
	ManaAbilities []string `json:"mana_abilities,omitempty"`
	Zones         []string `json:"zones,omitempty"`
	Faces         []int    `json:"faces,omitempty"`
	AttackTargets []string `json:"attack_targets,omitempty"`
	Blocks        []string `json:"blocks,omitempty"`
	Why           string   `json:"why"`
}

// legalActionsScenario is one board, one viewer, the frame that viewer
// receives, and the declared digest. Pass declares `legal_actions.pass`.
type legalActionsScenario struct {
	Name   string             `json:"name"`
	Note   string             `json:"note"`
	Viewer string             `json:"viewer"`
	View   protocol.GameView  `json:"view"`
	Pass   bool               `json:"pass"`
	Expect []readyExpectation `json:"expect"`
}

func buildLegalActionsScenarios(t *testing.T) []legalActionsScenario {
	t.Helper()
	var out []legalActionsScenario

	// 1. The active seat's main phase, with one of everything the
	//    digest can mark there: a land to play, an instant to cast, a
	//    permanent with an activated ability, a basic land's mana
	//    ability and a nonland, no-{T} mana source (#1621's Vivi).
	{
		g, active, _ := duel(t)
		mountain := battlefieldCard(g, active, basic("Mountain", "Mountain"))
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
		bombardment := battlefieldCard(g, active, game.Card{
			Name: "Goblin Bombardment", TypeLine: "Enchantment",
			ManaCost: "{1}{R}", OracleID: oracleGoblinBombardment,
		})
		bear := battlefieldCard(g, active, creature("Sacrificial Bear", "{1}{G}", 2, 2))
		vivi := battlefieldCard(g, active, game.Card{
			Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard",
			ManaCost: "{1}{U}{R}", Power: 1, Toughness: 3, OracleID: oracleViviOrnitier,
		})
		forest := handCard(active, basic("Forest", "Forest"))
		bolts := handCard(active, bolt())
		dear := handCard(active, sorcery("Concentrate", "{2}{U}{U}"))
		advanceTo(t, g, game.StepPrecombatMain)
		out = append(out, legalActionsScenario{
			Name: "main_phase_ready",
			Note: "active seat, precombat main, empty stack, land drop unused",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Pass: true,
			Expect: []readyExpectation{
				{InstanceID: forest.String(), Name: "Forest", Ready: true,
					Kinds: []string{"land"}, Zones: []string{"hand"},
					Why: "CR 305: main phase, empty stack, your turn, land drop unused"},
				{InstanceID: bolts.String(), Name: "Lightning Bolt", Ready: true,
					Kinds: []string{"cast"}, Zones: []string{"hand"}, Faces: []int{0},
					Why: "an instant, and {R} is payable off a Mountain"},
				{InstanceID: dear.String(), Name: "Concentrate", Ready: false,
					Why: "{2}{U}{U} is four mana; the board makes three"},
				{InstanceID: bombardment.String(), Name: "Goblin Bombardment", Ready: true,
					Kinds: []string{"activate"}, Abilities: []string{"own:0"},
					Why: "a creature to sacrifice is on the board"},
				{InstanceID: bear.String(), Name: "Sacrificial Bear", Ready: false,
					Why: "no ability, and no attack outside combat; being sacrificed is Bombardment's move"},
				{InstanceID: mountain.String(), Name: "Mountain", Ready: true,
					Kinds: []string{"mana"}, ManaAbilities: []string{"land:R"},
					Why: "an untapped land's mana ability is legal; drawing no pip for it is the client's rule (ADR 0105 §4)"},
				{InstanceID: vivi.String(), Name: "Vivi Ornitier", Ready: true,
					Kinds: []string{"mana"}, ManaAbilities: []string{"own:0"},
					Why: "{0}, your turn, not yet used this turn (#1621)"},
			},
		})
	}

	// 2. No decision owed: the other seat during the active seat's main
	//    phase. There is no digest, so nothing is ready.
	{
		g, _, other := duel(t)
		battlefieldCard(g, other, basic("Mountain", "Mountain"))
		bolts := handCard(other, bolt())
		advanceTo(t, g, game.StepPrecombatMain)
		out = append(out, legalActionsScenario{
			Name: "no_decision_owed",
			Note: "the non-active seat during the active seat's main phase: legal_actions is absent",
			View: protocol.ViewOfGameFor(g, other.ID.String()), Viewer: other.ID.String(),
			Expect: []readyExpectation{
				{InstanceID: bolts.String(), Name: "Lightning Bolt", Ready: false,
					Why: "CR 117.1: you cannot cast without priority"},
			},
		})
	}

	// 3. Declare attackers: the creature that may attack names whom.
	{
		g, active, other := duel(t)
		giant := battlefieldCard(g, active, creature("Hill Giant", "{3}{R}", 3, 3))
		advanceTo(t, g, game.StepDeclareAttackers)
		out = append(out, legalActionsScenario{
			Name: "declare_attackers",
			Note: "active seat in declare attackers with one untapped creature",
			View: protocol.ViewOfGameFor(g, active.ID.String()), Viewer: active.ID.String(),
			Pass: true,
			Expect: []readyExpectation{
				{InstanceID: giant.String(), Name: "Hill Giant", Ready: true,
					Kinds: []string{"attack"}, AttackTargets: []string{other.ID.String()},
					Why: "CR 508.1a: an untapped creature that has been under your control since the turn began"},
			},
		})
	}

	// 4. Declare blockers: each creature that may block names the
	//    attacker it may block.
	{
		g, active, other := duel(t)
		giant := battlefieldCard(g, active, creature("Hill Giant", "{3}{R}", 3, 3))
		wall := battlefieldCard(g, other, creature("Wall of Wood", "{G}", 0, 3))
		cub := battlefieldCard(g, other, creature("Bear Cub", "{1}{G}", 2, 2))
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(giant, other.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)
		out = append(out, legalActionsScenario{
			Name: "declare_blockers",
			Note: "the defending seat, one attacker pointed at it, two untapped creatures; it owes a declaration but holds no priority, so there is no pass move (#328)",
			View: protocol.ViewOfGameFor(g, other.ID.String()), Viewer: other.ID.String(),
			Pass: false,
			Expect: []readyExpectation{
				{InstanceID: wall.String(), Name: "Wall of Wood", Ready: true,
					Kinds: []string{"block"}, Blocks: []string{giant.String()},
					Why: "CR 509.1a: an untapped creature may block an attacker attacking its controller"},
				{InstanceID: cub.String(), Name: "Bear Cub", Ready: true,
					Kinds: []string{"block"}, Blocks: []string{giant.String()},
					Why: "the same, for the second creature"},
			},
		})
	}
	return out
}

// TestLegalActionsAgreement asserts the digest in each scenario's frame
// matches its declaration, and writes the fixture the client reads.
func TestLegalActionsAgreement(t *testing.T) {
	scenarios := buildLegalActionsScenarios(t)
	for _, sc := range scenarios {
		t.Run(sc.Name, func(t *testing.T) {
			d := sc.View.LegalActions
			if pass := d != nil && d.Pass; pass != sc.Pass {
				t.Errorf("legal_actions.pass = %v, scenario declares %v", pass, sc.Pass)
			}
			for _, e := range sc.Expect {
				var got *protocol.LegalSourceView
				if d != nil {
					got = d.Sources[e.InstanceID]
				}
				if (got != nil) != e.Ready {
					t.Errorf("%s: ready = %v, scenario declares %v (%s)", e.Name, got != nil, e.Ready, e.Why)
					continue
				}
				if got == nil {
					continue
				}
				kinds := make([]string, 0, len(got.Kinds))
				for _, k := range got.Kinds {
					kinds = append(kinds, string(k))
				}
				sameSet(t, e.Name, "kinds", kinds, e.Kinds)
				sameSet(t, e.Name, "abilities", got.Abilities, e.Abilities)
				sameSet(t, e.Name, "mana_abilities", got.ManaAbilities, e.ManaAbilities)
				sameSet(t, e.Name, "zones", got.Zones, e.Zones)
				sameSet(t, e.Name, "attack_targets", got.AttackTargets, e.AttackTargets)
				sameSet(t, e.Name, "blocks", got.Blocks, e.Blocks)
				gotFaces, wantFaces := slices.Clone(got.Faces), slices.Clone(e.Faces)
				slices.Sort(gotFaces)
				slices.Sort(wantFaces)
				if !slices.Equal(gotFaces, wantFaces) {
					t.Errorf("%s: faces = %v, scenario declares %v", e.Name, got.Faces, e.Faces)
				}
			}
		})
	}
	writeLegalActionsFixture(t, scenarios)
}

func sameSet(t *testing.T, card, field string, got, want []string) {
	t.Helper()
	g, w := slices.Clone(got), slices.Clone(want)
	sort.Strings(g)
	sort.Strings(w)
	if !slices.Equal(g, w) {
		t.Errorf("%s: %s = %v, scenario declares %v", card, field, got, want)
	}
}

// writeLegalActionsFixture is writeFixture for this file's scenarios:
// compare against the committed bytes, or rewrite them under -update.
func writeLegalActionsFixture(t *testing.T, scenarios []legalActionsScenario) {
	t.Helper()
	body, err := json.MarshalIndent(scenarios, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	body = canonicaliseIDs(body)
	body = append(body, '\n')

	if *updateFixture {
		if err := os.MkdirAll(filepath.Dir(legalActionsFixturePath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(legalActionsFixturePath, body, 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		t.Logf("wrote %s (%d scenarios, %d B)", legalActionsFixturePath, len(scenarios), len(body))
		return
	}
	have, err := os.ReadFile(legalActionsFixturePath)
	if err != nil {
		t.Fatalf("read %s (regenerate with `go test ./internal/legal/ -run TestLegalActionsAgreement -update`): %v", legalActionsFixturePath, err)
	}
	if string(have) != string(body) {
		t.Errorf("%s is stale — the frames the client's legal-actions test reads no longer match what the server produces.\n"+
			"Regenerate with: go test ./internal/legal/ -run TestLegalActionsAgreement -update", legalActionsFixturePath)
	}
}

// TestLegalActionsFixtureIsSeatSafe re-checks, on the committed bytes,
// that each frame's digest names only the viewer's own cards and that
// no frame carries another seat's move list.
func TestLegalActionsFixtureIsSeatSafe(t *testing.T) {
	body, err := os.ReadFile(legalActionsFixturePath)
	if err != nil {
		t.Skipf("no fixture yet: %v", err)
	}
	var scenarios []legalActionsScenario
	if err := json.Unmarshal(body, &scenarios); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(scenarios) == 0 {
		t.Fatal("fixture is empty")
	}
	for _, sc := range scenarios {
		owner := map[string]string{}
		zones := []protocol.ZoneView{sc.View.Battlefield, sc.View.Stack, sc.View.Exile}
		for _, s := range sc.View.Seats {
			zones = append(zones, s.Hand, s.Library, s.Graveyard, s.Command)
		}
		for _, z := range zones {
			for _, c := range z.Cards {
				owner[c.InstanceID] = c.Controller
			}
		}
		for _, m := range sc.View.LegalMoves {
			if m.Player.String() != sc.Viewer {
				t.Errorf("%s: frame for %s carries a move belonging to %s", sc.Name, sc.Viewer, m.Player)
			}
		}
		if sc.View.LegalActions == nil {
			continue
		}
		for id := range sc.View.LegalActions.Sources {
			if owner[id] != sc.Viewer {
				t.Errorf("%s: the digest for %s names %s, controlled by %q", sc.Name, sc.Viewer, id, owner[id])
			}
		}
	}
}
