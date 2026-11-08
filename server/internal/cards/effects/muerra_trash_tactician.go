package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muerra, Trash Tactician — Legendary Creature — Raccoon Warrior
// {1}{R}{G}, 2/4 (Bloomburrow, #4467):
//
//	"At the beginning of your first main phase, add {R} or {G} for each
//	 Raccoon you control.
//	 Whenever you expend 4, you gain 3 life.
//	 Whenever you expend 8, exile the top two cards of your library.
//	 Until the end of your next turn, you may play those cards."
//
// The main-phase trigger is Glóin the Mighty's, with the amount counted
// as it resolves: one mana per Raccoon you control then, Muerra
// included, each one {R} or {G} as you choose (a per-slot pick, so any
// mix). The two expend triggers are the shared expend vocabulary
// (expend.go) and Crimson Operative's impulse exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c115159c-6991-4812-b3b6-6ed80221cab4",
		Name:           "Muerra, Trash Tactician",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered: []game.TriggeredAbility{
			AtYourPrecombatMain("Muerra, Trash Tactician — add {R} or {G} for each Raccoon you control", muerraRaccoonMana),
			WheneverYouExpend(4, "Muerra, Trash Tactician — you gain 3 life", Do(GainLife{Amount: 3})),
			WheneverYouExpend(8, "Muerra, Trash Tactician — exile the top two cards of your library; you may play them until the end of your next turn",
				Do(Instruction(func(ctx *Context) error { return ExileTopNUntilYourNextTurn(ctx, 2) }))),
		},
	})
}

func muerraRaccoonMana(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := len(MatchingBattlefield(ctx, And(YouControl(), Subtype("Raccoon"))))
	if n == 0 {
		return nil
	}
	return AddMana{Amount: ManaN(n, "R", "G")}.Apply(ctx)
}
