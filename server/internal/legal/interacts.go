package legal

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// interacts.go — #2853, owner answer 2 (2026-10-09), and ADR 0142. An
// activated ability with no target can still be an answer to something
// on the stack: a sacrifice outlet in response to removal, a
// regeneration shield, Selfless Spirit's indestructible, Spore Frog's
// fog, a pump in response to a burn spell. Move.Interacts says so, and
// smart autopass stops for it as it stops for a targeted ability. Pure
// value (draw, mana, ramp, a fetch, tokens, scry, a Clue) stays
// unmarked. Move.CombatInteracts is the combat tier: an ability that
// changes a fight (a granted first strike, crew) and counts only in a
// combat window (#2871).
//
// ADR 0142: what an untargeted activation answers is DECLARED on the
// catalog row (game.Purpose.Answers), and untargetedFlags reads only
// that.
// Every catalog row smart autopass asks about declares (the ratchet,
// internal/cards/effects' TestEveryAnswersRowDeclares). Since S4 there
// is no printed-text read: a row that declares nothing (an ability
// carried on a card instance, built at run time, which is not catalog
// data) counts as interacting, following ADR 0009 §3's "a
// false-positive stop over a false-negative skip" (owner answer 3).
//
// Two inputs are not declared on an activated row:
//
//   - a mana ability reads ManaAbility.Answers, else its sacrifice cost
//     (manaAnswersOf): a creature sacrifice outlet answers, the mana
//     never does. A mana ability has no other source (owner answer 3);
//   - a "Sacrifice this creature: …" move interacts while its creature
//     is a target of an opponent's item on the stack, and
//     combat-interacts while it is attacking or blocking (owner answer
//     4, selfSacrificeThreatened). That is read from the game, not the
//     row.

// CostAnswers is what an activated ability's cost alone answers: a
// creature sacrifice, exile or return outlet (sac_outlet), or crew
// (animate). effects.Register reads it too: a row whose cost is an
// outlet must declare sac_outlet or protect (ADR 0142 §3), so the cost
// rule stays true for every declared row.
func CostAnswers(c game.AbilityCost) game.Answers {
	var a game.Answers
	if sacrificeInteracts(c.SacrificeOther) || exileCostInteracts(c.ExilePermanents) {
		a |= game.AnswerSacOutlet
	}
	if rc := c.ReturnToHand; !rc.Empty() && sacrificeInteracts(rc.Filter) {
		a |= game.AnswerSacOutlet
	}
	if c.Crew > 0 {
		a |= game.AnswerAnimate
	}
	return a
}

// ManaCostAnswers is the same question for a mana ability's sacrifice
// cost: only a creature sacrifice outlet (Ashnod's Altar, Phyrexian
// Altar) answers anything. The mana it makes never does.
func ManaCostAnswers(sacrificeOther *game.TargetSpec) game.Answers {
	if sacrificeInteracts(sacrificeOther) {
		return game.AnswerSacOutlet
	}
	return 0
}

// manaAnswersOf is what a mana ability can do in response: its
// declaration (sac_outlet or value), else its sacrifice cost.
func manaAnswersOf(m game.ManaAbilityShape) game.Answers {
	if m.Answers != 0 {
		return m.Answers
	}
	return ManaCostAnswers(m.SacrificeOther)
}

// selfSacrificeThreatened is ADR 0142 owner answer 4: "Sacrifice this
// creature: …" is an answer when the creature is about to be lost. It
// is read from the game: `stack` when the creature is a target of an
// item on the stack that its controller does not control, `combat` when
// it is attacking or blocking. A source that is not a creature on the
// battlefield, or a row that does not sacrifice its source, is neither.
func selfSacrificeThreatened(g *game.Game, source *game.Card, zone game.ZoneKind, sacrificesSelf bool) (stack, combat bool) {
	if !sacrificesSelf || source == nil || zone != game.ZoneBattlefield || !source.IsCreature() {
		return false, false
	}
	combat = source.AttackingTarget != uuid.Nil || source.BlockingTarget != uuid.Nil
	if g.Stack != nil {
		for i := range g.Stack.Cards {
			item := g.StackMeta[g.Stack.Cards[i].InstanceID]
			if item == nil || item.Controller == source.Controller {
				continue
			}
			for _, r := range item.Targets {
				if r.Kind == game.TargetCard && r.ID == source.InstanceID {
					return true, combat
				}
			}
		}
	}
	return false, combat
}

