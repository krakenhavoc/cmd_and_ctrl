package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siege of Towers — Sorcery {1}{R} (Ravnica Remastered, #5948):
//
//	"Replicate {1}{R} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.)
//	 Target Mountain becomes a 3/1 creature. It's still a land."
//
// Replicate is Gigadrowse's pair (replicateMana, Replicate(), CR 702.56a).
// The animation has no duration (CR 611.2a, 613): a scoped effect pinned to
// the Mountain object adds the Creature type (it stays a land) and sets base
// power and toughness 3/1 (layers 4 and 7b), lasting while that object is on
// the battlefield. A Mountain that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "d8eb0897-0d2f-4f2d-a09b-87d57d5a0fa9",
		Name:          "Siege of Towers",
		Completeness:  CompletenessFull,
		Targets:       TargetPermanent("target Mountain", Subtype("Mountain")),
		OptionalCosts: []game.AdditionalCost{replicateMana("{1}{R}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, id := range legalTargetCards(item, ctx.Game) {
				mods := append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(3, 1)...)
				if err := (ScopedEffectFor{
					Target:   id,
					Mods:     mods,
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Siege of Towers — becomes a 3/1 creature",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
