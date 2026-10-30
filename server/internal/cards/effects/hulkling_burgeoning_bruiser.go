package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hulkling, Burgeoning Bruiser — {2}{G} Legendary Creature — Kree Skrull
// Hero 2/3 (Marvel Super Heroes, #4455):
//
//	"Vigilance
//	 Whenever another creature you control enters, if it has greater
//	 power or toughness than Hulkling, put a +1/+1 counter on Hulkling."
//
// This is evolve's condition written out as its own ability, not the evolve
// keyword (Hulkling does not have evolve). The comparison is an intervening
// "if" (CR 603.4), so it is made when the creature enters and again as the
// trigger resolves, with the entered creature read as it last was
// (CR 603.10) and Hulkling read current: a pump or a counter in response
// changes the answer. Power is compared with power and toughness with
// toughness.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "03c934a6-be57-480d-8c84-4e46d9ccb406",
		Name:            "Hulkling, Burgeoning Bruiser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "the entered creature has greater power or toughness than Hulkling",
				Test: hulklingStillGreater,
			}, On(game.EventETB, hulklingEnteredGreater,
				"Hulkling — put a +1/+1 counter on Hulkling",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

// hulklingEnteredGreater is the trigger condition: another creature you
// control entered with greater power or toughness than Hulkling has now.
func hulklingEnteredGreater(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	entered, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !entered.IsCreature() {
		return false
	}
	return entered.PowerForComparison() > source.PowerForComparison() ||
		entered.CurrentToughness() > source.CurrentToughness()
}

// hulklingStillGreater is the same comparison made as the trigger
// resolves (CR 603.4): the entered creature as it last was, Hulkling as
// he is now.
func hulklingStillGreater(ctx *Context) bool {
	item := ctx.Item
	if item == nil || item.Trigger == nil || item.Trigger.Object == nil {
		return false
	}
	entered, ok := ctx.Game.PermanentForEffect(item.Trigger.Object.Ref())
	self, selfOK := ctx.Game.LookupCardForEffect(item.SourceCardID)
	if !ok || !selfOK || !slices.Contains(entered.Characteristic.Types, "Creature") {
		return false
	}
	return entered.Power > self.PowerForComparison() || entered.Toughness > self.CurrentToughness()
}
