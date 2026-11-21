package effectstest

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// land_play_restriction_cards_test.go — ADR 0109 §4 (#1895), the card
// half: each printed shape refused at the land play (CR 101.2) with its
// printed clause, the card left in hand, and the same answer from the
// gate the enumerator and the view ask. The three-way agreement is
// asserted in internal/legal and internal/protocol against the same cards.

const (
	LpAggressiveMiningOracle   = "ff321c13-03ae-4cb9-b371-1242290f8433"
	LpCityInABottleOracle      = "a83f25e3-4d84-4c9b-ab12-19b8d326e459"
	LpExperimentalFrenzyOracle = "3715e8b8-31df-499b-8a91-1bcd1199d6eb"
	LpMoonholdOracle           = "9b515bb8-6e37-4b2d-8d56-bb517c3b267c"
	LpPardicMinerOracle        = "e1788b54-4cd9-459f-b2c9-06773e68e9e3"
	LpRockJockeyOracle         = "09fff6a3-6a54-4101-a587-55d7113b9639"
	LpSolfataraOracle          = "6053a192-fcf9-4b06-9f46-e84eaadaa882"
	LpTerritorialDisputeOracle = "a1785817-f17b-471b-a63b-866e7972df1f"
	LpTurfWoundOracle          = "8dc48180-1482-40c4-8c5f-1724018e9c5b"
	LpWardOfBonesOracle        = "c3ea1497-63ac-46bb-95e9-d98b93a880b3"
)

// lpHandLand puts a land card in p's hand.
func LpHandLand(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	typeLine := "Basic Land — Forest"
	if name != "Forest" {
		typeLine = "Land"
	}
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// lpPlay plays a land out of hand through the engine's own path.
func LpPlay(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) error {
	t.Helper()
	AdvanceToMain(t, g)
	return g.CastSpell(p.ID, id, game.CastSpellParams{})
}

// lpGate asks the land-play gate for a land card in `zone`.
func LpGate(g *game.Game, p *game.Player, name string, zone game.ZoneKind) error {
	var err error
	g.WithWriteLock(func() {
		err = g.LandPlayGateLocked(p.ID, game.Card{Name: name, TypeLine: "Land"}, zone)
	})
	return err
}

// lpAssertRefused pins the error SHAPE: the sentinel, the printed clause,
// and the card still in hand with no land drop spent.
func LpAssertRefused(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the land play was allowed, want a refusal mentioning %q", want)
	}
	if !errors.Is(err, game.ErrCantPlayLand) {
		t.Fatalf("refusal is %v, want it to wrap game.ErrCantPlayLand", err)
	}
	var cant *game.CantPlayLandError
	if !errors.As(err, &cant) || !strings.Contains(cant.Reason, want) {
		t.Fatalf("refusal = %v, want a *game.CantPlayLandError mentioning %q", err, want)
	}
	if !p.Hand.Contains(id) {
		t.Error("a refused land play did not leave the card in hand")
	}
	if g.LandsPlayedThisTurnFor(p.ID) != 0 {
		t.Error("a refused land play spent a land drop")
	}
}
