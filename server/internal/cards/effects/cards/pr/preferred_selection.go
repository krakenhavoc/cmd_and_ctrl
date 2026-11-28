package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Preferred Selection — Enchantment {2}{G}{G} (Mirage, #5633):
//
//	"At the beginning of your upkeep, look at the top two cards of your
//	 library. You may sacrifice this enchantment and pay {2}{G}{G}. If you
//	 do, put one of those cards into your hand. If you don't, put one of
//	 those cards on the bottom of your library."
//
// The look is at the top two (CR 701.20 sense: private to you). With the
// enchantment still on the battlefield you are asked to pay {2}{G}{G} (a
// payment you cannot fund counts as declining); paying sacrifices the
// enchantment and you take one of the two into your hand, the other staying
// on top. Declining, or an enchantment that is gone, puts one of the two on
// the bottom. The two cards ride the prompt as plain data, so a table waiting
// on any answer is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d05d2d6-3ea1-4d8a-9f1c-6fc3f5e38ea9",
		Name:         "Preferred Selection",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Preferred Selection — look at the top two cards of your library", oldB813PreferredSelectionLook),
		},
	})
}

func oldB813PreferredSelectionLook(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Controller()
	cards := g.LookAtTopOfLibraryForEffect(player, 2)
	if len(cards) == 0 {
		return nil
	}
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return OldB813AskForBottom(g, player, item.SourceCardID, cards)
	}
	return effects.MayPay{
		Chooser:  player,
		Cost:     "{2}{G}{G}",
		Question: "Preferred Selection — sacrifice it and pay {2}{G}{G} to put one of the top two cards into your hand?",
		Then:     oldB813PreferredSelectionBranches,
		Carry:    cards,
	}.Apply(ctx)
}

// oldB813PreferredSelectionBranches is the answer: paid, the enchantment is
// sacrificed and one card goes to hand; declined, one goes to the bottom. An
// on-disk key: never renamed, never reused.
var oldB813PreferredSelectionBranches = effects.PayUnlessBranches("pay-unless/old-b813-preferred-selection",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		cards := r.Carry
		if !effects.OnBattlefield(ctx.Game, ctx.Source()) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: effects.WithCarry(ctx.ItemTail(oldB813PreferredSelectionTake), cards...)}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return OldB813AskForBottom(ctx.Game, r.Chooser, ctx.Source(), r.Carry)
	})

// oldB813PreferredSelectionTake is "put one of those cards into your hand"
// once the enchantment has been sacrificed.
var oldB813PreferredSelectionTake = effects.CardTail("then/old-b813-preferred-selection-take", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.TakeFromLibraryToHand{
		Player: ctx.Controller(),
		Cards:  o.Tail.Carry,
		Max:    1,
		Label:  "Preferred Selection — put one of those cards into your hand",
		Key:    oldB813PreferredSelectionHand,
	}.Apply(ctx)
})

// oldB813PreferredSelectionHand is the take, with nothing after it: the other
// card stays where it was.
var oldB813PreferredSelectionHand = effects.RegisterTake("choose-cards/old-b813-preferred-selection-hand", effects.TakeDef{})

// oldB813PreferredSelectionBottom is the pick for the bottom of the library.
var oldB813PreferredSelectionBottom = game.RegisterChooseCardsThen("choose-cards/old-b813-preferred-selection-bottom", func(g *game.Game, r game.CardsChosen) error {
	picked := effects.CardsStillInALibrary(g, r.Picked)
	if len(picked) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   r.Chooser,
		Source:    r.Source,
		Cards:     picked,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Preferred Selection — put it on the bottom of your library",
	})
})

// OldB813AskForBottom asks `player` for one of `cards` (still in their
// library) to put on the bottom of it.
func OldB813AskForBottom(g *game.Game, player, source uuid.UUID, cards []uuid.UUID) error {
	cards = effects.CardsStillInALibrary(g, cards)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: "Preferred Selection — put one of those cards on the bottom of your library",
		Cards:    cards,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  oldB813PreferredSelectionBottom,
	})
	return nil
}
