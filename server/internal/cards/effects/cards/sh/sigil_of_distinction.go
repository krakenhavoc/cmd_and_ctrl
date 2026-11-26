package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sigil of Distinction — Artifact — Equipment {X} (Shards of Alara, S82 OLD
// batch 518, #5588):
//
//	"This Equipment enters with X charge counters on it.
//	 Equipped creature gets +1/+1 for each charge counter on this Equipment.
//	 Equip—Remove a charge counter from this Equipment."
//
// Astral Cornucopia's entry counters from the announced X, Banshee's Blade's
// counter-scaled pump, and an equip whose cost is the counter itself
// (Equip—Remove a charge counter, paid as the ability is activated, CR 602.2b):
// an Equipment with no counter left can't be moved, and the pump shrinks as
// the counters go. Equip is sorcery speed as ever.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "3aa3526b-bd06-417f-af29-5a9ff0844923",
		Name:                       "Sigil of Distinction",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters(game.CounterCharge)},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, func(_ *game.Game, source *game.Card) int { return source.Counters[game.CounterCharge] }),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipPayingAbility("Equip—Remove a charge counter from this Equipment.", effects.RemoveCountersFromThis(game.CounterCharge, 1)),
		},
	})
}
