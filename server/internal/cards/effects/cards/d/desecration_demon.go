package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desecration Demon — Creature — Demon {2}{B}{B}, 6/6:
//
//	"Flying
//	 At the beginning of each combat, any opponent may sacrifice a
//	 creature of their choice. If a player does, tap this creature and
//	 put a +1/+1 counter on it."
//
// Innocent Traveler's offer (ADR 0150): "any opponent" skips the
// Demon's controller, so the sacrifice is offered to each opponent in
// APNAP order from the active player until one makes it, and an
// opponent with no creature can't (CR 118.3). It fires at the beginning
// of every combat, on every player's turn. The first opponent to pay
// taps the Demon and gives it a counter; nobody paying leaves it
// untapped and free to attack or block. A Demon that has left the
// battlefield by then is tapped and grown not at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef75be00-1a88-47a1-a1ae-fe2c9881a798",
		Name:            "Desecration Demon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Desecration Demon — any opponent may sacrifice a creature to tap it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.AnyPlayerPays{
						Pay:       effects.SacrificePayment(1, "creature", "creatures", effects.QueryType("creature")),
						Opponents: true,
						Question:  "Desecration Demon — sacrifice a creature to tap it and put a +1/+1 counter on it?",
						Then:      desecrationDemonTapped,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// desecrationDemonTapped is "if a player does", keyed (#4264).
var desecrationDemonTapped = effects.AnyPlayerConsequences("any-player/desecration-demon-tap",
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		demon := ctx.Item.SourceCardID
		if err := (effects.TapTarget{Target: demon}).Apply(ctx); err != nil {
			return err
		}
		return effects.AddCounter{Target: demon, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	},
	nil,
	nil)
