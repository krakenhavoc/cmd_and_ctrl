package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Souls of the Faultless — Creature — Spirit {W}{B}{B}, 0/4 (Guildpact):
//
//	"Defender
//	 Whenever this creature is dealt combat damage, you gain that much life
//	 and attacking player loses that much life."
//
// Brash Taunter's trigger waiting for the damage to finish, so the TOTAL of
// the combat damage it was dealt is "that much" (CR 120.3, CR 510.2). It
// triggers on combat damage only; damage from a spell does not. The attacking
// player is the player whose creatures attacked, the active player (CR 506.2),
// who loses the life. The Souls' controller gains the life whether or not it
// survived the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5a12e989-d6bc-441e-88dd-779e68a9fee3",
		Name:            "Souls of the Faultless",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Combat && effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Souls of the Faultless — you gain that much life and attacking player loses that much life",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
					return nil
				}
				n := item.Trigger.Event.Amount
				ctx := effects.NewContext(g, item)
				if err := (effects.GainLife{Amount: n}).Apply(ctx); err != nil {
					return err
				}
				attacking := g.Seats[g.Turn.ActiveSeat]
				if attacking == nil || attacking.Eliminated {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, attacking.ID, -n)
			},
		}},
	})
}
