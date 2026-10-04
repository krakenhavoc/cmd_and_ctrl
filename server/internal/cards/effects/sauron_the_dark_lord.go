package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sauron, the Dark Lord — Legendary Creature — Avatar Horror
// {3}{U}{B}{R}, 7/6:
//
//	"Ward—Sacrifice a legendary artifact or legendary creature.
//	 Whenever an opponent casts a spell, amass Orcs 1.
//	 Whenever an Army you control deals combat damage to a player, the
//	 Ring tempts you.
//	 Whenever the Ring tempts you, you may discard your hand. If you
//	 do, draw four cards."
//
// The commander of #2062's deck. Four abilities, each on an existing
// shape:
//
//   - The ward's sacrifice is one permanent that is legendary and an
//     artifact or a creature. An opponent's Ring-bearer is legendary
//     through the Ring (CR 701.54c), so it pays.
//   - Amass is the keyword action (CR 701.47, ADR 0087), once per spell
//     an opponent casts.
//   - "An Army you control deals combat damage to a player" triggers
//     once for each Army that connects (it is not "one or more").
//   - "Whenever the Ring tempts you" triggers on every temptation,
//     with or without a creature to choose (CR 701.54d). The "you may"
//     is asked as it resolves (MayChoice); "discard your hand" can be
//     chosen with an empty hand, and choosing it is what "if you do"
//     reads (CR 118.12), so the four cards are drawn then too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f8f29ba-7e64-455d-be05-ec7973f2bd0a",
		Name:         "Sauron, the Dark Lord",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardSacrifice("a legendary artifact or legendary creature", Legendary(), Or(Artifact(), Creature())),
				"Sauron, the Dark Lord — ward, sacrifice a legendary artifact or legendary creature"),
			On(game.EventCast, AnOpponentCast(nil), "Sauron, the Dark Lord — amass Orcs 1", Do(Amass{Subtype: "Orc", N: 1})),
			On(game.EventDealDamage, anArmyYouControlDealtCombatDamageToAPlayer,
				"Sauron, the Dark Lord — the Ring tempts you", Do(TheRingTemptsYou{})),
			WheneverTheRingTemptsYou("Sauron, the Dark Lord — you may discard your hand and draw four cards", sauronTheDarkLordWheel),
		},
	})
}

// anArmyYouControlDealtCombatDamageToAPlayer is "whenever an Army you
// control deals combat damage to a player": one trigger per Army and
// per damage event.
func anArmyYouControlDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && Army()(g, source.Controller, c)
}

// sauronTheDarkLordWheel is the last ability: "you may discard your
// hand. If you do, draw four cards."
func sauronTheDarkLordWheel(g *game.Game, item *game.StackItem) error {
	return MayChoice{
		Question: "Sauron, the Dark Lord — discard your hand and draw four cards?",
		OnYes: func(ctx *Context) error {
			if _, err := discardWholeHand(ctx.Game, ctx.Controller()); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 4}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
