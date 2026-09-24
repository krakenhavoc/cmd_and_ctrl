package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_requirements.go — the card-facing half of CR 508.1d attack
// requirements (#1571, ADR 0045 amendment of 2026-09-24). The engine
// half — what a requirement is, how a declaration is judged against
// the most requirements it could obey, where it is refused — lives in
// game/attack_requirements.go. A card file only says WHICH creatures
// a requirement reaches and for HOW LONG:
//
//	Static: []game.StaticAbility{AttacksEachCombat()},              // Zurgo Helmsmasher
//	Static: []game.StaticAbility{AttacksEachCombatWhere(allCreatures)}, // Grand Melee
//	OpponentsCreaturesAttackIfAble{Duration: DurationUntilEndOfTurn(ctx)}.Apply(ctx) // Bident of Thassa
//
// Goad needs none of this: goad stamps the engine's per-object marker
// (Card.GoadedBy — b33Goad), and the engine reads goad's two
// requirements off it.
//
// # Why a requirement is not a Restriction bit
//
// A restriction says a declaration may not contain something; a
// requirement says it should contain something IF it can. CR 508.1d
// counts requirements — two sources asking the same creature to
// attack are two requirements, and under a limit that decides which
// creature gets the slot — so a requirement is one entry per source in
// Characteristic.AttackRequirements rather than a bit.

// requirementFrom is the AttackRequirement a static or a floating
// effect writes, attributed to `source` so the refusal sentence can
// name the card ("… attacks each combat if able (Grand Melee).").
func requirementFrom(source *game.Card, otherThan uuid.UUID) game.AttackRequirement {
	r := game.AttackRequirement{OtherThan: otherThan}
	if source != nil {
		r.Source, r.SourceName = source.InstanceID, source.Name
	}
	return r
}

// AttacksEachCombat is "~ attacks each combat if able" — a requirement
// a creature prints on itself (Zurgo Helmsmasher).
//
// It is the creature's own ability, so it goes with the creature's
// abilities: a Zurgo that loses all abilities no longer has to attack
// (CR 613.1f — the catalog stops being asked once CatalogAbilityKey
// answers empty).
func AttacksEachCombat() game.StaticAbility {
	return AttacksEachCombatWhere(selfOnly)
}

// AttacksEachCombatWhere is "<creatures> attack each combat if able"
// from a permanent — Goblin Rabblemaster's "Other Goblin creatures you
// control", Grand Melee's "All creatures". `appliesTo` scopes it the
// way any static's AppliesTo does; only creatures are ever affected.
//
// The requirement belongs to the SOURCE, so a creature that loses all
// its abilities under Grand Melee still has to attack
// (Characteristic.AttackRequirements is never cleared by a layer-6
// removal, exactly like Restrictions).
func AttacksEachCombatWhere(appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && appliesTo(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			c.AttackRequirements = append(c.AttackRequirements, requirementFrom(source, uuid.Nil))
		},
	}
}

// OpponentsCreaturesAttackIfAble is a RESOLVING effect's requirement on
// "creatures your opponents control", for a duration:
//
//   - Bident of Thassa: "Creatures your opponents control attack this
//     turn if able." — DurationUntilEndOfTurn.
//   - Kardur, Doomscourge: "until your next turn, creatures your
//     opponents control attack each combat if able and attack a player
//     other than you if able." — OtherThanYou, DurationUntilYourNextTurn.
//   - The Akroan War, chapter II — DurationUntilYourNextTurn.
//
// The affected set is NOT locked when the effect begins. CR 611.2c
// locks the set only for an effect that changes characteristics or
// control; a requirement does neither ("it modifies the rules of the
// game"), so it reaches a creature an opponent casts, or gains control
// of, after the ability resolved. The record's scope
// (game.ScopeOpponentsCreatures) is read at every layer pass.
//
// A requirement on ONE permanent (Legion Warboss's token) is an
// ordinary pinned record instead:
//
//	ScopedEffectFor{Target: token, Mods: []game.Mod{game.AddAttackRequirementMod(uuid.Nil)}, …}
type OpponentsCreaturesAttackIfAble struct {
	// OtherThanYou adds "and attack a player other than you if able" —
	// a second requirement, which CR 508.1d counts.
	OtherThanYou bool
	// Duration is how long it lasts; the zero value is until end of
	// turn (game.Duration's own default).
	Duration game.Duration
	// Label is attribution for logs and tests.
	Label string
}

// Apply registers the requirement as a data record (ADR 0041 phase 3),
// so a table holding one keeps its restore point.
func (a OpponentsCreaturesAttackIfAble) Apply(ctx *Context) error {
	you := ctx.Controller()
	if you == uuid.Nil {
		return nil
	}
	mods := []game.Mod{game.AddAttackRequirementMod(uuid.Nil)}
	if a.OtherThanYou {
		mods = append(mods, game.AddAttackRequirementMod(you))
	}
	ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeOpponentsCreatures, you,
		mods, a.Duration, eotLabel(a.Label, "creatures your opponents control attack if able"))
	return nil
}
