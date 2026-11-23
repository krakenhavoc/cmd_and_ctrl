package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
// creature has left (CR 701.14b).
//
// "A land with an activated ability that isn't a mana ability" reads
// the land's activated abilities through game.ActivatedAbilitiesForCard,
// the accessor the activation path itself uses, which lists CR 602
// abilities and not mana abilities (CR 605): a fetch land, Strip Mine
// or a creature land qualifies, a basic land does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "40b872b5-3ec2-4fcc-b152-91966f67c2be",
		Name:            "Ravager Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
		Triggered:       []game.TriggeredAbility{ravagerWurmTrigger()},
	})
}

func ravagerWurmTrigger() game.TriggeredAbility {
	t := effects.WhenThisEnters(ravagerWurmLabel, func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseN("Choose up to one", 0, 1,
		effects.ModeDoing("This creature fights target creature you don't control.",
			effects.Harmful(effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
			effects.ThisFightsTheModesTarget),
		effects.ModeDoing("Destroy target land with an activated ability that isn't a mana ability.",
			effects.Harmful(effects.TargetPermanent("target land with an activated ability that isn't a mana ability", effects.Land(), LandWithNonManaActivatedAbility)),
			effects.DestroyTheModesTarget),
	)
	return t
}

// landWithNonManaActivatedAbility passes a permanent with at least one
// CR 602 activated ability. Mana abilities (CR 605) are not on the list
// game.ActivatedAbilitiesForCard returns.
func LandWithNonManaActivatedAbility(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return len(game.ActivatedAbilitiesForCard(c)) > 0
}
