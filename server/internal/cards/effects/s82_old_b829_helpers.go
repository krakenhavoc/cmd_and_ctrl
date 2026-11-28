package effects

import (
	"slices"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OpponentControlsColorPermanent reports "an opponent controls a <color>
// permanent" for source's controller (S82 OLD batch 829, #5636).
func OpponentControlsColorPermanentB829(g *game.Game, source *game.Card, color string) bool {
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

// ScarabAuraSpecB829 is the Ice Age Scarab Aura: "Enchanted creature can't be
// blocked by <color> creatures. Enchanted creature gets +2/+2 as long as an
// opponent controls a <color> permanent."
func ScarabAuraSpecB829(oracleID, name, color, word string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(2, 2, func(g *game.Game, source *game.Card) int {
				if OpponentControlsColorPermanentB829(g, source, color) {
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

// PreventDamageToAttachedFromChosenColorB829 is "Prevent all damage that would
// be dealt to enchanted creature by sources of the chosen color": a CR 615.1a
// prevention whose source colour is read as the damage would be dealt, and
// whose colour is the Aura's own chosen colour.
func PreventDamageToAttachedFromChosenColorB829(label string) game.ReplacementEffect {
	return PreventAllDamageDealtTo(PreventionStatic{
		To: ToEquippedCreature,
		From: func(_ *game.ReplacementEvent, _ *game.Game, lki *game.Characteristic, src *game.Card) bool {
			return lki != nil && src.ChosenColor != "" && slices.Contains(lki.Colors, src.ChosenColor)
		},
		Label: label,
	})
}

// EssenceFlareCounterB829 puts a -0/-1 counter on the creature the Aura
// enchants as the trigger resolves.
func EssenceFlareCounterB829(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := EnchantedCreatureOf(ctx)
	if host == uuid.Nil || !OnBattlefield(g, host) {
		return nil
	}
	return AddCounter{Target: host, Kind: "-0/-1", N: 1}.Apply(ctx)
}

// DamageEnchantedControllerUnlessPayB829 is "this Aura deals N damage to that
// creature's controller unless that player pays <cost>" (Seizures). The payer
// rides the prompt as plain data.
func DamageEnchantedControllerUnlessPayB829(name string, n int, cost string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if item == nil || item.Trigger == nil {
			return nil
		}
		c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
		if !ok {
			return nil
		}
		p := g.PlayerByIDForEffect(c.Controller)
		if p == nil || p.Eliminated {
			return nil
		}
		return PayUnless{
			Chooser:  p.ID,
			Cost:     cost,
			Question: name + " — pay " + cost + ", or take " + strconv.Itoa(n) + " damage?",
			Else:     seizuresDamage,
			Carry:    []uuid.UUID{p.ID},
		}.Apply(NewContext(g, item))
	}
}

var seizuresDamage = PayUnlessElse("pay-unless/old-b829-seizures-damage",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		return DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: 3}.Apply(ctx)
	})

// ErrantMinionPaymentB829 is Errant Minion's upkeep: its victim may pay any
// amount of mana; the Aura deals 2 damage to them and X of it is prevented.
func ErrantMinionPaymentB829(g *game.Game, item *game.StackItem) error {
	if item == nil || item.Trigger == nil {
		return nil
	}
	return MayPayX{
		Chooser: item.Trigger.Event.Actor,
		Symbol:  "{1}",
		Label:   "Errant Minion",
		Buys:    "prevent X of the 2 damage this Aura deals to you",
		Goal:    func(_ *Context, ceiling int) int { return min(ceiling, 2) },
		Unit:    game.PayAmountDamage,
		Then:    errantMinionDamage,
		Carry:   []uuid.UUID{item.Trigger.Event.Actor},
	}.Apply(NewContext(g, item))
}

var errantMinionDamage = PayUnlessBranches("pay-unless/old-b829-errant-minion-damage",
	func(ctx *Context, r game.PayUnlessAnswer) error { return errantMinionHit(ctx, r, r.Number) },
	func(ctx *Context, r game.PayUnlessAnswer) error { return errantMinionHit(ctx, r, 0) })

func errantMinionHit(ctx *Context, r game.PayUnlessAnswer, prevented int) error {
	if len(r.Carry) == 0 {
		return nil
	}
	dmg := 2 - prevented
	if dmg <= 0 {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: dmg}.Apply(ctx)
}

// EndStepOfEnchantedCreaturesControllerB829 is the When of "At the beginning of
// the end step of enchanted creature's controller".
func EndStepOfEnchantedCreaturesControllerB829(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	host := g.AttachedHostOf(source)
	return host != nil && ev.Kind == game.EventBeginEndStep && ev.Actor != uuid.Nil && ev.Actor == host.Controller
}

// AggressionDestroyB829 destroys the creature the Aura enchants if it didn't
// attack this turn.
func AggressionDestroyB829(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host := EnchantedCreatureOf(ctx)
	if host == uuid.Nil || !OnBattlefield(g, host) || g.TimesAttackedThisTurn(host) > 0 {
		return nil
	}
	return DestroyTarget{Target: host}.Apply(ctx)
}
