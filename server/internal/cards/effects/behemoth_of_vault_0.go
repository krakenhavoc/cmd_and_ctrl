package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Behemoth of Vault 0 — Artifact Creature — Robot {6}, 6/6:
//
//	"Trample
//	 When this creature enters, you get {E}{E}{E}{E} (four energy
//	 counters).
//	 When this creature dies, you may pay an amount of {E} equal to
//	 target nonland permanent's mana value. When you do, destroy that
//	 permanent."
//
// ADR 0129 §3 (#1995): the dies trigger targets the permanent as it
// goes on the stack (CR 603.3d). As it resolves, the amount is that
// permanent's mana value then, paid through the energy prompt (CR
// 118.12); "when you do" is a reflexive trigger (CR 603.12) that
// destroys the same permanent, carried in its payload rather than
// targeted again.
//
// No simplification.
func init() {
	dies := WhenThisDies("Behemoth of Vault 0 — you may pay {E} equal to target nonland permanent's mana value",
		func(g *game.Game, item *game.StackItem) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return MayPayEnergy{
				N:        firstTargetManaValue(g, item),
				Question: "Behemoth of Vault 0 — pay {E} equal to its mana value to destroy it?",
				OnPay: func(ctx *Context) error {
					return ReflexiveTrigger{
						Label: "Behemoth of Vault 0 — destroy that permanent",
						Cards: []uuid.UUID{ctx.Item.Targets[0].ID},
						Body:  behemothDestroyBody,
					}.Apply(ctx)
				},
			}.Apply(NewContext(g, item))
		})
	dies.Targets = TargetPermanent("target nonland permanent", Nonland())
	Register(Spec{
		OracleID:        "7d05ffe2-e52b-42af-87ca-6aae488b9f41",
		Name:            "Behemoth of Vault 0",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Energy: 4},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Behemoth of Vault 0", 4),
			dies,
		},
	})
}
