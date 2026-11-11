package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smoldering Stagecoach — Artifact — Vehicle {3}{R}, */5 (Outlaws of
// Thunder Junction Commander):
//
//	"Smoldering Stagecoach's power is equal to the number of instant and
//	 sorcery cards in your graveyard.
//	 Whenever this Vehicle attacks, the next instant spell and the next
//	 sorcery spell you cast this turn each have cascade.
//	 Crew 2"
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a), read
// live as Enigma Drake's is. The attack trigger sets up two CR 603.7b
// "when you next cast" watches, one narrowed to instants and one to
// sorceries, each firing once on its own spell: that spell cascades through
// the trigger the watch puts on the stack above it (cmdr2S41CascadeTheSpell),
// so a turn's instant and sorcery each cascade once, and a spell that is
// both (there is none) would use both. They end at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83f87884-f0f9-482e-887e-35a8a5c20eb0",
		Name:         "Smoldering Stagecoach",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Smoldering Stagecoach's power is equal to the number of instant and sorcery cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = fandanielInstantsAndSorceries(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Smoldering Stagecoach — the next instant spell and the next sorcery spell you cast this turn each have cascade",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range []string{"Instant", "Sorcery"} {
						if err := WhenYouNextCast(
							"Smoldering Stagecoach — cascade: exile until a nonland card with lesser mana value",
							game.CastFilter{Types: []string{t}},
							cmdr2S41CascadeTheSpellBody,
						).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Smoldering Stagecoach"),
		}},
	})
}
