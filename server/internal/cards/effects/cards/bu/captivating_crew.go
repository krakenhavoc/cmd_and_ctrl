package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captivating Crew — Creature — Human Pirate {3}{R}, 4/3:
//
//	"{3}{R}: Gain control of target creature an opponent controls
//	 until end of turn. Untap that creature. It gains haste until end
//	 of turn. Activate only as a sorcery."
//
// A repeatable Act of Treason nailed to a 4/3 body — the same three
// primitives (#756) that card proves out, gated SorcerySpeed
// (CR 602.5d, Twitching Doll's shape for the flag) and targeted at
// "a creature an opponent controls" rather than any creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1049bc06-83de-4ed6-ae38-f259e5038a95",
		Name:         "Captivating Crew",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{R}: Gain control of target creature an opponent controls until end of turn. Untap that creature. It gains haste until end of turn. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{R}"),
			Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			SorcerySpeed: true,
			Effect:       captivatingCrewEffect,
		}},
	})
}

// captivatingCrewEffect is the resolution body: steal, untap, grant
// haste, in printed order.
//
// Caller holds g.mu.
func captivatingCrewEffect(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := effects.NewContext(g, item)
	target := item.Targets[0].ID
	if err := (effects.GainControl{
		Target:   target,
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Captivating Crew — gain control until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"haste"},
		Label:    "Captivating Crew — haste until end of turn",
	}.Apply(ctx)
}
