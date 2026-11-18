package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kyodai, Soul of Kamigawa — {3}{W} Legendary Creature — Dragon Spirit 3/3
// (Kamigawa: Neon Dynasty, S82 NEO batch 20):
//
//	"Flash
//	 Flying
//	 When Kyodai enters, another target permanent gains indestructible for
//	 as long as you control Kyodai.
//	 {W}{U}{B}{R}{G}: Kyodai gets +5/+5 until end of turn."
//
// The grant is a layer-6 effect on the target that lasts while Kyodai's
// controller (the controller as the ability resolves) still controls
// Kyodai (CR 611.2b). If Kyodai has left or changed hands by resolution
// the effect never begins. The target is mandatory, another permanent of
// any controller; with no other permanent the trigger is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d89bd776-8489-4c34-94ea-db1b088d9379",
		Name:            "Kyodai, Soul of Kamigawa",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Kyodai, Soul of Kamigawa — another target permanent gains indestructible for as long as you control Kyodai",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					d, ok := DurationWhileYouControlSource(ctx, ctx.Source(), ctx.Controller())
					if !ok {
						return nil
					}
					return ScopedEffectFor{
						Target:   id,
						Mods:     []game.Mod{game.AddKeywordsMod("indestructible")},
						Duration: d,
						Label:    "Kyodai, Soul of Kamigawa — indestructible for as long as you control Kyodai",
					}.Apply(ctx)
				}),
				Another(TargetPermanent("another target permanent"))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}: Kyodai gets +5/+5 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{W}{U}{B}{R}{G}"),
			Effect: Do(Instruction(func(ctx *Context) error {
				if !b15OnBattlefield(ctx.Game, ctx.Item.SourceCardID) {
					return nil
				}
				return BoostUntilEOT{Target: ctx.Item.SourceCardID, Power: 5, Toughness: 5, Label: "Kyodai, Soul of Kamigawa — +5/+5"}.Apply(ctx)
			})),
		}},
	})
}
