package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arni, Renowned Champion — Legendary Creature — Human Berserker
// {3}{R}, 1/5:
//
//	"Trample
//	 Whenever another creature you control enters, Arni gets +X/+0
//	 until end of turn, where X is that creature's power."
//
// X is read as the trigger resolves, from the entering creature's
// current power (its last-known power if it has already left).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1a929c4f-ae1f-44d9-a0a7-2ca2dc928dee",
		Name:            "Arni, Renowned Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Arni, Renowned Champion — +X/+0 until end of turn, where X is that creature's power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					return BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  info.Power,
						Label:  "Arni, Renowned Champion — +X/+0",
					}.Apply(ctx)
				}),
		},
	})
}
