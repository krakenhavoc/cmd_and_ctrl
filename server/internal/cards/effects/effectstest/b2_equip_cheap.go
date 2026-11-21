package effectstest

// b2_equip_cheap_test.go covers the b2-equip-cheap batch (#1107):
// seven inexpensive Equipment. The shared helpers — advanceToMain,
// seedBear, seedEquipment, equipTo, effectivePower, effectiveToughness,
// effectiveAbilities, attachmentHostOf, assertRestrictions,
// dealCombatDamageToPlayer, advanceToNextSeatsTurn — all live in
// attachments_test.go and its siblings.

const (
	BoneSawOracle         = "16e555f2-5aa8-4100-a036-eed48db0e84a"
	CatharsShieldOracle   = "b270d091-ea4c-4dfd-9c50-48b45bbc396e"
	HerosBladeOracle      = "e6bcd25d-39b6-4619-8a27-9f3d87f4a17b"
	DragonfireBladeOracle = "e809b847-b712-4558-95ea-9bb7356cde91"
	ThranPowerSuitOracle  = "3952c288-5633-4e40-abe6-18940d079644"
	BilbosRingOracle      = "a7cc0f6b-6b17-4e76-b2a6-6a6ee2519b61"
	LeylineAxeOracle      = "4f597675-0f6d-438c-990c-337171927a5e"
)
