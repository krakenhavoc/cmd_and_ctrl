package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ice Cave — Enchantment {3}{U}{U} (Apocalypse, #5614):
//
//	"Whenever a player casts a spell, any other player may pay that spell's
//	 mana cost. If a player does, counter the spell. (Mana cost includes
//	 color.)"
//
// A cast trigger on every player's spells. The offer goes to every player but
// the caster, one prompt at a time in APNAP order (the Ice Cave's own
// controller included when someone else cast), guarding the spell while it is
// open (Dash Hopes' shape); the first to pay counters it, and if all decline
// it resolves. The price is the spell's mana cost as it sits on the stack,
// {X} read as announced; a spell with no mana cost is not offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c532ecd2-feeb-41f5-bda0-b05e86f86a45",
		Name:         "Ice Cave",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.S82OldB635IceCave("Ice Cave — any other player may pay that spell's mana cost; if one does, counter the spell"),
		},
	})
}
