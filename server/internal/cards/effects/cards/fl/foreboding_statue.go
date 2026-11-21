package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foreboding Statue // Forsaken Thresher — {3} Artifact Creature — Construct
// 1/2 // Artifact Creature — Construct 5/5 (Innistrad: Crimson Vow):
//
//	Front: "{T}: Add one mana of any color. Put an omen counter on this
//	        creature.
//	        At the beginning of your end step, if there are three or more
//	        omen counters on this creature, untap it, then transform it."
//	Back:  "At the beginning of your first main phase, add one mana of any
//	        color."
//
// The first ability is a mana ability whose counter follows the mana (a
// Rider). The end-step "if" is checked as the step begins and as the
// trigger resolves (CR 603.4). The Thresher's mana comes from a trigger
// that resolves in the main phase, so it is usable that phase.
//
// No simplification.
func init() {
	const oracle = "d8c0f103-4658-4f4f-8429-c06bdd017846"
	omen3 := func(g *game.Game, source *game.Card) bool {
		c, ok := g.LookupCardForEffect(source.InstanceID)
		return ok && c.Counters["omen"] >= 3
	}
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Foreboding Statue",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Rider: func(g *game.Game, _, source uuid.UUID) error {
				_, err := g.AddCounterMustSettleNowForEffect(source, "omen", 1)
				return err
			},
			Label: "{T}: Add one mana of any color. Put an omen counter on this creature.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.S82VowB14YourEndStepIf(omen3),
				"Foreboding Statue — untap it, then transform it", func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !effects.OnBattlefield(g, item.SourceCardID) || c.Counters["omen"] < 3 {
						return nil // CR 603.4
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
						return err
					}
					return effects.TransformThis{}.Apply(ctx)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Forsaken Thresher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Forsaken Thresher — add one mana of any color",
				effects.Do(effects.AddMana{Produced: "{W|U|B|R|G}"})),
		},
	})
}
