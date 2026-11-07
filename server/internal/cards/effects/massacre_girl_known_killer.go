package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Massacre Girl, Known Killer — Legendary Creature — Human Assassin
// {2}{B}{B}, 4/4 (Murders at Karlov Manor, #4469):
//
//	"Menace
//	 Creatures you control have wither. (They deal damage to creatures in
//	 the form of -1/-1 counters.)
//	 Whenever a creature an opponent controls dies, if its toughness was
//	 less than 1, draw a card."
//
// Wither is a layer 6 keyword grant over the creatures you control, this
// one included; the damage path reads the keyword off the source as it
// deals damage (CR 702.80a). The trigger is Mari, the Killing Quill's
// "a creature an opponent controls dies" (b18OpponentsCreatureDied) with
// an intervening "if" (CR 603.4) on the creature's toughness as it last
// existed on the battlefield (CR 603.10a), -1/-1 counters included. That
// value cannot change once the creature has died, so the check as the
// ability resolves always agrees with the check as it triggered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ccb39f4a-2684-4eb7-9066-aad81ae08962",
		Name:            "Massacre Girl, Known Killer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			KeywordGrant(CreaturesYouControl, game.KeywordWither),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !b18OpponentsCreatureDied(ev, source, g) {
					return false
				}
				info, ok := g.LastKnownPermanentForEffect(ev.CardID)
				return ok && info.Toughness < 1
			}, "Massacre Girl, Known Killer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
