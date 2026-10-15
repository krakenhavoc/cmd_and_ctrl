package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const danithaParagonOracle = "4b6377da-83e7-4519-9582-16a9c16b8faa"

// Danitha Capashen, Paragon has no equip clause: her Oracle text is
// "First strike, vigilance, lifelink / Aura and Equipment spells you
// cast cost {1} less to cast." (Scryfall, 2026-10-14). Equip abilities
// cost their printed {3} with her out, so {C}{C} is not enough.
// Guards against #3666's reverted equip discount coming back.
func TestDanithaParagonDoesNotDiscountEquip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Danitha Capashen, Paragon",
		"Legendary Creature — Human Knight", danithaParagonOracle, false)
	bear := seedBear(g, me.ID)
	hammer := seedEquipment(g, me.ID, "Loxodon Warhammer", warhammerOracle)

	err := strictEquip(t, g, me.ID, hammer, bear, "{C}{C}")
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("equip {3} with {2} and Danitha out: err = %v, want insufficient mana", err)
	}
}
