package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const (
	GraspingShadowsOracleID = "522a4b02-24c7-45d2-9097-2803cc9fffad"
	DreadCounter            = "dread"
)

// Grasping Shadows — Enchantment {3}{B}, the front of a transforming card
// whose back is Shadows' Lair (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever a creature you control attacks alone, it gains deathtouch and
//	 lifelink until end of turn. Put a dread counter on this enchantment.
//	 Then if there are three or more dread counters on it, transform it."
//
// The three sentences are one trigger (CR 608.2c, in order): the lone
// attacker, a creature of yours, gains deathtouch and lifelink until the
// end of the turn, a dread counter goes on the enchantment, and then, with
// the counters counted after that one, it transforms at three or more. The
// attacker is read from the event; if it has left the battlefield by
// resolution it gains nothing, and the counter and the transform still
// follow.
//
// Shadows' Lair:
//
//	"{T}: Add {B}.
//	 {B}, {T}, Remove a dread counter from this land: You draw a card and
//	 you lose 1 life."
//
// A transformed permanent keeps its counters (CR 712.18), so the Lair spends
// the dread counters the enchantment gathered. Removing one is a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     GraspingShadowsOracleID,
		Name:         "Grasping Shadows",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "Grasping Shadows — the lone attacker gains deathtouch and lifelink until end of turn; put a dread counter on it; transform it at three",
				effects.Do(effects.Instruction(graspingShadowsGrant),
					effects.CounterOnThis{Kind: DreadCounter, N: 1},
					effects.Instruction(func(ctx *effects.Context) error {
						c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
						if !ok || !effects.OnBattlefield(ctx.Game, ctx.Source()) || c.Counters[DreadCounter] < 3 {
							return nil
						}
						return effects.TransformThis{}.Apply(ctx)
					}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     GraspingShadowsOracleID + "#1",
		Name:         "Shadows' Lair",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "{T}: Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}, Remove a dread counter from this land: You draw a card and you lose 1 life.",
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost(), effects.RemoveCountersFromThis(DreadCounter, 1)),
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Effect:  effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1}),
		}},
	})
}

// graspingShadowsGrant is "it gains deathtouch and lifelink until end of
// turn": the creature that attacked alone.
func graspingShadowsGrant(ctx *effects.Context) error {
	if ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	attacker := ctx.Item.Trigger.Event.CardID
	if !effects.OnBattlefield(ctx.Game, attacker) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"deathtouch", "lifelink"}, Label: "Grasping Shadows — deathtouch and lifelink"}.Apply(ctx)
}
