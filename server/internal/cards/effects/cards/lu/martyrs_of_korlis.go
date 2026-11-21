package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Martyrs of Korlis — Creature — Human {3}{W}{W}, 1/6:
//
//	"As long as this creature is untapped, all damage that would be dealt
//	 to you by artifacts is dealt to this creature instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) to the
// Martyrs, while untapped, of damage from an artifact source, judged as
// the source is when the damage would be dealt (CR 609.7c). The ruling:
// with several, you choose which receives each event (CR 616.1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ca54a23-f8eb-4982-b4ee-7392e2f2a1b3",
		Name:         "Martyrs of Korlis",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.StaticRedirection("Martyrs of Korlis — damage to you from artifacts is dealt to it instead",
				effects.RedirectWhere{Applies: effects.UntappedAnd(func(_ *game.Game, src *game.Card, ev *game.ReplacementEvent) bool {
					return effects.DamageToYou(nil, src, ev) && ev.SourceLKI != nil && effects.LkiHasType(ev.SourceLKI, "Artifact")
				}), To: effects.ToThisPermanent}),
		},
	})
}
