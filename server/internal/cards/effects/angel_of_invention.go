package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Invention — Creature — Angel {3}{W}{W}, 2/1 (EDHREC rank
// 4216):
//
//	"Flying, vigilance, lifelink
//	 Fabricate 2 (When this creature enters, put two +1/+1 counters on
//	 it or create two 1/1 colorless Servo artifact creature tokens.)
//	 Other creatures you control get +1/+1."
//
// A five-mana Angel that arrives as either a 4/3 flying lifelinker or
// a 3/2 flanked by two 2/2 Servos, because the anthem applies to the
// Servos and not to itself. That choice is the card: the counters
// mode wins a race, the tokens mode survives a Doom Blade and feeds a
// sacrifice outlet.
//
// The anthem is a Layer 7c static scoped to OTHER creatures its
// controller controls — TribeFilter{Others: true, YoursOnly: true},
// Maja's shape with no tribe. It recomputes every pass, so a Servo
// made by the fabricate choice is a 2/2 the instant it exists, and
// the Angel itself stays a printed 2/1 plus whatever counters it took.
//
// FABRICATE is not an engine keyword; it is a printed triggered
// ability with a reminder text, and CR 702.122a spells it out as
// exactly that: "When this permanent enters, choose one — put N +1/+1
// counters on it; or create N 1/1 colorless Servo artifact creature
// tokens." So it is one ETB trigger with a two-branch choice, and
// MayChoice is the prompt for it — with both branches labelled, since
// this is a genuine choice between two effects rather than a yes/no.
//
// It is a CHOICE, not a "you may": one branch or the other always
// happens. Both labels are filled in so the client never renders
// "Yes / No" for a question that has no yes.
//
// An Angel that has left the battlefield before the trigger resolves
// can still make Servos — the tokens do not need it — but the
// counters branch has nothing to put counters on. That is the printed
// behaviour and needs no guard beyond AddCounter's own zone
// tolerance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "28c7f2b6-ba67-4ccf-9e1b-99c89a9d1f72",
		Name:            "Angel of Invention",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "lifelink"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Angel of Invention — fabricate 2", func(g *game.Game, item *game.StackItem) error {
				return MayChoice{
					Question: "Angel of Invention — fabricate 2",
					YesLabel: "Put two +1/+1 counters on it",
					NoLabel:  "Create two 1/1 colorless Servo artifact creature tokens",
					OnYes: func(ctx *Context) error {
						return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
					},
					OnNo: func(ctx *Context) error {
						return CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 2}.Apply(ctx)
					},
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
