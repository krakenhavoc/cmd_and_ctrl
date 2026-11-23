package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// equippedCreatureYouControlAttacks is the When of "whenever an equipped
// creature you control attacks" (Swordsman, Sharp Scoundrel; Nahiri, Forged
// in Fury): one attack event per declared attacker, which must be the
// source controller's and equipped as the attack is declared.
func EquippedCreatureYouControlAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == uuid.Nil || ev.Actor != source.Controller {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.CardID)
	return ok && attacker.Controller == source.Controller && IsEquippedCreature(g, attacker)
}
