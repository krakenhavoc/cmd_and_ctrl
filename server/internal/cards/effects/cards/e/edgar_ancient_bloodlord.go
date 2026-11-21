package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Edgar, Ancient Bloodlord — Legendary Creature — Vampire Noble {W}{B},
// 2/3:
//
//	"Whenever another creature or planeswalker you control dies, you gain
//	 1 life.
//	 {2}, Sacrifice another creature or planeswalker: Put a +1/+1 counter
//	 on Edgar. He gains menace until end of turn. (He can't be blocked
//	 except by two or more creatures.)"
//
// The trigger reads what the permanent WAS as it died (last-known types,
// CR 603.10a), so an animated planeswalker or a creature that lost its
// type still counts correctly, and it excludes Edgar himself. The
// sacrifice is part of the cost, so the dies trigger it causes goes on
// the stack above the ability and the life is gained before the counter
// lands. The counter and menace go on Edgar only while he is still on the
// battlefield; if he is gone the ability does nothing (CR 608.2b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cab2bc93-f38b-4301-ac42-05765615586b",
		Name:         "Edgar, Ancient Bloodlord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
					return false
				}
				dead, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || dead.Controller != source.Controller {
					return false
				}
				return effects.LeftAsType(ev, dead, "creature") || effects.LeftAsType(ev, dead, "planeswalker")
			},
			Key:     "Edgar, Ancient Bloodlord — you gain 1 life",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  effects.Do(effects.GainLife{Amount: 1}),
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature or planeswalker: Put a +1/+1 counter on Edgar. He gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet | game.AnswerCombatGrant},
			Cost: effects.Plus(effects.ManaCost("{2}"), game.AbilityCost{
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))),
			}),
			// The menace waits for a prompt the counter asks (a CR 616 order; CR 608.2c, #4155).
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					if effects.B09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
						return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
					}
					return nil
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.B09SourceStillOnBattlefield(ctx.Game, ctx.Item) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: ctx.Item.SourceCardID, Keywords: []string{"menace"},
						Label: "Edgar, Ancient Bloodlord — menace"}.Apply(ctx)
				}),
			),
		}},
	})
}
