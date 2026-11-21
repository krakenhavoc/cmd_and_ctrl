package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primal Boost — Instant {2}{G} (Dominaria Remastered, #5948):
//
//	"Target creature gets +4/+4 until end of turn.
//	 Cycling {2}{G}
//	 When you cycle this card, you may have target creature get +1/+1 until
//	 end of turn."
//
// The cycling trigger is an optional targeted trigger, chosen as it goes on
// the stack (CR 603.3d) and answered as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "902e7d4b-7ab4-40b5-b3a5-c83ef604989a",
		Name:         "Primal Boost",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 4, 4)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return BoostUntilEOT{Target: item.Targets[0].ID, Power: 4, Toughness: 4, Label: "Primal Boost — +4/+4"}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					InGraveyard(On(game.EventCycle, Self, "Primal Boost — cycled: you may have target creature get +1/+1 until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							id, ok := b16FirstLegalTargetCard(NewContext(g, item))
							if !ok {
								return nil
							}
							return BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Primal Boost — +1/+1"}.Apply(NewContext(g, item))
						})),
					"Primal Boost — have target creature get +1/+1 until end of turn?"),
				TargetCreature("target creature"),
			),
		},
	})
}
