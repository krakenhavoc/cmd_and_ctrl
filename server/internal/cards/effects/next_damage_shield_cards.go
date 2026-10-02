package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// next_damage_shield_cards.go — the shared rows of the ADR 0107 §6
// cards (#1860): the Circle of Protection and Rune of Protection
// families, and the one-ability shield every other card of the seam
// prints. Append-only, so the clone gate never sees the same body twice.

// nextDamageShieldRow is an activated ability whose whole effect is a
// next-damage shield.
func nextDamageShieldRow(label string, cost game.AbilityCost, targets *game.TargetSpec, shield PreventNextDamageFromSource) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: targets,
		Effect: func(g *game.Game, item *game.StackItem) error {
			return shield.Apply(NewContext(g, item))
		},
	}
}

// nextDamageShieldSpell is an instant's OnResolve whose whole effect is
// a next-damage shield.
func nextDamageShieldSpell(shield PreventNextDamageFromSource) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return shield.Apply(ctx)
	}
}

// circleOfProtection is the Circle of Protection enchantment: one row,
// "<cost>: The next time <what> of your choice would deal damage to you
// this turn, prevent that damage." `label` is the row as printed and
// `what` the source it names ("a red source"). The Runes add Cycling {2}.
func circleOfProtection(oracle, name, cost, label, what string, q game.PermanentQuery, more ...ActivatedAbility) Spec {
	shield := PreventNextDamageFromChosenSource(ShieldYou, q)
	shield.Question = name + " — choose " + what
	shield.Label = name + " — prevent the next damage from " + what
	return Spec{
		OracleID:     oracle,
		Name:         name,
		Completeness: CompletenessFull,
		Activated:    append([]ActivatedAbility{nextDamageShieldRow(label, ManaCost(cost), nil, shield)}, more...),
	}
}

// runeOfProtection is circleOfProtection with the Runes' {W} row and
// "Cycling {2}".
func runeOfProtection(oracle, name, label, what string, q game.PermanentQuery) Spec {
	return circleOfProtection(oracle, name, "{W}", label, what, q, Cycling("{2}"))
}

// chosenColorShieldRow is "<cost>: The next time a source of your choice
// of the chosen color would deal damage to you this turn, prevent that
// damage" (Story Circle, Prismatic Circle). The colour is the one chosen
// as the enchantment entered, read as it last existed if the enchantment
// has gone by resolution (CR 608.2h). Before a colour is chosen there is
// no source "of the chosen color", so nothing is shielded.
func chosenColorShieldRow(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label: label,
		Cost:  cost,
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			info, ok := ctx.SourcePermanent()
			if !ok || info.ChosenColor == "" {
				return nil
			}
			return PreventNextDamageFromChosenSource(ShieldYou, QueryColors(info.ChosenColor)).Apply(ctx)
		},
	}
}

// chosenTypeShieldRow is "<cost>: The next time a creature of the chosen
// type would deal damage to you this turn, prevent that damage" (Circle
// of Solace). Nothing is chosen at resolution: the first creature of
// that type to deal damage to you is the one (CR 615.8, 615.9), judged
// as it deals the damage. The type is read as for chosenColorShieldRow.
func chosenTypeShieldRow(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label: label,
		Cost:  cost,
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			info, ok := ctx.SourcePermanent()
			if !ok || info.NamedTribe == "" {
				return nil
			}
			return PreventNextDamageFromSource{
				Queries: []game.PermanentQuery{{Types: []string{"creature"}, Subtypes: []string{info.NamedTribe}}},
				Protect: ShieldYou,
			}.Apply(ctx)
		},
	}
}

// shieldAgainstTargetCreature is "The next time target creature would
// deal damage this turn, prevent that damage" (Awe Strike, Dazzling
// Reflection): the source is the targeted creature, the shield protects
// whatever it would deal the damage to. With gainPower, "You gain life
// equal to target creature's power" comes first, read as the spell
// resolves.
func shieldAgainstTargetCreature(then game.BodyRef, gainPower bool) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if gainPower {
				ctx.Game.RecomputeLayersIfStaleLocked()
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok && c.CurrentPower() > 0 {
					if err := (GainLife{Amount: c.CurrentPower()}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return PreventNextDamageFromSource{From: t.ID, Protect: ShieldAnything, Then: then}.Apply(ctx)
		}
		return nil
	}
}
