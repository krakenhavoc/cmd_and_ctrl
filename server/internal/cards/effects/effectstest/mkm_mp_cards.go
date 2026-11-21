package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_mp_cards_test.go — Murders at Karlov Manor cards M–P (#4469). Each
// test plays the real catalog entry and asserts what its text says.

const (
	MkmMPGlassOracle       = "f2ce9e55-07c0-4f36-a59e-127693fc0f62"
	MkmMPBindingOracle     = "942f3f06-1df0-40b4-8537-71cbec111852"
	MkmMPPhantomOracle     = "6b0d4192-bc76-4ef1-92b2-b5fca81970bd"
	MkmMPGuardianOracle    = "ef414347-bbe3-47cb-b900-88275561c43e"
	MkmMPWatchOracle       = "2d400c01-d2d0-442a-bc3d-cf106a754dab"
	MkmMPInspectorOracle   = "c206bdcb-d2c4-4bda-a979-6c0627ef8f73"
	MkmMPOnTheJobOracle    = "e4733889-4d4c-4bf7-8a9b-fb5ad7a38395"
	MkmMPNightwatchOracle  = "3891a914-e597-4126-846f-39bf8534eb92"
	MkmMPMoroiiOracle      = "07d830a7-f0a2-4d46-9a53-101dca681603"
	MkmMPOffenderOracle    = "c8c42851-1b1b-49dd-a97c-7eda57dadb13"
	MkmMPGardenerOracle    = "15dca6a7-8367-4b86-a86c-e44f10fb3c53"
	MkmMPPerformerOracle   = "78f9ec0f-6015-4024-aca6-9a14cba373fa"
	MkmMPPerformerTypeLine = "Creature — Lizard Assassin"
)

// --- shared helpers ------------------------------------------------------------

// mkmMPCastFaceDown casts a catalog card from the active player's hand
// face down for its disguise and resolves it.
func MkmMPCastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "disguise"}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	if c := FindCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

// mkmMPTurnFaceUp pays the disguise cost from conjured mana and turns
// the permanent face up.
func MkmMPTurnFaceUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	g.WithWriteLock(func() { p.ManaPool.AddMana(ManaTokens(mana...)...) })
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	if FindCardAnywhere(t, g, id).FaceDown {
		t.Fatal("still face down after the special action")
	}
}

// mkmMPPT is a battlefield permanent's power and toughness after the
// layers and its counters.
func MkmMPPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := TmtB1Card(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// mkmMPOpponents is every seat but the active one.
func MkmMPOpponents(g *game.Game) []*game.Player {
	var out []*game.Player
	for i, p := range g.Seats {
		if i != g.Turn.ActiveSeat {
			out = append(out, p)
		}
	}
	return out
}

// mkmMPLives reads each player's life total.
func MkmMPLives(ps []*game.Player) []int {
	out := make([]int, len(ps))
	for i, p := range ps {
		out[i] = p.Life
	}
	return out
}
