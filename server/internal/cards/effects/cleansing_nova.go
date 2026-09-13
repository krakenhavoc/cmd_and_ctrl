package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "aff34f28-f707-4458-8af3-1bd5b13a6b10",
		Name:         "Cleansing Nova",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Indestructible saves a permanent from single-target removal and from lethal damage, but a board wipe (\"destroy all\") still destroys it."},
		Modes: ChooseOne(
			Mode("Destroy all creatures."),
			Mode("Destroy all artifacts and enchantments."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			sweeps := []CardPredicate{
				Creature(),
				Or(Artifact(), Enchantment()),
			}
			for i, match := range sweeps {
				if !ctx.HasMode(i) {
					continue
				}
				if err := (DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
