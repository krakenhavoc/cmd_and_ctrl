package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB630UpkeepIfYouControlEitherTargeting is "At the beginning of your
// upkeep, if you control a <colour a> or <colour b> permanent, target creature
// ..." (Ana Sanctuary, S82 OLD batch 630, #5614): the CR 603.4 intervening if is
// read as the trigger is put on the stack and again as it resolves, and the
// target is chosen as it goes on the stack (CR 603.3d). The effect is told
// whether the controller also controls both colours as it resolves, for the
// "if you control a <a> permanent and a <b> permanent, ... instead" rider.
func OldB630UpkeepIfYouControlEitherTargeting(letterA, letterB, name, label string, targets *game.TargetSpec, effect func(g *game.Game, item *game.StackItem, both bool) error) game.TriggeredAbility {
	either := func(g *game.Game, you uuid.UUID) bool {
		return B18ControlsPermanentOfColor(g, you, letterA) || B18ControlsPermanentOfColor(g, you, letterB)
	}
	return InterveningIfYou(name, either, Targeting(AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		both := B18ControlsPermanentOfColor(g, item.Controller, letterA) && B18ControlsPermanentOfColor(g, item.Controller, letterB)
		return effect(g, item, both)
	}), targets))
}

// OldB630PumpPerOtherAttackingKavu is "Whenever this creature attacks, it gets
// +1/+1 until end of turn for each other attacking Kavu" (Kavu Mauler): the
// attacking Kavu other than the source, counted as the trigger resolves,
// through the layered view.
func OldB630PumpPerOtherAttackingKavu(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		n := 0
		// "Kavu" is a word of the source's text (CR 612.2, #3984).
		kavu := g.TextReaderForEffect(item).TextSubtype("Kavu")
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != item.SourceCardID && c.IsCreature() && c.HasSubtype(kavu) && c.AttackingTarget != uuid.Nil {
				n++
			}
		}
		if n == 0 {
			return nil
		}
		return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n, Label: label}.Apply(NewContext(g, item))
	}
}

// OldB630DiscipleBoostRow is "<mana>, {T}: Target creature gets <+/-N>/-0 until
// end of turn" (the Apocalypse Disciples). {T} is in the cost, so summoning
// sickness applies (CR 302.6); a target gone by resolution gets nothing
// (CR 608.2b).
func OldB630DiscipleBoostRow(name, mana string, power int) ActivatedAbility {
	size := "+2/+0"
	purpose := ForTargets(BenefitUntilEOT(0, power, 0))
	if power < 0 {
		size = "-2/-0"
		purpose = ForTargets(HarmUntilEOT(0, power, 0))
	}
	return ActivatedAbility{
		Label:   mana + ", {T}: Target creature gets " + size + " until end of turn.",
		Purpose: purpose,
		Cost:    Plus(ManaCost(mana), TapCost()),
		Targets: TargetCreature("target creature"),
		Effect:  BoostTargetUntilEOT(name+" — "+size, power, 0),
	}
}

// OldB630DiscipleFlyingRow is "<mana>, {T}: Target creature gains flying until
// end of turn" (Ana Disciple).
func OldB630DiscipleFlyingRow(name, mana string) ActivatedAbility {
	return ActivatedAbility{
		Label:   mana + ", {T}: Target creature gains flying until end of turn.",
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
		Cost:    Plus(ManaCost(mana), TapCost()),
		Targets: TargetCreature("target creature"),
		Effect:  EbTargetCreatureGainsUntilEOT("flying", name+" — flying until end of turn"),
	}
}
