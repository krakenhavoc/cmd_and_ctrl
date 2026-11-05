package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enter the Avatar State — Instant — Lesson {W} (Avatar: The Last
// Airbender, #4459):
//
//	"Until end of turn, target creature you control becomes an Avatar in
//	 addition to its other types and gains flying, first strike,
//	 lifelink, and hexproof."
//
// Stone by Sunlight's two grants on one target for one duration: a
// layer-4 subtype add (the type a word of the spell's text, CR 612.2)
// and the four keywords (layer 6). A target gone by resolution makes the
// spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e552be1-0702-4e18-822f-98611f8a8e70",
		Name:         "Enter the Avatar State",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "flying", "first strike", "lifelink", "hexproof")),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := legalTargetCards(item, ctx.Game)
			if len(targets) == 0 {
				return nil
			}
			if err := (ScopedEffectFor{
				Target:   targets[0],
				Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Avatar")...)},
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Enter the Avatar State — an Avatar in addition to its other types",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   targets[0],
				Keywords: []string{"flying", "first strike", "lifelink", "hexproof"},
				Label:    "Enter the Avatar State — flying, first strike, lifelink and hexproof until end of turn",
			}.Apply(ctx)
		},
	})
}
