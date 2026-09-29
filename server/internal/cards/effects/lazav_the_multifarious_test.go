package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lazav_the_multifarious_test.go — #1723's second seam at the card
// level: "target creature card in your graveyard with mana value X"
// is ManaValueEqualsX, bound to the ability's own announced X
// (CR 602.2b) and re-checked at resolution (CR 608.2b).

func pushLazav(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Lazav, the Multifarious",
		TypeLine:   "Legendary Creature — Shapeshifter",
		OracleID:   "c14bcef2-6d49-4430-86d0-e5a87ba442d9",
		Power:      1, Toughness: 3,
		Owner: controller, Controller: controller,
	})
}

// TestLazavTheMultifariousXBoundAnnounce is the announce-time half:
// X=3 refuses a mana value 2 or 4 target and accepts a mana value 3
// one.
func TestLazavTheMultifariousXBoundAnnounce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lazav := pushLazav(g, me.ID)
	two := b17GraveyardCard(me, "Two Drop", "Creature — Test", "{2}")
	three := b17GraveyardCard(me, "Three Drop", "Creature — Test", "{3}")
	four := b17GraveyardCard(me, "Four Drop", "Creature — Test", "{4}")

	if err := g.ActivateCatalogAbility(me.ID, lazav, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: two}},
		XValue:  3,
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("X=3 targeting mana value 2 = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, lazav, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: four}},
		XValue:  3,
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("X=3 targeting mana value 4 = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, lazav, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: three}},
		XValue:  3,
	}); err != nil {
		t.Fatalf("X=3 targeting mana value 3: %v, want no error", err)
	}
	passPriorityAroundTable(t, g)

	card := battlefieldCardFor(g, lazav)
	if card == nil {
		t.Fatal("Lazav left the battlefield")
	}
	if card.Name != "Lazav, the Multifarious" {
		t.Errorf("name = %q, want the except clause's name to survive the copy", card.Name)
	}
	if !card.HasSupertype("Legendary") {
		t.Error("the except clause's Legendary supertype did not land")
	}
	if card.TypeLine != "Legendary Creature — Test" {
		t.Errorf("TypeLine = %q, want the copied card's type line plus the except clause's Legendary", card.TypeLine)
	}
	if mv, ok := card.ParsedManaValue(); !ok || mv != 3 {
		t.Errorf("post-copy mana value = %d (ok=%v), want 3 — the copy carries the printed cost, not the card's own", mv, ok)
	}
}

// TestLazavTheMultifariousXBoundResolutionRecheck is CR 608.2b: a
// target whose mana value changes between announce and resolution is
// no longer legal, even though it was legal when the ability was
// announced with X=3.
func TestLazavTheMultifariousXBoundResolutionRecheck(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lazav := pushLazav(g, me.ID)
	three := b17GraveyardCard(me, "Three Drop", "Creature — Test", "{3}")

	if err := g.ActivateCatalogAbility(me.ID, lazav, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: three}},
		XValue:  3,
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}

	// Something changes the target's mana value while the ability
	// sits on the stack — a cost-reducing static leaving, Cloudstone
	// Curio bouncing and replaying it, doesn't matter what. The point
	// is only that it is no longer mana value 3 by the time the
	// ability resolves.
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == three {
				me.Graveyard.Cards[i].ManaCost = "{5}"
			}
		}
	})

	passPriorityAroundTable(t, g)

	card := battlefieldCardFor(g, lazav)
	if card == nil {
		t.Fatal("Lazav left the battlefield")
	}
	if card.Name != "Lazav, the Multifarious" || card.TypeLine != "Legendary Creature — Shapeshifter" {
		t.Errorf("Lazav copied a target whose mana value changed after announce: name=%q type=%q",
			card.Name, card.TypeLine)
	}
}
