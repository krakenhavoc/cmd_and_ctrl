package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// nghaHiveMindLabel is Zellix's Hive Mind trigger's stack label, which
// the "one or more" dedup keys on.
const nghaHiveMindLabel = "Zellix, Sanity Flayer — Hive Mind: create a 1/1 black Horror"

// Zellix, Sanity Flayer — Legendary Creature — Horror {2}{U}, 2/3:
//
//	"Hive Mind — Whenever a player mills one or more creature cards, you
//	 create a 1/1 black Horror creature token.
//	 {1}, {T}: Target player mills three cards.
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// Hive Mind watches EventMill, which the engine emits once per card, and
// fires on the first creature card a player mills in a batch, declining
// that player's later cards (OncePerBatch keyed per milling player,
// nghaOncePerMiller). So one mill of several creature cards is one
// Horror, and an instruction that makes several players mill at once
// (Fractured Sanity's "each opponent mills fourteen cards") is one
// Horror for each player who milled a creature card, per the card's
// ruling. A mill a replacement sends elsewhere (Rest in Peace's exile)
// is not milled into a graveyard and fires nothing. Whose mill it is
// does not matter, the controller's own included; the token is the
// trigger controller's.
//
// The activated ability targets a player. Its {T} is the creature's own,
// so it is not available the turn Zellix arrives (CR 302.6).
//
// Choose a Background is deck construction (ADR 0144), read off the
// oracle text by internal/deck.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4dde484a-8d5b-4eb6-ab4a-e30cc370431c",
		Name:         "Zellix, Sanity Flayer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			nghaOncePerMiller(On(game.EventMill, nghaMilledACreatureCard, nghaHiveMindLabel,
				Do(CreateToken{Template: TokenCard("1/1 black Horror"), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Target player mills three cards.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(3),
		}},
	})
}

// nghaMilledACreatureCard is Hive Mind's condition before the dedup: a
// creature card was milled into a graveyard, by any player.
func nghaMilledACreatureCard(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventMill || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// nghaOncePerMiller makes a mill trigger fire once per batch per milling
// player — "whenever a player mills one or more …" (CR 603.2c). EventMill
// names the milling player as its Actor.
func nghaOncePerMiller(t game.TriggeredAbility) game.TriggeredAbility {
	t.OncePerBatch = true
	t.BatchKey = func(ev game.Event, _ *game.Card, _ *game.Game) string { return ev.Actor.String() }
	return t
}
