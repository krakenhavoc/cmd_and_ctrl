package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rite of the Dragoncaller — Enchantment {4}{R}{R} (EDHREC rank
// 2916):
//
//	"Whenever you cast an instant or sorcery spell, create a 5/5 red
//	 Dragon creature token with flying."
//
// A Dragon per spell. One trigger on EventCast, the controller's
// instants and sorceries (instantOrSorceryCastByYou reads the
// spell off the stack), making the 5/5 red flying Dragon
// (b14RedDragonToken at five — Dragonmaster Outcast's). The trigger
// goes on the stack above the spell and resolves first, so the
// Dragon is there before the spell is, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5097f4e6-50af-4641-909f-db44abf0ce32",
		Name:         "Rite of the Dragoncaller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Rite of the Dragoncaller — create a 5/5 red Dragon with flying", effects.Do(effects.CreateToken{Template: effects.B14RedDragonToken(5), N: 1})),
		},
	})
}
