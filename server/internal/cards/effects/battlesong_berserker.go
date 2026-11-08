package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battlesong Berserker — Creature — Human Berserker {3}{R}, 3/4
// (Foundations, #4465):
//
//	"Whenever you attack, target creature you control gets +1/+0 and
//	 gains menace until end of turn."
//
// Haunted Hellride's "whenever you attack" (once per attack declaration
// of yours, OncePerBatch over attackDeclaredByYou) with a +1/+0 and
// menace in place of the deathtouch and untap. It triggers whether or
// not the Berserker attacks, and the target is any creature you
// control, attacking or not. A target gone by resolution gets nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "396cc8ce-9efa-4fc7-90a6-b439b2b021db",
		Name:         "Battlesong Berserker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Battlesong Berserker — target creature you control gets +1/+0 and gains menace until end of turn",
					battlesongBerserkerPump),
				TargetCreature("target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 1, 0, "menace"))),
		},
	})
}

// battlesongBerserkerPump gives the still-legal target +1/+0 and menace
// until end of turn.
func battlesongBerserkerPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (BoostUntilEOT{Target: target, Power: 1, Label: "Battlesong Berserker — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{Target: target, Keywords: []string{"menace"},
		Label: "Battlesong Berserker — menace"}.Apply(ctx)
}
