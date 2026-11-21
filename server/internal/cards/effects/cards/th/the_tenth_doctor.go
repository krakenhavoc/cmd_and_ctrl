package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Tenth Doctor — Legendary Creature — Time Lord Doctor {3}{U}{R}, 3/5 (Doctor Who):
//
//	"Allons-y! — Whenever you attack, exile cards from the top of your library
//	 until you exile a nonland card. Put three time counters on it. If it doesn't
//	 have suspend, it gains suspend.
//	 Timey-Wimey — {7}: Time travel three times. Activate only as a sorcery. (For
//	 each suspended card you own and each permanent you control with a time counter
//	 on it, you may add or remove a time counter. Then do it two more times.)"
//
// The attack trigger is Kang Prime's exile run with three counters, once per declaration
// of attackers (CR 603.2c: "whenever you attack"), the lands exiled on the way staying in
// exile. Time travel is the engine's (CR 701.56a, TimeTravel), repeated three times: each
// walk lists the permanents and suspended cards as it begins, so the second and third
// walks see the counters the first one changed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "23af0a0a-1d12-47c7-b191-2bd3f84eea93",
		Name:         "The Tenth Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Allons-y! — exile cards from the top of your library until you exile a nonland card; put three time counters on it; it gains suspend",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    effects.UntilNonland,
						ThenTail: ctx.TailN(effects.GapB88SuspendedWithTimeCounters, 3),
					}.Apply(ctx)
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "Timey-Wimey — {7}: Time travel three times. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{7}"),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.TimeTravel{Times: 3}),
		}},
	})
}
