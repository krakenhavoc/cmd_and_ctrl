package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Depthshaker Titan — Artifact Creature — Robot {5}{R}{R}, 5/5 (Edge of
// Eternities Commander):
//
//	"When this creature enters, any number of target noncreature artifacts
//	 you control become 3/3 artifact creatures. Sacrifice them at the
//	 beginning of the next end step.
//	 Each artifact creature you control has melee, trample, and haste.
//	 (Whenever a creature with melee attacks, it gets +1/+1 until end of
//	 turn for each opponent you attacked this combat.)"
//
// The animation is BecomeArtifactCreature with a 3/3 set at layer 7b and no
// duration printed (CR 611.2a), pinned to each artifact as the object it is
// (CR 400.7); Push the Limit's delayed sacrifice then takes exactly those
// artifacts at the next end step, whoever's turn it is, one that has left
// or changed hands being skipped. The grant is a layer-6 static over every
// artifact creature the controller has (the Titan and the animated
// artifacts included, Vehicles and station Spacecraft that are creatures
// too), and melee is the engine's keyword (CR 702.121). As the rulings
// say, a Vehicle animated this way is a 3/3 and was not crewed, and a
// station permanent keeps its counters.
//
// No simplification.
func init() {
	spec := TargetPermanent("any number of target noncreature artifacts you control", Artifact(), Not(Creature()), YouControl()).WithCount(0, 0)
	grant := game.StaticAbility{
		Label:     "Each artifact creature you control has melee, trample, and haste.",
		Layer:     game.Layer6Ability,
		AppliesTo: cmdrB9ArtifactCreaturesYouControl,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, []string{game.KeywordMelee, "trample", "haste"})
		},
		Keywords: []string{game.KeywordMelee, "trample", "haste"},
	}
	Register(Spec{
		OracleID:     "23bf0648-7097-41df-a539-1d36ac42cf9c",
		Name:         "Depthshaker Titan",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Depthshaker Titan — any number of target noncreature artifacts you control become 3/3 artifact creatures; sacrifice them at the beginning of the next end step",
					depthshakerTitanAnimate),
				spec),
		},
	})
}

// cmdrB9ArtifactCreaturesYouControl is the AppliesTo of a static over "each
// artifact creature you control".
func cmdrB9ArtifactCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.IsArtifact() && target.Controller == source.Controller
}

func depthshakerTitanAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var made []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BecomeArtifactCreature{Target: t.ID, SetPower: 3, SetToughness: 3, Label: "Depthshaker Titan — becomes a 3/3 artifact creature"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		made = append(made, t.ID)
	}
	if len(made) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Depthshaker Titan — sacrifice the animated artifacts",
		Cards: made,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}
