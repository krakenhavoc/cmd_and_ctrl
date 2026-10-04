package game

// annihilator.go — CR 702.86, the third KEYWORD TRIGGER the engine
// derives from an object's ability list rather than from its catalog
// entry (#2073, ADR 0113 §2). Prowess (prowess.go) and evolve
// (evolve.go) are the models; toxic (infect_wither_toxic.go) is the
// model for the number.
//
//	702.86a Annihilator is a triggered ability. "Annihilator N" means
//	        "Whenever this creature attacks, defending player
//	        sacrifices N permanents."
//	702.86b If a creature has multiple instances of annihilator, each
//	        triggers separately.
//
// THE TOKEN. Annihilator carries a number, so it is stored the way
// toxic is: KeywordAnnihilator is the FAMILY key canonicalKeywords
// holds, the wire tokens are "annihilator N" (minted only by
// CanonicalAnnihilatorToken), and CanonicalKeywords refuses a bare
// "annihilator". Scryfall's keywords array says only "Annihilator", so
// the deck importer reads the number off the oracle line, as it does
// for toxic. It is CUMULATIVE (CR 702.86b): AppendKeywordAbility keeps
// a granted "annihilator 2" beside a printed "annihilator 4", and each
// is its own trigger.
//
// THE TRIGGER. "Whenever this creature attacks" is CR 508.3a: it
// triggers when the creature is declared as an attacker, and never for
// a creature put onto the battlefield attacking. EventAttack is emitted
// for a declared attacker and for nothing else, so watching it for the
// source is the whole rule — the same predicate the catalog's
// ThisAttacked reads.
//
// THE DEFENDING PLAYER (CR 508.5, 508.5a). It is the player the
// creature is attacking, the controller of the planeswalker it is
// attacking, or the protector of the battle it is attacking, decided
// per attacking creature. The trigger records it as it triggers
// (Params.Player) and reads it again as it resolves: if the source is
// still the same object and still attacking, the live answer
// (defendingPlayerForAttackerLocked, which also covers a planeswalker
// or battle that has left combat, #1364); otherwise the recorded one,
// which is CR 508.5's "the player that creature was attacking before
// it was removed from combat". The source leaving the battlefield does
// not stop the ability: annihilator has no "if" clause.
//
// THE SACRIFICE. The defending player chooses N permanents they
// control in ONE prompt and they are sacrificed together, as one
// simultaneous exit (CR 701.21a: only permanents they control). With
// fewer than N they sacrifice what they have (CR 609.3), which the
// prompt's bounds already do. A player who has left the game controls
// nothing (CR 800.4a) and is asked nothing.
//
// ORDERING (CR 603.3b). Annihilator triggers do NOT commute: each
// sacrifice's dies triggers land between them, so the attacking player
// orders them. Two instances with the same N on one creature have the
// same label, so the existing same-source, same-label rule skips a
// prompt that would mean nothing.

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// KeywordAnnihilator is the bare word CR 702.86 spells with a number
// after it. Like KeywordToxic it is the family key, not a token: the
// tokens are "annihilator N".
const KeywordAnnihilator = "annihilator"

// maxAnnihilatorValue bounds what AnnihilatorValue accepts. Printed
// annihilator tops out at 6 (Emrakul, the Aeons Torn), so a longer
// number is a malformed line and is refused, which errs weaker.
const maxAnnihilatorValue = 999

// numberedKeywordValue parses one ability token as a numbered keyword
// `word N`: the word, exactly one space and a positive decimal integer
// no greater than max, with nothing after it. Case and surrounding
// whitespace are normalised, so a raw oracle line can be handed in.
// Leading zeros are accepted and normalised away.
//
// Shared by toxic (ToxicValue) and annihilator (AnnihilatorValue), so
// the two numbered keywords can never disagree on what a number is.
func numberedKeywordValue(token, word string, max int) (int, bool) {
	rest, ok := strings.CutPrefix(strings.ToLower(strings.TrimSpace(token)), word)
	if !ok {
		return 0, false
	}
	digits, ok := strings.CutPrefix(rest, " ")
	if !ok || digits == "" {
		return 0, false
	}
	n := 0
	for i := 0; i < len(digits); i++ {
		d := digits[i]
		if d < '0' || d > '9' {
			return 0, false
		}
		n = n*10 + int(d-'0')
		if n > max {
			return 0, false
		}
	}
	if n <= 0 {
		return 0, false
	}
	return n, true
}

