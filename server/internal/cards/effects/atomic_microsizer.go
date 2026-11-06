package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Atomic Microsizer — Artifact — Equipment {U} (Edge of Eternities,
// #4461):
//
//	"Equipped creature gets +1/+0.
//	 Whenever equipped creature attacks, choose up to one target
//	 creature. That creature can't be blocked this turn and has base
//	 power and toughness 1/1 until end of turn.
//	 Equip {2}"
//
// Argentum Armor's attack trigger (attachedCreatureAttacked, one
// trigger per declaration) with an "up to one" clause over any
// creature, the equipped one included. The pick gets Artful Dodge's
// "can't be blocked" restriction until the end of the turn and
// Perfected Theory's layer 7b base power and toughness of 1/1, so the
// equipped creature's own +1/+0 and any counters still apply on top
// (CR 613.4b, 613.4c): an equipped creature that picks itself attacks
// as an unblockable 2/1. A pick that is no longer a legal target as
// the trigger resolves gets neither.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "048c3b0b-48ef-42f5-a2af-90aa1d133296",
		Name:         "Atomic Microsizer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			Targets: TargetCreature("up to one target creature").WithCount(0, 1),
			Key:     "Atomic Microsizer — up to one target creature can't be blocked this turn and has base power and toughness 1/1 until end of turn",
			Effect:  eoeABMicrosize,
		}},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}

// eoeABMicrosize makes the chosen creature unblockable this turn and a
// 1/1 at base until end of turn.
func eoeABMicrosize(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (RestrictUntilEOT{
			Target:       t.ID,
			Restrictions: game.CantBeBlocked,
			Label:        "Atomic Microsizer — can't be blocked this turn",
		}).Apply(ctx); err != nil {
			return err
		}
		return untilEndOfTurn(ctx, t.ID, nil,
			"Atomic Microsizer — base power and toughness 1/1 until end of turn",
			game.SetBasePTMods(1, 1)...)
	}
	return nil
}
