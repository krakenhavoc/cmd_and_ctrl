package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Origin of Captain America — Enchantment — Saga {2}{W} (Marvel Super
// Heroes Commander, #4455):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Put a +1/+1 counter on target creature you control. It gains
//	 first strike and vigilance until end of turn.
//	 II — Create a colorless Equipment artifact token named Sturdy Shield
//	 with "Equipped creature gets +1/+2" and equip {2}.
//	 III — Tap up to one target creature and put a stun counter on it."
//
// I is two instructions in printed order (the keywords wait for the
// counter's replacement prompt, CR 608.2c); a target that left makes the
// whole chapter do nothing (CR 608.2b). II makes U.S.Agent's Sturdy Shield
// token (SturdyShieldEquipmentToken), unattached: the chapter does not
// attach it. III is Stunning Shot's tap-and-stun on up to one target
// creature (tapTargetAndStunIt), any creature, yours included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf4ce0f2-101e-46ec-9368-7395872910b3",
		Name:         "Origin of Captain America",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel("Origin of Captain America", 1, "+1/+1 counter on target creature you control; it gains first strike and vigilance until end of turn"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						return effects.PlusOneCounterOnChosenTarget(ctx.Game, ctx.Item)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"first strike", "vigilance"},
							Label:    "Origin of Captain America — first strike and vigilance",
						}.Apply(ctx)
					}),
				)),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("Origin of Captain America", 2, "create a Sturdy Shield Equipment token"),
				effects.Do(effects.CreateToken{Template: effects.SturdyShieldEquipmentToken(), N: 1})),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel("Origin of Captain America", 3, "tap up to one target creature and put a stun counter on it"),
				effects.TargetCreature("up to one target creature").WithCount(0, 1), effects.TapTargetAndStunIt),
		},
	})
}
