package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seedship Broodtender — Creature — Insect Citizen {B}{G}, 2/3 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, mill three cards.
//	 {3}{B}{G}, Sacrifice this creature: Return target creature or
//	 Spacecraft card from your graveyard to the battlefield. Activate only
//	 as a sorcery."
//
// The mill is its controller's own (CR 701.17). The activation is Doomed
// Necromancer's sacrifice-this reanimation over Stray from History's
// "creature or Spacecraft card", at sorcery speed. The target is chosen
// before the cost is paid (CR 602.2b, 601.2c), so the Broodtender
// cannot return itself. A returned Spacecraft enters with no charge
// counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15c994b9-901a-46d6-804b-24560a14a481",
		Name:         "Seedship Broodtender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Seedship Broodtender — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{B}{G}, Sacrifice this creature: Return target creature or Spacecraft card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}{B}{G}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			Targets: effects.TargetCardInGraveyard("target creature or Spacecraft card from your graveyard",
				effects.YouOwn(), effects.Or(effects.Creature(), effects.Subtype("Spacecraft"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, _ = effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
				return nil
			},
		}},
	})
}
