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
// Three inputs are not declared on an activated row:
//
//   - a mana ability reads ManaAbility.Answers, else its sacrifice cost
//     (manaAnswersOf): a creature sacrifice outlet answers, the mana
//     never does. A mana ability has no other source (owner answer 3);
//   - a "Sacrifice this creature: …" move interacts while its creature
//     is a target of an opponent's item on the stack, and
//     combat-interacts while it is attacking or blocking (owner answer
//     4, selfSacrificeThreatened). That is read from the game, not the
//     row;
//   - a protect, survive, pump, prevent or sac_outlet answer is gated (ADR
//     0009 amendment #3555, owner 2026-10-15): it sets interacts only
//     while an opponent's item on the stack threatens the seat
//     (stackThreatens), and otherwise counts in the combat tier. So a
//     free sacrifice outlet or Selfless Spirit stops you for removal, a
//     wipe or a fight, and not for every opponent spell. That is read
//     from the game too.
//
// answer_window.go (#4734) reads the same declarations a second way, for
// Move.AnswersNow: kind by kind, against the window the seat is in. These
// flags are unchanged by it, for every reader that does not know that
// field.

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

// threatGated are the stack answers that count only against a threat
// (ADR 0009 amendment #3555, owner answers Q1 (a) and Q2 (a),
// 2026-10-15). Each saves or trades away something of yours, so it
// answers an item that reaches for your things, not an opponent's
// Divination. survive is the half of protect that ADR 0142's 2026-10-15
// amendment split off (regenerate, indestructible), so it is gated with
// it. remove, restrict and graveyard (split from restrict by #4734's Q5)
// act on what an opponent has, may do next or can bring back, so they
// keep counting against any item.
const threatGated = game.AnswerProtect | game.AnswerSurvive | game.AnswerPump | game.AnswerPrevent | game.AnswerSacOutlet

// stackThreatens reports whether an item on the stack that `seat` does
// not control threatens it (ADR 0009 amendment #3555): the item targets
// `seat`, a permanent `seat` controls or an item on the stack `seat`
// controls, or it is an instant or sorcery spell with no target, which
// could be a wipe, an edict or a steal, or it is an activated or
// triggered ability whose catalog row declares a sweep or an edict that
// reaches something the seat controls (abilityReaches, #3825, #3882). A permanent spell does nothing
// to the board while it is on the stack; its enters trigger is an item
// of its own, judged when it is put there. With no game to read it says
// yes: a stop over a missed window (ADR 0009 §3).
func stackThreatens(g *game.Game, seat uuid.UUID) bool {
	if g == nil {
		return true
	}
	if g.Stack == nil {
		return false
	}
	for i := range g.Stack.Cards {
		card := &g.Stack.Cards[i]
		item := g.StackMeta[card.InstanceID]
		if item == nil || item.Controller == seat {
			continue
		}
		targeted := false
		for _, r := range item.Targets {
			switch r.Kind {
			case game.TargetPlayer:
				targeted = true
				if r.ID == seat {
					return true
				}
			case game.TargetCard:
				targeted = true
				if controlledBy(g, r.ID, seat) {
					return true
				}
			case game.TargetSelf:
				targeted = true
			}
		}
		if !targeted && item.Kind == game.StackItemSpell && (card.IsInstant() || card.IsSorcery()) {
			return true
		}
		if abilityReaches(g, seat, item) {
			return true
		}
	}
	return false
}

// abilityReaches reports whether an opponent's activated or triggered
// ability affects permanents `seat` controls without targeting them
// (#3825, #3882): a wipe, an edict, damage or -N/-N to each creature,
// which arrives as a planeswalker's minus, Nevinyrral's Disk, Pestilence,
// Pernicious Deed, a saga chapter or a "whenever ..., destroy all
// creatures" or "each opponent sacrifices a creature" trigger. A
// permanent spell tells nothing about this, and an ability has no card
// type, so it reads what the catalog row declares: Purpose.Sweep, else
// AnswerRemove (an edict or a sweep with no class named). A sweep whose
// class holds nothing the seat controls does not reach it. An ability
// with no declared row (a sandbox activation, a trigger built as an
// unkeyed closure) is not read.
func abilityReaches(g *game.Game, seat uuid.UUID, item *game.StackItem) bool {
	p, ok := item.AbilityPurpose(battlefieldCardByID(g, item.SourceCardID))
	if !ok {
		return false
	}
	if p.Sweep.IsZero() {
		return p.Answers.Has(game.AnswerRemove)
	}
	return sweepReaches(g, seat, p.Sweep.Matches)
}

