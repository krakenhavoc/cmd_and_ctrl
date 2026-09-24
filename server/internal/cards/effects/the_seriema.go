package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Seriema — Legendary Artifact — Spacecraft, {1}{W}{W}, 5/5:
//
//	When The Seriema enters, search your library for a legendary
//	creature card, reveal it, put it into your hand, then shuffle.
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Spacecraft. Station only as a sorcery.
//	It's an artifact creature at 7+.)
//	7+ | Flying
//	Other tapped legendary creatures you control have indestructible.
//
// #337, reported in-app as "the enter the battlefield effect did not
// trigger": there was no catalog entry at all, because ADR 0037 §5
// refuses to ship half a card. Every line is buildable now:
//
//   - "7+ | Flying" is ThresholdKeywords(7, "flying") — a layer-6
//     self-grant with a charge-counter gate (ADR 0071).
//   - "It's an artifact creature at 7+" is SpacecraftAt(7, 5, 5) — a
//     layer-4 type add plus a layer-7b base-P/T set, same gate. Since
//     ADR 0039 made layer 4 authoritative, that is a real creature to
//     combat, targeting and the state-based actions, and the printed
//     5/5 is its base P/T rather than a number on the wire.
//   - Station itself is Station() (#759, ADR 0071 addendum
//     2026-09-23): #758's tap-another cost, sorcery timing, and
//     charge counters equal to the tapped creature's power as the
//     ability resolves (CR 608.2h).
//
// The thresholds read the counter count LIVE (CR 721.2a is an "as
// long as"), so a Seriema whose charge counters are removed stops
// being a creature and loses flying in the same recompute.
//
// CR 721.2c — "a station card has no power or toughness outside the
// battlefield" — is not modelled: The Seriema shows its printed 5/5
// in hand and in the graveyard. That is a view and importer rule with
// no rules consequence anywhere a player can act on it, and it is
// noted in ADR 0071 rather than caveated.
func init() {
	Register(Spec{
		OracleID:     "a6bcec1f-f515-4e63-9e84-8eb04cc582ff",
		Name:         "The Seriema",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("The Seriema — search for a legendary creature card",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: MatchLegendaryCreature,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "The Seriema — search for a legendary creature card",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(7, 5, 5),
			ThresholdKeywords(7, "flying"),
			otherTappedLegendaryCreaturesYouControlHaveIndestructible(),
		),
	})
}

// otherTappedLegendaryCreaturesYouControlHaveIndestructible is The
// Seriema's last line — an ordinary layer-6 grant with no gate on it,
// because it is printed above the threshold bar and is on from the
// moment the Spacecraft lands.
//
// "Tapped" is an AppliesTo input like any other, and the layer
// listener already bumps on EventTapCard / EventUntapCard, so a
// creature that taps to attack — or to station The Seriema itself —
// gains indestructible in the same recompute rather than at the next
// unrelated event.
func otherTappedLegendaryCreaturesYouControlHaveIndestructible() game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID != source.InstanceID &&
				target.Controller == source.Controller &&
				target.Tapped && target.IsCreature() && target.IsLegendary()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, []string{"indestructible"})
		},
	}
}
