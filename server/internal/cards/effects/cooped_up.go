package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cooped Up — Enchantment — Aura {1}{W} (Wilds of Eldraine, #4471):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 {2}{W}: Exile enchanted creature."
//
// Pacifism's restriction plus an activated ability of the Aura itself: only its
// controller may activate it (the ruling, and the default). "Enchanted creature"
// is read as the ability resolves, from the Aura where it is, or, if the Aura
// has left the battlefield in response, from its last-known information
// (CR 608.2h, the creature it was attached to as it last existed).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e8bcb44-b875-4fe4-b96f-97a20cbfec93",
		Name:         "Cooped Up",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{W}: Exile enchanted creature.",
			Cost:    ManaCost("{2}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect: func(g *game.Game, item *game.StackItem) error {
				id := coopedUpHost(g, item.SourceCardID)
				if id == uuid.Nil {
					return nil
				}
				return ExileTarget{Target: id}.Apply(NewContext(g, item))
			},
		}},
	})
}

// coopedUpHost is the creature the Aura `auraID` enchants: the live
// host, or, once the Aura has left, the creature it was attached to as it last
// existed, if that creature is still on the battlefield. uuid.Nil for none.
func coopedUpHost(g *game.Game, auraID uuid.UUID) uuid.UUID {
	if host := attachedHostFor(g, auraID); host != nil && host.IsCreature() {
		return host.InstanceID
	}
	info, ok := g.LastKnownPermanentForEffect(auraID)
	if !ok || info.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	if z := g.FindCardZoneForEffect(info.AttachedTo.ID); z == nil || z.Kind != game.ZoneBattlefield {
		return uuid.Nil
	}
	return info.AttachedTo.ID
}
