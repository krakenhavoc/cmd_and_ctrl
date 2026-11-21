package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guidelight Matrix — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {2}, {T}: Target Mount you control becomes saddled until end of
//	 turn. Activate only as a sorcery.
//	 {2}, {T}: Target Vehicle you control becomes an artifact creature
//	 until end of turn."
//
// The saddle ability is BecomeSaddled on the chosen Mount (CR 702.171,
// ADR 0071 amendment 2026-10-08, #2695): the designation with no
// saddlers, so "creatures that saddled it" stays empty. It is sorcery
// speed, as printed. The Vehicle animation is BecomeCreatureUntilEOT on
// the chosen Vehicle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98005890-d566-4a27-906f-625513171e85",
		Name:         "Guidelight Matrix",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Key:       "Guidelight Matrix — draw a card",
			Effect:    effects.Do(effects.DrawCards{N: 1}),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}, {T}: Target Mount you control becomes saddled until end of turn. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets:      effects.TargetPermanent("target Mount you control", effects.OfSubtype("Mount"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BecomeSaddled{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}, {
			Label:   "{2}, {T}: Target Vehicle you control becomes an artifact creature until end of turn.",
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Vehicle you control", effects.OfSubtype("Vehicle"), effects.YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BecomeCreatureUntilEOT{Target: t.ID, Label: "Guidelight Matrix — becomes an artifact creature"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