// AnnihilatorValue parses one ability token as CR 702.86's numbered
// keyword, reporting the N it carries. "annihilator" alone,
// "annihilator 0" and "annihilator two" answer false.
func AnnihilatorValue(token string) (int, bool) {
	return numberedKeywordValue(token, KeywordAnnihilator, maxAnnihilatorValue)
}

// CanonicalAnnihilatorToken normalises one printed annihilator clause
// to the engine's wire form — "Annihilator 04" becomes "annihilator
// 4" — reporting whether it is one at all. The only thing that mints
// the token; CanonicalKeywords calls it.
func CanonicalAnnihilatorToken(s string) (string, bool) {
	n, ok := AnnihilatorValue(s)
	if !ok {
		return "", false
	}
	return KeywordAnnihilator + " " + strconv.Itoa(n), true
}

// AnnihilatorAmounts is the N of every annihilator instance the card
// has, in ability-list order, through the walk keywordTriggersFor
// reads. A read for tests, the bot and the view.
func AnnihilatorAmounts(c *Card) []int {
	var out []int
	forEachAbilityToken(c, func(a string) bool {
		if n, ok := AnnihilatorValue(a); ok {
			out = append(out, n)
		}
		return true
	})
	return out
}

// annihilatorLabel is the stack label of one annihilator N trigger.
// It names N, so two instances with different numbers on one creature
// are told apart by the CR 603.3b ordering prompt, and two with the
// same number are not.
func annihilatorLabel(n int) string {
	return "Annihilator " + strconv.Itoa(n) + " — defending player sacrifices " + permanentsPhrase(n)
}

// permanentsPhrase is "a permanent" / "N permanents".
func permanentsPhrase(n int) string {
	if n == 1 {
		return "a permanent"
	}
	return strconv.Itoa(n) + " permanents"
}

// annihilatorTriggerFor is the one TriggeredAbility a single instance
// of "annihilator N" is.
func annihilatorTriggerFor(n int) TriggeredAbility {
	label := annihilatorLabel(n)
	return TriggeredAbility{
		Keyword: KeywordAnnihilator,
		Watches: []EventKind{EventAttack},
		// CR 508.3a: declared as an attacker. EventAttack is never
		// emitted for a creature put onto the battlefield attacking.
		AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.CardID == source.InstanceID
		},
		Build: func(ev Event, source *Card, _ Characteristic, g *Game) *StackItem {
			// Keyed directly, like prowess and evolve: annihilator has
			// no catalog row, and a table with one waiting on the
			// stack is still a restore point.
			return NewKeyedTriggeredItem(source, label, annihilatorSacrificeBody, EffectParams{
				Player: g.annihilatorDefenderLocked(ev, source.InstanceID),
				Amount: n,
			})
		},
	}
}

// annihilatorSacrificeKey is the body every annihilator trigger names
// (ADR 0113 §2, ADR 0041 P9): an on-disk identity, never renamed or
// reused. An older binary refuses a restore point naming it with
// ErrUnknownEffectKey, which is the designed rollback case.
// Params.Player is the defending player when it triggered,
// Params.Amount is N.
const annihilatorSacrificeKey = "annihilator/sacrifice"

// annihilatorSacrificeBody is the registered body's reference. It is
// built from the key rather than from DelayedBody's return value, and
// the body is registered in init below, because the resolution reaches
// the trigger harvest (a sacrifice fires dies triggers, which asks
// keywordTriggersFor, which builds this item): a package-level var
// initialised with the function would be an initialization cycle.
var annihilatorSacrificeBody = BodyRef{key: annihilatorSacrificeKey}

func init() {
	DelayedBody(annihilatorSacrificeKey, func(g *Game, item *StackItem, p EffectParams) error {
		return g.annihilatorSacrificeLocked(item, p.Player, p.Amount)
	})
}

