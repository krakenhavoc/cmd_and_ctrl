package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const coramOracle = "51c511c0-783a-456d-a7e2-959f6e8ddfc8"

func TestCoramTheUndertakerIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(coramOracle)
	if !ok || spec.Name != "Coram, the Undertaker" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// coramTable puts Coram (no summoning sickness) on the active player's
// battlefield in their precombat main phase.
func coramTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	coram := effectstest.B12Push(g, me.ID, "Coram, the Undertaker", "Legendary Creature — Human Warrior", coramOracle, 0, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			g.Battlefield.Cards[i].SummonedThisTurn = false
		}
	})
	return g, me, coram
}

// coramMill puts `cards` on top of `p`'s library and mills them.
func coramMill(t *testing.T, g *game.Game, p *game.Player, cards ...game.Card) []uuid.UUID {
	t.Helper()
	ids := make([]uuid.UUID, len(cards))
	for i := range cards {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		ids[i] = cards[i].InstanceID
		p.Library.PushTop(cards[i])
	}
	g.WithWriteLock(func() { g.MillNForEffect(p.ID, len(cards)) })
	for _, id := range ids {
		if !p.Graveyard.Contains(id) {
			t.Fatal("the milled cards did not reach the graveyard")
		}
	}
	return ids
}

// Coram gets +X/+0 for the greatest power among creature cards in ALL
// graveyards, an opponent's included.
func TestCoramGetsTheGreatestPowerAmongCreatureCardsInAllGraveyards(t *testing.T) {
	g, me, coram := coramTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if got := effectstest.EffectivePower(t, g, coram); got != 0 {
		t.Fatalf("power with empty graveyards = %d, want 0", got)
	}
	coramMill(t, g, me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2})
	coramMill(t, g, opp, game.Card{Name: "Big Wurm", TypeLine: "Creature — Wurm", ManaCost: "{5}{G}", Power: 6, Toughness: 6},
		game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{9}"})
	if got := effectstest.EffectivePower(t, g, coram); got != 6 {
		t.Fatalf("power = %d, want 6 (the opponent's Wurm)", got)
	}
}

// Whenever Coram attacks, each player mills a card.
func TestCoramAttackMakesEachPlayerMillACard(t *testing.T) {
	g, me, coram := coramTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Graveyard.Size()
	}
	effectstest.DeclareAttack(t, g, opp.ID, coram)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Graveyard.Size() - before[i]; got != 1 {
			t.Errorf("seat %d milled %d, want 1", i, got)
		}
	}
	_ = me
}

// On its controller's turn, Coram lets them play one land and cast one
// spell from among cards put into ANY graveyard from a library this turn.
// A card discarded into a graveyard does not qualify; a second spell and a
// second land are refused.
func TestCoramPlaysALandAndCastsASpellFromGraveyardsMilledThisTurn(t *testing.T) {
	g, me, _ := coramTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	milled := coramMill(t, g, opp,
		game.Card{Name: "Their Sorcery", TypeLine: "Sorcery", ManaCost: "{U}"},
		game.Card{Name: "Their Second Sorcery", TypeLine: "Sorcery", ManaCost: "{U}"},
		game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Their Island", TypeLine: "Basic Land — Island"},
	)
	discarded := effectstest.GfDead(t, g, opp, "Their Discarded Sorcery", "Sorcery", "{U}")
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, discarded, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast a card that was not put into the graveyard from a library")
	}
	if !effectstest.EnumeratedFor(g, me.ID, milled[0]) {
		t.Fatal("the bot is not offered the opponent's milled sorcery")
	}
	if err := g.CastSpell(me.ID, milled[0], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the opponent's milled sorcery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, milled[1], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a second spell was cast this turn")
	}
	if err := g.CastSpell(me.ID, milled[2], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("play the opponent's milled land after the spell: %v", err)
	}
	if !effectstest.OnBattlefieldByID(g, milled[2]) {
		t.Fatal("the land is not on the battlefield")
	}
	if err := g.CastSpell(me.ID, milled[3], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a second land was played this turn")
	}
}

// Not on another player's turn: their milled instant is refused.
func TestCoramDoesNothingOnAnotherPlayersTurn(t *testing.T) {
	g, me, _ := coramTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	milled := coramMill(t, g, opp, game.Card{Name: "Their Instant", TypeLine: "Instant", ManaCost: "{U}"})
	effectstest.AdvanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	coramMill(t, g, opp, game.Card{Name: "Their New Instant", TypeLine: "Instant", ManaCost: "{U}"})
	effectstest.FloatMana(t, g, me, "{U}")
	if err := g.CastSpell(me.ID, milled[0], game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast from the graveyards off Coram's controller's turn")
	}
}
