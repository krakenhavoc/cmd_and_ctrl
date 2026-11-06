package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Momo, Friendly Flier — Legendary Creature — Lemur Bat Ally {W}, 1/1
// (Avatar: The Last Airbender, #4459):
//
//	"Flying
//	 The first non-Lemur creature spell with flying you cast during each
//	 of your turns costs {1} less to cast.
//	 Whenever another creature you control with flying enters, Momo gets
//	 +1/+1 until end of turn."
//
// The discount is a static cost modifier on your own creature spells with
// flying that are not Lemurs, during your turn only, while no such spell
// has been cast yet this turn (the cast tally records each creature spell
// with flying and its subtypes as it becomes cast). It reduces generic
// mana only (CR 601.2f): a first such spell with no generic mana in its
// cost is not reduced and still uses the discount up (the ruling). The
// trigger is Watcher of the Spheres': another creature entering under your
// control with flying, read through the layered view.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a3c56e71-0de6-447a-a9b4-0971f56c2362",
		Name:            "Momo, Friendly Flier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "The first non-Lemur creature spell with flying you cast during each of your turns costs {1} less to cast.",
				YourSpell(), b42CreatureSpellWithFlying(), avB24FirstSpellWithoutSubtypeOnYourTurn("Lemur")),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anotherFlierEnteredUnderYourControl, "Momo, Friendly Flier — +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  1, Toughness: 1,
						Label: "Momo, Friendly Flier — +1/+1 until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
