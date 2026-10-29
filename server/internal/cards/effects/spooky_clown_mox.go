package effects

const spookyClownMoxTap = "spooky-clown-mox/tap-tap-creature"

// Spooky Clown Mox — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Spooky", "Clown", "Mox"; three art stickers.
//	{TK}{TK} — Vigilance
//	{TK}{TK}{TK}{TK} — {1}, {T}: Tap target creature.
//	{TK}{TK} — 1/5
//	{TK}{TK}{TK} — 5/4
//
// The first sticker gives vigilance. The second gives an ability that taps the permanent and pays {1} to tap a target creature.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "785e6563-51c4-4063-9ad9-0392954e135e",
		Name:         "Spooky Clown Mox",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: spookyClownMoxTap,
				Activated: []ActivatedAbility{{
					Label:   "{1}, {T}: Tap target creature.",
					Cost:    Plus(ManaCost("{1}"), TapCost()),
					Targets: TargetCreature("target creature"),
					Effect:  tapLegalCardTargets,
				}},
				Text: "{1}, {T}: Tap target creature.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{"vigilance"}},
			{Grant: spookyClownMoxTap},
		}},
	})
}