// annihilatorDefenderLocked is the defending player for the attacker
// `attacker` as its attack is announced (CR 508.5): read off the
// creature's attack, and off the event's attack target if the creature
// cannot be found.
//
// Caller must hold g.mu.
func (g *Game) annihilatorDefenderLocked(ev Event, attacker uuid.UUID) uuid.UUID {
	if d := g.defendingPlayerForAttackerLocked(findBattlefieldCard(g, attacker)); d != uuid.Nil {
		return d
	}
	if ev.Target != uuid.Nil {
		return g.defendingPlayerForAttackLocked(ev.Target)
	}
	return uuid.Nil
}

// AnnihilatorDefenderForEffect is annihilatorDefenderLocked for a
// catalog trigger's fill-in Build: "Ulamog has annihilator X"
// (effects.AnnihilatorCounted) records the defending player the same
// way the keyword's own trigger does, in item.Params.Player.
//
// Caller must hold g.mu.
func (g *Game) AnnihilatorDefenderForEffect(ev Event, attacker uuid.UUID) uuid.UUID {
	return g.annihilatorDefenderLocked(ev, attacker)
}

// AnnihilatorSacrificeForEffect resolves an annihilator trigger whose
// N is read as it resolves rather than printed: Ulamog, the Defiler's
// "annihilator X, where X is the number of +1/+1 counters on it"
// (effects.AnnihilatorCounted). The defending player is the one
// recorded in item.Params.Player, re-read exactly as the keyword's
// body re-reads it.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) AnnihilatorSacrificeForEffect(item *StackItem, n int) error {
	if item == nil {
		return nil
	}
	return g.annihilatorSacrificeLocked(item, item.Params.Player, n)
}

// annihilatorSacrificeLocked is the resolution: the defending player
// chooses N permanents they control and sacrifices them together.
//
// Caller must hold g.mu in write mode (resolution frame).
func (g *Game) annihilatorSacrificeLocked(item *StackItem, recorded uuid.UUID, n int) error {
	if item == nil || n <= 0 {
		return nil
	}
	defender := recorded
	// CR 508.5: while the creature is still attacking, the defending
	// player is read off its attack now; once it has been removed from
	// combat (or has left the battlefield), it is the player it was
	// attacking, which is the one recorded when it triggered.
	if !g.AbilitySourceIsNewObjectForEffect(item) {
		if src := findBattlefieldCard(g, item.SourceCardID); src != nil && src.AttackingTarget != uuid.Nil {
			defender = g.defendingPlayerForAttackerLocked(src)
		}
	}
	if p := g.playerByIDLocked(defender); p == nil || p.Eliminated {
		// CR 800.4a: a player who has left the game controls nothing.
		return nil
	}
	source := item.SourceCardID
	_, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
		Chooser:    defender,
		Source:     source,
		Question:   item.Label + " — choose " + permanentsPhrase(n) + " to sacrifice",
		Of:         []uuid.UUID{defender},
		Candidates: annihilatorCandidates(n),
	}, annihilatorSacrificeThen(source))
	return err
}

// annihilatorCandidates offers every permanent the defending player
// controls, exactly N to be chosen. The engine clamps the bounds when
// they control fewer (CR 609.3). Nothing is a target, so hexproof and
// shroud do not matter, and a planeswalker being attacked may be
// chosen (the 2010-06-15 ruling).
func annihilatorCandidates(n int) func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	return func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
		var out []uuid.UUID
		for _, c := range g.Battlefield.Cards {
			if c.Controller == of {
				out = append(out, c.InstanceID)
			}
		}
		return out, n, n
	}
}

// annihilatorSacrificeThen sacrifices the chosen permanents as one
// simultaneous exit, stamped with the annihilator's source as the
// card that asked. A package-level constructor closing over a scalar,
// the continuation contract every prompted run keeps.
func annihilatorSacrificeThen(source uuid.UUID) func(g *Game, picked PromptedPicks) error {
	return func(g *Game, picked PromptedPicks) error {
		ids := picked.Cards()
		if len(ids) == 0 {
			return nil
		}
		return g.SacrificeAllThenForEffect(source, ids, nil)
	}
}
