package decks

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// realdump_spell_answers_manual_test.go — #2877, ADR 0142's amendment
// of 2026-10-30: the sweep list for what a spell answers. Gated on
// CMDCTRL_SCRYFALL_DUMP like the other real-dump audits, because a
// catalog Spec carries no type line and CI has no dump.
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/decks/ -run RealDumpSpellAnswers -v
//
// It LISTS, without failing, every catalog card that can be cast at
// instant speed by its own printed text (an instant, or a card with
// flash) and declares Purpose.Answers in no spell slot: not on the
// card, a mode or an alternative cost. Each such cast stops smart
// autopass on an opponent's stack item, as every instant did before
// #2877, because an undeclared cast counts as an answer (ADR 0142 owner
// answer 3). A sweep batch declares a slice of this list and runs it
// again; the count it prints is the batch's "before" and "after".
//
// It also prints the same count for the curated decks alone (the ones
// GET /decks serves), the listing ADR 0009's amendment of 2026-10-30
// (#4734) declares first: 56 undeclared before its slice 5, none after.
//
// That amendment's slice 2: a declared cast now
// carries answers_now, matched by kind, so a declaration that is too
// narrow can pass a window the card could answer. The listing also
// counts the declared cards by kind (a card counts once for each kind
// any of its slots declares), and every sweep PR quotes those counts
// before and after under "Verdicts changed".
func TestRealDumpSpellAnswers(t *testing.T) {
	idx := loadDumpForPurpose(t)
	curated := map[string]bool{}
	for _, d := range All() {
		for _, c := range d.WithCompanion() {
			if !c.Basic {
				curated[c.OracleID] = true
			}
		}
	}
	declared := map[string]bool{}
	kinds := map[string]game.Answers{}
	names := map[string]string{}
	for _, s := range effects.All() {
		base, _, _ := strings.Cut(s.OracleID, "#")
		if _, ok := names[base]; !ok {
			names[base] = s.Name
		}
		declared[base] = declared[base] || spellDeclaresAnswers(s)
		kinds[base] |= spellAnswers(s)
	}
	var instants, flash, curatedLeft []string
	byKind := map[string]int{}
	done, curatedDone := 0, 0
	for base, name := range names {
		id, err := uuid.Parse(base)
		if err != nil {
			continue // a token key, which has no printing
		}
		card, ok := idx.FindByOracleID(id)
		if !ok {
			continue
		}
		typeLine, _ := printed(card)
		instant := strings.Contains(typeLine, "Instant")
		hasFlash := slices.Contains(card.Keywords, "Flash")
		if !instant && !hasFlash {
			continue
		}
		if curated[base] {
			if declared[base] {
				curatedDone++
			} else {
				curatedLeft = append(curatedLeft, name)
			}
		}
		if declared[base] {
			done++
			for _, k := range kinds[base].Wire() {
				byKind[k]++
			}
			continue
		}
		if instant {
			instants = append(instants, name)
		} else {
			flash = append(flash, name)
		}
	}
	sort.Strings(instants)
	sort.Strings(flash)
	sort.Strings(curatedLeft)
	t.Logf("%d instant-speed catalog cards declare what they answer; %d instants and %d flash cards do not yet",
		done, len(instants), len(flash))
	counts := make([]string, 0, len(byKind))
	for _, k := range game.AnswerWireNames() {
		if n := byKind[k]; n > 0 {
			counts = append(counts, fmt.Sprintf("%s %d", k, n))
		}
	}
	t.Logf("declared, by kind: %s", strings.Join(counts, ", "))
	t.Logf("undeclared instants:\n\t%s", strings.Join(instants, "\n\t"))
	t.Logf("undeclared flash cards:\n\t%s", strings.Join(flash, "\n\t"))
	t.Logf("curated decks: %d instant-speed cards declare what they answer; %d do not yet%s",
		curatedDone, len(curatedLeft), listed(curatedLeft))
}

// spellAnswers is the union of what a spec declares on the card, its
// modes and its alternative costs.
func spellAnswers(s effects.Spec) game.Answers {
	a := s.Purpose.Answers
	if s.Modes != nil {
		for _, o := range s.Modes.Options {
			a |= o.Purpose.Answers
		}
	}
	for _, alt := range s.AlternativeCosts {
		a |= alt.Purpose.Answers
	}
	return a
}

// spellDeclaresAnswers reports whether a spec declares Answers on the
// card, a mode or an alternative cost.
func spellDeclaresAnswers(s effects.Spec) bool { return spellAnswers(s) != 0 }

// listed formats names for a log line: nothing for none, else a list.
func listed(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return ":\n\t" + strings.Join(names, "\n\t")
}
