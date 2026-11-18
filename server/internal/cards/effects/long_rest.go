package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Long Rest — Sorcery {X}{G}{G}{G} (Adventures in the Forgotten Realms):
//
//	"Return X target cards with different mana values from your
//	 graveyard to your hand. If eight or more cards were returned to your
//	 hand this way, your life total becomes equal to your starting life
//	 total. Exile Long Rest."
//
// Exactly X targets (CR 107.3a) with a set rule that no two share a mana
// value (Agadeem's Awakening's EachDifferentManaValue, CR 601.2c, judged
// again at resolution). A card that left the graveyard in response is not
// returned and does not count (CR 608.2b); the count is the cards that
// actually reached the hand. "Becomes equal" is a gain or a loss of the
// difference (CR 119.5), so life-gain and life-loss replacements see it.
// The last sentence exiles the spell itself, so it never reaches the
// graveyard.
//
// No simplification.
func init() {
	targets := TargetCardInGraveyard("X target cards with different mana values from your graveyard", YouOwn()).
		EachDifferent(EachDifferentManaValue())
	targets.CountFromX = true
	Register(Spec{
		OracleID:     "fc8df5c8-fda6-454c-a1ac-21c6e3d75b9c",
		Name:         "Long Rest",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      targets,
		OnResolve: ResolveSteps(
			Instruction(longRestReturn),
			Instruction(func(ctx *Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

func longRestReturn(ctx *Context) error {
	returned := 0
	for _, id := range legalTargetCardIDs(ctx) {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			returned++
		}
	}
	if returned < 8 {
		return nil
	}
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil {
		return nil
	}
	delta := startingLifeOf(ctx.Game) - p.Life
	if delta == 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), delta)
}
