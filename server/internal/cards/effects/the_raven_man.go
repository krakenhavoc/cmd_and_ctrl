package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Raven Man — Legendary Creature — Human Wizard {1}{B}, 2/1 (Dominaria
// United):
//
//	"At the beginning of each end step, if a player discarded a card this
//	 turn, create a 1/1 black Bird creature token with flying and "This token
//	 can't block."
//	 {3}{B}, {T}: Each opponent discards a card. Activate only as a
//	 sorcery."
//
// The first ability is Mysterio's Mirage's trigger for EVERY player's end
// step and any player's discard: its intervening "if" (CR 603.4) reads each
// player's turn tally of cards discarded, as the step begins and again as
// the ability resolves, and it triggers once however many were discarded.
// The Bird is a catalog token whose can't-block is a static of its own. The
// second is Ninja of the Hand's discard (each opponent picks their own card)
// at sorcery speed, with the tap in the cost.
//
// No simplification.
func init() {
	anyoneDiscarded := func(g *game.Game) bool {
		for _, p := range g.Seats {
			if g.TurnTallyFor(p.ID).CardsDiscarded > 0 {
				return true
			}
		}
		return false
	}
	Register(Spec{
		OracleID:     "19918740-f6e3-40cd-9848-3c7a852a6c2a",
		Name:         "The Raven Man",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "a player discarded a card this turn",
				Test: func(ctx *Context) bool { return anyoneDiscarded(ctx.Game) },
			}, On(game.EventBeginEndStep,
				func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool { return anyoneDiscarded(g) },
				"The Raven Man — create a 1/1 black Bird creature token with flying and \"This token can't block.\"",
				Do(CreateToken{Template: RavenManBirdToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}{B}, {T}: Each opponent discards a card. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}{B}"), TapCost()),
			SorcerySpeed: true,
			Effect:       eachOpponentDiscardsOne,
		}},
	})
}
