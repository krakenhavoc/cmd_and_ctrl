package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yes Man, Personal Securitron — Legendary Artifact Creature — Robot {2}{W},
// 2/2 (Fallout):
//
//	"{T}: Target opponent gains control of Yes Man. When they do, you draw
//	 two cards and put a quest counter on Yes Man. Activate only during your
//	 turn.
//	 Wild Card — When Yes Man leaves the battlefield, its owner creates a
//	 tapped 1/1 white Soldier creature token for each quest counter on it."
//
// The hand-over is Humble Defector's: a layer-2 control effect with no
// stated duration (CR 611.2a). "When they do" is a reflexive trigger (CR
// 603.12) raised only if the opponent really gained control (Yes Man had
// left, or come back as a new object, otherwise nothing happens): it is
// mine, so I draw two, and it puts the quest counter on Yes Man wherever
// he is controlled. The leave trigger reads the quest counters he had as he
// left (CR 608.2h) and the Soldiers are created tapped by his owner, not
// his controller.
//
// No simplification.
var yesManBody = game.SimpleDelayedBody("yes-man/draw-two-and-quest-counter", func(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{N: 2}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range ctx.PayloadCards() {
		if onBattlefield(g, id) {
			return AddCounter{Target: id, Kind: "quest", N: 1}.Apply(ctx)
		}
	}
	return nil
})

func init() {
	Register(Spec{
		OracleID:     "16b8eab0-ccbd-4dbf-9998-82a4d85d0f77",
		Name:         "Yes Man, Personal Securitron",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     "{T}: Target opponent gains control of Yes Man. When they do, you draw two cards and put a quest counter on Yes Man. Activate only during your turn.",
			Purpose:   game.Purpose{Draws: 2},
			Cost:      TapCost(),
			Condition: DuringYourTurn(),
			Targets:   TargetPlayer("target opponent", Opponent()),
			Effect: Do(Instruction(func(ctx *Context) error {
				opp, ok := firstLegalPlayerTarget(ctx)
				src := ctx.Source()
				if !ok || !onBattlefield(ctx.Game, src) || ctx.isNewSourceObject(src) {
					return nil
				}
				if err := (GainControl{
					Target:     src,
					Controller: opp,
					Duration:   game.IndefiniteDuration(),
					Label:      "Yes Man — target opponent gains control of it",
				}).Apply(ctx); err != nil {
					return err
				}
				if c, ok := ctx.Game.LookupCardForEffect(src); !ok || c.Controller != opp {
					return nil
				}
				return ReflexiveTrigger{
					Label: "Yes Man, Personal Securitron — you draw two cards and put a quest counter on Yes Man",
					Cards: []uuid.UUID{src},
					Body:  yesManBody,
				}.Apply(ctx)
			})),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisLeaves("Yes Man, Personal Securitron — Wild Card: its owner creates a tapped 1/1 white Soldier creature token for each quest counter on it",
				func(g *game.Game, item *game.StackItem) error {
					info, ok := g.LastKnownPermanentForEffect(item.SourceCardID)
					if !ok || info.Counters["quest"] <= 0 {
						return nil
					}
					owner := item.Controller
					if c, ok := g.LookupCardForEffect(item.SourceCardID); ok {
						owner = c.Owner
					}
					return CreateTokenAdvanced{Controller: owner, Spec: Token(TokenCard("1/1 white Soldier")).EntersTapped(), N: info.Counters["quest"]}.Apply(NewContext(g, item))
				}),
		},
	})
}
