package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82MatB01ExileGreatestEdict is "Target opponent exiles a creature or
// planeswalker they control with the greatest mana value among creatures
// and planeswalkers they control" as a spell's whole resolution (End of
// the Hunt, Blot Out): slot 0 is the targeted opponent, who is asked,
// choosing among the tied greatest.
func S82MatB01ExileGreatestEdict(label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		t, ok := FirstLegalTarget(ctx)
		if !ok || t.Kind != game.TargetPlayer {
			return nil
		}
		return Msc17ExileEdict(ctx, t.ID, label,
			B10GreatestManaValueCreatureOrPlaneswalkerYouControl(), DoomfallExile)
	}
}
