package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Linden, the Steadfast Queen — Legendary Creature — Human Noble {W}{W}{W},
// 3/3 (Foundations, #4465):
//
//	"Vigilance
//	 Whenever a white creature you control attacks, you gain 1 life."
//
// Sanctum Seeker's per-attacker trigger with a colour test in place of the
// creature type: one trigger for each white creature you attack with,
// Linden included. "White" is a word of the card's text (CR 612.2), and
// the attacker's colour is read through the layers as it attacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dbc56f3b-83b3-41c8-a8ae-d3774fe6ee01",
		Name:            "Linden, the Steadfast Queen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.HasColor(source.TextColor("W"))
			}, "Linden, the Steadfast Queen — a white creature you control attacked: you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
