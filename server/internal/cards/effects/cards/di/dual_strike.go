package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dual Strike — Instant {R}{R} (Kaldheim):
//
//	"When you next cast an instant or sorcery spell with mana value 4 or less
//	 this turn, copy that spell. You may choose new targets for the copy.
//	 Foretell {R} (During your turn, you may pay {2} and exile this card from
//	 your hand face down. Cast it on a later turn for its foretell cost.)"
//
// A CR 603.7 delayed trigger for the first matching spell its controller casts
// this turn (Rimefire Torque's copy), with the mana value ceiling as data on
// the cast filter (CastFilter.MaxManaValue, an {X} counted at the value
// chosen). It does not copy Dual Strike itself, whose cast is over, and ends
// with the turn if nothing qualifies. The copy is not cast, and may have new
// targets (CR 707.10, 707.10c).
//
// No simplification.
func init() {
	four := 4
	effects.Register(effects.Spec{
		OracleID:       "a498ca5f-7743-47b7-a57a-471efa1a99e8",
		Name:           "Dual Strike",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{R}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.WhenYouNextCast(
				"Dual Strike — copy that spell",
				game.CastFilter{Types: []string{"Instant", "Sorcery"}, MaxManaValue: &four},
				effects.CopyTheSpellBody,
			).Apply(ctx)
		},
	})
}
