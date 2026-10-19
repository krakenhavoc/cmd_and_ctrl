package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Masamune — Legendary Artifact — Equipment {3}:
//
//	"As long as equipped creature is attacking, it has first strike and
//	 must be blocked if able.
//	 Equipped creature has "If a creature dying causes a triggered ability
//	 of this creature or an emblem you own to trigger, that ability
//	 triggers an additional time."
//	 Equip {2}"
//
// The quoted ability is GIVEN to the equipped creature (CR 113.10), so
// "this creature" is the host and "you" its controller: a CR 603.2d
// trigger doubler carried by a granted bundle (#4137, ADR 0018's trigger
// doubling addendum, amended 2026-10-19). It doubles a trigger whose
// source is the host itself or an emblem its controller owns, and only
// when the cause is a creature dying (Teysa Karlov's cause: any creature,
// tokens included). The host's other triggers, and any other permanent's,
// are not doubled.
//
// The first half reads live combat state: first strike and the
// must-be-blocked requirement both apply only while the host has been
// declared as an attacker (DependsOnAttackingStatus wakes the layer cache
// as attackers are declared and again when combat ends).
//
// No simplification.
const masamuneDoublerGrant = "the-masamune/dying-doubler"

func init() {
	hostIsAttacking := func(host *game.Card) bool { return host.AttackingTarget != uuid.Nil }
	firstStrike := GrantToAttachedWhile(func(host *game.Card, _ *game.Game, _ *game.Card) bool {
		return hostIsAttacking(host)
	}, "first strike")
	firstStrike.Label = "As long as equipped creature is attacking, it has first strike."
	firstStrike.DependsOnAttackingStatus = true
	mustBeBlocked := BlockRequirementWhere(game.BlockRequirementMustBeBlocked, func(target *game.Card, g *game.Game, source *game.Card) bool {
		return AttachedToSource(target, g, source) && hostIsAttacking(target)
	})
	mustBeBlocked.Label = "As long as equipped creature is attacking, it must be blocked if able."
	mustBeBlocked.DependsOnAttackingStatus = true

	Register(Spec{
		OracleID:     "89eb4c39-35ae-405b-8af4-779880cb11c9",
		Name:         "The Masamune",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:             masamuneDoublerGrant,
			TriggerDoublers: []game.TriggerDoubler{masamuneDoubler()},
			Text:            "If a creature dying causes a triggered ability of this creature or an emblem you own to trigger, that ability triggers an additional time.",
		}},
		Static: []game.StaticAbility{
			firstStrike,
			mustBeBlocked,
			GrantAbilitiesToAttached(masamuneDoublerGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// masamuneDoubler is the granted quoted ability. DoublesDying asks for a
// permanent the doubler's controller controls; this one names the host
// itself and an emblem its controller owns, so it is its own predicate.
func masamuneDoubler() game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    "The Masamune — a creature dying causes a triggered ability to trigger an additional time",
		RowLabel: "If a creature dying causes a triggered ability of this creature or an emblem you own to trigger, that ability triggers an additional time.",
		Applies: func(_ *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.HasSubject || !isDying(q.Event) {
				return false
			}
			switch {
			case q.Source.InstanceID == q.Doubler.InstanceID:
				return true
			case q.Source.IsEmblem():
				return q.Source.Owner == q.DoublerLKI.Controller
			}
			return false
		},
	}
}
