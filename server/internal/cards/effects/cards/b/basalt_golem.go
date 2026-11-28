package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Basalt Golem — Artifact Creature — Golem {5}, 2/4 (Mirage, #5633):
//
//	"This creature can't be blocked by artifact creatures.
//	 Whenever this creature becomes blocked by a creature, that creature's
//	 controller sacrifices it at end of combat. If the player does, they
//	 create a 0/2 colorless Wall artifact creature token with defender."
//
// The block rule is Argothian Sprite's. The trigger fires once for each
// creature that blocks the Golem (CR 509.3d) and schedules, for the beginning
// of the end of combat step (CR 511.2), the blocker's sacrifice by whoever
// controls it then. The Wall is created by that player only if the sacrifice
// really happened (a blocker that left already is not sacrificed, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9155f78e-0ccc-4e08-ae79-ef5b2db1183f",
		Name:         "Basalt Golem",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.Artifact()), "artifact creatures"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, OldB813BecameBlocked, OldB813BasaltGolemLabel, OldB813SacrificeTheBlockerAtEndOfCombat),
		},
	})
}

// OldB813BasaltGolemLabel is the trigger's stack label.
const OldB813BasaltGolemLabel = "Basalt Golem — that creature's controller sacrifices it at end of combat"

// OldB813BecameBlocked is "whenever this creature becomes blocked by a
// creature": the block pair names the source as the attacker.
func OldB813BecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBlock || ev.Target != source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// OldB813SacrificeTheBlockerAtEndOfCombat schedules the sacrifice.
func OldB813SacrificeTheBlockerAtEndOfCombat(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	blocker := item.Trigger.Event.CardID
	if blocker == uuid.Nil {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: OldB813BasaltGolemLabel,
		Cards: []uuid.UUID{blocker},
		Body:  oldB813SacrificeBody,
	}.Apply(effects.NewContext(g, item))
}

// oldB813SacrificeBody is the delayed half: the creature's controller now
// sacrifices it, and "if the player does" they create the Wall. An on-disk key.
var oldB813SacrificeBody = game.SimpleDelayedBody("old-b813/basalt-golem-sacrifice-the-blocker", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		if err := (effects.SacrificePermanent{Target: t.ID, ThenTail: effects.WithCarry(ctx.ItemTail(oldB813BasaltGolemWall), c.Controller)}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// oldB813BasaltGolemWall is "if the player does, they create a 0/2 Wall":
// the sacrificer rides as Carry[0]; the tail is told the permanent only when
// it really left.
var oldB813BasaltGolemWall = effects.CardTail("then/old-b813-basalt-golem-wall", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.CreateToken{Controller: o.Tail.Carry[0], Template: effects.TokenCard("0/2 colorless Wall artifact with defender"), N: 1}.Apply(ctx)
})
