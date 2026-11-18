package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:          "3570001c-844f-4319-9850-ab997d62a83b",
		Name:              "Visions of Ruin",
		Completeness:      CompletenessFull,
		CastableZones:     []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts:  []game.AlternativeCost{visionsFlashbackAlternative("{8}{R}{R}")},
		SelfCostModifiers: []game.CostModifier{visionsFlashbackDiscount()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				ExceptController: true,
				Match:            Artifact(),
				Label:            "an artifact",
				Keyed:            visionsOfRuinEdicted,
			}.Apply(ctx)
		},
	})
}

// visionsOfRuinEdicted is "for each artifact sacrificed this way, you create
// a Treasure token", keyed (#4290).
var visionsOfRuinEdicted = RegisterEdictThen("prompt-run/visions-of-ruin-treasures", func(ctx *Context, sacrificed game.PromptedSacrifices) error {
	n := sacrificed.Count()
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TreasureToken(), N: n}.Apply(ctx)
})
