package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marooned — Enchantment — Aura {U} (Star Trek, #4973):
//
//	"Enchant artifact or creature
//	 When this Aura enters, you may pay {3}. If you do, tap enchanted
//	 permanent.
//	 Enchanted permanent doesn't untap during its controller's untap step
//	 and can't be beamed up."
//
// The entry trigger asks for the {3} as it resolves (CR 603.5's "you
// may", a CR 118.12 payment) and, paid, taps the permanent the Aura
// enchants as the trigger resolves, or the one it last enchanted if the
// Aura has left (CR 608.2h). That permanent rides the prompt with its
// epoch, so one that left and came back is a new object and is not
// tapped (CR 400.7).
//
// "Doesn't untap" is Claustrophobia's UntapStepRestriction. "Can't be
// beamed up" is AttachedCantBeBeamedUp (trek_b8_helpers.go): the host
// can't be returned to pay any beam me up cost, printed (Open
// Communications) or granted (Scotty, Miracle Worker). Both are read live
// from the Aura, so they end when it leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "74e02bdf-a074-4356-9d06-0df01dd2df66",
		Name:                  "Marooned",
		Completeness:          CompletenessFull,
		Targets:               TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Static:                []game.StaticAbility{AttachedCantBeBeamedUp()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Marooned — you may pay {3} to tap enchanted permanent", maroonedMayPayToTap),
		},
	})
}

// maroonedMayPayToTap is the entry trigger's body: the offer, carrying the
// permanent the Aura enchants (or last enchanted) as it resolves.
func maroonedMayPayToTap(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.TriggeringPermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	if !ok || !onBattlefield(g, host.InstanceID) {
		return nil
	}
	return MayPay{
		Chooser:      item.Controller,
		Cost:         "{3}",
		Question:     "Marooned — pay {3} to tap enchanted permanent?",
		Then:         maroonedPaid,
		CarryObjects: []game.ObjectRef{{ID: host.InstanceID, Epoch: host.ObjectEpoch}},
	}.Apply(ctx)
}

// maroonedPaid is the paid half, keyed (#4264): tap the carried permanent
// while it is still that object on the battlefield.
var maroonedPaid = PayUnlessBranches("pay-unless/marooned-tap-enchanted-permanent",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) == 0 {
			return nil
		}
		host := r.Objects[0]
		c, ok := ctx.Game.LookupCardForEffect(host.ID)
		if !ok || c.ObjectEpoch != host.Epoch || !onBattlefield(ctx.Game, host.ID) {
			return nil
		}
		return TapTarget{Target: host.ID}.Apply(ctx)
	}, nil)
