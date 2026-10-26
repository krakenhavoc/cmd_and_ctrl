package effects

const jetpackDeathSeltzerMonstrosity = "jetpack-death-seltzer/monstrosity-3"

// Jetpack Death Seltzer — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Jetpack", "Death", "Seltzer"; three art stickers.
//	{TK}{TK} — Trample
//	{TK}{TK}{TK} — {3}: Monstrosity 3.
//	{TK}{TK}{TK} — 2/7
//	{TK}{TK}{TK}{TK} — 6/5
//
// The first sticker gives trample. The second gives Monstrosity 3 as an activated ability (CR 701.37): "{3}: Monstrosity 3."
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe37c718-b3fb-4e95-a547-dd741b70421c",
		Name:         "Jetpack Death Seltzer",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key:       jetpackDeathSeltzerMonstrosity,
				Activated: []ActivatedAbility{Monstrosity(ManaCost("{3}"), 3)},
				Text:      "{3}: Monstrosity 3. (If this creature isn't monstrous, put three +1/+1 counters on it and it becomes monstrous.)",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Keywords: []string{"trample"}},
			{Grant: jetpackDeathSeltzerMonstrosity},
		}},
	})
}
