package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const squeeTheImmortalOracle = "481fb71a-11b9-462e-a4da-ca35ffd1bd22"

func squeeTheImmortalCard(owner uuid.UUID) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: "Squee, the Immortal", TypeLine: "Legendary Creature — Goblin",
		OracleID: squeeTheImmortalOracle, ManaCost: "{1}{R}{R}", Colors: []string{"R"},
		Power: 2, Toughness: 1, Owner: owner, Controller: owner,
	}
}

func squeeAddRed(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.ManaPool.AddMana(game.ManaToken{Color: "R"})
	}
}

func TestSqueeTheImmortalCastsFromTheGraveyardAtItsOwnCost(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	squee := squeeTheImmortalCard(me.ID)
	me.Graveyard.PushTop(squee)
	squeeAddRed(me, 2)
	if err := g.CastSpell(me.ID, squee.InstanceID, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("two mana cast a {1}{R}{R} Squee from the graveyard")
	}
	squeeAddRed(me, 1)
	if err := g.CastSpell(me.ID, squee.InstanceID, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Squee from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(squee.InstanceID) {
		t.Fatal("Squee did not resolve onto the battlefield")
	}
}

func TestSqueeTheImmortalCastsFromExile(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	squee := squeeTheImmortalCard(me.ID)
	g.Exile.PushTop(squee)
	squeeAddRed(me, 3)
	if err := g.CastSpell(me.ID, squee.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast Squee from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(squee.InstanceID) {
		t.Fatal("Squee did not resolve onto the battlefield")
	}
}

func TestSqueeTheImmortalCannotBeCastFromAnotherPlayersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	squee := squeeTheImmortalCard(other.ID)
	other.Graveyard.PushTop(squee)
	squeeAddRed(me, 3)
	if err := g.CastSpell(me.ID, squee.InstanceID, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast an opponent's Squee from their graveyard")
	}
}
