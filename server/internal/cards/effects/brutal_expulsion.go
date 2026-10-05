package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brutal Expulsion — Instant {2}{U}{R}, devoid:
//
//	"Devoid (This card has no color.)
//	 Choose one or both —
//	 • Return target spell or creature to its owner's hand.
//	 • Brutal Expulsion deals 2 damage to target creature or
//	   planeswalker. If that creature or planeswalker would die this
//	   turn, exile it instead."
//
// Devoid is declared in PrintedKeywords, and the engine reads it as CR
// 702.114a's colour-defining ability (#2152), so the card is colourless
// in every zone. The first
// bullet is Venser's clause narrowed to creatures on the battlefield: a
// spell is returned without being countered, a creature is bounced. The
// second marks its target whether or not the damage is dealt (ADR 0108
// §1); a planeswalker that would die is exiled the same way. Each
// bullet reads its own target, in printed order (CR 608.2c), so one
// creature may be both bullets' target: it is bounced, and the damage
// then finds nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "71d178b3-e5ca-4576-83f7-8dce74758acd",
		Name:            "Brutal Expulsion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Return target spell or creature to its owner's hand.",
				TargetSpellOrPermanent("target spell or creature", nil, Creature()), returnModesSpellOrPermanentToHand),
			ModeDoing("Brutal Expulsion deals 2 damage to target creature or planeswalker. If that creature or planeswalker would die this turn, exile it instead.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())), damageModesTargetExileIfItDies(2)),
		),
	})
}
