package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Luminate Primordial — Creature — Avatar {5}{W}{W}, 4/7 (Gatecrash, S82 OLD
// b408, #5565):
//
//	"Vigilance
//	 When this creature enters, for each opponent, exile up to one target
//	 creature that player controls and that player gains life equal to its
//	 power."
//
// Havoc Eater's per-opponent clauses (PerOpponentClauses): one "up to one
// target creature" for each opponent, chosen as the trigger goes on the stack
// (CR 603.3d). Each target still legal as it resolves (CR 608.2b) is exiled and
// its controller gains life equal to the power it had as it last existed
// (CR 608.2h), whether or not the exile worked (the sentence has no "if you do").
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Luminate Primordial — for each opponent, exile up to one target creature that player controls; that player gains life equal to its power",
		luminatePrimordialEnters)
	trigger.Targets = effects.PerOpponentClauses(game.MaxPlayers-1, "creature", effects.Creature())
	effects.Register(effects.Spec{
		OracleID:        "c1bc819a-f291-4b55-94f7-269f160ce76b",
		Name:            "Luminate Primordial",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}

func luminatePrimordialEnters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	type exiled struct {
		controller uuid.UUID
		power      int
	}
	var done []exiled
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || !effects.OnBattlefield(g, t.ID) {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		done = append(done, exiled{controller: c.Controller, power: c.CurrentPower()})
		if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, e := range done {
		if e.power <= 0 {
			continue
		}
		if err := (effects.GainLife{Player: e.controller, Amount: e.power}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
