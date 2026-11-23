package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Michiko's Reign of Truth // Portrait of Michiko — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493).
//
//	I, II — Target creature gets +1/+1 until end of turn for each
//	        artifact and/or enchantment you control.
//	III   — Exile this Saga, then return it to the battlefield
//	        transformed under your control.
//
// Portrait of Michiko, Enchantment Creature — Human Noble 0/0: "This
// creature gets +1/+1 for each artifact and/or enchantment you control."
//
// The pump is counted as the chapter resolves (CR 608.2h) and the
// number is then fixed; a permanent that is both an artifact and an
// enchantment counts once. The Portrait is itself an enchantment, so it
// counts itself: alone it is a 1/1.
func init() {
	chapter := func(n int) game.TriggeredAbility {
		label := effects.SagaChapterLabel("Michiko's Reign of Truth", n, "target creature gets +1/+1 until end of turn for each artifact and/or enchantment you control")
		return effects.ChapterTriggerTargeting(n, label,
			effects.TargetCreature("target creature"),
			func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					k := effects.S82NeoB18ArtifactsAndEnchantments(g, item.Controller)
					return effects.BoostUntilEOT{Target: t.ID, Power: k, Toughness: k, Label: label}.Apply(ctx)
				}
				return nil
			})
	}
	effects.Register(effects.Spec{
		OracleID:     MichikosReignOfTruthOracleID,
		Name:         "Michiko's Reign of Truth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1),
			chapter(2),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Michiko's Reign of Truth", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     MichikosReignOfTruthOracleID + "#1",
		Name:         "Portrait of Michiko",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82NeoB18SelfPT("This creature gets +1/+1 for each artifact and/or enchantment you control.",
				func(g *game.Game, source *game.Card) int {
					return effects.S82NeoB18ArtifactsAndEnchantments(g, source.Controller)
				}),
		},
	})
}

// michikosReignOfTruthOracleID is shared with the back face ("#1").
const MichikosReignOfTruthOracleID = "94794f7c-62e7-43c8-ba7f-ac5c0bc9a712"
