package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OpponentControlsColorPermanentB830 is "an opponent controls a <color>
// permanent" for source's controller (S82 OLD batch 830, #5636).
func OpponentControlsColorPermanentB830(g *game.Game, source *game.Card, color string) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != source.Controller && c.Controller != uuid.Nil && c.HasColor(color) {
			return true
		}
	}
	return false
}

// ScarabAuraSpecB830 is the Ice Age Scarab Aura: "Enchanted creature can't be
// blocked by <color> creatures. Enchanted creature gets +2/+2 as long as an
// opponent controls a <color> permanent."
func ScarabAuraSpecB830(oracleID, name, color, word string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(2, 2, func(g *game.Game, source *game.Card) int {
				if OpponentControlsColorPermanentB830(g, source, color) {
					return 1
				}
				return 0
			}),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnAttached(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.HasColor(color)
			}, word+" creatures"),
		},
	}
}

// PreventNextOneOnEnchantedB830 is "Prevent the next 1 damage that would be
// dealt to enchanted creature this turn" (Fylgja).
func PreventNextOneOnEnchantedB830(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := EnchantedCreatureOf(ctx)
	if host == uuid.Nil {
		return nil
	}
	return PreventNextDamage{Target: host, Amount: 1, Label: "Fylgja: prevent the next 1 damage"}.Apply(ctx)
}

// BoostFirstTargetB830 is "Target creature gets +P/+T until end of turn".
func BoostFirstTargetB830(name string, power, toughness int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, tg := range ctx.LegalTargets() {
			if tg.Kind == game.TargetCard {
				return BoostUntilEOT{Target: tg.ID, Power: power, Toughness: toughness, Label: name + " — " + boostWords(power, toughness) + " until end of turn"}.Apply(ctx)
			}
		}
		return nil
	}
}

// MindWhipB830 is Mind Whip's upkeep: the creature's controller may pay {3};
// if they don't, the Aura deals 2 damage to them and its controller taps the
// creature. The player and the creature ride the prompt as plain data.
func MindWhipB830(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	aura, ok := ctx.SourcePermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	host, ok := g.LookupCardForEffect(aura.AttachedTo.ID)
	if !ok || !OnBattlefield(g, host.InstanceID) {
		return nil
	}
	return UpkeepPayUnless{
		Chooser:  item.Trigger.Event.Actor,
		Cost:     "{3}",
		Question: "Mind Whip — pay {3}, or take 2 damage and have " + host.Name + " tapped?",
		Else:     mindWhipElseB830,
		Carry:    []uuid.UUID{item.Trigger.Event.Actor, host.InstanceID},
	}.Apply(ctx)
}

var mindWhipElseB830 = PayUnlessElse("pay-unless/old-b830-mind-whip",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) < 2 {
			return nil
		}
		if err := (DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: 2}).Apply(ctx); err != nil {
			return err
		}
		if !OnBattlefield(ctx.Game, r.Carry[1]) {
			return nil
		}
		return TapTarget{Target: r.Carry[1]}.Apply(ctx)
	})

// BlockingWithASnowLandB830 is Snow Devil's "as long as it's blocking and you
// control a snow land".
func BlockingWithASnowLandB830(host *game.Card, g *game.Game, source *game.Card) bool {
	if host.BlockingTarget == uuid.Nil {
		return false
	}
	return CountPermanentsYouControl(g, source, func(c *game.Card) bool {
		return c.IsLand() && c.HasSupertype("Snow")
	}) > 0
}

// CanOnlyAttackAloneAttached is "enchanted creature can only attack alone"
// (Errantry): CanOnlyAttackAlone bound to the creature the source is attached to.
func CanOnlyAttackAloneAttached() game.AttackLimit {
	l := CanOnlyAttackAlone()
	l.OnAttached = true
	return l
}

// SnowDevilFirstStrikeB830 is "Enchanted creature has first strike as long as
// it's blocking and you control a snow land". Whether the creature is blocking
// changes on a declared block and when it leaves combat (CR 506.4), so the
// layer cache is told to drop on both.
func SnowDevilFirstStrikeB830() game.StaticAbility {
	s := GrantToAttachedWhile(BlockingWithASnowLandB830, "first strike")
	s.DependsOnAttackingStatus = true
	s.DependsOnEvents = []game.EventKind{game.EventBlock}
	return s
}
