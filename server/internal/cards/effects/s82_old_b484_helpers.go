package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b484_helpers.go — shared pieces of Rise of the Eldrazi batch 484
// (S82, #5581).

// OldB484EnchantedAttacksOrBlocks is the condition of "whenever enchanted
// creature attacks or blocks" (Sinister Possession, Luminous Wake), for an
// ability that watches game.EventAttack and game.EventBlock: an attack event
// names the enchanted creature, and a block event names the blocker once per
// block declaration (Amount <= 1, so a creature that blocks two attackers
// triggers once, CR 509.3a). `source` is the Aura.
func OldB484EnchantedAttacksOrBlocks(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.Kind == game.EventBlock && ev.Amount > 1 {
		return false
	}
	return source.IsAttachedTo(ev.CardID)
}
