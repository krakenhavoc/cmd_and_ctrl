package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fugitive Droid — Artifact Creature — Robot Scientist {U}, 1/1
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"This creature can't be blocked if an artifact entered the
//	 battlefield under your control this turn.
//	 {U}, Sacrifice this creature: Counter target spell that targets an
//	 artifact or creature you control."
//
// The evasion is a block rule whose condition is Iron Man, Master of
// Machines' tally (Game.EnteredWithCardTypeThisTurn), read when the block
// is declared: an artifact that entered under your control this turn,
// this creature included, and whether or not it is still on the
// battlefield. The counter is Rebuff the Wicked's, with the targeted
// permanent narrowed to an artifact or creature you control on the
// battlefield (tmtB8SpellTargetsArtifactOrCreatureYouControl), checked as
// the ability is activated and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c04d4fd4-e5aa-431d-824b-8bc94245f103",
		Name:         "Fugitive Droid",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{CantBeBlockedWhile(OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return g.EnteredWithCardTypeThisTurn(c.Controller, "artifact") > 0
		})},
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice this creature: Counter target spell that targets an artifact or creature you control.",
			Cost:    Plus(ManaCost("{U}"), SacrificeThis()),
			Targets: TargetSpell("target spell that targets an artifact or creature you control", tmtB8SpellTargetsArtifactOrCreatureYouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
