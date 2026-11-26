package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captain of the Mists — {2}{U} Creature — Human Wizard 2/3 (Shadows over
// Innistrad via Duel Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"Whenever another Human you control enters, untap this creature.
//	 {1}{U}, {T}: You may tap or untap target permanent."
//
// The trigger is Champion of the Parish's watcher over Humans entering under
// your control; the activation is Janjeet Sentry's tap-or-untap pick, made as
// it resolves, over any permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7b6cd8b-1bed-4742-8c97-8be4edbfc8ee",
		Name:         "Captain of the Mists",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Human"))
			}, "Captain of the Mists — untap it", effects.UntapThisPermanentEffect),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: You may tap or untap target permanent.",
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.B38TapOrUntapTarget("Captain of the Mists"),
		}},
	})
}
