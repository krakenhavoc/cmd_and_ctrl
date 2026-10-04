package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Mycosynth Gardens — Land — Sphere:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {X}, {T}: This land becomes a copy of target nontoken artifact you
//	 control with mana value X."
//
// The two mana abilities are Prismatic Lens's. The third is Lazav, the
// Multifarious's shape on a land: a target clause bound to the
// ability's own announced X (WithManaValueEqualsX, re-read at
// resolution, CR 608.2b) and a duration copy with no stated duration
// (BecomeCopy with CopyIndefinite, CR 707.2 and CR 611.2a). It lasts
// until the Gardens leave the battlefield or something else makes them
// a copy (CR 707.4).
//
// There is no except clause, so the copy is the artifact and nothing
// else: it stops being a land (unless the artifact is one) and loses
// all three of these abilities, which is the card. It copies only the
// artifact's copiable values, not its counters, tapped state or
// attachments; an artifact that is itself a copy is copied as what it
// copies; an {X} in the artifact's mana cost counts as 0 for its mana
// value (all three from the rulings). A token can't be targeted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "03f5c566-825c-4c46-9c01-a2f9b1e70a13",
		Name:         "The Mycosynth Gardens",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
		},
		Activated: []ActivatedAbility{{
			Label: "{X}, {T}: This land becomes a copy of target nontoken artifact you control with mana value X.",
			Cost:  Plus(ManaCost("{X}"), TapCost()),
			Targets: TargetPermanent("target nontoken artifact you control with mana value X",
				Artifact(), Not(IsTokenPredicate()), YouControl()).WithManaValueEqualsX(),
			Effect: theMycosynthGardensBecomeACopy,
		}},
	})
}

// theMycosynthGardensBecomeACopy is the copy ability's resolution.
func theMycosynthGardensBecomeACopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	of := FirstLegalBattlefieldTarget(ctx)
	if of == uuid.Nil {
		return nil
	}
	return BecomeCopy{
		Targets:  []uuid.UUID{ctx.Source()},
		Of:       of,
		Duration: CopyIndefinite,
		Label:    "The Mycosynth Gardens — becomes a copy",
	}.Apply(ctx)
}
