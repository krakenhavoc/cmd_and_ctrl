package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razorkin Hordecaller — Creature — Human Clown Berserker {4}{R}, 4/4
// (Duskmourn: House of Horror, #4466):
//
//	"Haste
//	 Whenever you attack, create a 1/1 red Gremlin creature token."
//
// "Whenever you attack" is Fire Navy Trebuchet's shape: one trigger per
// attack declaration (OncePerBatch over EventAttack), however many
// creatures attack and whether or not the Hordecaller is one of them.
// The Gremlin is Midnight Mayhem's token. It enters untapped and is not
// attacking: the card never says it enters attacking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3fc4bdf1-7da3-4f9f-806f-f82cd18ea65f",
		Name:            "Razorkin Hordecaller",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Razorkin Hordecaller — create a 1/1 red Gremlin creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Gremlin"), N: 1}))),
		},
	})
}
