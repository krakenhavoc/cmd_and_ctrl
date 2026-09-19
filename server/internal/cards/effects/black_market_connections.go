package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Market Connections — Enchantment {2}{B}:
//
//	"At the beginning of your first main phase, choose one or more —
//	 • Sell Contraband — Create a Treasure token. You lose 1 life.
//	 • Buy Information — Draw a card. You lose 2 life.
//	 • Hire a Mercenary — Create a 3/2 colorless Shapeshifter
//	   creature token with changeling. You lose 3 life."
//
// The batch-01 first pass declared this one skipped, and named the
// two reasons exactly: "a beginning-of-main-phase trigger event (only
// upkeep and end step exist)" and "a resolution-time 'choose one or
// more' prompt for a trigger (the modal machinery is cast-time
// only)". Both are gone. `AtYourPrecombatMain` watches
// EventBeginPrecombatMain, and #764 made `TriggeredAbility.Modes` the
// same game.ModeSpec a modal spell declares — the choice is a
// mode_pick prompt to the controller as the ability goes on the
// stack (CR 603.3c), a full priority round before it resolves.
//
// "Choose one or more" is `ChooseOneOrMore`: min one, max all three,
// each chosen bullet running its own body in announce order. Taking
// all three is three life and is meant to be.
//
// "Your FIRST main phase" is the precombat main, so the trigger is
// once per turn and never fires in the postcombat main — the reason
// the constructor is AtYourPrecombatMain rather than a main-phase
// watch that would fire twice.
//
// The life is LOST, not damage: no prevention, no replacement, and
// nothing that watches for damage sees it. That matters for the card
// it is usually played beside.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2664f28-49e1-46f8-a863-b217e961a57c",
		Name:         "Black Market Connections",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			blackMarketConnectionsTrigger(),
		},
	})
}

// blackMarketConnectionsTrigger is the first-main trigger with its
// three bullets. Each declares its own body (ModeDoing), so the
// engine dispatches the chosen ones at resolution in announce order
// and the card file writes no switch.
func blackMarketConnectionsTrigger() game.TriggeredAbility {
	t := AtYourPrecombatMain("Black Market Connections — choose one or more",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneOrMore(
		ModeDoing("Sell Contraband — Create a Treasure token. You lose 1 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := (CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -1)
			}),
		ModeDoing("Buy Information — Draw a card. You lose 2 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := (DrawCards{Player: item.Controller, N: 1}.Apply(ctx)); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -2)
			}),
		ModeDoing("Hire a Mercenary — Create a 3/2 colorless Shapeshifter creature token with changeling. You lose 3 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if err := (CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("3/2 colorless Shapeshifter with changeling"),
					N:          1,
				}.Apply(ctx)); err != nil {
					return err
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -3)
			}),
	)
	return t
}
