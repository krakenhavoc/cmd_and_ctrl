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
// may" is the trigger's optional prompt; the land is the bounce-land
// choice (b04BounceLand's shape — the pick_target prompt is the one
// picker the engine has for a choice among your own permanents, and a
// land you control can never be an illegal pick); the search is the
// S22 chooser with Harrow's "up to two" and Cultivate's tapped flag.
//
// The ordering is the printed one, and it is why the land is chosen
// as a trigger target rather than through a sacrifice prompt queued
// at resolution: a sacrifice prompt has no continuation, so the
// search would have to be queued alongside it, and with a small
// library the search resolves synchronously — the basics would enter
// BEFORE the sacrifice, and the prompt would offer them as the land
// to sacrifice. Choosing at trigger time keeps sacrifice-then-search.
//
// Sandbox simplification, weaker than printed: the land is chosen
// when the trigger goes on the stack, not on resolution, so an
// opponent who removes that land in response fizzles the trigger
// (printed, you would pick another). With no land at all the trigger
// is removed with no prompt (CR 603.3d), which matches "you may
// sacrifice a land" with nothing to sacrifice.
func init() {
	Register(Spec{
		OracleID:     "788a4896-3414-40bf-b391-d2efea2fb5f9",
		Name:         "Springbloom Druid",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You pick the land to sacrifice when the trigger goes on the stack rather than on resolution, so opponents can respond to the choice."},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Springbloom Druid — sacrifice a land to search for up to two basic lands?"},
			Targets:        TargetPermanent("a land you control", And(Land(), YouControl())),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Springbloom Druid — sacrifice a land, then fetch up to two basic lands tapped",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						ctx := NewContext(g, item)
						if !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						if err := (SacrificePermanent{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
							return err
						}
						return SearchLibrary{
							Player:        item.Controller,
							Predicate:     IsBasicLand,
							Dest:          game.ZoneBattlefield,
							Limit:         2,
							Shuffle:       true,
							TappedOnEntry: true,
							Reason:        "Springbloom Druid — up to two basic land cards, tapped",
						}.Apply(ctx)
					})
			},
		}},
	})
}
