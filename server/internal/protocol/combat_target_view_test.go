package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_target_view_test.go — the wire half of S27's polymorphic
// attack target. The client cannot offer a planeswalker as an attack
// target it cannot see, and it cannot tell a seat id from an instance
// id without being told which it is holding.

func seatWalker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Walker",
		TypeLine:   "Legendary Planeswalker — Test",
		Owner:      owner,
		Controller: owner,
		Counters:   map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func seatSiege(g *game.Game, owner, protector uuid.UUID, defense int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:        id,
		Name:              "Test Siege",
		TypeLine:          "Battle — Siege",
		Owner:             owner,
		Controller:        owner,
		ProtectorPlayerID: protector,
		Counters:          map[string]int{game.CounterDefense: defense},
	})
	return id
}

func cardInView(t *testing.T, v GameView, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range v.Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("card %s missing from the battlefield view", id)
	return CardView{}
}

func TestBattleViewCarriesItsProtectorAndDefense(t *testing.T) {
	g := buildActiveGame(t)
	owner, protector := g.Seats[0].ID, g.Seats[1].ID
	id := seatSiege(g, owner, protector, 5)

	c := cardInView(t, ViewOfGame(g), id)
	if c.ProtectorPlayer != protector.String() {
		t.Errorf("protector_player = %q, want the protector", c.ProtectorPlayer)
	}
	if c.Defense != 5 {
		t.Errorf("defense = %d, want 5", c.Defense)
	}
}

func TestNonBattleCarriesNoDefense(t *testing.T) {
	g := buildActiveGame(t)
	id := seatWalker(g, g.Seats[0].ID, 4)

	c := cardInView(t, ViewOfGame(g), id)
	if c.Defense != 0 || c.ProtectorPlayer != "" {
		t.Errorf("a planeswalker carried battle fields: defense=%d protector=%q", c.Defense, c.ProtectorPlayer)
	}
}

// TestAttackingTargetKindDistinguishesSeatsFromPermanents is the
// reason the kind is on the wire at all: attacking_target is a seat
// id or an instance id, and a client that had to guess would have to
// try one lookup and fall back to the other.
func TestAttackingTargetKindDistinguishesSeatsFromPermanents(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	defender := g.Seats[1].ID
	walker := seatWalker(g, defender, 4)

	atPlayer := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:      atPlayer,
		Name:            "Attacks a player",
		TypeLine:        "Creature — Test",
		Power:           2,
		Toughness:       2,
		Owner:           owner,
		Controller:      owner,
		AttackingTarget: defender,
	})
	atWalker := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:      atWalker,
		Name:            "Attacks a walker",
		TypeLine:        "Creature — Test",
		Power:           2,
		Toughness:       2,
		Owner:           owner,
		Controller:      owner,
		AttackingTarget: walker,
	})

	v := ViewOfGame(g)
	if got := cardInView(t, v, atPlayer).AttackingTargetKind; got != "player" {
		t.Errorf("attacking a seat: kind = %q, want \"player\"", got)
	}
	if got := cardInView(t, v, atWalker).AttackingTargetKind; got != "planeswalker" {
		t.Errorf("attacking a planeswalker: kind = %q, want \"planeswalker\"", got)
	}
}

// TestAttackerCarriesItsDefendingPlayer — #1339. The block picker
// offers a creature only the attackers its controller defends
// (CR 802.4a), and a battle's defender is its PROTECTOR, which the
// client should not have to know. The server names the seat.
func TestAttackerCarriesItsDefendingPlayer(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0].ID, g.Seats[1].ID
	walker := seatWalker(g, other, 4)
	// Controlled by the attacker's own seat, protected by the other:
	// the controller is the WRONG answer here.
	siege := seatSiege(g, owner, other, 5)
	gone := uuid.New() // a planeswalker that has left (CR 506.4c)
	attackers := map[string]uuid.UUID{}
	for name, target := range map[string]uuid.UUID{"player": other, "walker": walker, "siege": siege, "gone": gone} {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: "At " + name, TypeLine: "Creature — Test",
			Power: 2, Toughness: 2, Owner: owner, Controller: owner, AttackingTarget: target,
		})
		attackers[name] = id
	}

	v := ViewOfGame(g)
	for _, name := range []string{"player", "walker", "siege"} {
		if got := cardInView(t, v, attackers[name]).DefendingPlayer; got != other.String() {
			t.Errorf("attacking the %s: defending_player = %q, want the other seat", name, got)
		}
	}
	if got := cardInView(t, v, attackers["gone"]).DefendingPlayer; got != "" {
		t.Errorf("attacking a walker that left: defending_player = %q, want omitted", got)
	}
	if got := cardInView(t, v, walker).DefendingPlayer; got != "" {
		t.Errorf("a permanent that is not attacking carries defending_player %q", got)
	}
}

// TestTurnViewPublishesTheActivePlayersAttackTargets — present only
// during declare_attackers, because that is the only step where it
// means anything.
func TestTurnViewPublishesTheActivePlayersAttackTargets(t *testing.T) {
	g := buildActiveGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := seatWalker(g, active.ID, 4)
	theirs := seatWalker(g, other.ID, 4)

	if v := ViewOfGame(g); len(v.Turn.AttackTargets) != 0 {
		t.Fatalf("attack_targets published outside declare_attackers: %d entries", len(v.Turn.AttackTargets))
	}

	for i := 0; i < 20 && g.Turn.Step != game.StepDeclareAttackers; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatal("never reached declare_attackers")
	}

	seen := map[string]string{}
	for _, tgt := range ViewOfGame(g).Turn.AttackTargets {
		seen[tgt.ID] = tgt.Kind
	}
	if seen[theirs.String()] != "planeswalker" {
		t.Error("the opponent's planeswalker was not published as an attack target")
	}
	if _, ok := seen[mine.String()]; ok {
		t.Error("the active player's own planeswalker was published as an attack target")
	}
	if seen[other.ID.String()] != "player" {
		t.Error("an opposing seat was not published as an attack target")
	}
	if _, ok := seen[active.ID.String()]; ok {
		t.Error("the active player was published as their own attack target")
	}
}
