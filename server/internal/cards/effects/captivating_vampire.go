package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captivating Vampire — Creature — Vampire {1}{B}{B}, 2/2 (Innistrad
// Remastered, #5948):
//
//	"Other Vampire creatures you control get +1/+1.
//	 Tap five untapped Vampires you control: Gain control of target creature. It
//	 becomes a Vampire in addition to its other types."
//
// The anthem is a layer-7c static over the other Vampire creatures its controller
// controls. The cost is Catapult Master's five-untapped-creatures tap (the
// Captivating Vampire may tap itself among the five, summoning sickness
// notwithstanding, since it is not a {T} cost, CR 302.6). The control change has
// no duration, as Beguiler of Wills', and the creature gains the Vampire type for
// as long as it is controlled this way (CR 613.1d): one indefinite effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "acd382ee-94fb-4a29-b14e-c2a3c6dde7e0",
		Name:         "Captivating Vampire",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			Labelled(b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
					target.IsCreature() && target.HasSubtype(source.TextSubtype("Vampire"))
			}, 1, 1), "Other Vampire creatures you control get +1/+1."),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap five untapped Vampires you control: Gain control of target creature. It becomes a Vampire in addition to its other types.",
			Cost:    tapFiveUntapped("five untapped Vampires you control", "Vampire"),
			Targets: TargetCreature("target creature"),
			Effect:  gapB08StealAndMakeAVampire,
		}},
	})
}

func gapB08StealAndMakeAVampire(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (GainControl{
		Target:     target,
		Controller: item.Controller,
		Duration:   game.IndefiniteDuration(),
		Label:      "Captivating Vampire — gain control (no stated duration)",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("Vampire"))},
		Duration: game.IndefiniteDuration(),
		Label:    "Captivating Vampire — it becomes a Vampire in addition to its other types",
	}.Apply(ctx)
}
