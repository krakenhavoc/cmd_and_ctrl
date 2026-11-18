package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shanid, Sleepers' Scourge — Legendary Creature — Human Knight
// {1}{R}{W}{B}, 2/4:
//
//	"Menace
//	 Other legendary creatures you control have menace.
//	 Whenever you play a legendary land or cast a legendary spell, you draw
//	 a card and you lose 1 life."
//
// The grant is Esika's recipient set (other legendary creatures you
// control, by instance, read off the effective supertypes). The trigger is
// ONE printed ability with two conditions, a land PLAYED (never one put
// onto the battlefield by an effect, b20LandPlayed) or a spell CAST. Both
// read the card's legendary supertype: the land as it entered, the spell as
// it was cast. Life is lost as a draw-then-lose pair, as Black Widow's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "70e03c58-194f-443d-ab64-646dd14d2f53",
		Name:            "Shanid, Sleepers' Scourge",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			Labelled(KeywordGrant(otherLegendaryCreaturesYouControl, "menace"),
				"Other legendary creatures you control have menace."),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventZoneMove, game.EventCast},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !b31YouPlayedALandOrCastASpell(ev, source, g) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsLegendary()
				}, "Shanid, Sleepers' Scourge — draw a card and lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
