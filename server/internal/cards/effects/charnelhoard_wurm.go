package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Charnelhoard Wurm — Creature — Wurm {4}{B}{R}{G}, 6/6 (Conflux,
// #5587):
//
//	"Trample
//	 Whenever this creature deals damage to an opponent, you may return
//	 target card from your graveyard to your hand."
//
// Any damage to an opponent, combat or not, triggers it. The target is
// chosen as the trigger goes on the stack (CR 603.3d) and the "you may"
// is asked as it resolves (CR 603.5); a card that left the graveyard in
// response is not returned (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4a430fa3-e693-424b-9981-d7d8193445e3",
		Name:            "Charnelhoard Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				On(game.EventDealDamage, ThisDealtDamageToAnOpponent,
					"Charnelhoard Wurm — return target card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target card from your graveyard", YouOwn())),
				"Charnelhoard Wurm — return target card from your graveyard to your hand?"),
		},
	})
}
