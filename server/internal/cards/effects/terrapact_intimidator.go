package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terrapact Intimidator — Creature — Kavu Scout {1}{R}, 2/1 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, target opponent may have you create two
//	 Lander tokens. If they don't, put two +1/+1 counters on this
//	 creature."
//
// Combustible Gearhulk's punisher: the question belongs to the targeted
// opponent and is asked as the trigger resolves (MayChoice with Player).
// A yes makes two Landers (Horizon Explorer's token, b16LanderToken) for
// the Intimidator's controller; a no puts two +1/+1 counters on it, and
// only while it is still the same object on the battlefield
// (CounterOnThis, CR 400.7). An opponent who has become an illegal target
// is never asked, and the trigger does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c08e0958-8fa2-49c7-ab17-a70dfc1ef33a",
		Name:         "Terrapact Intimidator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Terrapact Intimidator — target opponent may have you create two Landers", terrapactIntimidatorAsk),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}

// terrapactIntimidatorAsk asks the still-legal opponent.
func terrapactIntimidatorAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return MayChoice{
			Player:   t.ID,
			Question: "Terrapact Intimidator — let its controller create two Lander tokens? If you don't, it gets two +1/+1 counters.",
			YesLabel: "They create two Landers",
			NoLabel:  "It gets two +1/+1 counters",
			Then:     terrapactIntimidatorAnswer,
		}.Apply(ctx)
	}
	return nil
}

// terrapactIntimidatorAnswer is the opponent's answer. An on-disk key:
// never renamed, never reused.
var terrapactIntimidatorAnswer = MayYesNo("confirm/terrapact-intimidator-landers-or-counters",
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: b16LanderToken(), N: 2}.Apply(ctx)
	},
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return CounterOnThis{Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
	})
