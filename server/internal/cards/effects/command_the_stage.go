package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Command the Stage — Sorcery {2}{R}:
//
//	"Create a 2/2 colorless Wizard Soldier creature token named Cadet,
//	 then put a +1/+1 counter on each other Wizard token you control.
//	 At the beginning of each upkeep, if an opponent was dealt
//	 noncombat damage last turn, return this card from your graveyard
//	 to your hand."
//
// The spell is whole: the Cadet is created first, then every OTHER
// Wizard token its controller has (the new Cadet is excluded by
// instance ID) gets a counter, through AddCounter so a counter
// replacement applies.
//
// The graveyard-recursion trigger is the deferral. It needs a record of
// the previous turn's noncombat damage to opponents, and a triggered
// ability that works from the graveyard; the engine keeps neither (the
// turn tally resets as each turn begins). The card is weaker than
// printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "13bfc51f-d079-40ef-a4ba-47d06d3150a3",
		Name:         "Command the Stage",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The ability that returns this card from your graveyard to your hand isn't implemented — it stays in the graveyard once cast.",
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			me := ctx.Controller()
			made, err := ctx.Game.CreateTokensForEffect(me, TokenCard("2/2 colorless Wizard Soldier named Cadet"), 1, game.TokenEntryOptions{})
			if err != nil {
				return err
			}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != me || !IsToken(c) || !c.HasSubtype("Wizard") || slices.Contains(made, c.InstanceID) {
					continue
				}
				if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
