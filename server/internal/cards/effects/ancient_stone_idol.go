package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Stone Idol — Artifact Creature — Golem {10}, 12/12:
//
//	"Flash
//	 This spell costs {1} less to cast for each attacking creature.
//	 Trample
//	 When this creature dies, create a 6/12 colorless Construct
//	 artifact creature token with trample."
//
// #746: the reduction is a self cost modifier counting every attacking
// creature, whoever controls it — flash is what makes it matter.
func init() {
	Register(Spec{
		OracleID:        "5f981cca-5cd9-49e4-ab1a-bbbf6fc7e737",
		Name:            "Ancient Stone Idol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(PermanentsOnBattlefield(AttackingCreature()),
				"This spell costs {1} less to cast for each attacking creature."),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Ancient Stone Idol — create a 6/12 Construct",
				Do(CreateToken{Template: TokenCard("6/12 colorless Construct artifact with trample"), N: 1})),
		},
	})
}
