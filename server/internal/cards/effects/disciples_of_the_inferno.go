package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disciples of the Inferno — Creature — Human Monk 4/4, the BACK FACE of
// Invasion of Regatha (face 1):
//
//	"Prowess
//	 If a noncreature source you control would deal damage to a
//	 creature, battle, or opponent, it deals that much damage plus 2
//	 instead."
//
// Registered under "<oracle_id>#1". The source is read as the damage
// event snapshotted it (a spell on the stack, a permanent as it last
// was), so a creature that has died is still a creature source and an
// animated Vehicle is not a noncreature one. The recipient is a creature
// or battle on the battlefield (yours included) or any player but you;
// a planeswalker is not on the list. Every source's damage is +2, each
// Disciples adding its own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        invasionOfRegathaOracleID + "#1",
		Name:            "Disciples of the Inferno",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
					return false
				}
				ch, ok := damageSourceCharacteristics(ev, g)
				if !ok || ch.Controller != src.Controller || slices.Contains(ch.Types, "Creature") {
					return false
				}
				return damageHitsCreatureBattleOrOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount += 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Disciples of the Inferno: a noncreature source you control deals 2 more damage",
		}},
	})
}

// damageHitsCreatureBattleOrOpponentOf is "a creature, battle, or
// opponent" of `you`: any player but you, or a creature or battle on the
// battlefield whoever controls it.
func damageHitsCreatureBattleOrOpponentOf(ev *game.ReplacementEvent, g *game.Game, you uuid.UUID) bool {
	if ev.DamageTarget == uuid.Nil {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.DamageTarget); p != nil {
		return p.ID != you
	}
	c, ok := g.LookupCardForEffect(ev.DamageTarget)
	return ok && onBattlefield(g, c.InstanceID) && (c.IsCreature() || c.IsBattle())
}
