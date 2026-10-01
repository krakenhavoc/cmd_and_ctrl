package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cast Through Time — Enchantment {4}{U}{U}{U}:
//
//	"Instant and sorcery spells you control have rebound. (Exile the
//	 spell as it resolves if you cast it from your hand. At the
//	 beginning of your next upkeep, you may cast that card from exile
//	 without paying its mana cost.)"
//
// Every instant and sorcery its controller casts from hand comes back
// for free a turn later. It is a static over SPELLS (CR 613.1f, ADR
// 0107 §3): the stack step of the layer pass gives each such spell on
// the stack rebound, and the resolution reads it the way it reads a
// printed rebound (game/rebound.go). A static is never locked in
// (CR 611.3a), so a spell cast after Cast Through Time entered has
// rebound, and one still on the stack when it leaves loses it.
//
// The rebound itself does the rest: only a spell cast from your hand
// is exiled (a rebound recast from exile goes to the graveyard), a
// copy is never exiled, and a spell that also has buyback or is an
// Adventure asks you which replacement applies (CR 616.1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a911cb2-caae-4378-bf09-1c5b0751dd35",
		Name:         "Cast Through Time",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			SpellsYouControlHave(Or(Instant(), Sorcery()), game.KeywordRebound),
		},
	})
}
