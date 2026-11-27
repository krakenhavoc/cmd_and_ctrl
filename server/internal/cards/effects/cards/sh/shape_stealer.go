package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shape Stealer — Creature — Shapeshifter Spirit {U}{U}, 1/1 (Saviors of
// Kamigawa, #5601):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this
//	 creature's base power and toughness become equal to that creature's
//	 power and toughness until end of turn."
//
// One trigger on the block event, which is one per blocker and attacker pair
// (CR 509.3b, 509.3d): when Shape Stealer blocks, "that creature" is the
// attacker it blocks; when it is blocked, each blocker triggers it. The other
// creature's power and toughness are read as the ability resolves, or as it
// last existed if it has gone (CR 608.2h). They become Shape Stealer's base
// values (layer 7b), so counters and pumps on Shape Stealer still apply, and
// the effect ends in the cleanup step (CR 514.2). A creature blocked by no
// creature (an effect made it blocked) does not trigger it.
//
// No simplification.
func init() {
	const label = "Shape Stealer — its base power and toughness become equal to that creature's until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "fa9a6617-fd89-490e-9ca5-8e6da99d89ed",
		Name:         "Shape Stealer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				_, ok := effects.OldB592PairOtherCreature(ev, source.InstanceID)
				return ok
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				other, ok := effects.OldB592PairOtherCreature(ctx.Trigger().Event, item.SourceCardID)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.SetBasePowerMod(effects.PowerNowOrLastKnown(g, other)),
						game.SetBaseToughnessMod(effects.ToughnessNowOrLastKnown(g, other)),
					},
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    label,
				}.Apply(ctx)
			}),
		},
	})
}
