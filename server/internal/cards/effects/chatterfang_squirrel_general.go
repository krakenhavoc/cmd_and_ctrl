package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chatterfang, Squirrel General — Legendary Creature — Squirrel Warrior
// {2}{G}, 3/3 (Bloomburrow Commander):
//
//	"Forestwalk
//	 If one or more tokens would be created under your control, those
//	 tokens plus that many 1/1 green Squirrel creature tokens are
//	 created instead.
//	 {B}, Sacrifice X Squirrels: Target creature gets +X/-X until end of
//	 turn."
//
// The replacement is Xorn's additive group (CR 614.1, 701.7b) on the
// creation instruction: the instruction's own groups are left alone and
// one more group of Squirrels is appended, as many as the instruction
// creates across every group ("that many"). Two Chatterfangs each add
// their own. The activated ability is Grim Hireling's SacrificeX with
// the X announced with the activation; its effect reads ctx.X().
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f15b3b76-d38a-48db-bb44-3296183c8641",
		Name:            "Chatterfang, Squirrel General",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
		XMatters:        true,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller && ev.TokenCount() > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{
					Template: TokenCard("1/1 green Squirrel"), Count: ev.TokenCount(),
				})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Chatterfang: those tokens plus that many Squirrels",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice X Squirrels: Target creature gets +X/-X until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, 0, 0)),
			Cost:    Plus(ManaCost("{B}"), SacrificeX("X Squirrels", HasSubtype("Squirrel"))),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return BoostUntilEOT{
							Target: t.ID, Power: x, Toughness: -x,
							Label: "Chatterfang — +X/-X",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
