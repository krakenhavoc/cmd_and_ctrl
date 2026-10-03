package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Refraction Trap — Instant — Trap {3}{W}:
//
//	"If an opponent cast a red instant or sorcery spell this turn, you may
//	 pay {W} rather than pay this spell's mana cost.
//	 Prevent the next 3 damage that a source of your choice would deal to
//	 you and/or permanents you control this turn. If damage is prevented
//	 this way, Refraction Trap deals that much damage to any target."
//
// ADR 0108 §7 and §9 (#1904, #1905): Healing Grace's charged shield
// against a source chosen as it resolves, protecting you and every
// permanent you control (CR 615.7, divided by you when it meets several
// at once — divide_shield), whose CR 615.5 follow-up has Refraction Trap
// deal the amount prevented to its target, carried on the shield (Mod.To)
// from the cast. The rulings: it is not a redirection; the new damage is
// Refraction Trap's, not combat damage; no damage prevented (another
// effect got there first, or it can't be prevented) is no damage dealt; a
// target that can't be dealt damage by then still leaves the shield. The
// trap cost reads the opponents' casts this turn, by colour, as each spell
// became cast (game.CastTally).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "0fc0554d-da45-49d2-941b-3f6c6458e823",
		Name:         "Refraction Trap",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "trap",
			Label:     "If an opponent cast a red instant or sorcery spell this turn, pay {W}",
			ManaCost:  "{W}",
			Condition: anOpponentCastAnInstantOrSorceryOfColor("R"),
		}},
		OnResolve: sourceShieldSpell(PreventDamageFromChosenSource(ShieldYouAndPermanentsYouControl).Charged(3).
			WithThen(dealThatMuchToTheChosenTargetBody).DealingTo(0)),
	})
}

// anOpponentCastAnInstantOrSorceryOfColor is a trap's "if an opponent
// cast a <colour> instant or sorcery spell this turn".
func anOpponentCastAnInstantOrSorceryOfColor(letter string) func(*game.Game, uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		for _, p := range g.Seats {
			if p == nil || p.ID == controller {
				continue
			}
			if g.CastTallyFor(p.ID).CastInstantOrSorceryOfColor(letter) {
				return true
			}
		}
		return false
	}
}
