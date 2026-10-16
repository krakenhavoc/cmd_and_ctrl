package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caretaker's Talent — Enchantment — Class {2}{W} (EDHREC rank 633):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 Whenever one or more tokens you control enter, draw a card. This
//	 ability triggers only once each turn.
//	 {W}: Level 2
//	 When this Class becomes level 2, create a token that's a copy of
//	 target token you control.
//	 {3}{W}: Level 3
//	 Creature tokens you control get +2/+2."
//
// ADR 0071 Class machinery, each level's ability gated so it exists
// only from that level on (CR 716.2a):
//
//   - Level 1 watches token entries under your control, once per batch
//     (one draw for a whole Secure the Wastes), and "only once each
//     turn" is the per-turn trigger record (b11TriggeredThisTurn), so a
//     countered draw still used the turn's trigger, as printed.
//   - Level 2 is BecomesLevel with a target: a token copy of a token
//     you control (CreateTokenCopy, CR 707.2), the target chosen as the
//     trigger goes on the stack.
//   - Level 3 is Intangible Virtue's anthem at +2/+2.
//
// No simplification.
func init() {
	const drawLabel = "Caretaker's Talent — draw a card (tokens entered)"
	Register(Spec{
		OracleID:     "c97957a2-8310-4cff-8aad-871b7901d124",
		Name:         "Caretaker's Talent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsToken() && !b11TriggeredThisTurn(g, source.InstanceID, drawLabel)
			}, drawLabel, Do(DrawCards{N: 1}))),
			Targeting(BecomesLevel(2, "Caretaker's Talent — create a token that's a copy of target token you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return CreateTokenCopy{Controller: ctx.Controller(), Copy: t.ID, N: 1}.Apply(ctx)
					}
					return nil
				}),
				TargetPermanent("target token you control", IsTokenPredicate(), YouControl())),
		},
		Static: []game.StaticAbility{{
			Label:      "Creature tokens you control get +2/+2.",
			Layer:      game.Layer7PT,
			SubLayer:   game.SubLayer7C_Modify,
			AppliesTo:  b05CreatureTokenYouControl,
			ActiveWhen: Level(3),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 2
				c.Toughness += 2
			},
		}},
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{W}")),
			LevelUp(3, ManaCost("{3}{W}")),
		},
	})
}
