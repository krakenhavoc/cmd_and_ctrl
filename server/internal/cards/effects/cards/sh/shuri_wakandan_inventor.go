package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shuri, Wakandan Inventor — {1}{U} Legendary Creature — Human Artificer
// Hero, 2/1 (Marvel Super Heroes, #4455):
//
//	"Artifact spells you cast cost {1} less to cast.
//	 {1}, {T}: Target artifact you control becomes a copy of a second
//	 target artifact you control until end of turn, except it isn't
//	 legendary. Activate only as a sorcery."
//
// The reduction is Foundry Inspector's. The copy is a duration BecomeCopy
// (CR 707.2, 611.2) over two clauses (CR 601.2c): slot 0 is the artifact
// that changes, slot 1 the one it copies, and the second is Distinct. Each
// slot is checked at resolution (CR 608.2b); with either gone nothing
// happens. The "except" clause removes the Legendary supertype from the
// copiable values.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8c047fc-20a9-4c62-b8a6-9ead56d60ad0",
		Name:         "Shuri, Wakandan Inventor",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Artifact spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ArtifactSpell()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Target artifact you control becomes a copy of a second target artifact you control until end of turn, except it isn't legendary. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			SorcerySpeed: true,
			Targets: effects.Clauses(
				effects.PermanentYouControl("target artifact you control (it becomes the copy)", effects.Artifact()),
				effects.Distinct(effects.PermanentYouControl("a second target artifact you control (the one it copies)", effects.Artifact())),
			),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				changing, ok := ctx.ClauseTarget(0)
				if !ok {
					return nil
				}
				original, ok := ctx.ClauseTarget(1)
				if !ok {
					return nil
				}
				return effects.BecomeCopy{
					Targets: []uuid.UUID{changing.ID},
					Of:      original.ID,
					Except:  func(v *game.PrintedValues) { v.RemoveSupertype("Legendary") },
					Label:   "Shuri, Wakandan Inventor — becomes a copy until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