// combatFlags is an activate move's combat_interacts and
// combat_defender_only.
type combatFlags struct {
	any, defenderOnly bool
}

// untargetedFlags is an activate move's interacts and combat flags: the
// row's answers by tier, and owner answer 4's threatened self-sacrifice.
// A move with a target sets none (has_targets already stops you).
func untargetedFlags(g *game.Game, source *game.Card, zone game.ZoneKind, ab game.ActivatedAbilityShape, targets []game.TargetRef) (bool, combatFlags) {
	if hasTargets(targets) {
		return false, combatFlags{}
	}
	answers := ab.Purpose.Answers
	if answers == 0 {
		// Undeclared (ADR 0142 owner answer 3): a stop over a missed
		// window. Every catalog row in the ratchet's scope declares, so
		// this is a row carried on a card instance, or a catalog row the
		// guard refuses a declaration: a sorcery-speed row (offered in
		// its controller's main phase, which the client reads as a play
		// whatever this says) or an "up to X targets" row announced with
		// none.
		return true, combatFlags{}
	}
	threatened, fighting := selfSacrificeThreatened(g, source, zone, ab.Cost.SacrificeSelf)
	return answerFlags(answers, threatened, fighting)
}

// answerFlags maps a row's answers and its threatened self-sacrifice
// onto the move's flags. combat_interacts is never set beside interacts,
// which already counts in every window (#2871), and an ability whose
// only combat answer is a token blocker (makes_blocker) matters only to
// a defending player (combat_defender_only).
func answerFlags(answers game.Answers, threatened, fighting bool) (bool, combatFlags) {
	if answers.HasTier(game.TierStack) || threatened {
		return true, combatFlags{}
	}
	if !answers.HasTier(game.TierCombat) && !fighting {
		return false, combatFlags{}
	}
	blockerOnly := answers&(game.AnswerCombatGrant|game.AnswerAnimate|game.AnswerMakesBlocker) == game.AnswerMakesBlocker
	return false, combatFlags{any: true, defenderOnly: blockerOnly && !fighting}
}

// manaMoveInteracts is a mana move's interacts: a declared or read
// sacrifice outlet, or (owner answer 4) a creature that sacrifices
// itself while an opponent's item targets it. A mana move never sets
// combat_interacts.
func manaMoveInteracts(g *game.Game, source *game.Card, zone game.ZoneKind, ab game.ManaAbilityShape) bool {
	if manaAnswersOf(ab).HasTier(game.TierStack) {
		return true
	}
	threatened, _ := selfSacrificeThreatened(g, source, zone, ab.SacrificeCost)
	return threatened
}

// valueSacrifices are the printed sacrifice objects that are a price,
// not an answer: lands and the food, treasure, clue and blood tokens.
var valueSacrifices = []string{
	"land", "swamp", "forest", "island", "mountain", "plains", "desert",
	"food", "treasure", "clue", "blood", "artifact", "room", "token",
}

// sacrificeInteracts reads a sacrifice cost's printed object ("a
// creature", "another creature", "a Food"). A creature, a permanent,
// or a creature type the list does not know (a Goblin, an Eldrazi) is
// an outlet; a land, a token kind or a plain artifact is not.
func sacrificeInteracts(spec *game.TargetSpec) bool {
	if spec == nil {
		return false
	}
	label := strings.ToLower(strings.ReplaceAll(spec.Label, "noncreature", ""))
	if strings.Contains(label, "creature") || strings.Contains(label, "permanent") {
		return true
	}
	for _, v := range valueSacrifices {
		if strings.Contains(label, v) {
			return false
		}
	}
	return true
}

// exileCostInteracts: exiling a creature (or any permanent) you control
// as a cost saves it from what is on the stack; exiling a land does not.
func exileCostInteracts(c *game.ExilePermanentsCost) bool {
	if c.Empty() {
		return false
	}
	switch strings.ToLower(c.CardType) {
	case "", "creature", "planeswalker":
		return true
	}
	return false
}
