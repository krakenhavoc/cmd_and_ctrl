package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rent Is Due — Enchantment {W} (Marvel's Spider-Man, #4460):
//
//	"At the beginning of your end step, you may tap two untapped
//	 creatures and/or Treasures you control. If you do, draw a card.
//	 Otherwise, sacrifice this enchantment."
//
// Caparocti Sunborn's tap-two shape: a card-set pick of exactly two (or
// none, the "you may" declined) over the untapped creatures and Treasures
// you control. Each pick is re-read before it is tapped, because the board
// can move under an open prompt. The draw happens only if both were really
// tapped ("if you do"); in every other case, with fewer than two to tap
// included, the enchantment is sacrificed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3fe1530d-bcf2-440b-a905-5a2216daa7eb",
		Name:         "Rent Is Due",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Rent Is Due — tap two creatures and/or Treasures to draw a card, otherwise sacrifice this enchantment", rentIsDueEndStep),
		},
	})
}

func rentIsDueEndStep(g *game.Game, item *game.StackItem) error {
	candidates := rentIsDueUntapped(g, item.Controller)
	if len(candidates) < 2 {
		return nothingIfGone(SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item)))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Rent Is Due — tap two untapped creatures and/or Treasures you control to draw a card? (Otherwise sacrifice Rent Is Due.)",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		ThenKey:  rentIsDueTap,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// rentIsDueTap is the pick being tapped, then the draw or the sacrifice,
// keyed (#4253): the asking item rides the prompt as data, and "exactly
// two" is a registered set rule rather than a closure.
var rentIsDueTap = askedPickRule("choose-cards/rent-is-due-tap-two",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(ctx *Context, r game.CardsChosen) error {
		tapped := 0
		if len(r.Picked) == 2 {
			still := rentIsDueUntapped(ctx.Game, ctx.Controller())
			for _, id := range r.Picked {
				if !slices.Contains(still, id) {
					continue
				}
				if err := ctx.Game.TapTargetForEffect(id); err != nil {
					return err
				}
				tapped++
			}
		}
		if tapped == 2 {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		}
		return nothingIfGone(SacrificePermanent{Target: ctx.Source()}.Apply(ctx))
	})

// rentIsDueUntapped is every untapped creature or Treasure `controller`
// controls. "Treasures" is a word of Rent Is Due's text (CR 612.2).
func rentIsDueUntapped(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || c.Tapped {
			continue
		}
		if c.IsCreature() || c.HasSubtypeWord("Treasure") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
