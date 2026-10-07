package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// energy_cards_test.go — ADR 0129 PR 1 cards outside the four batch
// files: Chthonian Nightmare, whose "Pay X {E}" meets #2028's
// return-this cost.

const chthonianNightmare = "c7a360f5-725e-4864-93cc-87f96f95975e"

func TestChthonianNightmarePaysXEnergyAndReanimatesManaValueX(t *testing.T) {
	g, me, _ := p7Table(t)
	nightmare := apaPush(g, me.ID, me.ID, game.Card{Name: "Chthonian Nightmare", OracleID: chthonianNightmare, TypeLine: "Enchantment"})
	fodder := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	two := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{B}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	target := []game.TargetRef{{Kind: game.TargetCard, ID: two}}

	setEnergy(t, g, me, 1)
	err := g.ActivateCatalogAbility(me.ID, nightmare, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: target})
	if !errors.Is(err, game.ErrInsufficientEnergy) {
		t.Fatalf("X=2 with one energy: err = %v, want ErrInsufficientEnergy", err)
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(nightmare) {
		t.Fatal("a refused activation paid its sacrifice or its return")
	}

	setEnergy(t, g, me, 3)
	p7Activate(t, g, me, nightmare, 0, game.ActivateAbilityParams{
		XValue: 2, SacrificeIDs: []uuid.UUID{fodder}, Targets: target})
	if energyOf(me) != 1 {
		t.Errorf("energy = %d, want 1 after paying X=2", energyOf(me))
	}
	if !g.Battlefield.Contains(two) {
		t.Error("the mana value 2 creature card was not returned to the battlefield")
	}
	if !me.Hand.Contains(nightmare) {
		t.Error("Chthonian Nightmare is not in its owner's hand")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

// The target must have mana value exactly X (CR 601.2c via 602.2b).
func TestChthonianNightmareRefusesAnotherManaValue(t *testing.T) {
	g, me, _ := p7Table(t)
	nightmare := apaPush(g, me.ID, me.ID, game.Card{Name: "Chthonian Nightmare", OracleID: chthonianNightmare, TypeLine: "Enchantment"})
	fodder := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	two := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: two, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{B}",
			Owner: me.ID, Controller: me.ID})
	})
	setEnergy(t, g, me, 5)
	err := g.ActivateCatalogAbility(me.ID, nightmare, 0, game.ActivateAbilityParams{
		XValue: 3, SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: two}}})
	if err == nil {
		t.Fatal("X=3 targeting a mana value 2 card was accepted")
	}
	if energyOf(me) != 5 {
		t.Errorf("energy = %d after a refusal, want 5", energyOf(me))
	}
}

// ADR 0129 §7: the enumerator tries every X the seat's energy pays for
// and offers only the X a creature card in the graveyard matches, with
// MoveCost.Energy naming it; the move is one the engine accepts (#544).
func TestChthonianNightmareIsEnumeratedAtTheMatchingX(t *testing.T) {
	g, me, _ := p7Table(t)
	nightmare := apaPush(g, me.ID, me.ID, game.Card{Name: "Chthonian Nightmare", OracleID: chthonianNightmare, TypeLine: "Enchantment"})
	apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{B}",
			Owner: me.ID, Controller: me.ID})
	})
	setEnergy(t, g, me, 3)
	var moves []legal.Move
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == "activate_ability" && m.Source == nightmare {
			moves = append(moves, m)
		}
	}
	if len(moves) == 0 {
		t.Fatal("no Chthonian Nightmare activation offered")
	}
	for _, m := range moves {
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if p.XValue != 2 || m.Cost == nil || m.Cost.Energy != 2 {
			t.Errorf("move %q: x %d cost %+v, want X=2 paying 2 energy", m.Label, p.XValue, m.Cost)
		}
		clone := g.Clone()
		if err := actions.Dispatch(clone, actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: me.ID, Params: m.Params}); err != nil {
			t.Errorf("move %q rejected: %v", m.Label, err)
		}
	}
}

func TestChthonianNightmareEntersWithThreeEnergy(t *testing.T) {
	spec, ok := Lookup(chthonianNightmare)
	if !ok {
		t.Fatal("Chthonian Nightmare is not registered")
	}
	if spec.Purpose.Energy != 3 {
		t.Errorf("purpose energy = %d, want 3", spec.Purpose.Energy)
	}
	if c := spec.Activated[0].Cost; !c.EnergyX || !c.ReturnSelf || c.SacrificeOther == nil {
		t.Errorf("cost = %+v, want pay X energy, a sacrifice and return this", c)
	}
}
