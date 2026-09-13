package effects

// talismans.go — the Talismans (Mirrodin's allied cycle was
// reprinted alongside Modern Horizons' enemy cycle). Six landed with
// the mana-rider batch (#267); Talisman of Curiosity and Talisman of
// Resilience arrive with the roadmap's batch 04 (#297):
//
//	Artifact {2}
//	"{T}: Add {C}."
//	"{T}: Add {A} or {B}. This artifact deals 1 damage to you."
//
// The painlands' shape on a two-mana rock, and the reason the two
// cycles share one helper. A Talisman is the second-best two-mana
// rock in two-colour Commander precisely because the drawback is a
// drawback you choose: ramping into a colourless cost is free, and
// only the colored line costs a point.
//
// Unlike the painlands these are ordinary artifacts, so nothing here
// was previously broken — an unregistered Talisman was merely inert
// rather than actively misleading. It is registered with the
// painlands anyway because the machinery is identical and the play
// rate is comparable.
//
// No simplification. The mana cost, both abilities and the damage
// rider are all as printed.
func init() {
	for _, t := range []struct{ oracleID, name, a, b string }{
		{"4c0a0448-b9d6-43a0-8549-64066dac63f0", "Talisman of Dominance", "U", "B"},
		{"14d2979d-5728-42d7-a027-0eb1f754655d", "Talisman of Creativity", "U", "R"},
		{"1d9aeaaa-66f6-41cb-9bac-162d6fd8662c", "Talisman of Indulgence", "B", "R"},
		{"b693c3de-2eaf-4850-b405-e79d00adefda", "Talisman of Hierarchy", "W", "B"},
		{"00e35322-1a9a-41e3-9ce1-359c8eaa3bc7", "Talisman of Progress", "W", "U"},
		{"6c326439-5620-4ec6-a56a-fe9c3d5d2a46", "Talisman of Conviction", "R", "W"},
		// Roadmap batch 04 (#297) — two more of the ten.
		{"8c34b089-aad1-476e-958a-3077bf1bbb51", "Talisman of Curiosity", "G", "U"},
		{"42b8aa14-10bc-4bd6-88d9-4bb287eadd19", "Talisman of Resilience", "B", "G"},
		// Roadmap batch 05 (#298).
		{"f2ccc9e8-8e92-4f8c-8728-8c748630e0dd", "Talisman of Impulse", "R", "G"},
		// Roadmap batch 09 (#302).
		{"e5fcc5d7-6a60-4a5b-9d02-6c30041a95b9", "Talisman of Unity", "G", "W"},
	} {
		Register(Spec{
			OracleID:     t.oracleID,
			Name:         t.name,
			Completeness: CompletenessFull,
			ManaAbilities: []ManaAbility{
				painlessColorless(),
				painDual(t.a, t.b, "This artifact"),
			},
		})
	}
}
