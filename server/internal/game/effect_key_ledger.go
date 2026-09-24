package game

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// effect_key_ledger.go reads and appends the effect-key ledger,
// testdata/effect_keys.txt (ADR 0041 phase 3, tier 2, #1497). It lives
// in production code only because two packages' tests need it: the
// engine's (its own bodies) and the catalog's (every body, so the full
// check). Nothing at runtime calls it.

// effectKeyTestPrefix is the namespace tests register throwaway bodies
// under. Never in a restore point, so never in the ledger.
const effectKeyTestPrefix = "test/"

// CheckEffectKeyLedger compares the registry with the ledger at path.
// missing is every registered key the ledger lacks (after appending
// them, when update is set). stale is every ledger line that no longer
// resolves; it is computed only when complete is set, because a package
// that has not loaded the catalog would see every catalog key as stale.
func CheckEffectKeyLedger(path string, update, complete bool) (missing, stale []string, err error) {
	registered := map[string]bool{}
	for _, line := range RegisteredEffectKeys() {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.HasPrefix(fields[1], effectKeyTestPrefix) {
			continue
		}
		registered[line] = true
	}
	ledger, err := readEffectKeyLedger(path)
	if err != nil {
		return nil, nil, err
	}
	for line := range registered {
		if !ledger[line] {
			missing = append(missing, line)
		}
	}
	sort.Strings(missing)
	if update && len(missing) > 0 {
		if err := appendEffectKeyLedger(path, missing); err != nil {
			return nil, nil, err
		}
		for _, line := range missing {
			ledger[line] = true
		}
		missing = nil
	}
	if !complete {
		return missing, nil, nil
	}
	for line := range ledger {
		if registered[line] {
			continue
		}
		if ledgerLineResolves(line) {
			continue // still resolves, through an alias
		}
		stale = append(stale, line)
	}
	sort.Strings(stale)
	return missing, stale, nil
}

func readEffectKeyLedger(path string) (map[string]bool, error) {
	out := map[string]bool{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[line] = true
	}
	return out, sc.Err()
}

// ledgerLineResolves reports whether a ledger line still resolves, by
// its OWN kind: a "body X" line is not kept alive by a condition named
// X (#1568 review).
func ledgerLineResolves(line string) bool {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return false
	}
	switch fields[0] {
	case "body":
		return KnownEffectBody(fields[1])
	case "condition":
		return KnownEffectCondition(fields[1])
	case "alias":
		effectRegistryMu.RLock()
		defer effectRegistryMu.RUnlock()
		to, ok := effectAliases[fields[1]]
		return ok && (len(fields) < 3 || to == fields[2])
	}
	return false
}

// appendEffectKeyLedger appends lines, writing the header first only
// when the file does not exist or is empty — never again into a file
// that already holds it, even one holding nothing but the header.
func appendEffectKeyLedger(path string, lines []string) error {
	fresh := true
	if st, err := os.Stat(path); err == nil && st.Size() > 0 {
		fresh = false
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if fresh {
		if _, err := fmt.Fprint(f, effectKeysHeader); err != nil {
			_ = f.Close()
			return err
		}
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			_ = f.Close()
			return err
		}
	}
	return f.Close()
}

const effectKeysHeader = `# effect_keys.txt — ADR 0041 phase 3's effect-key ledger (tier 2, #1497).
#
# Every delayed-trigger body, event condition and alias key a production
# binary registers, one per line: "body <key>", "condition <key>" or
# "alias <old> <new>". A restore point can name any of them, so they are
# on-disk identities: APPEND-ONLY. Never delete or edit a line; rename a
# key with game.EffectAlias(old, new), which keeps the old one resolving.
#
# Append new keys with:
#   go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys

`
