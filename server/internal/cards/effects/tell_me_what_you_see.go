package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tell Me What You See — Enchantment {1}{B} (Star Trek, #4973):
//
//	"When this enchantment enters, target opponent may have you draw a
//	 card. If they don't, they lose 4 life and you surveil 1.
//	 Whenever you face a dilemma, you may return this enchantment to its
//	 owner's hand. (You face a dilemma as you choose one or more modes for
//	 a spell or ability.)"
//
// The enters trigger is Combustible Gearhulk's shape (#568, #796): the
// question belongs to the target opponent and is asked as the trigger
// resolves. "Yes" is you drawing a card; "no" is that opponent losing 4
// life (a life change, no damage) and then you surveilling 1. A target
// that is gone by resolution is asked nothing and the trigger does
// nothing (CR 608.2b), not the "don't" branch.
//
// The dilemma trigger (EventFacedDilemma, batch 7) is a CR 603.5 "you
// may" asked as it resolves; "yes" returns this enchantment to its
// owner's hand, unless it has left the battlefield or become a new
// object (CR 400.7) in the meantime.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "abfdfae5-61aa-42d6-9c8e-83fda59de233",
		Name:         "Tell Me What You See",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Tell Me What You See — target opponent may have you draw a card", tellMeWhatYouSeeAsk),
				TargetPlayer("target opponent", Opponent()),
			),
			Optional(On(game.EventFacedDilemma, YouFacedADilemma,
				"Tell Me What You See — you may return it to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
				}),
				"Return Tell Me What You See to its owner's hand?"),
		},
	})
}

// tellMeWhatYouSeeAsk asks the target opponent, if it is still a legal
// target.
func tellMeWhatYouSeeAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return MayChoice{
			Player:   t.ID,
			Question: "Tell Me What You See — let its controller draw a card? If you don't, you lose 4 life and they surveil 1.",
			Draws:    1, Drawer: ctx.Controller(),
			YesLabel: "They draw a card",
			NoLabel:  "Lose 4 life; they surveil 1",
			Then:     tellMeWhatYouSeeAnswer,
		}.Apply(ctx)
	}
	return nil
}

// tellMeWhatYouSeeAnswer is the opponent's answer. Keyed: the key is an
// on-disk identity, never renamed or reused.
var tellMeWhatYouSeeAnswer = MayYesNo("confirm/tell-me-what-you-see-draw-or-lose",
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	},
	func(ctx *Context, r game.ConfirmAnswer) error {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -4); err != nil {
			return err
		}
		return Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
	})
