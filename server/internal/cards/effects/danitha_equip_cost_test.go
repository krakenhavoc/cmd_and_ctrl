package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const danithaParagonOracle = "4b6377da-83e7-4519-9582-16a9c16b8faa"

// #3666: Danitha's third clause, "Equip abilities you activate cost {1}
// less to activate". Loxodon Warhammer's equip is {3}, so it costs {2}.
func TestDanithaParagonDiscountsYourEquipByOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Danitha Capashen, Paragon",
		"Legendary Creature — Human Knight", danithaParagonOracle, false)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	if err := strictEquip(t, g, me.ID, hammer, bear, "{C}{C}{C}"); err != nil {
		t.Fatalf("equip with Danitha out: %v", err)
	}
	if got := poolSize(g, me.ID); got != 1 {
		t.Errorf("pool after equip = %d, want 1 (equip {3} - 1 = {2})", got)
	}
}

func TestDanithaParagonDoesNotDiscountWithoutHer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	err := strictEquip(t, g, me.ID, hammer, bear, "{C}{C}")
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip {3} with {2}, no Danitha: err = %v, want insufficient mana", err)
	}
}

func TestDanithaParagonDoesNotDiscountAnOpponentsEquip(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, opp.ID, "Danitha Capashen, Paragon",
		"Legendary Creature — Human Knight", danithaParagonOracle, false)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	err := strictEquip(t, g, me.ID, hammer, bear, "{C}{C}")
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("opponent's Danitha discounted my equip: err = %v", err)
	}
}
