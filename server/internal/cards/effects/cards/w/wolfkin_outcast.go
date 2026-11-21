package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfkin Outcast // Wedding Crasher — {5}{G} Creature — Human Werewolf
// 5/4 // Creature — Werewolf 6/5 (#2586, ADR 0132):
//
//	Front: "This spell costs {2} less to cast if you control a Wolf or
//	        Werewolf.
//	        Daybound"
//	Back:  "Whenever this creature or another Wolf or Werewolf you
//	        control dies, draw a card.
//	        Nightbound"
//
// The discount is a self cost modifier (it applies to the spell in hand,
// CR 601.2a), reading the caster's battlefield. The dies trigger is
// Pashalik Mons's shape: the source's own death, or another Wolf or
// Werewolf the controller controlled as it died — judged on the
// permanent as it last existed (CR 603.10a), so a changeling counts.
//
// No simplification.
func init() {
	const oracle = "ac3d07cd-88c4-4e22-9d86-f92d34f406d4"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Wolfkin Outcast",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "This spell costs {2} less to cast if you control a Wolf or Werewolf.", controlsAWolfOrWerewolfCost),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Wedding Crasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if effects.CardDied(ev, source) {
					return true
				}
				dead, ok := effects.DiedCreature(ev, g)
				return ok && dead.InstanceID != source.InstanceID &&
					effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					// "Wolf or Werewolf" are words of the source's text (CR 612.2, #3984).
					(effects.LeftAsSubtype(ev, dead, source.TextSubtype("Wolf")) || effects.LeftAsSubtype(ev, dead, source.TextSubtype("Werewolf")))
			}, "Wedding Crasher — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

// controlsAWolfOrWerewolfCost is the cost condition "if you control a
// Wolf or Werewolf": the caster controls such a creature on the
// battlefield. Read-only, under the cast path's lock.
func controlsAWolfOrWerewolfCost(q game.CostQuery) bool {
	if q.Game == nil {
		return false
	}
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && c.IsCreature() && (c.HasSubtype("Wolf") || c.HasSubtype("Werewolf")) {
			return true
		}
	}
	return false
}
