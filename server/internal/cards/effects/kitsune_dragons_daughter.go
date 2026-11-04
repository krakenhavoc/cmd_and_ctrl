package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitsune, Dragon's Daughter — Legendary Creature — Fox Warlock Avatar
// {4}{U}{U}, 6/6 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Vigilance
//	 Whenever Kitsune enters or deals combat damage to a player, you may
//	 exchange control of two other target creatures controlled by
//	 different players."
//
// One ability with two conditions (WhenThisEntersOrAttacks' shape). The
// two targets are one clause of two creatures other than Kitsune with
// Run Away Together's set rule, "controlled by different players"
// (EachDifferentController), judged as the trigger goes on the stack and
// again at resolution; with no such pair the trigger is removed (CR
// 603.3d). The "you may" is asked as it resolves (CR 603.5). The exchange
// is Switcheroo's: all or nothing (CR 701.12a), so it does nothing when
// either creature is no longer a legal target, and it lasts indefinitely.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9be19bdf-db1e-433e-8be4-8b985a568211",
		Name:            "Kitsune, Dragon's Daughter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			kitsuneTrigger(),
		},
	})
}

// kitsuneTrigger is the enters-or-combat-damage trigger.
func kitsuneTrigger() game.TriggeredAbility {
	t := OnAny([]game.EventKind{game.EventETB, game.EventDealDamage},
		func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
			if ev.Kind == game.EventETB {
				return Self(ev, source, ch, g)
			}
			return ThisDealtCombatDamageToAPlayer(ev, source, ch, g)
		},
		"Kitsune, Dragon's Daughter — you may exchange control of two other target creatures controlled by different players",
		kitsuneExchange)
	t.OptionalPrompt = &game.TriggerOptionalPrompt{
		Question: "Kitsune, Dragon's Daughter — exchange control of the two creatures?",
	}
	t.Targets = Another(TargetCreature("two other target creatures controlled by different players").
		WithCount(2, 2).EachDifferent(EachDifferentController()))
	return t
}

// kitsuneExchange exchanges control of the two targets when both are
// still legal.
func kitsuneExchange(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) != 2 {
		return nil
	}
	return ExchangeControl{
		A:     legal[0].ID,
		B:     legal[1].ID,
		Label: "Kitsune, Dragon's Daughter — exchange control",
	}.Apply(ctx)
}
