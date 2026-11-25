package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Answered Prayers — Enchantment {1}{W}{W} (Modern Horizons):
//
//	"Whenever a creature you control enters, you gain 1 life. If this
//	 enchantment isn't a creature, it becomes a 3/3 Angel creature with flying
//	 in addition to its other types until end of turn."
//
// One printed trigger, one registered. The life comes first, then the
// animation: the enchantment is read as the ability resolves (CR 608.2c), and it
// becomes a creature only when it is not one then and is still the object the
// ability came from (CR 400.7). The effect is a layer 4 type addition, a layer
// 4 subtype addition (Angel) and a layer 7b base power and toughness of 3/3,
// with flying a layer 6 grant, all until end of turn; the enchantment keeps its
// other types. A creature it becomes can't trigger it by entering, since it was
// already on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b238254-51b4-4e68-a8c6-93ab1180012e",
		Name:         "Answered Prayers",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Answered Prayers — gain 1 life; if this isn't a creature, it becomes a 3/3 Angel creature with flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					if me, ok := g.LookupCardForEffect(item.SourceCardID); !ok || me.IsCreature() {
						return nil
					}
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Angel")...),
							game.AddKeywordsMod("flying"),
						}, game.SetBasePTMods(3, 3)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Answered Prayers — a 3/3 Angel creature with flying until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
