package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_wwk_b1_helpers.go — S82 Worldwake batch 1 (#5582): the Ally
// trigger, the per-kick enters trigger and the landfall spells'
// condition. Append-only.

// wwkThisOrAnotherAllyEntered is Worldwake's Ally trigger, "Whenever
// this creature or another Ally you control enters" (South Pole
// Voyager's condition): the source's own entry, or another Ally entering
// under its controller's control. "Ally" is read through the source's
// text, so a text-changing effect retypes the trigger.
func wwkThisOrAnotherAllyEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.CardID == source.InstanceID ||
		b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Ally"))
}

// wwkAllyMayCounter is the whole Ally row "Whenever this creature or
// another Ally you control enters, you may put a +1/+1 counter on this
// creature" (Hada Freeblade, Graypelt Hunter, Bojuka Brigand). The "you
// may" is asked as the trigger resolves (CR 603.5).
func wwkAllyMayCounter(name string) game.TriggeredAbility {
	return Optional(On(game.EventETB, wwkThisOrAnotherAllyEntered,
		name+" — put a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		name+" — put a +1/+1 counter on this creature?")
}

// wwkWhenEntersPerKick is "When this creature enters, <effect> for each
// time it was kicked" (Lightkeeper of Emeria, Deathforge Shaman,
// Bloodhusk Ritualist). Stronghold Arena's shape: `per` times the
// number of kicks is read off the record the resolution carried onto
// the permanent (CR 400.7d, game.CardKickedTimes) as the trigger is put
// on the stack, and stamped on item.Params.Amount, so the effect sees it
// even when the creature has left by then. `targets` may be nil.
func wwkWhenEntersPerKick(label string, per int, targets *game.TargetSpec, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Targets:   targets,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = per * game.CardKickedTimes(*source)
			return item
		},
		Effect: effect,
	}
}

// wwkHadLandfallThisTurn is the Worldwake landfall spells' "if you had a
// land enter the battlefield under your control this turn" (Tomb Hex,
// Rest for the Weary, Mysteries of the Deep): the controller's turn
// tally of lands that entered under them, which still counts a land
// that has since left or changed control (PlayerTurnTally.LandsEntered).
func wwkHadLandfallThisTurn(ctx *Context) bool {
	return ctx.Game.TurnTally.Players[ctx.Controller()].LandsEntered > 0
}
