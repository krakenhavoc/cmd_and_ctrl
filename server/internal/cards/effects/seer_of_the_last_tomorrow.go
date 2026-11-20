package effects

// Seer of the Last Tomorrow — Creature — Snake Cleric {2}{U}, 1/4 (Hour of
// Devastation, #5534):
//
//	"{U}, {T}, Discard a card: Target player mills three cards."
//
// The spellshaper cost (mana, {T} and DiscardACard, the card discarded as
// the ability is activated, CR 602.2b) on the shared TargetPlayerMills
// body. The {T} is a creature's, so it waits out summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc5cb189-9eff-4b4e-a46d-7a8e62529c89",
		Name:         "Seer of the Last Tomorrow",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: Target player mills three cards.",
			Cost:    Plus(ManaCost("{U}"), TapCost(), DiscardACard()),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(3),
		}},
	})
}
