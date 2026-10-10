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
// Simplification, declared: hexproof FROM planeswalkers is a
// source-qualified hexproof the targeting gate has no shape for
// (ADR 0038 §7 sets out why protection-style keywords are per-card),
// so the card ships without it. Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "06b7987d-c876-4369-9c2d-ce1d35a01097",
		Name:         "Eradicator Valkyrie",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Hexproof from planeswalkers isn't implemented — planeswalkers' abilities can target it.",
		},
		PrintedKeywords: []string{"flying", "lifelink"},
		Activated: []ActivatedAbility{
			BoastAnswering(
				game.AnswerSacOutlet|game.AnswerRemove,
				"{1}{B}, Sacrifice a creature: Each opponent sacrifices a creature or planeswalker.",
				Plus(ManaCost("{1}{B}"), SacrificeACreature()),
				eachOpponentSacrificesACreatureOrPlaneswalker),
		},
	})
}
