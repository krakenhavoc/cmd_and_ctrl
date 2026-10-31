package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// answer_window.go — ADR 0009's amendment of 2026-10-30 (#4734),
// accepted 2026-10-31 with the owner's answers Q1 to Q5 (a): a held
// answer stops smart auto-pass only when it can answer now.
//
// Every cast, activation and mana move whose declared answers (ADR
// 0142: an activated row's Purpose.Answers, a mana ability's Answers or
// its sacrifice cost, a cast's castAnswersOf) are something other than
// `value` alone carries Move.AnswersNow: true when one of those answers
// matches the window the seat holds priority in, false when none does.
// A move that declares several kinds matches when one of them does, the
// way a modal card declares the union (ADR 0142 sweep ruling 4). The
// server computes it once, for the seat being enumerated, so the browser
// and the MCP seat read one answer and keep no copy of the rule (Q3).
//
// A move it does not judge carries none, and every reader classes it as
// before (ADR 0009 §3, a stop over a missed window):
//
//   - a move that declares nothing: an undeclared cast (most of the
//     catalog's instant-speed cards, until ADR 0142's sweep declares
//     them), a row carried on a card instance that declares nothing, a
//     targeted row that declares nothing. A targeted row that declares (Q2, from slice 7) is judged
//     by its kind like any other;
//   - a move whose only answer is `value`: #2877's value_only covers a
//     cast, and a value activation was never a response;
//   - a move that declares `remove`, which keeps today's windows (Q1);
//   - a move that declares animate or combat_grant, at the seat's own
//     beginning of combat with no opponent's item on the stack, when none
//     of its answers matches (the owner's decision of 2026-10-31, below).
//
// The old flags (interacts, combat_interacts, combat_defender_only,
// value_only) are still computed as before, beside this, so a reader
// that does not know answers_now behaves as it did. In the viewer's own
// main phase with an empty stack every cast and activation is a play
// whatever this says (the client's sorcery window), so the verdict is
// set there too and read by nobody.
//
// What each kind matches, by window (the amendment's tables):
//
//   - protect, survive, pump, prevent, sac_outlet (savingAnswers): an
//     opponent's item on the stack that threatens you or yours, or a
//     combat that involves you. The threat is per kind (Q4): an item
//     that targets a permanent you control, an instant or sorcery spell
//     with no target that its declaration does not say misses you
//     (CR 115.10), or an ability that declares a sweep reaching you or an
//     edict (abilityReaches) threatens every one of them; an item that
//     targets you only protect, prevent and sac_outlet; an item that
//     targets a spell or ability you control only protect. A permanent
//     spell threatens nothing: its enters trigger is judged as an item of
//     its own (CR 603.3). Your own items never threaten you;
//   - restrict: an opponent's upkeep (the first step anyone gets
//     priority in, CR 502.4 and 503.1), or an opponent's spell on the
//     stack, a copy included (CR 707.10). Never an opponent's ability;
//   - graveyard (Q5, split from restrict): an opponent's spell on the
//     stack, or an opponent's item, spell or ability, that targets a card
//     in a graveyard. A reanimation or regrowth trigger is such an item,
//     and exiling the card in response leaves it with an illegal target
//     (CR 608.2b). Not an opponent's upkeep, and not an opponent's
//     ability that targets nothing in a graveyard;
//   - a counter (targets_stack, read from the announcement, not a kind):
//     an opponent's item on the stack;
//   - combat_grant, animate, makes_blocker: a combat that involves you,
//     in the steps where they can still change it (Q1). Not beginning of
//     combat, your own or anyone's. They matched your own beginning of
//     combat (CR 507.2: the last priority before attacks) until the owner
//     decided on 2026-10-31 (#4734) that a manland, a Crew row or a
//     keyword grant must not stop smart auto-pass there every turn. An
//     opponent's item on the stack still matches there, as in any step.
//     At your own beginning of combat, with no opponent's item on the
//     stack, an animate or combat_grant move that nothing else matches
//     carries no verdict rather than false: it keeps the class it had
//     before the amendment. So it is no stop at the default settings,
//     where the step is not ticked, and a player who ticks the step still
//     stops there through the stops grid, as the owner decided, even with
//     "only when I can do something" on, which a false verdict would end.
//     Animating in the first main phase is a play and unaffected.
//
// A combat that involves you (CR 506.3, 802.2: the rule asks whether
// you are attacked, not whether you are a defending player):
//
//   - declare attackers, after the declaration (CR 508.2): you or yours
//     is attacked, or you attacked, for the saving kinds and
//     combat_grant; only you or yours attacked for animate and
//     makes_blocker;
//   - declare blockers (CR 509.2) and the first-strike damage step
//     (CR 510.4): a creature you control blocking or blocked (CR 509.1h)
//     for protect, survive and sac_outlet, and for protect also an
//     unblocked creature attacking you or yours; a creature you control
//     blocking or attacking, blocked or not, for pump and combat_grant;
//     that, or an unblocked creature attacking you or yours, for prevent;
//     never animate or makes_blocker, since blockers are declared
//     (CR 509.1).
//
// The regular combat damage step and end of combat match nothing for
// combat: the damage is dealt and state-based actions are checked before
// anyone gets priority (CR 117.5, 704.5g).
//
// Owner answer 4 is a match too, whatever the row declares: a
// "Sacrifice this creature: …" move matches while its creature is a
// target of an opponent's item on the stack, at any step, or is
// attacking or blocking in a step where combat counts (declare attackers,
// declare blockers, first-strike damage).
//
// The stack is read from g.StackMeta, where an activated or triggered
// ability lives with no card of its own (game/stack.go), and from the
// spell cards in g.Stack. A spell card that carries no metadata (a bare
// test fixture) reads as an untargeted spell.

