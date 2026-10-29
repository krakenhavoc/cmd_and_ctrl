package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// hobB6AttackWithAnElf is the When of "whenever you attack with one or
// more Elves" (Arboreal Alliance, Celeborn the Wise): one of the
// source's controller's creatures was declared as an attacker and it is an
// Elf. The engine emits one attack event per attacker, so the card wraps
// it in OncePerBatch. "Elf" is a word of the source's text (CR 612.2,
// #3984) and subtypes are read effective, so a changeling is an Elf.
func hobB6AttackWithAnElf(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Elf"))
}
