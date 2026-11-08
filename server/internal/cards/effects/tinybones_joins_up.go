package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tinybones Joins Up — Legendary Enchantment {B} (Outlaws of Thunder
// Junction, #4468):
//
//	"When Tinybones Joins Up enters, any number of target players each
//	 discard a card.
//	 Whenever a legendary creature you control enters, any number of
//	 target players each mill a card and lose 1 life."
//
// "Any number of target players" is Court of Cunning's (0, 0) count: none
// is a legal choice, and you may pick yourself. Each pick is re-checked as
// the ability resolves (CR 608.2b).
//
//   - The entry discard is Ral Zarek, Guest Lecturer's −1 (ralZarekDiscard):
//     one ordinary discard prompt per chosen player, each player choosing
//     their own card, and an empty hand discards nothing.
//   - The legendary trigger reads the entering permanent's effective type
//     line, so a creature made legendary counts (Hero's Blade's reading).
//     The mills run in APNAP order (millEachPlayerInTurn), then each chosen
//     player loses 1 life, as printed: mill, then lose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d2bd4f6-174d-411a-b930-9161aeefd33a",
		Name:         "Tinybones Joins Up",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Tinybones Joins Up — any number of target players each discard a card", ralZarekDiscard),
				TargetPlayer("any number of target players").WithCount(0, 0)),
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsCreature() && c.IsLegendary()
				}, "Tinybones Joins Up — any number of target players each mill a card and lose 1 life", tinybonesJoinsUpMill),
				TargetPlayer("any number of target players").WithCount(0, 0)),
		},
	})
}

// tinybonesJoinsUpMill mills each chosen player one card, then the tail
// takes 1 life from each.
func tinybonesJoinsUpMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return millEachPlayerInTurn(ctx, kozilekTargetPlayers(ctx), constantMill(1), ctx.ItemTail(tinybonesJoinsUpLoseLife))
}

// tinybonesJoinsUpLoseLife is "and lose 1 life" after the mills: each
// chosen player still legal loses 1 life. An on-disk key.
var tinybonesJoinsUpLoseLife = CardTail("then/tinybones-joins-up-lose-life", func(ctx *Context, _ game.TailOutcome) error {
	for _, p := range kozilekTargetPlayers(ctx) {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p, -1); err != nil {
			return err
		}
	}
	return nil
})
