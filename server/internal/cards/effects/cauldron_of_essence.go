package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cauldron of Essence — Artifact {1}{B}{G} (EDHREC rank 2345):
//
//	"Whenever a creature you control dies, each opponent loses 1 life
//	 and you gain 1 life.
//	 {1}{B}{G}, {T}, Sacrifice a creature: Return target creature
//	 card from your graveyard to the battlefield. Activate only as a
//	 sorcery."
//
// A Blood Artist that recurs. The drain is Zulaport Cutthroat's body
// on a "creature you control dies" condition read post-move (the
// Cauldron is not a creature, so there is no self case). The
// reanimation is a CR 602 activated ability at sorcery speed whose
// sacrifice is paid at announce, so the sacrificed creature's own
// drain trigger sits ABOVE the ability and resolves first. The
// engine validates the target before it pays the cost (CR 601.2c
// before 601.2h), so the creature being sacrificed is not yet in the
// graveyard when the target is picked and cannot be its own
// reanimation target — exactly as in paper. The card returns under
// its owner's control; "from your graveyard" means the controller
// owns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a2f8cde8-bf7b-4234-89f0-a95f9dc937e3",
		Name:         "Cauldron of Essence",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b22CreatureYouControlDied(ev, source, g)
			}, "Cauldron of Essence — each opponent loses 1 life and you gain 1 life", func(g *game.Game, item *game.StackItem) error {
				return b07DrainEachOpponent(g, item)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{B}{G}, {T}, Sacrifice a creature: Return target creature card from your graveyard to the battlefield.",
			Cost:         Plus(ManaCost("{1}{B}{G}"), TapCost(), SacrificeACreature()),
			Targets:      TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
