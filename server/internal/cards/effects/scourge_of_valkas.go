package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scourge of Valkas — Creature — Dragon {2}{R}{R}{R}, 4/4:
//
//	"Flying
//	 Whenever this creature or another Dragon you control enters, it
//	 deals X damage to any target, where X is the number of Dragons you
//	 control.
//	 {R}: This creature gets +1/+0 until end of turn."
//
// The Dragon that entered is the damage source, and X is counted as the
// ability resolves, so a Dragon that left in response still deals the
// damage (last-known information, CR 608.2h) for the Dragons you control
// then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0a147009-a57c-4ac1-ab81-d8bd8b737631",
		Name:            "Scourge of Valkas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b14SelfOrDragonYouControlEntered(ev, source, g)
			},
			Targets: TargetAny(),
			Key:     "Scourge of Valkas — it deals X damage to any target, where X is the number of Dragons you control",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				x := cmdrB5CountYouControl(g, item.Controller, func(c game.Card) bool {
					return c.IsCreature() && c.HasSubtypeWord("Dragon")
				})
				return DealDamage{Source: item.Trigger.Event.CardID, Target: ts[0].ID, Amount: x}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Scourge of Valkas — +1/+0"}.Apply(ctx)
			},
		}},
	})
}
