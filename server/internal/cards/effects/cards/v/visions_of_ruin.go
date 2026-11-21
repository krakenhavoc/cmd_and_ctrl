package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Visions of Ruin — Sorcery {3}{R} (Midnight Hunt Commander):
//
//	"Each opponent sacrifices an artifact. For each artifact sacrificed this
//	 way, you create a Treasure token.
//	 Flashback {8}{R}{R}. This spell costs {X} less to cast this way, where X
//	 is the greatest mana value of a commander you own on the battlefield or
//	 in the command zone."
//
// An edict (CR 701.21, EachPlayerSacrifices): each opponent chooses their
// own artifact, and the Treasures are made once every chosen artifact has
// gone, one per artifact that really left. A commander that took the
// command zone instead was not sacrificed this way. The flashback discount
// is the Visions cycle's (visionsFlashbackDiscount).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "3570001c-844f-4319-9850-ab997d62a83b",
		Name:              "Visions of Ruin",
		Completeness:      effects.CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{effects.VisionsFlashbackAlternative("{8}{R}{R}")},
		SelfCostModifiers: []game.CostModifier{effects.VisionsFlashbackDiscount()},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Artifact(),
				Label:            "an artifact",
				Keyed:            visionsOfRuinEdicted,
			}.Apply(ctx)
		},
	})
}

// visionsOfRuinEdicted is "for each artifact sacrificed this way, you create
// a Treasure token", keyed (#4290).
var visionsOfRuinEdicted = effects.RegisterEdictThen("prompt-run/visions-of-ruin-treasures", func(ctx *effects.Context, sacrificed game.PromptedSacrifices) error {
	n := sacrificed.Count()
	if n == 0 {
		return nil
	}
	return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(ctx)
})
