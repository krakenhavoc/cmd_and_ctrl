package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avalanche of Sector 7 — Legendary Creature — Human Rebel {2}{R}, */3
// (#3656):
//
//	"Menace
//	 Avalanche of Sector 7's power is equal to the number of artifacts
//	 your opponents control.
//	 Whenever an opponent activates an ability of an artifact they
//	 control, Avalanche of Sector 7 deals 1 damage to that player."
//
// The power is a layer 7a characteristic-defining ability over every
// artifact an opponent controls. The trigger is Harsh Mentor's watch on
// somebody else's activation (mana abilities included, since the card
// says "an ability"), narrowed to an artifact the activator controls as
// well as the source being a battlefield artifact. It fires once per
// activation, and the damage goes to the activating player.
//
// No simplification.
func init() {
	const label = "Avalanche of Sector 7 — 1 damage to that player"
	trigger := WheneverAnOpponentActivates(label, Artifact(), true,
		func(activator uuid.UUID) Effect {
			return func(g *game.Game, item *game.StackItem) error {
				return DealDamage{Source: item.SourceCardID, Target: activator, Amount: 1}.Apply(NewContext(g, item))
			}
		})
	base := trigger.AppliesTo
	trigger.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !base(ev, source, lki, g) {
			return false
		}
		return activationSourceMatches(g, ev, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.Controller == ev.Actor
		})
	}
	Register(Spec{
		OracleID:        "4eb1414c-ba12-4bae-b410-79d51925d6fa",
		Name:            "Avalanche of Sector 7",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of artifacts your opponents control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.Battlefield.Cards {
					if p.Controller != source.Controller && p.IsArtifact() {
						n++
					}
				}
				c.Power = n
			},
		}},
		Triggered: []game.TriggeredAbility{trigger},
	})
}
