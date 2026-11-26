package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ever After — {4}{B}{B} Sorcery (Commander 2020):
//
//	"Return up to two target creature cards from your graveyard to the
//	 battlefield. Each of those creatures is a black Zombie in addition to its
//	 other colors and types. Put Ever After on the bottom of its owner's
//	 library."
//
// Grimoire of the Dead's return-together with an indefinite black-Zombie
// effect pinned to each returned object (CR 400.7). The spell goes to the
// bottom of its owner's library as the last instruction, whether or not
// anything came back (Spell Crumple's PutIntoLibrary); a copy of it is not a
// card and goes nowhere.
func init() {
	everAfterZombies = effects.TogetherTail("ever-after/black-zombies-then-library",
		func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
			for _, id := range entered {
				if err := (effects.ScopedEffectFor{
					Target: id,
					Mods: []game.Mod{
						game.AddSubtypesMod(ctx.TextSubtypes("Zombie")...), game.AddColorsMod(ctx.TextColors("B")...),
					},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Ever After — a black Zombie in addition to its other colors and types",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return everAfterToLibrary(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "86be501c-6f05-4ce0-adb0-7adb135bfb93",
		Name:         "Ever After",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.Creature(), effects.YouOwn()).WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return everAfterToLibrary(ctx)
			}
			return effects.ReturnFromGraveyardTogether{Targets: ids, Controller: item.Controller, ThenTail: everAfterZombies}.Apply(ctx)
		},
	})
}

var everAfterZombies game.TailRef

func everAfterToLibrary(ctx *effects.Context) error {
	if ctx.Item != nil && ctx.Item.IsCopy {
		return nil
	}
	return effects.PutIntoLibrary{Card: ctx.Source(), ToBottom: true}.Apply(ctx)
}
