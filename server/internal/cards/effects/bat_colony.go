package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bat Colony — Enchantment {2}{W} (The Lost Caverns of Ixalan, #4470):
//
//	"When this enchantment enters, create a 1/1 black Bat creature token with
//	 flying for each mana from a Cave spent to cast it.
//	 Whenever a Cave you control enters, put a +1/+1 counter on target
//	 creature you control."
//
// The count is the mana from a Cave source among what paid for the spell this
// permanent was, read as it enters (CR 400.7d, game.ManaSourceCave: the
// subtype the producing permanent had when the mana was made, so a Cave that
// is gone still counts). The second ability sees a Cave under your control
// entering, Bat Colony's own entry excluded (it is not a Cave); with no
// creature to target it is not put on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c2022f95-2c26-4a2a-b83b-60efa937cdde",
		Name:           "Bat Colony",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceCave,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Bat Colony — create a 1/1 black Bat creature token with flying for each mana from a Cave spent to cast it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := ctx.ManaSpentToCastThis().CountFrom(game.ManaSourceCave)
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 black Bat with flying"), N: n}.Apply(ctx)
				}),
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.HasSubtype(source.TextSubtype("Cave"))
				}, "Bat Colony — put a +1/+1 counter on target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
