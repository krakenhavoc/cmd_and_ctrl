package cardsch

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Circle of Flame — Enchantment {1}{R} (Magic 2015, #5556):
//
//	"Whenever a creature without flying attacks you or a planeswalker you
//	 control, this enchantment deals 1 damage to that creature."
//
// Blood Reckoning's attack trigger (one per attacking creature, CR 508.3a,
// aimed at the controller or a planeswalker they control) narrowed to
// attackers without flying as they attack. The Circle deals the damage, as
// it last existed if it has left (CR 608.2h); an attacker that has left
// combat's battlefield by resolution takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c31801e6-2444-4158-8e19-2afe751d7cff",
		Name:         "Circle of Flame",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, circleOfFlameWatches,
				"Circle of Flame — 1 damage to the attacking creature without flying",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					attacker := ctx.Trigger().Event.CardID
					if !effects.OnBattlefield(g, attacker) {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: attacker, Amount: 1}.Apply(ctx)
				}),
		},
	})
}

// circleOfFlameWatches is "a creature without flying attacks you or a
// planeswalker you control".
func circleOfFlameWatches(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil || !effects.AttacksYouOrYourPlaneswalker(g, ev.Target, source.Controller) {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.CardID)
	return ok && attacker.IsCreature() && !game.HasKeyword(&attacker, "flying")
}
