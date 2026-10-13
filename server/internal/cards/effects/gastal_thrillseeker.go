package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gastal Thrillseeker — Creature — Lizard Berserker {B}{R}, 2/3:
//
//	"Start your engines!
//	 When this creature enters, it deals 1 damage to target opponent and
//	 you gain 1 life.
//	 Max speed — This creature has deathtouch and haste."
//
// ADR 0138 (#2122, #2711). The enters trigger targets an opponent; damage
// and life are one ability, so a target that is gone by resolution skips
// both (CR 608.2b). Max speed is two layer 6 grants switched on by the
// controller's speed (CR 702.178a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d185ad9f-c7fc-41f0-b852-4fe2cde0a8ae",
		Name:            "Gastal Thrillseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Gastal Thrillseeker — 1 damage to target opponent and you gain 1 life",
					oneDamageToTargetOpponentAndGainOne),
				TargetPlayer("target opponent", Opponent())),
		},
		Static: MaxSpeedSelfKeywords("deathtouch", "haste"),
	})
}

// oneDamageToTargetOpponentAndGainOne is "<this> deals 1 damage to target
// opponent and you gain 1 life" (Massacre Girl, Most Wanted; Gastal
// Thrillseeker): the targeted opponent is read from the item as it
// resolves, and damage and life are one ability, so a target that is gone
// skips both (CR 608.2b).
func oneDamageToTargetOpponentAndGainOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: victim, Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Amount: 1}.Apply(ctx)
}
