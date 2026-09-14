package effects

// Heartless Hidetsugu — Legendary Creature — Ogre Shaman {3}{R}{R},
// 4/3 (EDHREC rank 3043):
//
//	"{T}: Heartless Hidetsugu deals damage to each player equal to
//	 half that player's life total, rounded down."
//
// The table-halver. One CR 602 activation with a tap cost — a
// creature source, so summoning sickness applies as printed — and
// a body that reads every life total before dealing any of the
// damage, so the amounts are simultaneous (b29DamageEachPlayerHalfTheirLife).
// The damage is the Ogre's, so a Furnace of Rath doubles it and a
// damage-dealt trigger on it sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b68e965a-96f4-447a-99be-35dbd27846e9",
		Name:         "Heartless Hidetsugu",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{T}: Heartless Hidetsugu deals damage to each player equal to half that player's life total, rounded down.",
			Cost:   TapCost(),
			Effect: b29DamageEachPlayerHalfTheirLife,
		}},
	})
}
