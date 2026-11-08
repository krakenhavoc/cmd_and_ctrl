package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Due Diligence — Enchantment — Aura {2}{W} (Murders at Karlov Manor,
// #4469):
//
//	"Enchant creature
//	 When this Aura enters, target creature you control other than
//	 enchanted creature gets +2/+2 and gains vigilance until end of turn.
//	 Enchanted creature gets +2/+2 and has vigilance."
//
// The statics are Lead Pipe's PumpAttached and a GrantToAttached.
// "Other than enchanted creature" is Secret Invasion's clause, built as
// the trigger goes on the stack from the creature the Aura enchants then
// (TargetsFrom). With no other creature you control the trigger has no
// legal target and is removed (CR 603.3d). If the Aura has become
// attached to the target by the time the trigger resolves, it does
// nothing (the ruling); otherwise the target gets Get a Leg Up's
// +2/+2 and a vigilance grant until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d412150d-3c70-44cf-97cd-965c438e7303",
		Name:         "Due Diligence",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			GrantToAttached("vigilance"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       "Due Diligence — target creature you control other than enchanted creature gets +2/+2 and vigilance until end of turn",
			TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
				return TargetCreature("target creature you control other than enchanted creature",
					YouControl(), NotSelf(source.AttachedTo.ID))
			},
			Effect: dueDiligencePump,
		}},
	})
}

// dueDiligencePump gives the target +2/+2 and vigilance until end of
// turn, unless the Aura now enchants it.
func dueDiligencePump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if host, ok := enchantedCreatureAsItResolves(ctx); ok && host == target {
		return nil
	}
	return pumpAndGrantFirstTarget(2, 2, "Due Diligence — +2/+2 and vigilance until end of turn", "vigilance")(g, item)
}
