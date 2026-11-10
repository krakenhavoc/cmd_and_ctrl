package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deepfathom Echo — Creature — Merfolk Spirit {2}{G}{U}, 4/4 (Lost Caverns
// of Ixalan, #4470):
//
//	"At the beginning of combat on your turn, this creature explores. Then
//	 you may have it become a copy of another creature you control until
//	 end of turn."
//
// One trigger, two halves in order: the explore (CR 701.44) and, once it has
// finished (its graveyard question answered), the optional copy. The second
// half is the explore's registered continuation (ExploresTimes.Then), so it
// waits for the explore's prompt and a table saved on it still resumes. The
// copy is a duration copy of a creature chosen as it happens (CR 608.2:
// "another creature you control", not a target), so hexproof does not stop
// it; the counter the explore put stays (a non-copy effect), and the copy
// has no trigger of its own to fire again until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c0d6c7e-a7cc-49e0-bf84-3fc33dd10b90",
		Name:         "Deepfathom Echo",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Deepfathom Echo — explores, then you may become a copy of another creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ExploresTimes{Explorer: item.SourceCardID, N: 1, Then: deepfathomEchoThen}.Apply(ctx)
				}),
		},
	})
}

// deepfathomEchoThen is the half of the trigger after the explore: offer the
// other creatures the controller has. Carry[0] is the Echo. An on-disk key:
// never renamed, never reused.
var deepfathomEchoThen = game.RegisterTail("card/deepfathom-echo-then", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	echo := o.Tail.Carry[0]
	var candidates []uuid.UUID
	echoHere := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == echo {
			echoHere = true
		} else if c.Controller == o.Tail.Player && c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if !echoHere || len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  o.Tail.Player,
		Source:   o.Tail.Source,
		Question: "Deepfathom Echo — you may have it become a copy of another creature you control until end of turn",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  deepfathomEchoCopy,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{echo},
	})
	return nil
})

// deepfathomEchoCopy is the keyed answer: the Echo becomes a copy of the pick.
var deepfathomEchoCopy = game.RegisterChooseCardsThen("choose-cards/deepfathom-echo-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	ctx := NewContext(g, &game.StackItem{Controller: r.Chooser, SourceCardID: r.Source})
	return BecomeCopy{
		Targets: []uuid.UUID{r.Carry[0]},
		Of:      r.Picked[0],
		Label:   "Deepfathom Echo — becomes a copy until end of turn",
	}.Apply(ctx)
})
