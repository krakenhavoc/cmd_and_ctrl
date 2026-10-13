package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archmage Emeritus — Creature — Human Wizard {2}{U}{U}, 2/2 (EDHREC
// rank 255):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery
//	 spell, draw a card."
//
// The spellslinger deck's card-draw engine; Storm-Kiln Artist's
// trigger with a draw where the Treasure was. The trigger goes on
// the stack above the spell that caused it and resolves first, so
// the card is in hand before the spell resolves — as in paper.
//
// The "or copy" half rides EventSpellCopied (createSpellCopyLocked), so
// a copy of an instant or sorcery you control draws too (CR 207.2c). No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "8305d576-21d8-4ce7-8eda-a7cd9793aca5",
		Name:         "Archmage Emeritus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny(magecraftEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return instantOrSorceryCastOrCopiedByYou(ev, source, g)
			}, "Archmage Emeritus — draw a card (magecraft)", Do(DrawCards{N: 1})),
		},
	})
}
