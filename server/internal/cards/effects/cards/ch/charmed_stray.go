package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charmed Stray — Creature — Cat {W}, 1/1 (War of the Spark):
//
//	"Lifelink
//	 When this creature enters, put a +1/+1 counter on each other creature you
//	 control named Charmed Stray."
//
// "Named" is the card name (CR 201.2, every face checked), so a creature that
// merely is a copy of one counts and a token named Charmed Stray does too. The
// source itself is excluded ("other"), and each counter goes on through the
// counter pipeline as one batch's member.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6781166b-f6ab-4076-8e97-0cd9342f1b41",
		Name:            "Charmed Stray",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Charmed Stray — put a +1/+1 counter on each other creature you control named Charmed Stray",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller || c.InstanceID == item.SourceCardID || !c.IsCreature() || !game.CardNameMatches(c, "Charmed Stray") {
							continue
						}
						if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
