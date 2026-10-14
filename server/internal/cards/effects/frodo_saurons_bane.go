package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frodo, Sauron's Bane — Legendary Creature — Halfling Citizen {W}, 1/2:
//
//	"{W/B}{W/B}: If Frodo is a Citizen, it becomes a Halfling Scout
//	 with base power and toughness 2/3 and lifelink.
//	 {B}{B}{B}: If Frodo is a Scout, it becomes a Halfling Rogue with
//	 'Whenever this creature deals combat damage to a player, that
//	 player loses the game if the Ring has tempted you four or more
//	 times this game. Otherwise, the Ring tempts you.'"
//
// Each "becomes" is a ScopedEffect record with no stated duration, so
// it lasts as long as Frodo is this object (CR 611.2a, 400.7), and its
// creature types are a SET, not an add (game.SetCreatureTypesMod,
// #2144, CR 205.1a): the Scout is no longer a Citizen and the Rogue is
// no longer a Scout. That is what makes the two abilities a ladder.
// The "if" is checked as the ability resolves (CR 608.2c), against the
// creature types Frodo has then, so activating the second ability on
// a Citizen does nothing.
//
// The Rogue's text is a granted trigger (ADR 0093's duration grant),
// in the same record as its new types, one timestamp. A later record
// does not take the earlier one's abilities: each effect changes only
// what it names, so the Rogue keeps the Scout's base 2/3 and lifelink.
//
// The trigger reads "you" as the trigger's controller, and the tempt
// count as it resolves. A player who can't lose the game (Platinum
// Angel) doesn't, and the Ring doesn't tempt you instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8be62f5c-0ab2-4869-97a2-d8a69a3fde95",
		Name:         "Frodo, Sauron's Bane",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: frodoSauronsBaneRogueGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Frodo, Sauron's Bane — that player loses the game, or the Ring tempts you",
					frodoSauronsBaneRogueHits),
			},
			Text: "Whenever this creature deals combat damage to a player, that player loses the game if the Ring has tempted you four or more times this game. Otherwise, the Ring tempts you.",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{W/B}{W/B}: If Frodo is a Citizen, it becomes a Halfling Scout with base power and toughness 2/3 and lifelink.",
				Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
				Cost:    ManaCost("{W/B}{W/B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !frodoSauronsBaneIs(g, item, "Citizen") {
						return nil
					}
					self := item.SourceCardID
					return ScopedEffectFor{
						Target: self,
						Mods: []game.Mod{
							game.SetCreatureTypesMod("Halfling", "Scout"),
							game.AddKeywordsMod("lifelink"),
							game.SetBasePowerMod(2),
							game.SetBaseToughnessMod(3),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Frodo, Sauron's Bane — a 2/3 Halfling Scout with lifelink",
					}.Apply(NewContext(g, item))
				},
			},
			{
				// Its trigger pays off only on a later hit, so it answers
				// nothing on the stack.
				Label:   "{B}{B}{B}: If Frodo is a Scout, it becomes a Halfling Rogue with \"Whenever this creature deals combat damage to a player, that player loses the game if the Ring has tempted you four or more times this game. Otherwise, the Ring tempts you.\"",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{B}{B}{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !frodoSauronsBaneIs(g, item, "Scout") {
						return nil
					}
					self := item.SourceCardID
					return GrantAbilitiesFor{
						Target:   self,
						Keys:     []string{frodoSauronsBaneRogueGrant},
						Also:     []game.Mod{game.SetCreatureTypesMod("Halfling", "Rogue")},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Frodo, Sauron's Bane — a Halfling Rogue",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

const frodoSauronsBaneRogueGrant = "frodo-saurons-bane/rogue"

// frodoSauronsBaneIs is the abilities' "if Frodo is a <type>", read as
// the ability resolves: Frodo must still be the permanent that
// activated it (CR 400.7) and have that creature type now.
func frodoSauronsBaneIs(g *game.Game, item *game.StackItem, subtype string) bool {
	if !sourceIsStillThisPermanent(g, item) {
		return false
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	return ok && c.HasSubtype(subtype)
}

// frodoSauronsBaneRogueHits is the granted trigger's resolution: the
// damaged player loses if the Ring has tempted the trigger's controller
// four or more times this game, and otherwise the Ring tempts them.
func frodoSauronsBaneRogueHits(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if game.RingTemptCount(g, ctx.Controller()) >= 4 {
		return LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
	}
	return TheRingTemptsYou{}.Apply(ctx)
}
