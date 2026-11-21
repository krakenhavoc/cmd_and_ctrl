package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greatsword of Tyr — Artifact — Equipment {1}{W} (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Whenever equipped creature attacks, put a +1/+1 counter on it and tap up
//	 to one target creature defending player controls.
//	 Equip {W}"
//
// One trigger on the equipped creature's attack declaration, with a target
// chosen as it goes on the stack: up to one creature controlled by the player
// that creature attacks (CR 506.2a), none being fine. The counter goes on the
// attacker named by the event, so it is put even with no target; a target
// that has become illegal is not tapped (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Greatsword of Tyr — put a +1/+1 counter on the creature and tap up to one target creature defending player controls"
	trigger := effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.AttachedCreatureAttacked(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		if id := item.Trigger.Event.CardID; effects.OnBattlefield(g, id) {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.Msc13TapTheLegalTargets(g, item)
	})
	trigger.TargetsFrom = effects.UpToOne(func(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
		equipped, ok := g.LookupCardForEffect(tc.Event.CardID)
		if !ok {
			return nil
		}
		defendersCreature := effects.TargetOfDefendingPlayer("creature", "target creature defending player controls", effects.Creature())
		return defendersCreature(tc, &equipped, g)
	})
	effects.Register(effects.Spec{
		OracleID:     "81a602b6-a4f9-4534-a84e-2fa7556c2309",
		Name:         "Greatsword of Tyr",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{W}")},
	})
}
