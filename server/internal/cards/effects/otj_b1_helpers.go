package effects

// otj_b1_helpers.go — the shared pieces behind Outlaws of Thunder
// Junction batch 1 (#4468). Every package-level name carries the otjB1
// prefix because other batches land beside this one.

// otjB1LootCounter is Bandit's Haul's loot counter. Counter kinds are
// free-form strings (game/counter_types.go); this one has no engine
// meaning beyond its name.
const otjB1LootCounter = "loot"

// otjB1SacrificedWasOutlaw reports whether the permanent the ability's
// cost sacrificed was an outlaw as it last existed on the battlefield
// (CR 608.2h): an Assassin, Mercenary, Pirate, Rogue or Warlock, or a
// creature with every creature type (CR 702.73a). The same batch as
// b40Outlaw, read off the cost's last-known record rather than a live
// card.
func otjB1SacrificedWasOutlaw(ctx *Context) bool {
	info, ok := ctx.SacrificedPermanent()
	if !ok {
		return false
	}
	if info.Characteristic.AllCreatureTypes {
		return true
	}
	for _, t := range []string{"Assassin", "Mercenary", "Pirate", "Rogue", "Warlock"} {
		if hasFold(info.Characteristic.Subtypes, t) {
			return true
		}
	}
	return false
}
