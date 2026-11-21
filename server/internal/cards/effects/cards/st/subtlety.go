package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Subtlety — Creature — Elemental Incarnation {2}{U}{U}, 3/3 (Special
// Guests):
//
//	"Flash
//	 Flying
//	 When this creature enters, choose up to one target creature spell or
//	 planeswalker spell. Its owner puts it on their choice of the top or
//	 bottom of their library.
//	 Evoke—Exile a blue card from your hand."
//
// Sudden Setback's placement for a spell: the choice is the OWNER's, so the
// top-or-bottom prompt goes to the owner, and the spell leaves the stack for
// the library WITHOUT being countered (CR 701.6a), so one that can't be
// countered still moves. The target is chosen as the trigger goes on the
// stack (up to one, so none is allowed) and a spell that has gone by
// resolution does nothing (CR 608.2b). Evoke is the shared exile-a-card
// alternative cost (Endurance): the creature enters, its trigger goes on the
// stack, and the sacrifice trigger follows (CR 702.74a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "377179d5-ac83-4d34-b5b1-f3d8caa60f79",
		Name:            "Subtlety",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		AlternativeCosts: []game.AlternativeCost{
			effects.EvokePitch(effects.CardInYourHand("a blue card from your hand", effects.OfColor("U")), "a blue card from your hand"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Subtlety — its owner puts the target spell on their choice of the top or bottom of their library",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						var id uuid.UUID
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								id = t.ID
								break
							}
						}
						c, ok := g.LookupCardForEffect(id)
						if !ok || c.Owner == uuid.Nil {
							return nil
						}
						return effects.PutInLibraryInAnyOrder{
							Chooser:   c.Owner,
							Cards:     []uuid.UUID{id},
							From:      game.ZoneStack,
							Placement: game.LibraryPlaceTopOrBottom,
							Label:     "Subtlety — put it on your choice of the top or bottom of your library",
						}.Apply(ctx)
					}),
				effects.TargetSpell("up to one target creature spell or planeswalker spell", effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1)),
		},
	})
}
