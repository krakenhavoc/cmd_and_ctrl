package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// LosesAbilitiesAndBasePTUntilEOT is "Until end of turn, target creature loses
// all abilities and has base power and toughness P/T" (Ovinize, Humble): one
// continuous effect on the first legal target, layer 6 removing the abilities and
// layer 7b setting the base values (CR 613.1), so counters and later pumps still
// add. A target gone by resolution does nothing (CR 608.2b).
func LosesAbilitiesAndBasePTUntilEOT(ctx *Context, name string, power, toughness int) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	mods := append([]game.Mod{game.LoseAllAbilitiesMod()}, game.SetBasePTMods(power, toughness)...)
	return UntilEndOfTurn(ctx, id, nil, fmt.Sprintf("%s — loses all abilities, base power and toughness %d/%d", name, power, toughness), mods...)
}
