package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warmaker Gunship — Artifact — Spacecraft {2}{R}, 4/3 (Edge of
// Eternities, #4461):
//
//	"When this Spacecraft enters, it deals damage equal to the number of
//	 artifacts you control to target creature an opponent controls.
//	 Station
//	 6+ | Flying"
//
// Explosive Prodigy's enters damage with the artifact count in place of
// the colours: the count is read as the trigger resolves (CR 608.2h), and
// the Gunship counts itself while it is still on the battlefield. The
// damage is the Gunship's, read as it last existed if it has left. With
// no creature an opponent controls, the trigger is removed. Station and
// the 6+ line are Susurian Dirgecraft's helpers: at six or more charge
// counters it is a 4/3 artifact creature with flying (CR 702.184).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c947171b-ed9e-4b83-af45-bd595a8d84ee",
		Name:         "Warmaker Gunship",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Warmaker Gunship — damage equal to the number of artifacts you control to target creature an opponent controls",
					warmakerGunshipDamage),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(6, 4, 3),
			effects.ThresholdKeywords(6, "flying"),
		),
	})
}

// warmakerGunshipDamage deals the controller's artifact count, read now,
// to the announced creature if it is still legal.
func warmakerGunshipDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	n := effects.CountControlled(g, item.Controller, effects.MatchArtifact)
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}
