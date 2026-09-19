package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abzan Charm — Instant {W}{B}{G} (EDHREC rank 4433):
//
//	"Choose one —
//	 • Exile target creature with power 3 or greater.
//	 • You draw two cards and you lose 2 life.
//	 • Distribute two +1/+1 counters among one or two target
//	   creatures."
//
// The Khans charm that has stayed relevant: exile (not destroy) for
// the big thing, or a Sign in Blood when there is no big thing.
// Roadmap batch 42 (#449), "no new machinery".
//
// Mode 0 EXILES, which is the mode's whole point — indestructible,
// regeneration and death triggers all miss. "Power 3 or greater" is
// the layered power, checked at announce (CR 601.2c) and again at
// resolution (CR 608.2b), so shrinking the creature in response is a
// real answer.
//
// Mode 1 draws and then loses, in that order. It matters at 2 life or
// less: you draw the cards first and the loss can kill you, which is
// the printed card.
//
// DECLARED SIMPLIFICATION (#259) on mode 2. The printed bullet
// distributes two counters among ONE OR TWO target creatures; this
// spec offers one target and puts both counters there. A single
// target is one of the printed card's own legal distributions, so the
// mode is never stronger than printed — only less flexible. The
// missing half is multi-target distribution, tracked by #764; when
// that lands, this mode becomes a second target slot and the caveat
// comes off.
func init() {
	Register(Spec{
		OracleID:     "4137a22c-f793-4e27-a9b2-72740a6e2121",
		Name:         "Abzan Charm",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The +1/+1 counter mode puts both counters on a single target creature; you cannot split them between two creatures.",
		},
		Modes: ChooseOne(
			Mode("Exile target creature with power 3 or greater.",
				TargetCreature("target creature with power 3 or greater", PowerGE(3))),
			Mode("You draw two cards and you lose 2 life."),
			Mode("Put two +1/+1 counters on target creature.",
				TargetCreature("target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
			case ctx.HasMode(1):
				if err := (DrawCards{Player: ctx.Controller(), N: 2}).Apply(ctx); err != nil {
					return err
				}
				return GainLife{Player: ctx.Controller(), Amount: -2}.Apply(ctx)
			case ctx.HasMode(2):
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return AddCounter{Target: item.Targets[0].ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}
			return nil
		},
	})
}
