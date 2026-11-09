package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Molten Exhale — Sorcery {1}{R} (Tarkir: Dragonstorm, #4463):
//
//	"You may cast this spell as though it had flash if you behold a
//	 Dragon as an additional cost to cast it. (To behold a Dragon,
//	 choose a Dragon you control or reveal a Dragon card from your
//	 hand.)
//	 Molten Exhale deals 4 damage to target creature or planeswalker."
//
// CR 601.3c: a spell that may be cast as though it had flash only if an
// additional cost is paid may BEGIN to be cast at instant speed. The
// engine's shape for that is FlashForMore's: an offer that opens the
// instant-speed window for its own claim (AlternativeCost.AsThoughFlash).
// Here the offer charges the printed {1}{R} and carries the behold as its
// card component (AlternativeCost.Behold, CR 701.4a, the one candidate
// walk Kindle the Inner Flame's flashback reads): the caster names one
// Dragon they control or one Dragon card in their hand on alt_cost_ids, a
// card from hand is revealed, and nothing moves. Cast for its printed
// cost alone, the spell keeps sorcery timing.
//
// Modelling the additional cost as an offer has FlashForMore's one edge:
// the offer replaces the mana cost, so the flash cast cannot also take a
// second alternative cost (CR 118.9a). On the printed card the behold is
// an additional cost and the flash a timing permission (CR 601.3c), so a
// flash cast could also take an alternative cost; here it can't. The
// price is otherwise the same, and nothing on the card reads whether a
// Dragon was beheld.
//
// Simplification: that one edge, declared as the caveat. It closes when
// the engine grants "as though it had flash" for an additional cost
// without an alternative-cost offer.
func init() {
	Register(Spec{
		OracleID:     "82aa1402-4b1c-4343-aba5-dd3271cc1ac5",
		Name:         "Molten Exhale",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Beholding a Dragon to cast this at instant speed can't be combined with another alternative cost.",
		},
		Purpose: ForTargets(DamageToTarget(0, 4)),
		AlternativeCosts: []game.AlternativeCost{{
			Key:           "flash-behold",
			Label:         "Cast with flash — behold a Dragon ({1}{R})",
			ManaCost:      "{1}{R}",
			AsThoughFlash: true,
			Behold:        &game.RevealCost{Subtype: "Dragon", Behold: true},
			BeholdCount:   1,
			PayLabel:      "a Dragon",
		}},
		Targets:   TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: damageToFirstTarget(4),
	})
}
