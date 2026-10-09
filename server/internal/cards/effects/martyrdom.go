package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const martyrdomGrant = "martyrdom/redirect"

// martyrdomRedirectRow is the granted "{0}: The next 1 damage that would
// be dealt to target creature, planeswalker, or player this turn is dealt
// to this creature instead." with its "Only you may activate this
// ability" rider.
func martyrdomRedirectRow() ActivatedAbility {
	row := redirectRow("{0}: The next 1 damage that would be dealt to target creature, planeswalker, or player this turn is dealt to this creature instead. Only you may activate this ability.",
		ManaCost("{0}"), targetCreaturePlaneswalkerOrPlayer("target creature, planeswalker, or player"),
		RedirectDamage{Protect: ShieldClause(0), Amount: 1, To: RedirectToThis})
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
	Register(Spec{
		OracleID:     "f6a6da20-52c8-4921-9884-29d3a3051b0d",
		Name:         "Martyrdom",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Grants: []AbilityGrant{{
			Key:       martyrdomGrant,
			Activated: []ActivatedAbility{martyrdomRedirectRow()},
			Text:      "{0}: The next 1 damage that would be dealt to target creature, planeswalker, or player this turn is dealt to this creature instead. Only you may activate this ability.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{martyrdomGrant},
				Label:  "Martyrdom — until end of turn, the creature can take damage for a target",
			}.Apply(ctx)
		},
	})
}
