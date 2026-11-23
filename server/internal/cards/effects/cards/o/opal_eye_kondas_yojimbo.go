package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Opal-Eye, Konda's Yojimbo — Legendary Creature — Fox Samurai {1}{W}{W}, 1/4:
//
//	"Defender
//	 Bushido 1 (Whenever this creature blocks or becomes blocked, it gets
//	 +1/+1 until end of turn.)
//	 {T}: The next time a source of your choice would deal damage this
//	 turn, that damage is dealt to Opal-Eye instead.
//	 {1}{W}: Prevent the next 1 damage that would be dealt to Opal-Eye
//	 this turn."
//
// ADR 0108 §9 (#1905): a "next time" redirection from a source chosen as
// the ability resolves (CR 609.7a), whatever it would deal the damage to,
// to Opal-Eye as it is now. Bushido 1 is CR 702.45a's trigger
// (bushido.go). The last ability is the charged shield (CR 615.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4acac65-112b-48fc-bea3-44747c1389a3",
		Name:         "Opal-Eye, Konda's Yojimbo",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(1, "Opal-Eye, Konda's Yojimbo")},
		Activated: []effects.ActivatedAbility{
			effects.RedirectRow("{T}: The next time a source of your choice would deal damage this turn, that damage is dealt to Opal-Eye instead.",
				effects.TapCost(), nil, effects.RedirectDamage{Choose: true, Protect: effects.ShieldAnything, Next: true, To: effects.RedirectToThis}),
			{
				Label:   "{1}{W}: Prevent the next 1 damage that would be dealt to Opal-Eye this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    effects.ManaCost("{1}{W}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PreventNextDamage{Target: item.SourceCardID, Amount: 1, Label: "Opal-Eye — prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
