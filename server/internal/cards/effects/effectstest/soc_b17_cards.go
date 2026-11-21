package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b17_cards_test.go — Secrets of Strixhaven Commander, seventeenth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	SocB17LocustSprayOracle  = "f9dce026-2653-4793-b949-357b08a936ea"
	SocB17PrismaticOracle    = "2cb98ca9-d7bb-416b-a17e-ee5f8e4d78f2"
	SocB17RepelOracle        = "ba58f508-8036-496c-8a2f-c1e410165e93"
	SocB17SleightOracle      = "05a039ad-1689-4d13-b393-6c1219583b5d"
	SocB17SmallpoxOracle     = "7bd22c98-0782-447f-a5fa-4281585a3eab"
	SocB17StargazeOracle     = "db8b4869-f898-40ad-bbad-62390a02d1b1"
	SocB17TremorsOracle      = "f0881756-de74-43b1-a9fe-c2fde93faf5a"
	SocB17StallionOracle     = "47ee6837-2e2f-4344-ab9f-6b8887874f77"
	SocB17GreatHallOracle    = "a8c70dab-1e27-4a9c-bd2d-910d5720d02d"
	SocB17InkwrightOracle    = "d2a372d8-3eb2-4ba4-a75f-7951654de383"
	SocB17StallionTypeLine   = "Creature — Elemental Horse"
	SocB17InkwrightTypeLine  = "Creature — Human Wizard"
	SocB17PlainSorceryOracle = ""
)

// --- fixtures ----------------------------------------------------------------

// socB17Perm puts a permanent on a battlefield with a mana cost and a body.
func SocB17Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle, cost string, p, tough int) uuid.UUID {
	return S58p6Push(g, owner, name, typeLine, oracle, cost, p, tough)
}

// socB17Hand puts a card with a mana cost into a hand.
func SocB17Hand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	})
	return id
}

func SocB17Names(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

// socB17RemoveCards returns cs without the cards with the given ids.
func SocB17RemoveCards(cs []game.Card, ids ...uuid.UUID) []game.Card {
	drop := map[uuid.UUID]bool{}
	for _, id := range ids {
		drop[id] = true
	}
	out := cs[:0:0]
	for _, c := range cs {
		if !drop[c.InstanceID] {
			out = append(out, c)
		}
	}
	return out
}

// socB17CastSpellFor casts a plain instant or sorcery with the given cost,
// paid with exactly `pool`, and settles the table.
func SocB17CastSpellFor(t *testing.T, g *game.Game, typeLine, cost, pool string) {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := SocB17Hand(me, "Test Spell", typeLine, cost, SocB17PlainSorceryOracle)
	FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", cost, err)
	}
	// Two Opus triggers from one cast open an ordering prompt.
	for AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
	}
	PassPriorityAroundTable(t, g)
}
