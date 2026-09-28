package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archetype of Endurance — Enchantment Creature — Boar {6}{G}{G}, 6/5
// (EDHREC rank 2876):
//
//	"Creatures you control have hexproof.
//	 Creatures your opponents control lose hexproof and can't have or
//	 gain hexproof."
//
// #1651's static proof card (ADR 0038's amendment of 2026-09-28, B2),
// and Arcane Lighthouse's inverse. Two layer-6 statics: a keyword grant
// over its controller's creatures (the Archetype included — the clause
// has no "other"), and LoseAndCantHave over everyone else's. The
// can't-have beats every grant whatever its timestamp, so an opponent's
// Heroic Intervention or Swiftfoot Boots gives their creatures nothing.
//
// Two opposing Archetypes of Endurance cancel out: each player's
// creatures have hexproof from one and can't have it from the other,
// and "can't" wins (CR 101.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79254223-3fa6-4b7c-8163-1e48bf5cb708",
		Name:         "Archetype of Endurance",
		Completeness: CompletenessFull,
		Static:       archetypeStatics("hexproof"),
	})
}

// archetypeStatics is the Archetype cycle's two lines for one keyword:
// "Creatures you control have <kw>. Creatures your opponents control
// lose <kw> and can't have or gain <kw>." (Aggression, Courage,
// Endurance, Finality, Imagination.)
func archetypeStatics(keyword string) []game.StaticAbility {
	return []game.StaticAbility{
		b16GrantKeywords(b16CreaturesYouControl, keyword),
		LoseAndCantHave(b27CreaturesOpponentsControl, keyword),
	}
}
