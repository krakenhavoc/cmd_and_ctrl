package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Huskburster Swarm — Creature — Elemental Insect {7}{B}, 6/6
// (Bloomburrow, #4467):
//
//	"This spell costs {1} less to cast for each creature card you own in
//	 exile and in your graveyard.
//	 Menace, deathtouch"
//
// Cyan, Vengeful Samurai's graveyard count with the caster's face-up
// creature cards in exile added. A face-down exiled card has no
// characteristics (CR 406.3a), so it is no creature card; the Swarm
// itself never counts, wherever it is cast from. The reduction is
// generic mana only and never takes the {B} (CR 118.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "faefb352-575c-49f7-9f7a-96f2279a077b",
		Name:            "Huskburster Swarm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(huskbursterCreatureCards,
				"This spell costs {1} less to cast for each creature card you own in exile and in your graveyard."),
		},
	})
}

// huskbursterCreatureCards counts the creature cards the caster owns in
// exile (face up) and in their graveyard, the Swarm itself aside.
func huskbursterCreatureCards(q game.CostQuery) int {
	n := creatureCardsInYourGraveyard(q)
	if q.Game == nil || q.Game.Exile == nil {
		return n
	}
	for _, c := range q.Game.Exile.Cards {
		if c.InstanceID != q.Card.InstanceID && c.Owner == q.Controller && !c.FaceDown && c.IsCreature() {
			n++
		}
	}
	return n
}
