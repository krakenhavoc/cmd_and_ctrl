package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guidelight Matrix — Artifact {2}:
//
//	"When this artifact enters, draw a card.
//	 {2}, {T}: Target Mount you control becomes saddled until end of
//	 turn. Activate only as a sorcery.
//	 {2}, {T}: Target Vehicle you control becomes an artifact creature
//	 until end of turn."
//
// The saddle ability is left out: the engine has no saddled state
// (CR 702.171) and no Mount in the catalog could use one, so the
// ability would be a button that does nothing. The Vehicle animation is
// BecomeCreatureUntilEOT on the chosen Vehicle.
func init() {
	Register(Spec{
		OracleID:     "98005890-d566-4a27-906f-625513171e85",
		Name:         "Guidelight Matrix",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Saddle isn't implemented — the ability that saddles a Mount is not offered."},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Key:       "Guidelight Matrix — draw a card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Target Vehicle you control becomes an artifact creature until end of turn.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetPermanent("target Vehicle you control", OfSubtype("Vehicle"), YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return BecomeCreatureUntilEOT{Target: t.ID, Label: "Guidelight Matrix — becomes an artifact creature"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