// sweepReaches reports whether `seat` controls a permanent of the class a
// sweep matches. Matches is an upper bound (Sweep.Partial), so a miss is
// certain and a hit is a possible. An unnamed or unknown class says yes.
func sweepReaches(g *game.Game, seat uuid.UUID, m game.SweepMatch) bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != seat {
			continue
		}
		switch m {
		case game.SweepCreatures:
			if c.IsCreature() {
				return true
			}
		case game.SweepCreaturesManaValue3OrLess:
			if c.IsCreature() && c.ManaValue() <= 3 {
				return true
			}
		case game.SweepCreaturesManaValue4OrMore:
			if c.IsCreature() && c.ManaValue() >= 4 {
				return true
			}
		case game.SweepNonlandPermanents:
			if !c.IsLand() {
				return true
			}
		case game.SweepArtifacts:
			if c.IsArtifact() {
				return true
			}
		case game.SweepEnchantments:
			if c.IsEnchantment() {
				return true
			}
		case game.SweepArtifactsAndEnchantments:
			if c.IsArtifact() || c.IsEnchantment() {
				return true
			}
		default:
			return true
		}
	}
	return false
}

// battlefieldCardByID is the permanent with this id, or nil.
func battlefieldCardByID(g *game.Game, id uuid.UUID) *game.Card {
	if g.Battlefield == nil || id == uuid.Nil {
		return nil
	}
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
			return c
		}
	}
	return nil
}

// controlledBy reports whether id is a permanent or an item on the
// stack that `seat` controls.
func controlledBy(g *game.Game, id, seat uuid.UUID) bool {
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
				return c.Controller == seat
			}
		}
	}
	if item := g.StackMeta[id]; item != nil {
		return item.Controller == seat
	}
	return false
}

// inCombatStep is the server's half of the client's combat window
// (inCombatWindow, #2871): beginning of combat, declare attackers or
// declare blockers. A mana move carries no combat_interacts, so a
// threat-gated sacrifice outlet that makes mana reads the step here.
func inCombatStep(g *game.Game) bool {
	if g == nil {
		return false
	}
	switch g.Turn.Step {
	case game.StepBeginCombat, game.StepDeclareAttackers, game.StepDeclareBlockers:
		return true
	}
	return false
}

// combatFlags is an activate move's combat_interacts and
// combat_defender_only.
type combatFlags struct {
	any, defenderOnly bool
}

// untargetedFlags is an activate move's interacts and combat flags for
// `seat`: the row's answers by tier, owner answer 4's threatened
// self-sacrifice, and #3555's threat gate. A move with a target sets
// none (has_targets already stops you).
func untargetedFlags(g *game.Game, seat uuid.UUID, source *game.Card, zone game.ZoneKind, ab game.ActivatedAbilityShape, targets []game.TargetRef) (bool, combatFlags) {
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
	threat := answers.HasAny(threatGated) && stackThreatens(g, seat)
	return answerFlags(answers, threatened, fighting, threat)
}

// answerFlags maps a row's answers, its threatened self-sacrifice and
// whether the stack threatens the seat onto the move's flags.
// combat_interacts is never set beside interacts, which already counts
// in every window (#2871), and an ability whose only combat answer is a
// token blocker (makes_blocker) matters only to a defending player
// (combat_defender_only).
//
// A threat-gated answer (#3555) sets interacts only with `threat`.
// Without one it counts as a combat answer: a pump, a regeneration
// shield or a sacrifice outlet still changes a fight, so combat is
// unchanged. It is never defender-only, since it helps an attacker too.
func answerFlags(answers game.Answers, threatened, fighting, threat bool) (bool, combatFlags) {
	gated := answers.HasAny(threatGated)
	if (answers &^ threatGated).HasTier(game.TierStack) || threatened || (gated && threat) {
		return true, combatFlags{}
	}
	if !answers.HasTier(game.TierCombat) && !gated && !fighting {
		return false, combatFlags{}
	}
	blockerOnly := answers&(game.AnswerCombatGrant|game.AnswerAnimate|game.AnswerMakesBlocker|threatGated) == game.AnswerMakesBlocker
	return false, combatFlags{any: true, defenderOnly: blockerOnly && !fighting}
}

// manaMoveInteracts is a mana move's interacts for `seat`: a declared
// or read sacrifice outlet while the stack threatens the seat or in a
// combat step (#3555's gate; a mana move never sets combat_interacts,
// so the combat window is read here), or (owner answer 4) a creature
// that sacrifices itself while an opponent's item targets it.
func manaMoveInteracts(g *game.Game, seat uuid.UUID, source *game.Card, zone game.ZoneKind, ab game.ManaAbilityShape) bool {
	if answers := manaAnswersOf(ab); answers.HasTier(game.TierStack) {
		if !answers.HasAny(threatGated) || inCombatStep(g) || stackThreatens(g, seat) {
			return true
		}
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
