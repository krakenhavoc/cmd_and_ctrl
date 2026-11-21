package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stable Wormhole — {2}{G} Enchantment (Star Trek, #4973):
//
//	"Whenever you attack, create a tapped Planet land token named New
//	 Planet with "{T}: Add one mana of any color.""
//
// "Whenever you attack" triggers once per declaration of attackers, however
// many creatures attack and whomever they attack (CR 508.3; Adeline's
// OncePerBatch shape). The New Planet matches the printed token (ttrk 27).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dba55309-cb21-42db-8968-8da07227a1be",
		Name:         "Stable Wormhole",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Stable Wormhole — create a tapped New Planet", effects.Do(effects.TrekB22CreateTappedNewPlanet()))),
		},
	})
}
