package cardsbl

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood on the Snow — Snow Sorcery {4}{B}{B} (Kaldheim):
//
//	"Choose one —
//	 • Destroy all creatures.
//	 • Destroy all planeswalkers.
//	 Then return a creature or planeswalker card with mana value X or less
//	 from your graveyard to the battlefield, where X is the amount of {S}
//	 spent to cast this spell. ({S} is mana from a snow source.)"
//
// Either mode destroys everything it names at once (CR 701.8), then the
// controller names one creature or planeswalker card of their graveyard whose
// mana value is at most the snow mana spent (CR 107.4h) — including a card the
// wipe just put there — and it returns to the battlefield. No target: the card
// is chosen as the spell resolves. With no snow mana, only a mana value 0 card
// is eligible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "45f386ba-ad62-4c0e-b6ad-ae02ca8f2bec",
		Name:           "Blood on the Snow",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsPoolManaOnly,
		WantsManaFrom:  game.ManaSourceSnow,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Destroy all creatures.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.DestroyAllMatching{Match: effects.Creature()}).Apply(ctx); err != nil {
						return err
					}
					return bloodOnTheSnowReturn(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
			effects.ModeDoing("Destroy all planeswalkers.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.DestroyAllMatching{Match: effects.Planeswalker()}).Apply(ctx); err != nil {
						return err
					}
					return bloodOnTheSnowReturn(ctx)
				}),
		),
	})
}

// bloodOnTheSnowReturn is "Then return a creature or planeswalker card with
// mana value X or less from your graveyard to the battlefield, where X is the
// amount of {S} spent to cast this spell".
func bloodOnTheSnowReturn(ctx *effects.Context) error {
	x := ctx.ManaSpent().CountFrom(game.ManaSourceSnow)
	return effects.ReturnChosenFromGraveyard{
		Question: "Blood on the Snow — return a creature or planeswalker card with mana value " + strconv.Itoa(x) + " or less from your graveyard to the battlefield",
		Match: func(c game.Card) bool {
			return (c.IsCreature() || c.IsPlaneswalker()) && c.ManaValue() <= x
		},
		Min: 1,
		Max: 1,
	}.Apply(ctx)
}
