package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ria Ivor, Bane of Bladehold — Legendary Creature — Phyrexian Knight
// {2}{W}{B}, 3/4 (ONE):
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 At the beginning of combat on your turn, the next time target
//	 creature would deal combat damage to one or more players this
//	 combat, prevent that damage. If damage is prevented this way,
//	 create that many 1/1 colorless Phyrexian Mite artifact creature
//	 tokens with toxic 1 and "This token can't block.""
//
// Battle cry is the shared keyword trigger. The second ability makes a
// next-damage shield (CR 615.8) on the target, as the trigger resolves:
// combat damage only, and only damage to players, so damage to a blocker
// or a planeswalker neither meets nor spends it, and it ends with this
// combat (CR 611.2b). The damage is prevented whole, and the follow-up
// (CR 615.5) creates one Mite per point prevented, as many as the shield
// prevented across every player the creature hurt in that one instance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "785f55bb-d47c-470c-94d4-95f69055e35f",
		Name:         "Ria Ivor, Bane of Bladehold",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			BattleCry("Ria Ivor, Bane of Bladehold"),
			Targeting(
				AtBeginningOfYourCombat("Ria Ivor, Bane of Bladehold — prevent the next combat damage the creature would deal to players; create that many Phyrexian Mites",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id := firstLegalCardTarget(ctx)
						if id == uuid.Nil {
							return nil
						}
						return PreventNextDamageFromSource{
							From:          id,
							Protect:       ShieldAnything,
							CombatOnly:    true,
							ToPlayersOnly: true,
							ThisCombat:    true,
							Then:          riaIvorMitesBody,
							Label:         "Ria Ivor, Bane of Bladehold — prevent the next combat damage to players",
						}.Apply(ctx)
					}),
				TargetCreature("target creature")),
		},
	})
}
