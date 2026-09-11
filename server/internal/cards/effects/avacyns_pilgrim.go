package effects

// Avacyn's Pilgrim — Creature — Human Monk {G}, 1/1 (EDHREC rank
// 514):
//
//	"{T}: Add {W}."
//
// A one-drop dork that fixes into white — the Selesnya deck's Elf.
// Same shape as Llanowar Elves with a different colour; CR 302.1
// summoning sickness applies through the engine's mana-ability gate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "069f6530-e65c-4d52-85f3-e0a2acd148c5",
		Name:     "Avacyn's Pilgrim",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
