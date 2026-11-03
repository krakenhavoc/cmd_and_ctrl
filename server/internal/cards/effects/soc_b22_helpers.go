package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b22_helpers.go — shared pieces for the twenty-second Secrets of
// Strixhaven batch (#4456).

// socB22EachPlayerLosesOneLife is "each player loses 1 life": the caster
// first, then each opponent in seat order, each a life change that names the
// item's source and no damage (Smallpox, Conciliator's Duelist).
func socB22EachPlayerLosesOneLife(ctx *Context) error {
	for _, id := range tablePlayers(ctx) {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -1); err != nil {
			return err
		}
	}
	return nil
}

// socB22FlickerFirstTargetWithACounter is "exile target <permanent>, then
// return that card to the battlefield under its owner's control with a
// +1/+1 counter on it" (Spaceshift, Daydream): the first target, if it is
// still legal as the spell resolves (CR 608.2b), blinks at once and returns
// with the counter already on it.
func socB22FlickerFirstTargetWithACounter(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
		return nil
	}
	return Flicker{Target: item.Targets[0].ID, PlusOneCounters: 1}.Apply(ctx)
}
