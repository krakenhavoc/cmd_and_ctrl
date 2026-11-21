package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inner Demon — Enchantment — Aura {2}{B}{B} (Battlebond, #5529):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2, has flying, and is a Demon in addition
//	 to its other types.
//	 When this Aura enters, all non-Demon creatures get -2/-2 until end of
//	 turn."
//
// The three statics are Angelic Destiny's: a layer 7c pump, a layer-6
// flying grant and a layer-4 Demon added to the creature types (CR 205.1b),
// "Demon" a word of the Aura's text (CR 612.2).
//
// The enters trigger is Noxious Ghoul's shrink at -2/-2, yours included:
// every creature that is not a Demon as it resolves, snapshotted then (CR
// 611.2c), so the enchanted creature, a Demon by then, is spared, and a
// creature that enters later in the turn is not shrunk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2cee861f-d28c-47e6-baee-2da9b7fa89d8",
		Name:         "Inner Demon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
			effects.AttachedIsAlso("Demon"),
		},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WhenThisEnters("Inner Demon — all non-Demon creatures get -2/-2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					demon := ctx.TextSubtype("Demon")
					return effects.BoostUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsCreature() && !c.HasSubtype(demon)
						},
						Power:     -2,
						Toughness: -2,
						Label:     "Inner Demon — non-Demons get -2/-2",
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2, Partial: true}}),
		},
	})
}
