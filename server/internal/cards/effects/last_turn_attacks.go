package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// last_turn_attacks.go — the card-facing half of ADR 0108 §6 (#1882):
// "if it attacked during your last turn" and "if a player attacked you
// during their last turn". The engine half is game/last_turn_attacks.go,
// where Player.LastTurnAttacks is written as each turn ends. A card file
// says only which form it prints:
//
//	UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapIfItAttackedDuringYourLastTurn(selfOnly)}, // Goblin Rock Sled
//	UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapIfItAttackedDuringYourLastTurn(AttachedToSource)}, // Tangle Kelp
//	Static: []game.StaticAbility{CantAttackIfItAttackedDuringYourLastTurn()},                       // Giant Turtle
//	SelfCostModifiers: []game.CostModifier{CostsLess(2, "…", APlayerAttackedYouDuringTheirLastTurn())}, // Avenge

// doesntUntapIfItAttackedDuringYourLastTurn is "<it> doesn't untap during
// <its controller's> untap step if it attacked during <its controller's>
// last turn" (CR 502.3). `which` says which permanent the clause is about:
// the source itself (Goblin Rock Sled) or the creature it enchants (Tangle
// Kelp). The controller is the permanent's controller now, which for the
// untap step is the player whose step it is.
func doesntUntapIfItAttackedDuringYourLastTurn(which func(*game.Card, *game.Game, *game.Card) bool) game.UntapStepRestriction {
	return game.UntapStepRestriction{
		Label: "doesn't untap if it attacked during its controller's last turn",
		Restricts: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target != nil && source != nil && which(target, g, source) && g.AttackedDuringControllersLastTurn(target)
		},
	}
}

// CantAttackIfItAttackedDuringYourLastTurn is "This creature can't attack
// if it attacked during your last turn" (Giant Turtle, CR 508.1c). It is
// the creature's own ability and goes with its abilities (CR 613.1f). The
// answer changes only as a turn ends, where the layer cache is already
// invalidated as the next turn begins (onTurnBeganLocked).
func CantAttackIfItAttackedDuringYourLastTurn() game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfOnly(target, g, source) && g.AttackedDuringControllersLastTurn(target)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Restrictions |= game.CantAttack
		},
	}
}

// APlayerAttackedYouDuringTheirLastTurn passes when some other player
// attacked the caster during that player's last turn: "This spell costs
// {2} less to cast if a player attacked you during their last turn"
// (Avenge, CR 601.2f).
func APlayerAttackedYouDuringTheirLastTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.AnyPlayerAttackedYouDuringTheirLastTurn(q.Controller)
	}
}

// TargetNonlandPermanentOfThePlayerHit is the TargetsFrom clause for "target
// nonland permanent that player controls", where "that player" is the
// player a creature just dealt combat damage to (O-Kagachi): a fact about
// what happened, so it is read off the trigger's own event (CR 603.10),
// not off the caster. Returning nil, for a context with no player, drops
// the clause (CR 603.3d).
func TargetNonlandPermanentOfThePlayerHit(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	hit := tc.Event.Target
	if hit == uuid.Nil {
		return nil
	}
	return &game.TargetSpec{
		Mode: "permanent", Label: "target nonland permanent that player controls",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return !c.IsLand() && c.Controller == hit
		},
		Min: 1, Max: 1,
	}
}
