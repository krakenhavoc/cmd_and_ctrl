package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ravagerWurmLabel is the enters trigger's stack label.
const ravagerWurmLabel = "Ravager Wurm — choose up to one"

// Ravager Wurm — Creature — Wurm {3}{R}{G}{G}, 4/5:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 When this creature enters, choose up to one —
//	 • This creature fights target creature you don't control.
//	 • Destroy target land with an activated ability that isn't a mana
//	   ability."
//
// Riot is the engine's keyword (ADR 0109 §10), asked before the Wurm
// enters, so the fight already counts a riot counter. The modal trigger
// is a real mode pick (#764); "up to one" lets the controller choose
// neither. The fight is b10Fight, which deals nothing if either
// creature has left (CR 701.12b).
//
// "A land with an activated ability that isn't a mana ability" reads
// the land's activated abilities through game.ActivatedAbilitiesForCard,
// the accessor the activation path itself uses, which lists CR 602
// abilities and not mana abilities (CR 605): a fetch land, Strip Mine
// or a creature land qualifies, a basic land does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40b872b5-3ec2-4fcc-b152-91966f67c2be",
		Name:            "Ravager Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
		Triggered:       []game.TriggeredAbility{ravagerWurmTrigger()},
	})
}

func ravagerWurmTrigger() game.TriggeredAbility {
	t := WhenThisEnters(ravagerWurmLabel, func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseN("Choose up to one", 0, 1,
		ModeDoing("This creature fights target creature you don't control.",
			TargetCreature("target creature you don't control", OpponentControls()),
			func(item *game.StackItem, ctx *Context, occ int) error {
				target, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return b10Fight(ctx, item.SourceCardID, target.ID)
			}),
		ModeDoing("Destroy target land with an activated ability that isn't a mana ability.",
			TargetPermanent("target land with an activated ability that isn't a mana ability", Land(), landWithNonManaActivatedAbility),
			DestroyTheModesTarget),
	)
	return t
}

// landWithNonManaActivatedAbility passes a permanent with at least one
// CR 602 activated ability. Mana abilities (CR 605) are not on the list
// game.ActivatedAbilitiesForCard returns.
func landWithNonManaActivatedAbility(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return len(game.ActivatedAbilitiesForCard(c)) > 0
}
