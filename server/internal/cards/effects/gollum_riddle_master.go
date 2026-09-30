package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gollumRiddleMasterLabel is the trigger's stack label, and with it the
// key its "hasn't been chosen" memory is kept under.
const gollumRiddleMasterLabel = "Gollum, Riddle Master — an opponent cast a spell of the chosen quality"

// Gollum, Riddle Master — Legendary Creature — Halfling Horror {1}{B},
// 3/1:
//
//	"As Gollum enters, choose odd or even. (Zero is even.)
//	 Whenever an opponent casts a spell with mana value of the chosen
//	 quality, choose one that hasn't been chosen —
//	 • Put a +1/+1 counter on Gollum.
//	 • Each opponent loses 2 life and you gain 2 life.
//	 • Draw a card."
//
// "Choose odd or even" is a CR 614.12 choice of one of two printed
// words, the same prompt the Sieges' "choose Khans or Dragons" uses
// (ChooseOptionAsEnters), stored on this permanent. The trigger reads
// it live: before the controller answers, no quality is chosen and no
// spell matches.
//
// The spell's mana value is read as it is on the stack, X included (CR
// 202.3e), and zero is even. ChooseOneNotChosen (ADR 0097) makes each
// bullet available once for this Gollum; a Gollum that leaves and
// returns chooses its quality again and remembers no modes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "514b451b-814d-45fb-a2ba-8fe6f0bdad60",
		Name:         "Gollum, Riddle Master",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Gollum, Riddle Master", "odd", "even"),
		Triggered: []game.TriggeredAbility{
			gollumRiddleMasterTrigger(),
		},
	})
}

func gollumRiddleMasterTrigger() game.TriggeredAbility {
	t := On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || !ByAnOpponent(ev, source, game.Characteristic{}, g) {
			return false
		}
		quality := ChosenOptionOf(g, source.InstanceID)
		if quality == "" {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		mv, ok := g.ManaValueForEffect(c)
		if !ok {
			return false
		}
		return (mv%2 == 0) == (quality == "even")
	}, gollumRiddleMasterLabel, func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOneNotChosen(
		ModeDoing("Put a +1/+1 counter on Gollum.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		ModeDoing("Each opponent loses 2 life and you gain 2 life.", nil,
			eachOpponentLosesTwoYouGainTwo),
		ModeDoing("Draw a card.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
	)
	return t
}
