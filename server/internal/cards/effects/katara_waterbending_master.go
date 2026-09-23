package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara, Waterbending Master — Legendary Creature — Human Warrior
// Ally {1}{U}, 1/3:
//
//	"Whenever you cast a spell during an opponent's turn, you get an
//	 experience counter.
//	 Whenever Katara attacks, you may draw a card for each experience
//	 counter you have. If you do, discard a card."
//
// The cast trigger is Brineborn Cutthroat's condition — the caster is
// Katara's controller and is not the active player — paying out in
// the player counter Aang, Airbending Master also uses
// (youGetAnExperienceCounter).
//
// The attack trigger's "you may" is asked at RESOLUTION (MayChoice,
// #796), where the card prints it, and the count is read there too,
// so a counter gained in response is included. The discard is linked
// to the draw (CR 607.2, Mask of Memory's shape): it happens only on
// the branch where the cards were drawn, and the player picks it from
// the hand the draws landed in. With no experience counters there is
// nothing to draw, so nothing is asked and nothing is discarded —
// drawing zero cards is not having drawn.
//
// No simplification.
func init() {
	const (
		castLabel   = "Katara, Waterbending Master — you get an experience counter"
		attackLabel = "Katara, Waterbending Master — draw a card per experience counter, then discard"
	)
	Register(Spec{
		OracleID:     "2e3af413-6706-4023-871a-e5e49da5beab",
		Name:         "Katara, Waterbending Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !isActivePlayer(g, source.Controller)
			}, castLabel, youGetAnExperienceCounter),
			WheneverThisAttacks(attackLabel, kataraMayDrawPerExperience),
		},
	})
}

// kataraMayDrawPerExperience is "you may draw a card for each
// experience counter you have. If you do, discard a card."
//
// Caller holds g.mu.
func kataraMayDrawPerExperience(g *game.Game, item *game.StackItem) error {
	n := experienceCounters(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return MayChoice{
		Question: "Katara, Waterbending Master — draw a card for each experience counter, then discard a card?",
		YesLabel: "Draw",
		NoLabel:  "Decline",
		OnYes: func(ctx *Context) error {
			return b16DrawThenDiscard(ctx.Game, ctx.Item, experienceCounters(ctx.Game, ctx.Controller()), 1)
		},
	}.Apply(NewContext(g, item))
}
