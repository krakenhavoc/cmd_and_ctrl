package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aziza, Mage Tower Captain — Legendary Creature — Djinn Sorcerer {R}{W},
// 2/2 (Secrets of Strixhaven):
//
//	"Whenever you cast an instant or sorcery spell, you may tap three
//	 untapped creatures you control. If you do, copy that spell. You may
//	 choose new targets for the copy."
//
// The "may tap" is the pick itself (Caparocti Sunborn's shape): a prompt
// over the untapped creatures you control that takes exactly three or
// nothing, so declining, or having fewer than three, copies nothing. Any
// untapped creature counts, Aziza and a summoning-sick one included,
// because tapping them is an effect, not a {T} cost (CR 302.6). The three
// are re-checked as the answer lands; the copy is then made from
// last-known information if the spell has already left the stack
// (CR 608.2h), with the new-targets choice the copy's controller is owed.
func init() {
	Register(Spec{
		OracleID:     "70e10cb9-ae27-4dba-9a0d-34a78f924774",
		Name:         "Aziza, Mage Tower Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Aziza, Mage Tower Captain — tap three untapped creatures you control to copy that spell",
				socB20AzizaTapThree),
		},
	})
}

func socB20AzizaTapThree(g *game.Game, item *game.StackItem) error {
	candidates := socB20UntappedCreaturesYouControl(g, item.Controller)
	if len(candidates) < 3 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Aziza, Mage Tower Captain — tap three untapped creatures you control to copy that spell?",
		Cards:    candidates,
		Min:      0,
		Max:      3,
		ThenKey:  socB20AzizaTap,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

var socB20AzizaTap = askedPickRule("choose-cards/aziza-mage-tower-captain-tap-three",
	func(picked []game.Card) bool { return len(picked) == 3 },
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) != 3 {
			return nil
		}
		still := socB20UntappedCreaturesYouControl(ctx.Game, ctx.Controller())
		for _, id := range r.Picked {
			if !slices.Contains(still, id) {
				return nil
			}
		}
		for _, id := range r.Picked {
			if err := ctx.Game.TapTargetForEffect(id); err != nil {
				return err
			}
		}
		return CopySpell{
			StackID:          ctx.Trigger().Event.CardID,
			Controller:       ctx.Controller(),
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(ctx)
	})
