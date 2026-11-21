package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const martyrdomGrant = "martyrdom/redirect"

// martyrdomRedirectRow is the granted "{0}: The next 1 damage that would
// be dealt to target creature, planeswalker, or player this turn is dealt
// to this creature instead." with its "Only you may activate this
// ability" rider.
func martyrdomRedirectRow() effects.ActivatedAbility {
	row := effects.RedirectRow("{0}: The next 1 damage that would be dealt to target creature, planeswalker, or player this turn is dealt to this creature instead. Only you may activate this ability.",
		effects.ManaCost("{0}"), effects.TargetCreaturePlaneswalkerOrPlayer("target creature, planeswalker, or player"),
		effects.RedirectDamage{Protect: effects.ShieldClause(0), Amount: 1, To: effects.RedirectToThis})
	row.GrantorOnly = true
	return row
}

// Martyrdom — Instant {1}{W}{W}:
//
//	"Until end of turn, target creature you control gains "{0}: The next
//	 1 damage that would be dealt to target creature, planeswalker, or
//	 player this turn is dealt to this creature instead." Only you may
//	 activate this ability."
//
// A duration grant (ADR 0093 PR 4) of a bundle whose one row is the
// redirection of ADR 0108 §9, the reverse of Personal Incarnation's
// direction: it protects the ability's TARGET and sends the damage to
// the creature that has the ability (RedirectToThis).
//
// "Only you may activate this ability" is ADR 0106 §1's amendment of
// 2026-10-09 (#1947): the row is GrantorOnly, so its one activator is
// the controller of the spell when it resolved, recorded on the grant
// (game.GrantedAbility.Activator). A creature that changes hands
// afterwards gives its new controller nothing. The activator is "you"
// in the effect, so the redirection is theirs: the creature takes the
// damage, whoever controls it.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f6a6da20-52c8-4921-9884-29d3a3051b0d",
		Name:         "Martyrdom",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Grants: []effects.AbilityGrant{{
			Key:       martyrdomGrant,
			Activated: []effects.ActivatedAbility{martyrdomRedirectRow()},
			Text:      "{0}: The next 1 damage that would be dealt to target creature, planeswalker, or player this turn is dealt to this creature instead. Only you may activate this ability.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{martyrdomGrant},
				Label:  "Martyrdom — until end of turn, the creature can take damage for a target",
			}.Apply(ctx)
		},
	})
}
