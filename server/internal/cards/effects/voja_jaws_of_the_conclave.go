package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voja, Jaws of the Conclave — Legendary Creature — Wolf {2}{R}{G}{W}
// (Murders at Karlov Manor, #4469):
//
//	"Vigilance, trample, ward {3}
//	 Whenever Voja attacks, put X +1/+1 counters on each creature you
//	 control, where X is the number of Elves you control. Draw a card for
//	 each Wolf you control."
//
// Both counts are read as the ability resolves (the layered view: a
// changeling is an Elf and a Wolf, a face-down creature is neither, CR
// 708.2a). Voja is a Wolf and counts for the draw. The counters go on every
// creature you control then, one batch of placements, each its own; an X of 0
// puts none. The draw is for each Wolf, 0 drawing nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d94ff0b8-5499-4eeb-80f0-20c7fbf07499",
		Name:            "Voja, Jaws of the Conclave",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Voja, Jaws of the Conclave — ward {3}"),
			WheneverThisAttacks("Voja — put X +1/+1 counters on each creature you control, where X is the number of Elves you control; draw a card for each Wolf you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					elves, wolves := 0, 0
					var mine []game.Card
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != ctx.Controller() || !c.IsCreature() {
							continue
						}
						mine = append(mine, c)
						if c.HasSubtype(ctx.TextSubtype("Elf")) {
							elves++
						}
						if c.HasSubtype(ctx.TextSubtype("Wolf")) {
							wolves++
						}
					}
					if elves > 0 {
						for _, c := range mine {
							if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: elves}).Apply(ctx.asGroupMember()); err != nil {
								return err
							}
						}
					}
					if wolves == 0 {
						return nil
					}
					return DrawCards{N: wolves}.Apply(ctx)
				}),
		},
	})
}
