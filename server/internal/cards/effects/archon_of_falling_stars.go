package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of Falling Stars — Creature — Archon {4}{W}{W}, 4/4 (Foundations
// Jumpstart, #5929):
//
//	"Flying
//	 When this creature dies, you may return target enchantment card from
//	 your graveyard to the battlefield."
//
// Artisan of Kozilek's optional return of a card from your own graveyard,
// on a dies trigger. The card is chosen as the trigger goes on the stack,
// so the Archon (not an enchantment) is never a pick, and the target is
// judged again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	dies := WhenThisDies("Archon of Falling Stars — return target enchantment card from your graveyard to the battlefield",
		returnFirstLegalGraveyardTargetToBattlefield)
	dies.Targets = TargetCardInGraveyard("target enchantment card from your graveyard", Enchantment(), YouOwn())
	Register(Spec{
		OracleID:        "38e0a1d0-2a4c-4510-9610-abea7fa1de7c",
		Name:            "Archon of Falling Stars",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(dies, "Return an enchantment card from your graveyard to the battlefield?"),
		},
	})
}
