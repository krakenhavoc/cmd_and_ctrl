package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82OldB561BecomesElemental is "target land becomes a 2/2 blue Elemental
// creature with flying. It's still a land." with no duration (Balduvian
// Frostwaker): it adds the creature type, the Elemental subtype, blue, flying
// and a base 2/2, lasting while the land stays the same object (CR 611.2a).
func S82OldB561BecomesElemental(ctx *Context, target uuid.UUID) error {
	mods := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
		game.SetColorsMod(ctx.TextColors("U")...),
		game.AddKeywordsMod("flying"),
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     append(mods, game.SetBasePTMods(2, 2)...),
		Duration: game.IndefiniteDuration(),
		Label:    "Balduvian Frostwaker — becomes a 2/2 blue Elemental creature with flying",
	}.Apply(ctx)
}