// savingAnswers are the kinds that keep or trade away something of
// yours (ADR 0009 amendment #4734, Q1): the same five #3555 gated on a
// threat (threatGated).
const savingAnswers = threatGated

// targetOnYou is what an item that targets you threatens (Q4 (a)):
// protection or hexproof stops the targeting, damage can be prevented,
// and a sacrifice outlet trades away what an edict would take.
// Indestructible and a pump do nothing for a player.
const targetOnYou = game.AnswerProtect | game.AnswerPrevent | game.AnswerSacOutlet

// answerWindow is what the window the seat holds priority in lets a
// declared answer do, read once per enumeration (enumerator.answerWindow).
type answerWindow struct {
	// matched is every declared kind that matches now.
	matched game.Answers
	// oppItem is set when an item an opponent controls is on the stack:
	// a move that targets the stack answers it.
	oppItem bool
	// combatStep is set in declare attackers, declare blockers and the
	// first-strike damage step: where owner answer 4's attacking or
	// blocking creature can still be saved.
	combatStep bool
	// targetedByOpp holds every card an opponent's item on the stack
	// targets: owner answer 4's threatened creature.
	targetedByOpp map[uuid.UUID]bool
	// ownBeginCombat is set at the seat's own beginning of combat with no
	// opponent's item on the stack: where an unmatched animate or
	// combat_grant move carries no verdict (the owner's decision of
	// 2026-10-31, #4734) instead of false.
	ownBeginCombat bool
}

// answerWindow is the enumerator's window, read on first use.
func (e *enumerator) answerWindow() *answerWindow {
	if e.window == nil {
		w := readAnswerWindow(e.g, e.seat)
		e.window = &w
	}
	return e.window
}

// readAnswerWindow reads the window `seat` is in.
func readAnswerWindow(g *game.Game, seat uuid.UUID) answerWindow {
	w := answerWindow{}
	active := isActiveSeat(g, seat)
	w.readStack(g, seat)
	w.readCombat(g, seat)
	if g.Turn.Step == game.StepUpkeep && !active {
		w.matched |= game.AnswerRestrict
	}
	w.ownBeginCombat = active && g.Turn.Step == game.StepBeginCombat && !w.oppItem
	return w
}

// stackEntry is one item on the stack: its metadata, the spell card
// when it is a spell, and who controls it now.
type stackEntry struct {
	item       *game.StackItem
	card       *game.Card
	controller uuid.UUID
}

// spell reports whether the entry is a spell (a copy included).
func (s stackEntry) spell() bool {
	return s.item == nil || s.item.Kind == game.StackItemSpell
}

