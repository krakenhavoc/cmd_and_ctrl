package decks

import (
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
func TestRealDumpSpellAnswers(t *testing.T) {
	idx := loadDumpForPurpose(t)
	declared := map[string]bool{}
	names := map[string]string{}
	for _, s := range effects.All() {
		base, _, _ := strings.Cut(s.OracleID, "#")
		if _, ok := names[base]; !ok {
			names[base] = s.Name
		}
		declared[base] = declared[base] || spellDeclaresAnswers(s)
	}
	var instants, flash []string
	done := 0
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
		if declared[base] {
			done++
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
	t.Logf("%d instant-speed catalog cards declare what they answer; %d instants and %d flash cards do not yet",
		done, len(instants), len(flash))
	t.Logf("undeclared instants:\n\t%s", strings.Join(instants, "\n\t"))
	t.Logf("undeclared flash cards:\n\t%s", strings.Join(flash, "\n\t"))
}

// spellDeclaresAnswers reports whether a spec declares Answers on the
// card, a mode or an alternative cost.
func spellDeclaresAnswers(s effects.Spec) bool {
	if s.Purpose.Answers != 0 {
		return true
	}
	if s.Modes != nil {
		for _, o := range s.Modes.Options {
			if o.Purpose.Answers != 0 {
				return true
			}
		}
	}
	for _, a := range s.AlternativeCosts {
		if a.Purpose.Answers != 0 {
			return true
		}
	}
	return false
}
