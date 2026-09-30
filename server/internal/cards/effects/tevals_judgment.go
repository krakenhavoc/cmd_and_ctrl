package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// b16TevalsJudgmentLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen this turn" memory is kept under.
const b16TevalsJudgmentLabel = "Teval's Judgment — cards left your graveyard"

// Teval's Judgment — Enchantment {2}{B} (EDHREC rank 1773):
//
//	"Whenever one or more cards leave your graveyard, choose one that
//	 hasn't been chosen this turn —
//	 • Draw a card.
//	 • Create a Treasure token.
//	 • Create a 2/2 black Zombie Druid creature token."
//
// The graveyard-recursion payoff. The condition is
// b16CardLeftYourGraveyard — a move out of a graveyard for a card
// the controller owns, whether an EventZoneMove or a flashback's
// EventCast — and "one or more" is the per-label dedup
// (OncePerBatch): a Bojuka Bog on your graveyard emits
// one move per card and fires this once.
//
// ADR 0097 (#1749): the mode is the controller's choice, made through
// the trigger's mode_pick prompt as it goes on the stack (CR 603.3c),
// and "that hasn't been chosen this turn" is the engine's memory of
// what this enchantment's ability has chosen, recorded at the choice.
// The old shape took the modes in printed order by counting this
// trigger's RESOLUTIONS, which was the wrong event twice over: a
// countered trigger had not used its mode, and two triggers on the
// stack at once could both read the same count. Once all three are
// used, a later trigger that turn is removed with no effect (the
// Breeches ruling's reading of CR 700.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71fc2393-f55c-4b06-897c-fb7d4199b5f5",
		Name:         "Teval's Judgment",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			tevalsJudgmentTrigger(),
		},
	})
}

func tevalsJudgmentTrigger() game.TriggeredAbility {
	t := OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b16CardLeftYourGraveyard(ev, source, g)
	}, b16TevalsJudgmentLabel, func(g *game.Game, item *game.StackItem) error { return nil }))
	t.Modes = ChooseOneNotChosenThisTurn(
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
		ModeDoing("Create a Treasure token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
			}),
		ModeDoing("Create a 2/2 black Zombie Druid creature token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 black Zombie Druid"), N: 1}.Apply(ctx)
			}),
	)
	return t
}
