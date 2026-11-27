package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB396FateCounter is the counter kind Triad of Fates puts on creatures
// (Oblivion Stone's kind, spelled out here so this file stands on its own).
const OldB396FateCounter = "fate"

// OldB396HasFateCounter is "creature that has a fate counter on it".
func OldB396HasFateCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters[OldB396FateCounter] > 0 }
}

// OldB396PutFateCounterOnTarget is "put a fate counter on another target
// creature": the first target still legal as the ability resolves.
func OldB396PutFateCounterOnTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return AddCounter{Target: t.ID, Kind: OldB396FateCounter, N: 1}.Apply(ctx)
	}
	return nil
}

// OldB396ExileTargetItsControllerDraws is "exile target creature. Its
// controller draws N cards": the controller is read before the exile moves
// the creature (CR 608.2h), the exile comes first and the draw follows
// (CR 608.2c).
func OldB396ExileTargetItsControllerDraws(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			controller := c.Controller
			if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: controller, N: n}.Apply(ctx)
		}
		return nil
	}
}

// OldB396MonstrousMayAttackDespiteDefender is "as long as this creature is
// monstrous, it can attack as though it didn't have defender" (Colossus of
// Akros): a self-only layer-6 DefenderWaiver behind the Monstrous gate. The
// keyword stays (CR 609.4).
func OldB396MonstrousMayAttackDespiteDefender() game.StaticAbility {
	s := DefenderWaiverStatic("As long as this creature is monstrous, it can attack as though it didn't have defender.",
		func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && SelfOnly(target, g, source)
		})
	s.ActiveWhen = Monstrous()
	return s
}

// OldB396PolisCrusherTrigger is Polis Crusher's "Whenever this creature deals
// combat damage to a player, if this creature is monstrous, destroy target
// enchantment that player controls": the CR 603.4 intervening "if" is read as
// the creature last existed, as it triggers and again as it resolves.
func OldB396PolisCrusherTrigger() game.TriggeredAbility {
	monstrous := func(g *game.Game, source *game.Card) bool {
		info, ok := g.PermanentForEffect(game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch})
		return ok && info.Monstrous
	}
	t := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		Key:     "Polis Crusher — destroy target enchantment that player controls",
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return ev.Source == source.InstanceID && CombatDamageToPlayerBy(ev, source.Controller, g) && monstrous(g, source)
		},
		Targets: ControlledByThatPlayer(TargetPermanent("target enchantment that player controls", Enchantment())),
		Effect:  DestroyFirstLegalTarget,
	}
	return InterveningIf(CondBoard{Name: "this creature is monstrous", Test: func(ctx *Context) bool {
		info, ok := ctx.SourcePermanent()
		return ok && info.Monstrous
	}}, t)
}
