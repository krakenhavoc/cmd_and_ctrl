package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reservoir Kraken — Creature — Kraken {2}{U}{U}, 6/6 (Streets of New
// Capenna, #5491):
//
//	"Trample, ward {2}
//	 At the beginning of each combat, if this creature is untapped, any
//	 opponent may tap an untapped creature they control. If they do, tap
//	 this creature and create a 1/1 blue Fish creature token with "This
//	 token can't be blocked.""
//
// Desecration Demon's offer (ADR 0150) with a tap for the sacrifice: "any
// opponent" skips the Kraken's controller, so the tap is offered to each
// opponent in APNAP order until one makes it, and an opponent with no
// untapped creature can't (CR 118.3). The "if untapped" is an intervening
// "if" (CR 603.4), checked as the trigger fires and again as it resolves.
// The first opponent to tap a creature taps the Kraken and gives its
// controller the Fish; nobody tapping leaves the Kraken untapped.
//
// No simplification.
func init() {
	kraken := On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return StepBegan(game.StepBeginCombat, false)(ev, source, lki, g) && !source.Tapped
	}, "Reservoir Kraken — any opponent may tap an untapped creature they control",
		func(g *game.Game, item *game.StackItem) error {
			return AnyPlayerPays{
				Pay:       TapPayment(1, "creature", "creatures", QueryTypes("creature")),
				Opponents: true,
				Question:  "Reservoir Kraken — tap an untapped creature you control to tap it and give its controller a Fish?",
				Then:      reservoirKrakenTapped,
			}.Apply(NewContext(g, item))
		})
	Register(Spec{
		OracleID:        "42e4d4d8-e4e0-4cf8-a6d0-887a17513280",
		Name:            "Reservoir Kraken",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Reservoir Kraken — ward {2}"),
			InterveningIf(CondBoard{Name: "this creature is untapped", Test: func(ctx *Context) bool {
				self, ok := ctx.SourcePermanent()
				return ok && !self.Tapped
			}}, kraken),
		},
	})
}

// reservoirKrakenTapped is "if they do", keyed (#4264): the Kraken is
// tapped and its controller creates the Fish.
var reservoirKrakenTapped = AnyPlayerConsequences("any-player/reservoir-kraken-tap",
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		if err := (TapTarget{Target: ctx.Item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Controller: ctx.Item.Controller, Template: S82SncB07FishToken(), N: 1}.Apply(ctx)
	},
	nil,
	nil)
