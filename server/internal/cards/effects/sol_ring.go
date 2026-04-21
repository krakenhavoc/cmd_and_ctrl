package effects

// Sol Ring — "{T}: Add {C}{C}." Colorless Medallion-tier artifact.
//
// S14 sandbox: mana abilities are activated abilities, and the
// activated-ability pipeline lands in S19. The card is registered
// here (no OnResolve, no OnETB) so its CardView gets the catalog
// AUTO badge for UX consistency with other "known-handled" cards
// — even though the handling is "nothing special at resolution."
// Mana is tracked on paper today.
//
// Sol Ring's combat-tempo impact is real; catalog-badging it also
// marks the card for future auto-fire treatment without a rewrite.
func init() {
	Register(Spec{
		OracleID: "6ad8011d-3471-4369-9d68-b264cc027487",
		Name:     "Sol Ring",
	})
}
