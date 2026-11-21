package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// next_damage_shield_cards.go — the shared rows of the ADR 0107 §6
// cards (#1860): the Circle of Protection and Rune of Protection
// families, and the one-ability shield every other card of the seam
// prints. Append-only, so the clone gate never sees the same body twice.

// nextDamageShieldRow is an activated ability whose whole effect is a
// shield — a next-damage shield, or (ADR 0108 §7, sourceShieldRow) a
// shield that is not one-use.
func NextDamageShieldRow(label string, cost game.AbilityCost, targets *game.TargetSpec, shield Applier) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: targets,
		Purpose: PreventsDamage(),
		Effect: func(g *game.Game, item *game.StackItem) error {
			return shield.Apply(NewContext(g, item))
		},
	}
}

// preventsDamage is the shield rows' declaration (ADR 0142): they
// prevent or redirect damage. A row with a target clause declares it
// too (#4734 slice 7, ADR 0142's amendment of 2026-11-05), so smart
// autopass matches its move by kind instead of stopping at every window
// through has_targets.
func PreventsDamage() game.Purpose {
	return game.Purpose{Answers: game.AnswerPrevent}
}

// alsoAnswers is a shared row with more answers declared beside its
// helper's: a shield whose cost sacrifices a creature is a sacrifice
// outlet too (Circle of Despair, Martyr's Cause), which the guard's cost
// coupling requires, and a redirection that can send the damage to an
// opponent's creature or to any target is removal too (Soltari
// Guerrillas, Zhalfirin Crusader), whose `remove` keeps the move's old
// stops.
func AlsoAnswers(a game.Answers, ab ActivatedAbility) ActivatedAbility {
	ab.Purpose.Answers |= a
	return ab
}

// nextDamageShieldSpell is an instant's OnResolve whose whole effect is
// a next-damage shield.
func NextDamageShieldSpell(shield PreventNextDamageFromSource) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return shield.Apply(ctx)
	}
}

// circleOfProtection is the Circle of Protection enchantment: one row,
// "<cost>: The next time <what> of your choice would deal damage to you
// this turn, prevent that damage." `label` is the row as printed and
// `what` the source it names ("a red source"). The Runes add Cycling {2}.
func CircleOfProtection(oracle, name, cost, label, what string, q game.PermanentQuery, more ...ActivatedAbility) Spec {
	shield := PreventNextDamageFromChosenSource(ShieldYou, q)
	shield.Question = name + " — choose " + what
	shield.Label = name + " — prevent the next damage from " + what
	return Spec{
		OracleID:     oracle,
		Name:         name,
		Completeness: CompletenessFull,
		Activated:    append([]ActivatedAbility{NextDamageShieldRow(label, ManaCost(cost), nil, shield)}, more...),
	}
}

// runeOfProtection is circleOfProtection with the Runes' {W} row and
// "Cycling {2}".
func RuneOfProtection(oracle, name, label, what string, q game.PermanentQuery) Spec {
	return CircleOfProtection(oracle, name, "{W}", label, what, q, Cycling("{2}"))
}

// chosenColorShieldRow is "<cost>: The next time a source of your choice
// of the chosen color would deal damage to you this turn, prevent that
// damage" (Story Circle, Prismatic Circle). The colour is the one chosen
// as the enchantment entered, read as it last existed if the enchantment
// has gone by resolution (CR 608.2h). Before a colour is chosen there is
// no source "of the chosen color", so nothing is shielded.
func ChosenColorShieldRow(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Purpose: PreventsDamage(),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			info, ok := ctx.SourcePermanent()
			if !ok || info.ChosenColor == "" {
				return nil
			}
			return PreventNextDamageFromChosenSource(ShieldYou, QueryColors(info.ChosenColor)).WithChosenQueries().Apply(ctx)
		},
	}
}

// chosenTypeShieldRow is "<cost>: The next time a creature of the chosen
// type would deal damage to you this turn, prevent that damage" (Circle
// of Solace). Nothing is chosen at resolution: the first creature of
// that type to deal damage to you is the one (CR 615.8, 615.9), judged
// as it deals the damage. The type is read as for chosenColorShieldRow.
func ChosenTypeShieldRow(label string, cost game.AbilityCost) ActivatedAbility {
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Purpose: PreventsDamage(),
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			info, ok := ctx.SourcePermanent()
			if !ok || info.NamedTribe == "" {
				return nil
			}
			return PreventNextDamageFromSource{
				Queries:       []game.PermanentQuery{{Types: []string{"creature"}, Subtypes: []string{info.NamedTribe}}},
				Protect:       ShieldYou,
				ChosenQueries: true,
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
func ShieldAgainstTargetCreature(then game.BodyRef, gainPower bool) func(*game.StackItem, *Context) error {
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
