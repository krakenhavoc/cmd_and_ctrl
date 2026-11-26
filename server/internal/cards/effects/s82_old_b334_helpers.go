package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// DanceOfTheSkywiseBecomes is Dance of the Skywise's record: "Until end of
// turn, <target> becomes a blue Dragon Illusion with base power and toughness
// 4/4, loses all abilities, and gains flying."
func DanceOfTheSkywiseBecomes(ctx *Context, target uuid.UUID) error {
	mods := []game.Mod{
		game.SetColorsMod(ctx.TextColor("U")),
		game.SetCreatureTypesMod(ctx.TextSubtype("Dragon"), ctx.TextSubtype("Illusion")),
		game.SetBasePowerMod(4), game.SetBaseToughnessMod(4),
		game.LoseAllAbilitiesMod("flying"),
	}
	return UntilEndOfTurn(ctx, target, nil, "Dance of the Skywise — becomes a blue Dragon Illusion 4/4 with flying and no other abilities", mods...)
}
