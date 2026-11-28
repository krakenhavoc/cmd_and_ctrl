package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wellspring — Enchantment — Aura {1}{G}{W} (Mirage, S82 OLD batch 811, #5633):
//
//	"Enchant land
//	 When this Aura enters, gain control of enchanted land until end of turn.
//	 At the beginning of your upkeep, untap enchanted land. You gain control of
//	 that land until end of turn."
//
// Two printed triggers with the same two-step body: the Aura's controller
// gains control of the land it is attached to until end of turn (CR 611.2b),
// and at their upkeep also untaps it first. A land the Aura has left, or that
// is no longer on the battlefield, is not touched (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8650bd97-ecf6-420f-9b72-a749970afcd4",
		Name:         "Wellspring",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wellspring — gain control of enchanted land until end of turn", func(g *game.Game, item *game.StackItem) error {
				return wellspringTake(effects.NewContext(g, item), false)
			}),
			effects.AtYourUpkeep("Wellspring — untap enchanted land and gain control of it until end of turn", func(g *game.Game, item *game.StackItem) error {
				return wellspringTake(effects.NewContext(g, item), true)
			}),
		},
	})
}

// wellspringTake gains control of the enchanted land until end of turn,
// untapping it first when `untap`.
func wellspringTake(ctx *effects.Context, untap bool) error {
	land := effects.EnchantedCreatureOf(ctx)
	if land == uuid.Nil || !effects.OnBattlefield(ctx.Game, land) {
		return nil
	}
	if untap {
		if err := (effects.UntapTarget{Target: land}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.GainControl{
		Target:   land,
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Wellspring — gain control until end of turn",
	}.Apply(ctx)
}
