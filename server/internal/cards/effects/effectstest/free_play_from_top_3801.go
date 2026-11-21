package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #3801: Ziatora's Envoy and Rashmi, Eternities Crafter. "Look at (or
// reveal) the top card of your library. You may play a land / cast a spell
// from the top of your library without paying its mana cost. If you
// don't, put that card into your hand."

const (
	ziatorasEnvoyOracle = "93fe1d5b-3ab3-431a-9d78-175550ba9d43"
	RashmiOracle        = "01a79429-7d57-4a51-9815-a20905ed4707"
)

// envoyTable has the Envoy ready to connect for 5 and a chosen card on top
// of the controller's library. attack() runs combat to the damage step and
// settles the trigger, leaving the table at the prompt (if there is one).
type envoyTable struct {
	G       *game.Game
	Me, Opp *game.Player
	Envoy   uuid.UUID
}

func NewEnvoyTable(t *testing.T) *envoyTable {
	t.Helper()
	g := NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	AdvanceToMain(t, g)
	envoy := PushDiesCreatureForTest(g, me.ID, "Ziatora's Envoy", ziatorasEnvoyOracle,
		"Creature — Lizard Warrior", 5, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return &envoyTable{G: g, Me: me, Opp: opp, Envoy: envoy}
}

// top puts a card on top of the controller's library.
func (tb *envoyTable) Top(name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	tb.G.WithWriteLock(func() {
		tb.Me.Library.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: tb.Me.ID, Controller: tb.Me.ID,
		})
	})
	return id
}

func (tb *envoyTable) Attack(t *testing.T) {
	t.Helper()
	AttackWith(t, tb.G, tb.Opp.ID, tb.Envoy)
	PassPriorityAroundTable(t, tb.G)
}

// mayCast is the offer waiting on the controller: the land's may_cast
// (play_or_hand), or the spell's resolution_cast (free_or_hand, cast
// during the trigger's resolution, #4250).
func (tb *envoyTable) MayCast() *game.PendingChoice {
	if c := LatestChoiceOfKindFor(tb.G, game.PendingChoiceMayCast, tb.Me.ID); c != nil {
		return c
	}
	return KeywordCastPromptFor(tb.G, tb.Me.ID)
}

// answer answers the land's may_cast either way, or declines the spell's
// resolution_cast (its cast needs targets, so the tests cast it).
func (tb *envoyTable) Answer(t *testing.T, apply bool) {
	t.Helper()
	c := tb.MayCast()
	if c == nil {
		t.Fatal("no offer is waiting")
	}
	if c.Kind == game.PendingChoiceResolutionCast {
		if apply {
			t.Fatal("cast the spell with its targets; answer only declines it")
		}
		if err := tb.G.ResolvePendingChoice(c.ID, tb.Me.ID, nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
		return
	}
	if err := tb.G.ResolveMayCast(c.ID, tb.Me.ID, apply); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
}

func (tb *envoyTable) InLibraryTop(id uuid.UUID) bool {
	n := tb.Me.Library.Size()
	return n > 0 && tb.Me.Library.Cards[n-1].InstanceID == id
}
