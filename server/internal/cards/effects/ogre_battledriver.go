package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ogre Battledriver — Creature — Ogre Warrior {2}{R}{R}, 3/3 (Tarkir:
// Dragonstorm Commander):
//
//	"Whenever another creature you control enters, that creature gets
//	 +2/+0 and gains haste until end of turn."
//
// The creature that entered is the trigger's object (Aunt May's read):
// both the boost and the haste go to it, and only if it is still on the
// battlefield as the ability resolves. Tokens count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c834c4b5-c122-4c63-81c7-884ca4c0dc3b",
		Name:         "Ogre Battledriver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Ogre Battledriver — that creature gets +2/+0 and gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					id := ctx.Trigger().Object.Ref().ID
					if err := (BoostUntilEOT{Target: id, Power: 2, Label: "Ogre Battledriver — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Ogre Battledriver — haste"}.Apply(ctx)
				}),
		},
	})
}
