package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raff, Weatherlight Stalwart — Legendary Creature — Human Wizard {W}{U},
// 1/3 (Dominaria United, #5487):
//
//	"Whenever you cast an instant or sorcery spell, you may tap two untapped
//	 creatures you control. If you do, draw a card.
//	 {3}{W}{W}: Creatures you control get +1/+1 and gain vigilance until end
//	 of turn."
//
// The tap is a choice made as the trigger resolves (Caparocti Sunborn's
// shape): a pick of exactly two untapped creatures you control, or none
// (the "you may" declined). Raff himself may be one, being untapped. Each
// pick is re-read before it is tapped, and the card is drawn only if both
// were really tapped ("if you do"). The activated ability is the
// creatures-you-control pump (Battershield Warrior's) with vigilance
// granted to the same set, read as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	const name = "Raff, Weatherlight Stalwart"
	pump := ActivatedAbility{
		Label:   "{3}{W}{W}: Creatures you control get +1/+1 and gain vigilance until end of turn.",
		Cost:    ManaCost("{3}{W}{W}"),
		Purpose: game.Purpose{Answers: game.AnswerPump, TeamPump: &game.Pump{Power: 1, Toughness: 1, Keywords: []string{"vigilance"}}},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: name + " — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"vigilance"}, Label: name + " — vigilance"}.Apply(ctx)
		},
	}
	Register(Spec{
		OracleID:     "41f3acbf-6092-435f-a5af-4d80dd6e1aef",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				name+" — you may tap two untapped creatures you control; if you do, draw a card",
				raffOfferTap),
		},
		Activated: []ActivatedAbility{pump},
	})
}

func raffOfferTap(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	candidates := untappedCreaturesControlledBy(g, controller)
	if len(candidates) < 2 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Raff, Weatherlight Stalwart — tap two untapped creatures you control to draw a card?",
		Cards:    candidates,
		Min:      0,
		Max:      2,
		ThenKey:  raffTap,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// raffTap taps the two chosen creatures, then draws if both were tapped.
var raffTap = askedPickRule("choose-cards/raff-weatherlight-stalwart-tap-two",
	func(picked []game.Card) bool { return len(picked) == 2 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 2 {
			return nil
		}
		still := untappedCreaturesControlledBy(ctx.Game, ctx.Controller())
		tapped := 0
		for _, id := range r.Picked {
			if !slices.Contains(still, id) {
				continue
			}
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
			tapped++
		}
		if tapped != 2 {
			return nil
		}
		return DrawCards{N: 1}.Apply(ctx)
	})
