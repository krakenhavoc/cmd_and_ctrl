package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lizard, Connors's Curse — Legendary Creature — Lizard Villain {2}{G}{G},
// 5/5 (Marvel's Spider-Man, #4460):
//
//	"Trample
//	 Lizard Formula — When Lizard, Connors's Curse enters, up to one other
//	 target creature loses all abilities and becomes a green Lizard
//	 creature with base power and toughness 4/4."
//
// An enter trigger with an "up to one other target creature" clause (Rescue,
// Pepper Potts' shape, so Lizard is never her own target). The change is one
// indefinite continuous effect pinned to that object (Hunted by The Family's
// shape, CR 611.2a, 400.7), so a creature that blinks is a new object and is
// no longer a Lizard, and a creature that enters later is untouched. Its mods
// by layer, as printed:
//
//   - layer 4, "becomes a Lizard creature": SetTypesMod replaces its card
//     types and creature types (CR 205.1a), so an artifact creature stops
//     being an artifact;
//   - layer 5, "green";
//   - layer 6, "loses all abilities" (CR 613.1f);
//   - layer 7b, "base power and toughness 4/4": a set, so counters and
//     anthems still apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "48d1b217-b068-468c-bc6c-4bef9e44cc6e",
		Name:            "Lizard, Connors's Curse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Lizard Formula — up to one other target creature loses all abilities and becomes a green Lizard creature with base power and toughness 4/4",
					lizardFormula),
				Another(TargetCreature("up to one other target creature").WithCount(0, 1))),
		},
	})
}

// lizardFormula is the trigger's body: the chosen creature, if it is still a
// legal target, becomes a 4/4 green Lizard with no abilities.
func lizardFormula(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
			continue
		}
		return lizardFormulaOn(ctx, t.ID)
	}
	return nil
}

func lizardFormulaOn(ctx *Context, id uuid.UUID) error {
	return ScopedEffectFor{
		Target: id,
		Mods: []game.Mod{
			game.SetTypesMod([]string{"Creature"}, ctx.TextSubtypes("Lizard")),
			game.SetColorsMod(ctx.TextColors("G")...),
			game.LoseAllAbilitiesMod(),
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
		},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
		Label:    "Lizard Formula — becomes a green 4/4 Lizard and loses all abilities",
	}.Apply(ctx)
}
