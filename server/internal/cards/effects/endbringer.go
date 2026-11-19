package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Endbringer — Creature — Eldrazi {5}{C}, 5/5:
//
//	"Untap this creature during each other player's untap step.
//	 {T}: This creature deals 1 damage to any target.
//	 {C}, {T}: Target creature can't attack or block this turn.
//	 {C}{C}, {T}: Draw a card."
//
// The untap is Thousand Moons Infantry's (an UntapStepPermission aimed at
// the source). The {C} in the costs is colourless mana only (CR 107.4c).
// The three tap abilities are a ping (Apprentice Sorcerer's), Off Balance's
// restriction on a target creature until end of turn, and Snapping
// Voidcraw's draw.
func init() {
	const name = "Endbringer"
	Register(Spec{
		OracleID:     "db1ea6e0-4076-49dd-92f7-7e6d7eb65a27",
		Name:         name,
		Completeness: CompletenessFull,
		UntapStep: []game.UntapStepPermission{
			untapSelfDuringEachOtherPlayersUntapStep(name + " — untap this creature"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: This creature deals 1 damage to any target.",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 1)),
				Effect:  sourceDealsDamageToEachLegalTarget(1),
			},
			{
				Label:   "{C}, {T}: Target creature can't attack or block this turn.",
				Cost:    Plus(ManaCost("{C}"), TapCost()),
				Targets: TargetCreature("target creature"),
				Effect:  nemTargetRestrictedThisTurn(game.CantAttackOrBlock, name+" — can't attack or block this turn"),
			},
			{
				Label:   "{C}{C}, {T}: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{C}{C}"), TapCost()),
				Effect:  b36DrawOne,
			},
		},
	})
}
