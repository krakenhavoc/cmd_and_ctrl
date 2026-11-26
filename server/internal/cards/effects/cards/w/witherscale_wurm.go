package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witherscale Wurm — Creature — Wurm {4}{G}{G}, 9/9 (Shadowmoor, S82 OLD b530,
// #5591):
//
//	"Whenever this creature blocks or becomes blocked by a creature, that
//	 creature gains wither until end of turn. (It deals damage to creatures in
//	 the form of -1/-1 counters.)
//	 Whenever this creature deals damage to an opponent, remove all -1/-1
//	 counters from it."
//
// The first trigger is Ornery Goblin's per-pair block trigger (CR 509.3b,
// 509.3d): one block event for each blocker-attacker pair that names the Wurm,
// so it fires once for each creature it blocks or is blocked by, and the other
// creature of that pair gains wither. The second fires on any damage the Wurm
// deals to an opponent, combat or not, and takes every -1/-1 counter off the
// Wurm.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37b7f044-f57b-4d13-82be-46fd01f1dba6",
		Name:         "Witherscale Wurm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SkewerSlingerFoe(ev, source.InstanceID) != uuid.Nil
			}, "Witherscale Wurm — that creature gains wither until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					foe := effects.SkewerSlingerFoe(item.Trigger.Event, item.SourceCardID)
					if foe == uuid.Nil || !effects.OnBattlefield(g, foe) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: foe, Keywords: []string{game.KeywordWither},
						Label: "Witherscale Wurm — wither"}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToAnOpponent,
				"Witherscale Wurm — remove all -1/-1 counters from it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.OldB530RemoveAllMinusOneCounters(effects.NewContext(g, item), item.SourceCardID)
				}),
		},
	})
}
