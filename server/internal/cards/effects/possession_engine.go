package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Possession Engine — Artifact — Vehicle {3}{U}{U}, 5/5 (Aetherdrift,
// #4464):
//
//	"When this Vehicle enters, gain control of target creature an
//	 opponent controls for as long as you control this Vehicle. That
//	 creature can't attack or block for as long as you control this
//	 Vehicle.
//	 Crew 3"
//
// Mind Flayer's theft (GainControl with DurationWhileYouControlSource)
// and, under the same duration, a can't-attack-or-block restriction
// record pinned to the stolen creature as the object it is now (Kang,
// Suspicious Captain's ScopedEffectFor shape). Both end the moment the
// Engine leaves or someone else takes it. An Engine that has already
// left, or changed hands, by the time the trigger resolves takes nothing
// and restricts nothing: the condition is false as the effects would
// begin, so they never begin. Crew 3 is the ordinary crew ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab7c0256-eea9-4e0d-932d-c25460eed836",
		Name:         "Possession Engine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Possession Engine — gain control of target creature an opponent controls; it can't attack or block while you control this Vehicle",
					possessionEngineTakes),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Possession Engine"),
		}},
	})
}

// possessionEngineTakes steals the target and bars it from attacking and
// blocking, both for as long as the trigger's controller controls the
// Engine.
func possessionEngineTakes(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	d, ok := DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
	if !ok {
		return nil
	}
	if err := (GainControl{
		Target:     target,
		Controller: item.Controller,
		Duration:   d,
		Label:      "Possession Engine — control for as long as you control Possession Engine",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddRestrictionsMod(game.CantAttackOrBlock)},
		Duration: d,
		Label:    "Possession Engine — can't attack or block for as long as you control Possession Engine",
	}.Apply(ctx)
}
