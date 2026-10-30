package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Secret Invasion — Enchantment — Aura {1}{U}{U} (Marvel Super Heroes,
// #4455):
//
//	"Enchant creature you control
//	 When this Aura enters, exile up to one target creature other than
//	 enchanted creature until this Aura leaves the battlefield. Enchanted
//	 creature becomes a copy of that creature until this Aura leaves the
//	 battlefield.
//	 Enchanted creature has ward {2}."
//
// "Enchant creature you control" is EnchantCreature(YouControl()). The
// "other than enchanted creature" is built per trigger (TargetsFrom, which
// is handed the Aura and so its host). One resolution does both halves, in printed order of
// consequence: the copiable values are read from the creature on the
// battlefield, as it is now (CR 707.2), the copy begins on the host with the
// Aura as its source and lasts as long as the Aura stays on the battlefield
// (a copy of nothing, or an Aura that left and came back, begins nothing),
// and the creature is exiled until the Aura leaves (CR 610.3, Oblivion
// Ring's record, which returns it as a new object under its owner's
// control). Ward {2} is an ability the Aura gives its host, which a copy
// effect (layer 1) does not take away.
//
// No simplification.
func init() {
	const label = "Secret Invasion — exile up to one target creature other than enchanted creature, and enchanted creature becomes a copy of it"
	Register(Spec{
		OracleID:     "25443ea4-3177-4e29-a2b1-f3fee60b3205",
		Name:         "Secret Invasion",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       label,
				TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
					return TargetCreature("up to one target creature other than enchanted creature",
						NotSelf(source.AttachedTo.ID)).WithCount(0, 1)
				},
				Effect: secretInvasionResolve,
			},
			WardAttached(WardMana("{2}"), "Secret Invasion — enchanted creature has ward {2}"),
		},
	})
}

// secretInvasionResolve is the enter trigger's resolution.
func secretInvasionResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := firstLegalCardTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok || host == target {
		return nil
	}
	aura := ctx.Source()
	d, ok := DurationWhileSourceRemains(ctx, aura)
	if !ok {
		return nil // the Aura left (and maybe came back) first: nothing begins
	}
	// The copy reads the creature as it is on the battlefield, before it
	// is exiled (CR 707.2, last-known information).
	if v, ok := g.CopiableValuesForEffect(target); ok {
		g.BecomeCopyForEffect(aura, target, []uuid.UUID{host}, v, d, "Secret Invasion — enchanted creature becomes a copy")
	}
	return ExileUntil{
		Target:     target,
		ThisLeaves: true,
		Label:      "Secret Invasion — the exiled creature returns when Secret Invasion leaves the battlefield",
	}.Apply(ctx)
}
