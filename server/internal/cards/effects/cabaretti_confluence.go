package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cabaretti Confluence — Sorcery {3}{R}{G}{W} (Streets of New Capenna
// Commander):
//
//	"Choose three. You may choose the same mode more than once.
//	 • Create a token that's a copy of target creature you control. It
//	   gains haste. Sacrifice it at the beginning of the next end step.
//	 • Exile target artifact or enchantment.
//	 • Creatures target player controls get +1/+1 and gain first strike
//	   until end of turn."
//
// Mystic Confluence's shape (CR 700.2d, ChooseNRepeating): each occurrence
// has its own target group and runs in the order chosen. The copy bullet
// is Reflection of Kiki-Jiki's token (haste pinned to the token, one
// delayed sacrifice). The team bullet locks the target player's creatures
// as it resolves (CR 611.2c), so two of them stack to +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "99c32f20-da98-477d-9eb2-9c0b189d8d88",
		Name:         "Cabaretti Confluence",
		Completeness: CompletenessFull,
		Modes: ChooseNRepeating("Choose three (you may choose the same mode more than once)", 3, 3,
			ModeDoing("Create a token that's a copy of target creature you control. It gains haste. Sacrifice it at the beginning of the next end step.",
				TargetCreature("target creature you control", YouControl()),
				s82SncB15CopyWithHasteSacrifice("Cabaretti Confluence")),
			ModeDoing("Exile target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				ExileTheModesTarget),
			ModeDoing("Creatures target player controls get +1/+1 and gain first strike until end of turn.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					theirs := s82SncB15ControlledBy(t.ID)
					if err := (BoostUntilEOT{Match: theirs, Power: 1, Toughness: 1, Label: "Cabaretti Confluence — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: theirs, Keywords: []string{"first strike"}, Label: "Cabaretti Confluence — first strike"}.Apply(ctx)
				}),
		),
	})
}
