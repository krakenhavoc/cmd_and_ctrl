package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Riku of Many Paths — Legendary Creature — Human Wizard {G}{U}{R}, 3/3
// (Outlaws of Thunder Junction, #4468):
//
//	"Whenever you cast a modal spell, choose up to X, where X is the
//	 number of times you chose a mode for that spell —
//	 • Exile the top card of your library. Until the end of your next
//	   turn, you may play it.
//	 • Put a +1/+1 counter on Riku. It gains trample until end of turn.
//	 • Create a 1/1 blue Bird creature token with flying."
//
// A modal spell is one whose controller chose modes for it as it was
// cast (CR 700.2, 700.2a), and every modal spell has at least one chosen,
// so "modal" is "the cast spell's stack item carries chosen modes"
// (StackItem.Modes). A spree spell is modal and is covered too (the
// first ruling); a modal double-faced card is not a modal spell and
// carries none.
//
// X is the number of times a mode was chosen for that spell, repeats
// included (the third ruling: a mode chosen three times is 3), which is
// len(StackItem.Modes). It is read off the triggering spell as the
// trigger is put on the stack, when its own modes are chosen
// (CR 603.3c), through game.RegisterModeCountOfEvent (#4468) - the
// counted "choose up to X" Bumi, King of Three Trials uses, given the
// event. X is capped at the three bullets, none chosen twice (the fourth
// ruling, CR 700.2d).
//
// The bullets run in printed order (CR 608.2c): the impulse exile
// (b20ExileTopUntilEndOfNextTurn: "play", so a land can be played,
// following the normal timing rules, the fifth ruling); the counter and
// trample only while Riku is still the object that triggered
// (CounterOnThis); and the Bird.
//
// No simplification.
func init() {
	t := On(game.EventCast, rikuYouCastAModalSpell, "Riku of Many Paths — choose up to X",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseUpToCount(rikuModesChosen,
		ModeDoing("Exile the top card of your library. Until the end of your next turn, you may play it.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return b20ExileTopUntilEndOfNextTurn(ctx.Game, item)
			}),
		ModeDoing("Put a +1/+1 counter on Riku. It gains trample until end of turn.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				src := ctx.Source()
				if !onBattlefield(ctx.Game, src) {
					return nil
				}
				if err := (CounterOnThis{Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{
					Target:   src,
					Keywords: []string{"trample"},
					Label:    "Riku of Many Paths — trample",
				}.Apply(ctx)
			}),
		ModeDoing("Create a 1/1 blue Bird creature token with flying.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 blue Bird with flying"), N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "1cfe6e3a-f8c9-4105-ad03-0faf1fcc4626",
		Name:         "Riku of Many Paths",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// rikuModesChosen is X: how many times a mode was chosen for the spell
// whose cast fired the trigger. An on-disk-stable registry key.
var rikuModesChosen = game.RegisterModeCountOfEvent("riku-of-many-paths/modes-chosen-for-that-spell",
	func(g *game.Game, _ uuid.UUID, ev game.Event) int {
		return rikuModesChosenFor(g, ev)
	})

// rikuModesChosenFor is the number of modes chosen for the spell an
// EventCast names, 0 when it is not on the stack or is not modal.
func rikuModesChosenFor(g *game.Game, ev game.Event) int {
	if ev.Kind != game.EventCast || ev.CardID == uuid.Nil {
		return 0
	}
	item := g.StackItemForEffect(ev.CardID)
	if item == nil {
		return 0
	}
	return len(item.Modes)
}

// rikuYouCastAModalSpell is "whenever you cast a modal spell".
func rikuYouCastAModalSpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	return rikuModesChosenFor(g, ev) > 0
}
