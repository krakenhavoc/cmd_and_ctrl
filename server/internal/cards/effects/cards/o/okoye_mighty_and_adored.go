package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Okoye, Mighty and Adored — Legendary Creature — Human Warrior Hero
// {2}{G}{W}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"When Okoye enters, you become the monarch.
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature. Whenever that creature attacks the monarch this
//	 turn, it gains double strike and trample until end of turn."
//
// The second sentence is a delayed triggered ability (CR 603.7) that the
// beginning-of-combat trigger schedules for the creature it targeted. It
// fires for that object only (a creature that left and came back is a
// new object and never fires it), on every attack it makes this turn
// (Repeats), and only when the attacked player is the monarch AT THE
// MOMENT OF THE ATTACK: the condition reads the crown when the creature
// is declared as an attacker. Attacking a planeswalker or a battle
// controlled by the monarch is not attacking the monarch.
//
// No simplification.
func init() {
	const label = "Okoye, Mighty and Adored — put a +1/+1 counter on target creature"
	effects.Register(effects.Spec{
		OracleID:     "8fbfc5b1-2146-4d07-a5fe-5f025da31bd6",
		Name:         "Okoye, Mighty and Adored",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Okoye, Mighty and Adored"),
			effects.Targeting(effects.AtYourStep(game.StepBeginCombat, label, okoyeCounterAndWatch), effects.TargetCreature("target creature")),
		},
	})
}

var (
	// "Whenever that creature attacks the monarch": the one creature
	// CondParams.Object names, declared as an attacker of the player who
	// holds the crown.
	okoyeAttacksTheMonarchCondition = game.DelayedEachCondition(
		"attack/that-creature-attacks-the-monarch", okoyeAttacksTheMonarch)

	// "It gains double strike and trample until end of turn."
	okoyeGrantBody = game.SimpleDelayedBody("grant/double-strike-and-trample-until-eot", okoyeGrantDoubleStrikeAndTrample)
)

func okoyeAttacksTheMonarch(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
	if ev.Kind != game.EventAttack || p.Object.ID == uuid.Nil || ev.CardID != p.Object.ID {
		return nil
	}
	c, ok := effects.BattlefieldCardForEffect(g, p.Object.ID)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !effects.YoureTheMonarch(g, ev.Target) {
		return nil
	}
	return []uuid.UUID{c.InstanceID}
}

func okoyeGrantDoubleStrikeAndTrample(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	it := effects.FirstPayloadCard(ctx)
	if it == uuid.Nil {
		return nil
	}
	return effects.GrantKeywordUntilEOT{
		Target:   it,
		Keywords: []string{"double strike", "trample"},
		Label:    "Okoye, Mighty and Adored — double strike and trample until end of turn",
	}.Apply(ctx)
}

// okoyeCounterAndWatch puts the counter on the target and schedules the
// attack watcher for that same creature.
func okoyeCounterAndWatch(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	c, ok := effects.BattlefieldCardForEffect(g, id)
	if !ok {
		return nil
	}
	return effects.DelayedOnEvent{
		Label:      "Okoye, Mighty and Adored — it gains double strike and trample",
		On:         []game.EventKind{game.EventAttack},
		Condition:  okoyeAttacksTheMonarchCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
		Body:       okoyeGrantBody,
		Repeats:    true,
	}.Apply(ctx)
}
