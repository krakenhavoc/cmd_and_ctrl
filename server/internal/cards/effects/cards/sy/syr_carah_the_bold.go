package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Syr Carah, the Bold — Legendary Creature — Human Knight {3}{R}{R}, 3/3
// (Throne of Eldraine, #5516):
//
//	"Whenever Syr Carah or an instant or sorcery spell you control deals
//	 damage to a player, exile the top card of your library. You may play
//	 that card this turn.
//	 {T}: Syr Carah deals 1 damage to any target."
//
// The trigger reads each damage event: Syr Carah as the source, or an
// instant or sorcery spell its controller controls that is mid-resolution
// (Bloodfeather Phoenix's read), dealing positive damage to any player.
// The exile is Kirk's impulse exile. The ping is Rootwater Hunter's.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3da0550-b816-4aed-8174-ddde199b6589",
		Name:         "Syr Carah, the Bold",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Syr Carah deals 1 damage to any target.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, syrCarahDamage,
				"Syr Carah — exile the top card of your library; you may play that card this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := effects.B12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}

// syrCarahDamage is the trigger event: positive damage to a player from
// Syr Carah herself or from an instant or sorcery spell her controller
// controls.
func syrCarahDamage(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil || ev.Target == uuid.Nil {
		return false
	}
	if g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	if ev.Source == source.InstanceID {
		return true
	}
	spell, item, ok := g.ResolvingSpellForEffect(ev.Source)
	return ok && item.Kind == game.StackItemSpell && item.Controller == source.Controller &&
		(spell.IsInstant() || spell.IsSorcery())
}