// stackEntries is every item on the stack, in no particular order: the
// spell cards in g.Stack with their metadata, and every ability, which
// has only metadata.
func stackEntries(g *game.Game) []stackEntry {
	var out []stackEntry
	onZone := map[uuid.UUID]bool{}
	if g.Stack != nil {
		for i := range g.Stack.Cards {
			c := &g.Stack.Cards[i]
			e := stackEntry{item: g.StackMeta[c.InstanceID], controller: c.Controller}
			if e.item != nil {
				e.controller = e.item.Controller
			}
			if e.spell() {
				e.card = c
			}
			out = append(out, e)
			onZone[c.InstanceID] = true
		}
	}
	for id, item := range g.StackMeta {
		if item == nil || onZone[id] {
			continue
		}
		out = append(out, stackEntry{item: item, controller: item.Controller})
	}
	return out
}

// readStack reads what the opponents' items on the stack let each kind
// answer.
func (w *answerWindow) readStack(g *game.Game, seat uuid.UUID) {
	for _, e := range stackEntries(g) {
		if e.controller == seat {
			continue
		}
		w.oppItem = true
		if e.spell() {
			w.matched |= game.AnswerRestrict | game.AnswerGraveyard
		}
		var targets []game.TargetRef
		if e.item != nil {
			targets = e.item.Targets
		}
		targeted := false
		for _, r := range targets {
			switch r.Kind {
			case game.TargetPlayer:
				targeted = true
				if r.ID == seat {
					w.matched |= targetOnYou
				}
			case game.TargetCard:
				targeted = true
				if w.targetedByOpp == nil {
					w.targetedByOpp = map[uuid.UUID]bool{}
				}
				w.targetedByOpp[r.ID] = true
				if c := battlefieldCardByID(g, r.ID); c != nil {
					if c.Controller == seat {
						w.matched |= savingAnswers
					}
				} else if it := g.StackMeta[r.ID]; it != nil && it.Controller == seat {
					w.matched |= game.AnswerProtect
				} else if inAGraveyard(g, r.ID) {
					w.matched |= game.AnswerGraveyard
				}
			case game.TargetSelf:
				targeted = true
			}
		}
		if !targeted && e.spell() && untargetedSpellReaches(g, seat, e) {
			w.matched |= savingAnswers
		}
		if e.item != nil && abilityReaches(g, seat, e.item) {
			w.matched |= savingAnswers
		}
	}
}

// inAGraveyard reports whether the card with this id is in any player's
// graveyard: the target of a reanimation or regrowth item that exiling
// graveyards answers (the graveyard kind, Q5).
func inAGraveyard(g *game.Game, id uuid.UUID) bool {
	if id == uuid.Nil {
		return false
	}
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for i := range p.Graveyard.Cards {
			if p.Graveyard.Cards[i].InstanceID == id {
				return true
			}
		}
	}
	return false
}

// untargetedSpellReaches reports whether an opponent's spell with no
// target could reach what `seat` has (CR 115.10). An instant or sorcery
// could be a wipe, an edict or a steal, unless its declaration says it
// is not: exactly `value` for the modes it chose, or a sweep whose class
// holds nothing the seat controls. One that declares nothing reaches
// (#3555). A permanent spell does nothing while it is on the stack.
func untargetedSpellReaches(g *game.Game, seat uuid.UUID, e stackEntry) bool {
	c := e.card
	if c == nil {
		return true
	}
	if !c.IsInstant() && !c.IsSorcery() {
		return false
	}
	p := game.CardPurposeOfPtr(c)
	answers := p.Answers
	if e.item != nil {
		answers = castAnswersOf(p.Answers, game.ModeSpecFor(game.CatalogKey(*c)), e.item.Modes, nil)
	}
	if answers == game.AnswerValue {
		return false
	}
	if !p.Sweep.IsZero() {
		return sweepReaches(g, seat, p.Sweep.Matches)
	}
	return true
}

