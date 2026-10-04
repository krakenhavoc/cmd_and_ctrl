package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pain for All — Enchantment — Aura {2}{R}:
//
//	"Enchant creature you control
//	 When this Aura enters, enchanted creature deals damage equal to
//	 its power to any other target.
//	 Whenever enchanted creature is dealt damage, it deals that much
//	 damage to each opponent."
//
// Two triggers. The enters trigger is targeted: its clause is built
// per source (TargetsFrom) as "any target" minus the creature this
// Aura enchants, so the host is never a legal pick. That reads the
// source's own attachment, which cannot change while the trigger
// waits on the stack. The Aura is read at resolution as it is, or as
// it last was if it has left, so "enchanted creature" still names its
// host, and the host's power is read then too, as it is or as it last
// was (CR 608.2h). The damage is the host's.
//
// The reflect trigger is Screaming Nemesis's and Brash Taunter's shape
// aimed at each opponent of the Aura's controller: no target, one
// simultaneous batch of damage from the enchanted creature, each
// opponent hit for the amount the creature was dealt.
//
// One declared simplification, weaker than printed and the same one
// Screaming Nemesis carries: the engine emits one damage event per
// SOURCE, so an enchanted creature blocked by two creatures reflects
// twice, once per blocker's damage, where the printed card reflects
// once for the total. The same damage is dealt either way; only the
// split differs.
func init() {
	const etbLabel = "Pain for All — enchanted creature deals damage equal to its power to any other target"
	Register(Spec{
		OracleID:     "4eedf21c-0ad1-48da-a9fe-2ffdbf8371e9",
		Name:         "Pain for All",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If two or more sources damage the enchanted creature at the same time, it reflects each source's damage separately instead of the total in one go."},
		Targets:      EnchantCreature(YouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       etbLabel,
				TargetsFrom: func(_ game.TriggerContext, source *game.Card, _ *game.Game) *game.TargetSpec {
					spec := b35TargetAnyOther()
					host := source.AttachedTo.ID
					anyOK := spec.CardOK
					spec.CardOK = func(g *game.Game, caster uuid.UUID, c game.Card, z game.ZoneKind) bool {
						return c.InstanceID != host && anyOK(g, caster, c, z)
					}
					return spec
				},
				Effect: painForAllEnchantedCreatureBites,
			},
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Amount > 0 && source.AttachedTo.Kind == game.TargetCard && source.AttachedTo.ID == ev.Target
				},
				Key:    "Pain for All — enchanted creature deals that much damage to each opponent",
				Effect: painForAllReflect,
			},
		},
	})
}

// painForAllEnchantedCreatureBites is the enters trigger's body: the
// creature the Aura enchants deals damage equal to its power to the
// announced target.
func painForAllEnchantedCreatureBites(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	aura, ok := g.PermanentForEffect(ref)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host := aura.AttachedTo.ID
	power := b43PowerNowOrLastKnown(g, host)
	if power <= 0 {
		return nil
	}
	targets := ctx.LegalTargets()
	if len(targets) == 0 || (targets[0].Kind == game.TargetCard && targets[0].ID == host) {
		return nil
	}
	return DealDamage{Source: host, Target: targets[0].ID, Amount: power}.Apply(ctx)
}

// painForAllReflect is the second trigger's body: the damaged creature,
// named by the triggering event, deals that much to each opponent of the
// Aura's controller at once.
func painForAllReflect(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	ctx := NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := (DealDamage{Source: ev.Target, Target: opp, Amount: ev.Amount}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
