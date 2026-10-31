package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cf4ce0f2-101e-46ec-9368-7395872910b3",
		Name:         "Origin of Captain America",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel("Origin of Captain America", 1, "+1/+1 counter on target creature you control; it gains first strike and vigilance until end of turn"),
				TargetCreature("target creature you control", YouControl()),
				Do(
					Instruction(func(ctx *Context) error {
						return PlusOneCounterOnChosenTarget(ctx.Game, ctx.Item)
					}),
					Instruction(func(ctx *Context) error {
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"first strike", "vigilance"},
							Label:    "Origin of Captain America — first strike and vigilance",
						}.Apply(ctx)
					}),
				)),
			ChapterTrigger(2, SagaChapterLabel("Origin of Captain America", 2, "create a Sturdy Shield Equipment token"),
				Do(CreateToken{Template: SturdyShieldEquipmentToken(), N: 1})),
			ChapterTriggerTargeting(3, SagaChapterLabel("Origin of Captain America", 3, "tap up to one target creature and put a stun counter on it"),
				TargetCreature("up to one target creature").WithCount(0, 1), tapTargetAndStunIt),
		},
	})
}
