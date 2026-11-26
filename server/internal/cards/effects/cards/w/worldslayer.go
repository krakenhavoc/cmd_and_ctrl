package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worldslayer — Artifact — Equipment {5} (Magic 2012, S82 OLD batch 452,
// #5573):
//
//	"Whenever equipped creature deals combat damage to a player, destroy
//	 all permanents other than this Equipment.
//	 Equip {5}"
//
// Rogue's Gloves' trigger condition (equipped creature, combat damage,
// to a player) with a sweep of every other permanent. Indestructible
// permanents survive and regeneration applies (it is "destroy", not "can't
// be regenerated).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31be715f-9ad3-435c-9e52-e108fe96b969",
		Name:         "Worldslayer",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Sweep:   game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true},
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Worldslayer — destroy all permanents other than this Equipment",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyAllMatching{Match: effects.OtherThan(item.SourceCardID)}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{5}"),
		},
	})
}
