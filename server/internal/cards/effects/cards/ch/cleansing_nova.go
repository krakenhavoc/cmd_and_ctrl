package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleansing Nova — Sorcery {3}{W}{W}:
//
//	"Choose one —
//	 • Destroy all creatures.
//	 • Destroy all artifacts and enchantments."
//
// Five mana for whichever wrath the table needs. Austere Command's
// flexibility at half the setup and half the choices — you get one
// mode, not two, and the artifact/enchantment half is a single sweep
// rather than two separate ones.
//
// That "and" is worth being precise about: the second mode is ONE
// sweep over Or(Artifact(), Enchantment()), not an artifact sweep
// followed by an enchantment sweep. An artifact enchantment is
// destroyed once either way, but the simultaneous batch is the whole
// point of the primitive and splitting it would put two events where
// the card prints one.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aff34f28-f707-4458-8af3-1bd5b13a6b10",
		Name:         "Cleansing Nova",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Destroy all creatures."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
			effects.ModeWithPurpose(effects.Mode("Destroy all artifacts and enchantments."), game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy}}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			sweeps := []effects.CardPredicate{
				effects.Creature(),
				effects.Or(effects.Artifact(), effects.Enchantment()),
			}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (effects.DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
