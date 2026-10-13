package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thespian's Stage — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}: This land becomes a copy of target land, except it has
//	 this ability."
//
// Mizzium Transreliquat's indefinite line on a land. The copy is
// permanent (CR 611.2a — no duration is printed), takes the target
// land's copiable values (CR 707.2) and keeps functioning as Thespian's
// Stage by granting itself the ability back as a named bundle (CR
// 707.9a), so a Stage that copied a Command Tower can copy again. It
// does NOT keep "{T}: Add {C}" — the first ability is replaced by the
// copied land's own, exactly as printed.
//
// The tap is part of the cost, so the Stage is tapped as it becomes the
// copy and stays tapped (a copy of an untapped land is still a tapped
// permanent). A target that left before resolution does nothing (CR
// 608.2b); the Stage may target itself, which copies nothing new.
//
// No simplification.
const thespiansStageGrant = "thespians-stage/this-ability"

var thespiansStageAbility = ActivatedAbility{
	Label:   "{2}, {T}: This land becomes a copy of target land, except it has this ability.",
	Cost:    Plus(ManaCost("{2}"), TapCost()),
	Targets: TargetPermanent("target land", Land()),
	Effect: func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		targets := ctx.LegalTargets()
		if len(targets) == 0 {
			return nil
		}
		return BecomeCopy{
			Targets:  []uuid.UUID{ctx.Source()},
			Of:       targets[0].ID,
			Duration: CopyIndefinite,
			Except:   func(v *game.PrintedValues) { v.GrantAbility(thespiansStageGrant) },
			Label:    "Thespian's Stage — becomes a copy",
		}.Apply(ctx)
	},
}

func init() {
	Register(Spec{
		OracleID:     "b01e698b-608a-4fc7-8073-b01d044743ec",
		Name:         "Thespian's Stage",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       thespiansStageGrant,
			Text:      "{2}, {T}: This land becomes a copy of target land, except it has this ability.",
			Activated: []ActivatedAbility{thespiansStageAbility},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{thespiansStageAbility},
	})
}
