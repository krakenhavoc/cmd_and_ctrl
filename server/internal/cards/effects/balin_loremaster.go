package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Balin, Loremaster — Legendary Creature — Dwarf Bard {3}{R}{R}, 4/4:
//
//	"Storied (If you control three or more artifacts, legendaries,
//	 and/or Sagas, you have an enduring story for the rest of the game.)
//	 Whenever Balin or another Dwarf you control enters, you may discard
//	 your hand. Draw X cards, where X is the number of cards discarded
//	 this way. If you have an enduring story, Balin deals X damage to
//	 each opponent."
//
// Storied is the engine's (CR 702.195a). The "you may" is the trigger's
// own yes/no; on a yes the whole hand is discarded (no choice of cards),
// X is filed as a step record, and the draw and the damage are later
// steps that each wait for the one before (CR 608.2c). "If you have an
// enduring story" is a clause of the effect, read as it resolves, not an
// intervening "if": the story told by the Dwarf that just entered counts.
// Balin deals the damage as he last existed if he has left (CR 608.2h).
//
// No simplification. The Hobbit (#4475).
const balinLoremasterDiscarded = "balin-loremaster/discarded"

func init() {
	Register(Spec{
		OracleID:        "0d420e41-43e9-41d6-832c-5a9f410c994e",
		Name:            "Balin, Loremaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordStoried},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, balinOrAnotherDwarfEntered, "Balin, Loremaster — discard your hand, draw that many",
				Do(
					Instruction(func(ctx *Context) error {
						n, err := discardWholeHand(ctx.Game, ctx.Controller())
						ctx.RecordStep(balinLoremasterDiscarded, nil, n)
						return err
					}),
					Instruction(func(ctx *Context) error {
						x := ctx.StepRecord(balinLoremasterDiscarded).N()
						if x <= 0 {
							return nil
						}
						return DrawCards{N: x}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						x := ctx.StepRecord(balinLoremasterDiscarded).N()
						if x <= 0 || !YouHaveAnEnduringStory(ctx.Game, ctx.Controller()) {
							return nil
						}
						return DamageEachOpponent{N: x}.Apply(ctx)
					}),
				)),
				"Balin, Loremaster — discard your hand? (Draw that many cards; with an enduring story, Balin deals that much damage to each opponent.)"),
		},
	})
}

// balinOrAnotherDwarfEntered is "Balin or another Dwarf you control
// enters".
func balinOrAnotherDwarfEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Dwarf"))
}
