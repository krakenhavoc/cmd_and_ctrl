package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soulbright Seeker — Creature — Elemental Sorcerer {R}, 2/1:
//
//	"As an additional cost to cast this spell, behold an Elemental or pay {2}. (To behold an Elemental, choose an Elemental you control or reveal an Elemental card from your hand.)
//	 {R}: Target creature you control gains trample until end of turn. If this is the third time this ability has resolved this turn, add {R}{R}{R}{R}."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the behold branch added by its 2026-10-07 amendment
// (BeholdOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "762679dd-4345-43c3-929b-15959443aeb0",
		Name:           "Soulbright Seeker",
		Completeness:   CompletenessFull,
		AdditionalCost: BeholdOrPay("an", "Elemental", "{2}"),
		Activated: []ActivatedAbility{{
			Label:   soulbrightSeekerLabel,
			Cost:    ManaCost("{R}"),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect:  soulbrightSeekerAbility,
		}},
	})
}

const soulbrightSeekerLabel = "{R}: Target creature you control gains trample until end of turn; the third time it resolves this turn, add {R}{R}{R}{R}"

// soulbrightSeekerAbility is the ability's body. The count is the
// per-turn tally of this Seeker's resolutions of this ability, which
// includes the one resolving (Game.ResolvedThisTurn, Elrond's count),
// so the third resolution is the one that adds the mana. The trample
// is granted first, to the target if it is still legal; the mana
// does not depend on the target (the ability was not countered).
func soulbrightSeekerAbility(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if id, ok := b16FirstLegalTargetCard(ctx); ok {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"trample"},
			Label:    "Soulbright Seeker — trample until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if g.ResolvedThisTurn(item.SourceCardID, soulbrightSeekerLabel) == 3 {
		return AddMana{Produced: "{R}{R}{R}{R}"}.Apply(ctx)
	}
	return nil
}
