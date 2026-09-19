package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Springbloom Druid — Creature — Elf Druid {2}{G}, 1/1 (EDHREC rank
// 779):
//
//	"When this creature enters, you may sacrifice a land. If you do,
//	 search your library for up to two basic land cards, put them
//	 onto the battlefield tapped, then shuffle."
//
// Harrow on a body: net one land, two colours fixed, two landfall
// triggers, and a creature left over for the sacrifice deck. The "you
// may" is the trigger's optional prompt; the land is chosen AT
// RESOLUTION through the ordinary sacrifice prompt, over the lands the
// controller has when the trigger resolves; the search is the S22
// chooser with Harrow's "up to two" and Cultivate's tapped flag.
//
// THE LAND IS CHOSEN AT RESOLUTION (#1026). It used to be a trigger
// TARGET, and the comment here said why: a sacrifice prompt had no
// continuation, so the search would have had to be queued alongside
// it, and with a small library the search resolves synchronously — the
// basics would have entered BEFORE the sacrifice, and the prompt would
// have offered them as the land to sacrifice. #1019 gave the prompt a
// continuation, so the search is the run's continuation
// (PlayerSacrificesThenForEffect, ADR 0013 §5x) and the ordering falls
// out. Planar Engineering is the same "sacrifice, THEN search" shape
// and was migrated with the mechanic.
//
// Two things that changes about the card, both towards printed:
//
//   - WHICH land can be sacrificed. The choice is made when the
//     ability resolves, over the board as it then is, so a land played
//     in response is a legal answer and a land removed in response no
//     longer fizzles anything. There is no Targets clause any more, so
//     nothing is announced and an opponent cannot respond to the pick.
//   - "IF YOU DO" is the run's answer. A controller who says yes to
//     the trigger and then has no land at all when it resolves — every
//     one of them removed in response — is asked nothing and searches
//     for nothing, which is what "you may sacrifice a land. If you do"
//     says. The old shape fizzled the whole trigger instead (CR 603.3d,
//     its only target gone), which reached the same board by a
//     different rule and only because the trigger did nothing else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "788a4896-3414-40bf-b391-d2efea2fb5f9",
		Name:         "Springbloom Druid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Springbloom Druid — sacrifice a land to search for up to two basic lands?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Springbloom Druid — sacrifice a land, then fetch up to two basic lands tapped",
					springbloomDruidSacrificeThenSearch)
			},
		}},
	})
}

// springbloomDruidSacrificeThenSearch is the ability's body: one
// sacrifice prompt over the controller's lands, and the search as its
// continuation.
//
// The context is rebuilt from the live *Game inside the continuation
// rather than captured, the contract every continuation in the tree
// signs: an undo restores this game's fields in place, so a captured
// *Game would be the wrong one.
//
// Caller holds g.mu.
func springbloomDruidSacrificeThenSearch(g *game.Game, item *game.StackItem) error {
	return g.PlayerSacrificesThenForEffect(item.SourceCardID, item.Controller,
		sacrificeSpec("a land you control", Land()),
		"Springbloom Druid — sacrifice a land", 1,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			// "If you do" — CR 701.17a's sacrifice really happened for
			// this seat. A prompt withdrawn because the land left
			// under it, or a CR 614 window that cancelled the move,
			// searches for nothing.
			if !sacrificed.Sacrificed(item.Controller) {
				return nil
			}
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Springbloom Druid — up to two basic land cards, tapped",
			}.Apply(NewContext(g, item))
		})
}
