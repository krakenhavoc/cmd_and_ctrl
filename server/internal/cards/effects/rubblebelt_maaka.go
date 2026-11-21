package effects

// Rubblebelt Maaka — Creature — Cat {3}{R}, 3/3 (Ravnica Remastered, #5948):
//
//	"Bloodrush — {R}, Discard this card: Target attacking creature gets
//	 +3/+3 until end of turn."
//
// Bloodrush is Pyrewild Shaman's shape (gtcBloodrush): an activated ability
// that works from the hand, with discarding this card as part of the cost. A
// target that stopped attacking or left is illegal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "185db9c7-3936-473c-9220-ae255527aaf9",
		Name:         "Rubblebelt Maaka",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{gtcBloodrush("Rubblebelt Maaka", "{R}", 3, 3, "")},
	})
}
