package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Green Slime — Creature — Ooze {2}{G}, 2/2 (Commander Legends: Battle for
// Baldur's Gate):
//
//	"Flash
//	 When this creature enters, counter target activated or triggered ability
//	 from an artifact or enchantment source. If a permanent's ability is
//	 countered this way, destroy that permanent.
//	 Foretell {G}"
//
// Flash is printed and foretell is the shared special action (CR 702.143). The
// target is Echo's AbilityOnStack narrowed by the source's type (AbilityFromSource,
// read as the clause is checked). The ability's source object is named before the
// counter (CR 113.7a); once the ability is countered (CR 701.6a), the permanent it
// came from is destroyed only if that same object is still on the battlefield
// (CR 400.7). An ability that is not countered, or one whose source was not a
// permanent, destroys nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ec5a688-a884-410c-adee-ab784789d6e8",
		Name:            "Green Slime",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Green Slime — counter target activated or triggered ability from an artifact or enchantment source; if a permanent's ability is countered this way, destroy that permanent", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 {
						return nil
					}
					id := targets[0].ID
					ability := g.StackItemForEffect(id)
					if ability == nil {
						return nil
					}
					ref := ability.SourceObject
					if ref.ID == uuid.Nil {
						if src, ok := g.LookupCardForEffect(ability.SourceCardID); ok {
							ref = game.ObjectRef{ID: src.InstanceID, Epoch: src.ObjectEpoch}
						}
					}
					if err := (effects.CounterTarget{StackID: id}).Apply(ctx); err != nil {
						return err
					}
					if g.StackItemForEffect(id) != nil || ref.ID == uuid.Nil {
						return nil
					}
					if info, ok := g.PermanentForEffect(ref); ok && !info.Left {
						return effects.DestroyTarget{Target: ref.ID}.Apply(ctx)
					}
					return nil
				}),
				effects.AbilityOnStack("target activated or triggered ability from an artifact or enchantment source",
					effects.AbilityFromSource(effects.Or(effects.Artifact(), effects.Enchantment()))),
			),
		},
	})
}
