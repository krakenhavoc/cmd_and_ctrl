package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzium Transreliquat — Artifact {3}:
//
//	"{3}: This artifact becomes a copy of target artifact until end of
//	 turn.
//	 {1}{U}{R}: This artifact becomes a copy of target artifact,
//	 except it has this ability."
//
// Two duration copies (#1593, become_copy.go) over the SAME clause
// ("target artifact") at two different prices and two different
// durations, which is why they are two ActivatedAbility entries
// rather than one with a mode:
//
//   - The {3} line is Mirage Mirror's exact shape —
//     selfBecomesCopyOfTarget, an ordinary until-end-of-turn copy with
//     no except clause.
//   - The {1}{U}{R} line is Dimir Doppelganger's shape: an INDEFINITE
//     copy (CR 611.2a — no stated duration) that keeps functioning as
//     Mizzium Transreliquat by granting itself back as a named
//     ability bundle (CR 707.9a), so an artifact this became can
//     activate the {1}{U}{R} line again.
//
// The two abilities do not share a grant: "except it has this
// ability" on the printed card refers to the ability it is PART OF,
// not to the {3} line — copying with the cheap ability gives up the
// expensive one, exactly as printed. Activating the {3} ability on an
// indefinite copy still works (a duration copy on top of another,
// timestamp-ordered, CR 613.7) and reverts to the indefinite copy at
// cleanup rather than to plain Mizzium Transreliquat.
//
// No simplification.
const mizziumTransreliquatGrant = "mizzium-transreliquat/this-ability"

var mizziumTransreliquatKeepAbility = ActivatedAbility{
	Label:   "{1}{U}{R}: This artifact becomes a copy of target artifact, except it has this ability.",
	Cost:    ManaCost("{1}{U}{R}"),
	Targets: TargetPermanent("target artifact", Artifact()),
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		self := ctx.Source()
		return BecomeCopy{
			Targets:    []uuid.UUID{self},
			Of:         targets[0].ID,
			Indefinite: true,
			Except:     func(v *game.PrintedValues) { v.GrantAbility(mizziumTransreliquatGrant) },
			Label:      "Mizzium Transreliquat — becomes a copy",
		}.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:     "a1f73421-cda5-4029-9bcf-f2baee179e5c",
		Name:         "Mizzium Transreliquat",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       mizziumTransreliquatGrant,
			Text:      "{1}{U}{R}: This artifact becomes a copy of target artifact, except it has this ability.",
			Activated: []ActivatedAbility{mizziumTransreliquatKeepAbility},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{3}: This artifact becomes a copy of target artifact until end of turn.",
				Cost:    ManaCost("{3}"),
				Targets: TargetPermanent("target artifact", Artifact()),
				Effect:  selfBecomesCopyOfTarget("Mizzium Transreliquat — becomes a copy until end of turn"),
			},
			mizziumTransreliquatKeepAbility,
		},
	})
}
