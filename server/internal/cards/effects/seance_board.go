package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Séance Board — Artifact {2}:
//
//	"Morbid — At the beginning of each end step, if a creature died this
//	 turn, put a soul counter on this artifact.
//	 {T}: Add X mana of any one color, where X is the number of soul
//	 counters on this artifact. Spend this mana only to cast instant,
//	 sorcery, Demon, and Spirit spells."
//
// The counter is Bennie Bracks' "at the beginning of each end step"
// with Barad-dûr's creature-died tally as its intervening if (CR 603.4;
// the tally cannot fall within the turn, so no recheck is owed). The
// mana is Astral Cornucopia's "choose a color, one mana for each
// counter" reading soul counters, carrying the CR 106.12 spend
// restriction as one tag: a cast of an instant or sorcery spell, or a
// cast of a Demon or Spirit spell (ManaRestrictAnyOf, ADR 0103).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8331d2ff-3729-4014-8dcb-278644d2632a",
		Name:         "Séance Board",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15EndStepBegan(ev) && b11CreaturesDiedThisTurn(g) > 0
			}, "Séance Board — put a soul counter on this artifact", Do(CounterOnThis{Kind: "soul", N: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: ProducedOneColor(cmdr2B01SoulCountersOn),
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Demon|Spirit")},
			)},
			Label: "Add X mana of any one color, where X is the number of soul counters on this artifact. Spend this mana only to cast instant, sorcery, Demon, and Spirit spells",
		}},
	})
}
