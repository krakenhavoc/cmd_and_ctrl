package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eradicator Valkyrie — Creature — Angel Berserker {2}{B}{B}, 4/3:
//
//	"Flying, lifelink, hexproof from planeswalkers
//	 Boast — {1}{B}, Sacrifice a creature: Each opponent sacrifices a creature or planeswalker.
//	 (Activate only if this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go).
// "Each opponent sacrifices a creature or planeswalker" is an edict:
// it does not target, so hexproof and shroud are irrelevant and each
// opponent chooses their own (EachPlayerSacrifices).
//
// Hexproof from planeswalkers is game.HexproofFromPlaneswalkers
// (#3172, ADR 0038's amendment of 2026-10-13): hexproof that refuses
// only a source with the quality, tested at the targeting choke point
// the way protection is and only against an opponent's source. Its
// controller's own planeswalkers can still target it.
func init() {
	Register(Spec{
		OracleID:        "06b7987d-c876-4369-9c2d-ce1d35a01097",
		Name:            "Eradicator Valkyrie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink", game.HexproofFromPlaneswalkers},
		Activated: []ActivatedAbility{
			BoastAnswering(
				game.AnswerSacOutlet|game.AnswerRemove,
				"{1}{B}, Sacrifice a creature: Each opponent sacrifices a creature or planeswalker.",
				Plus(ManaCost("{1}{B}"), SacrificeACreature()),
				eachOpponentSacrificesACreatureOrPlaneswalker),
		},
	})
}
