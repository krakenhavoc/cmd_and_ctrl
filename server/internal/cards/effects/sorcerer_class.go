package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const sorcererClassGrant = "sorcerer-class/spell-or-level-mana"

// Sorcerer Class — {U}{R} Enchantment — Class (Adventures in the Forgotten
// Realms):
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 When this Class enters, draw two cards, then discard two cards.
//	 {U}{R}: Level 2
//	 Creatures you control have "{T}: Add {U} or {R}. Spend this mana only to
//	 cast an instant or sorcery spell or to gain a Class level."
//	 {3}{U}{R}: Level 3
//	 Whenever you cast an instant or sorcery spell, that spell deals damage to
//	 each opponent equal to the number of instant and sorcery spells you've cast
//	 this turn."
//
// Level 1 is the draw-then-discard of a plain enters trigger (CR 716.3: ungated,
// the Class enters at level 1, CR 716.2b). Level 2 is a layer-6 bundle grant
// (ADR 0093) gated at level 2 on the grantor's static: each creature gets one
// mana ability that adds {U} or {R}. Its mana carries two restriction clauses
// admitted either way (ManaRestrictAnyOf): a cast of an instant or sorcery spell,
// or an activation that is a Class level-up (ManaRestrictClassLevel, read off
// ActivatedAbility.ClassLevel, which LevelUp sets), including another Class's
// or this one's own level 3. Level 3 is a cast trigger gated at level 3; the spell
// is the damage source, and the count is read as the trigger resolves (CR 608.2h:
// "the number of ... you've cast this turn" is not locked in as it triggers), so
// it includes the spell that triggered it and any cast in response. The damage is
// dealt to each opponent separately (CR 120.3).
//
// No simplification.
func init() {
	grant := GrantAbilities(b16CreaturesYouControl, sorcererClassGrant)
	grant.Label = "Creatures you control have \"{T}: Add {U} or {R}. Spend this mana only to cast an instant or sorcery spell or to gain a Class level.\""
	grant.ActiveWhen = Level(2)
	Register(Spec{
		OracleID:     "19ba0f2d-4de3-49a0-be7d-da0aa220cd8a",
		Name:         "Sorcerer Class",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{U}{R}")),
			LevelUp(3, ManaCost("{3}{U}{R}")),
		},
		Grants: []AbilityGrant{{
			Key: sorcererClassGrant,
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{U|R}",
				Label:    "Add {U} or {R}. Spend this mana only to cast an instant or sorcery spell or to gain a Class level.",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
					[]string{game.ManaRestrictClassLevel},
				)},
			}},
			Text: "{T}: Add {U} or {R}. Spend this mana only to cast an instant or sorcery spell or to gain a Class level.",
		}},
		Static: []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sorcerer Class — draw two cards, then discard two cards",
				func(g *game.Game, item *game.StackItem) error {
					return drawThenDiscard(g, item.Controller, item.SourceCardID, 2, 2, "")
				}),
			AtLevel(3, WheneverYouCast(Or(Instant(), Sorcery()),
				"Sorcerer Class — that spell deals damage to each opponent equal to the number of instant and sorcery spells you've cast this turn",
				sorcererClassDamage)),
		},
	})
}

// sorcererClassDamage is the level-3 body: the cast spell is the source.
func sorcererClassDamage(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	spell := item.Trigger.Event.CardID
	n := g.CastTallyFor(item.Controller).InstantSorcery
	if n <= 0 {
		return nil
	}
	// One instruction, one instance of damage (CR 615.8, ADR 0108 PR 0).
	return g.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := (DealDamage{Source: spell, Target: opp, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
