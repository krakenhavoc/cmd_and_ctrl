package effects

import (
	"fmt"
	"slices"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b664_helpers.go — shared pieces of Invasion batch 664 (#5616).

// OldB664CantAttackUnlessACreatureOfColorAlsoAttacks is "This creature can't
// attack unless a <color> or <color> creature also attacks" (Scarred Puma;
// CR 508.1c): Okk's companion limit with the companion narrowed to the colors
// named (game.AttackLimit.OthersColors), read through the layers at
// declaration. `colors` are the letters ("B", "G").
func OldB664CantAttackUnlessACreatureOfColorAlsoAttacks(colors ...string) game.AttackLimit {
	return game.AttackLimit{Scope: game.AttackLimitNotAlone, Max: 1, OthersColors: slices.Clone(colors)}
}

// OldB664CostsMoreMana is "<matching> spells you cast cost <unit> more to
// cast" for an increase that is colored mana (Jade Leech's {G}, Alabaster
// Leech's {W}): CostsMore with a Unit, so the symbol is added to the cost as
// that color (CR 601.2f). Panics at init on a string the cost parser rejects or
// that carries {X}.
func OldB664CostsMoreMana(unit, label string, when ...CostPredicate) game.CostModifier {
	u, err := game.ParseCost(unit)
	if err != nil || u.XSlots > 0 || u.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.OldB664CostsMoreMana: %q is not a fixed mana amount (%v)", unit, err))
	}
	m := CostsMore(1, label, when...)
	m.Unit = &u
	return m
}

// OldB664CreaturesOfColorOnTheBattlefield counts the creatures on the
// battlefield, whoever controls them, that are `color` ("R", "G"), after the
// layers. A creature of two colors counts once.
func OldB664CreaturesOfColorOnTheBattlefield(g *game.Game, color string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if !c.IsCreature() {
			continue
		}
		for _, have := range c.EffectiveColors() {
			if strings.EqualFold(have, color) {
				n++
				break
			}
		}
	}
	return n
}

// OldB664NoOpponentControlsACreatureOfColor is the condition of Kavu Runner's
// "as long as no opponent controls a white or blue creature", read as the
// layers run with `source` the Kavu: no creature controlled by anyone but
// the source's controller has one of `colors`.
func OldB664NoOpponentControlsACreatureOfColor(g *game.Game, source *game.Card, colors ...string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == source.Controller || !c.IsCreature() {
			continue
		}
		for _, have := range c.EffectiveColors() {
			for _, want := range colors {
				if strings.EqualFold(have, want) {
					return false
				}
			}
		}
	}
	return true
}

// OldB664BlockedByAGreenCreature is the trigger condition of "Whenever this
// creature becomes blocked by a green creature" (CR 509.3d: once for each
// creature that blocks it): the block event names the blocker in CardID and the
// attacker in Target, and the blocker is green as the layers have it.
func OldB664BlockedByAGreenCreature(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventBlock || ev.Target != source.InstanceID || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasColor(source.TextColor("G"))
}

// OldB664DestroyTheBlockerNoRegen destroys the creature named by the block
// event, which can't be regenerated, as long as it is still on the battlefield.
func OldB664DestroyTheBlockerNoRegen(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	if !OnBattlefield(g, id) {
		return nil
	}
	return DestroyTarget{Target: id, CantBeRegenerated: true}.Apply(NewContext(g, item))
}
