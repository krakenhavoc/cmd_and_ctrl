package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alisaie Leveilleur — Legendary Creature — Elf Wizard {2}{W}, 3/2:
//
//	"Partner with Alphinaud Leveilleur (When this creature enters,
//	 target player may put Alphinaud Leveilleur into their hand from
//	 their library, then shuffle.)
//	 First strike
//	 Dualcast — The second spell you cast each turn costs {2} less to
//	 cast."
//
// Partner with is the shared entry trigger (PartnerWith); the
// deck-construction half is read off the oracle text. The discount reads
// the per-turn cast tally: the spell being priced is not in it yet
// (OtherSpellsCastThisTurn says why), so a tally of exactly one means
// "this is the second spell". It takes {2} off the generic part of the
// cost only, as a cost reduction does (CR 601.2f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4bd1d079-314e-4fb5-9b43-c9c83e5528bc",
		Name:            "Alisaie Leveilleur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Alisaie Leveilleur", "Alphinaud Leveilleur"),
		},
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Dualcast — the second spell you cast each turn costs {2} less to cast.",
				YourSpell(), secondSpellYouCastThisTurn()),
		},
	})
}

// secondSpellYouCastThisTurn passes while the caster has cast exactly one
// spell this turn, so the spell being priced is their second.
func secondSpellYouCastThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.CastTallyFor(q.Controller).Total == 1
	}
}
