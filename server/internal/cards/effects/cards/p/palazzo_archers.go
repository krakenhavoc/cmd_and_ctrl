package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Palazzo Archers — Creature — Human Archer {2}{G}, 2/3 (Assassin's Creed,
// #5948):
//
//	"Reach
//	 Whenever a creature with flying attacks you or a planeswalker you
//	 control, this creature deals damage equal to its power to that
//	 creature."
//
// Barbed Foliage's "attacks you" read, widened to a planeswalker you control
// (an EventAttack names the attacked player or planeswalker in Target; a
// battle does not count). Flying is read as the creature attacks (CR 603.2).
// The damage is the Archers' power as the trigger resolves, or as it last
// existed if it has left (CR 608.2h), dealt by the Archers as an object; an
// attacker that has left the battlefield is dealt none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4887b9e1-ea84-4543-b272-c106a13a33dc",
		Name:            "Palazzo Archers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, palazzoArchersFlierAttacksYouOrAWalker,
				"Palazzo Archers — damage equal to its power to that creature", palazzoArchersDamage),
		},
	})
}

func palazzoArchersFlierAttacksYouOrAWalker(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil || ev.Target == uuid.Nil {
		return false
	}
	attacksYou := ev.Target == source.Controller
	if !attacksYou {
		walker, ok := g.LookupCardForEffect(ev.Target)
		if !ok || !effects.OnBattlefield(g, ev.Target) || !walker.IsPlaneswalker() || walker.Controller != source.Controller {
			return false
		}
	}
	attacker, ok := g.LookupCardForEffect(ev.CardID)
	return ok && game.HasKeyword(&attacker, "flying")
}

func palazzoArchersDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	self, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	return effects.DealDamage{SourceObject: &ref, Target: ctx.Trigger().Event.CardID, Amount: self.Power}.Apply(ctx)
}
