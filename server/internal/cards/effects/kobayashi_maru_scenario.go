package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// kobayashiMaruLabel is the trigger's stack label, and with it the key
// its "hasn't been chosen" memory is kept under.
const kobayashiMaruLabel = "Kobayashi Maru Scenario — choose one that hasn't been chosen"

// Kobayashi Maru Scenario — Enchantment {1}{U}{U} (Star Trek, #4973):
//
//	"At the beginning of your upkeep, choose one that hasn't been
//	 chosen —
//	 • Hail — Return up to one target creature an opponent controls
//	   with mana value 3 or less to its owner's hand.
//	 • Sensors — Draw two cards.
//	 • Transport — Create two 1/1 red Officer creature tokens.
//	 • Abandon Ship — You lose the game."
//
// Demonic Pact's shape (ADR 0097): ChooseOneNotChosen keeps the memory
// on this object, so the fourth upkeep offers only Abandon Ship and the
// mandatory trigger must take it; a second Scenario, or this one after
// it leaves and returns, is a new object that remembers nothing
// (CR 400.7). The memory follows the enchantment, not its controller.
// "Hail" is "up to one" target, so the mode may be chosen
// with no creature to name. The bullets' leading words are flavour, with
// no rules meaning (CR 207.2d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b7ac3e8-0a7d-4b8a-a6a6-4cacc9b4f9de",
		Name:         "Kobayashi Maru Scenario",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			kobayashiMaruTrigger(),
		},
	})
}

func kobayashiMaruTrigger() game.TriggeredAbility {
	hail := TargetCreature("up to one target creature an opponent controls with mana value 3 or less",
		OpponentControls(), ManaValueLE(3))
	hail.Min = 0
	t := AtYourUpkeep(kobayashiMaruLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Hail — Return up to one target creature an opponent controls with mana value 3 or less to its owner's hand.",
			hail, BounceTheModesTarget),
		ModeWithPurpose(ModeDoing("Sensors — Draw two cards.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
			}), game.Purpose{Draws: 2}),
		ModeWithPurpose(ModeDoing("Transport — Create two 1/1 red Officer creature tokens.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: TokenCard("1/1 red Officer"), N: 2}.Apply(ctx)
			}), game.Purpose{Tokens: 2}),
		ModeDoing("Abandon Ship — You lose the game.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return LoseTheGame{Player: item.Controller}.Apply(ctx)
			}),
	)
	return t
}