// readCombat reads what a combat that involves `seat` lets each kind
// answer, by step. Beginning of combat has no case: no combat exists
// yet, and animate and combat_grant no longer match it on the seat's own
// turn (the owner's decision of 2026-10-31, #4734).
func (w *answerWindow) readCombat(g *game.Game, seat uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	switch g.Turn.Step {
	case game.StepDeclareAttackers:
		w.combatStep = true
		attacked, attacking := false, false
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.AttackingTarget == uuid.Nil {
				continue
			}
			if c.Controller == seat {
				attacking = true
			}
			if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == seat {
				attacked = true
			}
		}
		if attacked || attacking {
			w.matched |= savingAnswers | game.AnswerCombatGrant
		}
		if attacked {
			w.matched |= game.AnswerAnimate | game.AnswerMakesBlocker
		}
	case game.StepDeclareBlockers, game.StepFirstStrikeDamage:
		w.combatStep = true
		var blocking, blocked, attacking, unblockedAtYou bool
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == seat && c.BlockingTarget != uuid.Nil {
				blocking = true
			}
			if c.AttackingTarget == uuid.Nil {
				continue
			}
			if c.Controller == seat {
				attacking = true
				if g.BlockedAttackerForEffect(c.InstanceID) {
					blocked = true
				}
			}
			if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == seat && g.UnblockedAttackerForEffect(c.InstanceID) {
				unblockedAtYou = true
			}
		}
		if blocking || blocked {
			w.matched |= game.AnswerProtect | game.AnswerSurvive | game.AnswerSacOutlet
		}
		if blocking || attacking {
			w.matched |= game.AnswerPump | game.AnswerCombatGrant | game.AnswerPrevent
		}
		if unblockedAtYou {
			w.matched |= game.AnswerProtect | game.AnswerPrevent
		}
	}
}

// verdict is Move.AnswersNow for a move whose announcement declares
// `answers`: nil for a move that keeps today's class, else whether one
// of its answers matches the window. targetsStack is the announcement's
// counter half; self is owner answer 4's creature, nil when the move
// does not sacrifice a creature on the battlefield as its own cost.
//
// At the seat's own beginning of combat with no opponent's item on the
// stack, a move that declares animate or combat_grant and matches nothing
// else carries no verdict (ownBeginCombat), not false: false would end a
// ticked step's stop under "only when I can do something", which the
// owner kept (2026-10-31, #4734).
func (w *answerWindow) verdict(answers game.Answers, targetsStack bool, self *game.Card) *bool {
	if answers == 0 || answers == game.AnswerValue || answers.HasAny(game.AnswerRemove) {
		return nil
	}
	now := answers.HasAny(w.matched) || (targetsStack && w.oppItem) || w.selfSacrificeMatches(self)
	if !now && w.ownBeginCombat && answers.HasAny(game.AnswerCombatGrant|game.AnswerAnimate) {
		return nil
	}
	return &now
}

// selfSacrificeMatches is owner answer 4 (ADR 0142): the creature a
// "Sacrifice this creature: …" move gives up is a target of an
// opponent's item, or is attacking or blocking where combat counts.
func (w *answerWindow) selfSacrificeMatches(self *game.Card) bool {
	if self == nil {
		return false
	}
	if w.targetedByOpp[self.InstanceID] {
		return true
	}
	return w.combatStep && (self.AttackingTarget != uuid.Nil || self.BlockingTarget != uuid.Nil)
}

// selfSacrificed is owner answer 4's creature for a move: its source,
// when the move sacrifices it as a cost and it is a creature on the
// battlefield. Nil otherwise.
func selfSacrificed(source *game.Card, zone game.ZoneKind, sacrificesSelf bool) *game.Card {
	if !sacrificesSelf || source == nil || zone != game.ZoneBattlefield || !source.IsCreature() {
		return nil
	}
	return source
}

// castAnswersNow is a cast announcement's AnswersNow. A cast whose costs
// give up a creature is a sacrifice outlet beside whatever it declares
// (castValueOnly's rule, ADR 0142 owner answer 4); one that declares
// nothing, or exactly `value`, keeps today's class.
func (w *answerWindow) castAnswersNow(g *game.Game, answers game.Answers, targetsStack bool, costIDs ...[]uuid.UUID) *bool {
	if answers != 0 && answers != game.AnswerValue && givesUpCreature(g, costIDs...) {
		answers |= game.AnswerSacOutlet
	}
	return w.verdict(answers, targetsStack, nil)
}
