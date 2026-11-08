package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "15c994b9-901a-46d6-804b-24560a14a481",
		Name:         "Seedship Broodtender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Seedship Broodtender — mill three cards", Do(MillCards{N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}{G}, Sacrifice this creature: Return target creature or Spacecraft card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{B}{G}"), SacrificeThis()),
			SorcerySpeed: true,
			Targets: TargetCardInGraveyard("target creature or Spacecraft card from your graveyard",
				YouOwn(), Or(Creature(), Subtype("Spacecraft"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				_, _ = reanimateSingleTarget(NewContext(g, item), item.Controller)
				return nil
			},
		}},
	})
}
