package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const darkbladeAgentGrant = "darkblade-agent/draw"

// Darkblade Agent — Creature — Human Assassin {1}{U}{B}, 2/3 (Guilds of
// Ravnica):
//
//	"As long as you've surveilled this turn, this creature has deathtouch
//	 and "Whenever this creature deals combat damage to a player, you draw a
//	 card.""
//
// Two layer-6 statics on the creature itself, both gated on the controller
// having surveilled this turn (an EventSurveil with them as the actor, the
// way Desperate Futurescribe reads it): the deathtouch keyword, and an
// ADR 0093 grant of the draw trigger. The trigger exists only while the
// condition holds, so a surveil made after the damage does not draw.
//
// No simplification.
func init() {
	onlyThis := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && effects.S82OldB185YouveSurveilledThisTurn(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:     "9a293c87-ad40-4c6c-9fa4-9a8b5f081ac9",
		Name:         "Darkblade Agent",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: darkbladeAgentGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Darkblade Agent — you draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, you draw a card.",
		}},
		Static: []game.StaticAbility{
			{
				Label:           "As long as you've surveilled this turn, this creature has deathtouch.",
				Layer:           game.Layer6Ability,
				DependsOnEvents: []game.EventKind{game.EventSurveil},
				AppliesTo:       onlyThis,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					effects.AppendKeywordsTo(c, []string{"deathtouch"})
				},
			},
			darkbladeAgentGrantStatic(onlyThis),
		},
	})
}

// darkbladeAgentGrantStatic is the grant of the draw trigger, re-evaluated
// when a surveil is made (the layer cache is told to drop on EventSurveil).
func darkbladeAgentGrantStatic(onlyThis func(*game.Card, *game.Game, *game.Card) bool) game.StaticAbility {
	ab := effects.GrantAbilities(onlyThis, darkbladeAgentGrant)
	ab.Label = "As long as you've surveilled this turn, this creature has \"Whenever this creature deals combat damage to a player, you draw a card.\""
	ab.DependsOnEvents = []game.EventKind{game.EventSurveil}
	return ab
}
