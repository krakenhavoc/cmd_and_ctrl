package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Savior of Ollenbock — Creature — Human Soldier {1}{W}{W}, 1/2 (Crimson Vow):
//
//	"Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)
//	 Whenever this creature trains, exile up to one other target creature from
//	 the battlefield or creature card from a graveyard.
//	 When this creature leaves the battlefield, put the exiled cards onto the
//	 battlefield under their owners' control."
//
// "Trains" is CR 702.149c: a resolving training ability put a +1/+1 counter on
// this creature. It is modelled as a CR 603.12 reflexive trigger the training
// ability creates once the counter has gone on (vowB09TrainingThen, built like
// exploit.go's), which chooses its up-to-one target when it goes on the stack.
// The exile is linked to this object (ExileTarget.Linked, ADR 0155 §8), so the
// leave trigger returns every card still in exile that this object exiled, as
// one entry, under its owner's control; a token exiled this way is gone for
// good. A Savior that comes back is a new object with no link (CR 400.7), and
// one that leaves before the exile resolves exiles for good (the ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1857dc32-8ff1-4519-b1ca-480a5c2d3b3c",
		Name:         "Savior of Ollenbock",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.VowB09TrainingThen("Savior of Ollenbock — training: put a +1/+1 counter on it", saviorOfOllenbockExileBody),
			effects.On(game.EventLTB, effects.Self, "Savior of Ollenbock — put the exiled cards onto the battlefield under their owners' control", effects.ReturnAllExiledWithThisTogether),
		},
	})
}

// saviorOfOllenbockExileBody is "whenever this creature trains, exile up to
// one other target creature from the battlefield or creature card from a
// graveyard".
var saviorOfOllenbockExileBody = game.ReflexiveBody("savior-of-ollenbock/exile", effects.SimpleBody(saviorOfOllenbockExile),
	func(sourceID uuid.UUID, _ game.EffectParams) *game.TargetSpec {
		spec := effects.TargetCreature("up to one other target creature or creature card in a graveyard").WithCount(0, 1)
		spec.ExcludeSource = true
		spec.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
		return spec
	})

func saviorOfOllenbockExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == item.SourceCardID {
			continue
		}
		return effects.ExileTarget{Target: t.ID, Linked: true}.Apply(ctx)
	}
	return nil
}
