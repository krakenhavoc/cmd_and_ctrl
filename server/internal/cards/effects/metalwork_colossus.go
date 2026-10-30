package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metalwork Colossus — {11} Artifact Creature — Construct 10/10 (Marvel Super
// Heroes Commander, #4455):
//
//	"This spell costs {X} less to cast, where X is the total mana value of
//	 noncreature artifacts you control.
//	 Sacrifice two artifacts: Return this card from your graveyard to your
//	 hand."
//
// The reduction is a self cost modifier (Excalibur, Sword of Eden's shape)
// summing the printed mana value of each noncreature artifact the caster
// controls; it spends generic mana only and stops at zero. The return is
// Blessed Ghoul's graveyard ability with Breya's cost of two artifacts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64ffe529-ba53-48ec-8af1-b046fbd1f0cc",
		Name:         "Metalwork Colossus",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(totalManaValueOfNoncreatureArtifactsYouControl,
				"This spell costs {X} less to cast, where X is the total mana value of noncreature artifacts you control."),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice two artifacts: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    SacrificeN(2, "two artifacts", Artifact()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

// totalManaValueOfNoncreatureArtifactsYouControl is the amount behind "the
// total mana value of noncreature artifacts you control": the printed mana
// value of each one (CR 202.3), so a token adds nothing.
func totalManaValueOfNoncreatureArtifactsYouControl(q game.CostQuery) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	total := 0
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && c.IsArtifact() && !c.IsCreature() {
			total += c.ManaValue()
		}
	}
	return total
}
