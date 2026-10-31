package game

import (
	"bufio"
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
	if update {
		// Always rewrite on update, so a hand-merged ledger is put back in
		// sorted order even when nothing is missing.
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
	case "pick":
		return KnownRevealedPickThen(fields[1])
	case "option":
		return KnownOptionPickThen(fields[1])
	case "vote":
		return KnownVoteThen(fields[1])
	case "choose":
		return KnownChooseCardsThen(fields[1])
	case "payunless":
		return KnownPayUnlessElse(fields[1])
	case "confirm":
		return KnownConfirmThen(fields[1])
	case "maycast":
		// #4361: a keyed may_cast's answer (RegisterMayCastThen).
		return KnownMayCastThen(fields[1])
	case "castfollowup":
		// #4361 S13: a cast permission's FollowUp or IfNotCast body
		// (RegisterCastFollowUp).
		return KnownCastFollowUp(CastFollowUpKey(ledgerLineKey(line)))
	case "anyplayer":
		return KnownAnyPlayerPays(fields[1])
	case "number":
		return KnownNumberThen(fields[1])
	// #4253: these registries' keys may hold spaces (a "then discard"
	// body is keyed by its question), so the key is the rest of the line.
	case "resume":
		return KnownReplacementResume(ledgerLineKey(line))
	case "drawinstead":
		return KnownDrawInstead(ledgerLineKey(line))
	case "drawthen":
		return KnownDrawThen(ledgerLineKey(line))
	case "keywordinstead":
		return KnownKeywordInstead(ledgerLineKey(line))
	case "tail":
		// #4253 slice 3: a keyed event continuation (RegisterTail).
		return KnownTail(ledgerLineKey(line))
	case "searchrule":
		// #4253 slice 5: a search prompt's set rule (RegisterSearchRule).
		return KnownSearchRule(fields[1])
	case "milluntil":
		// #4253 slice 9: a keyed mill run's stop clause.
		return KnownMillUntil(ledgerLineKey(line))
	case "coinflip":
		// #4253 slice 10b: a keyed coin flip's continuation.
		return KnownCoinFlipThen(fields[1])
	case "run":
		// #4290: a keyed prompted run's continuation (RegisterPromptRunThen).
		return KnownPromptRunThen(fields[1])
	case "steps":
		// A keyed continuation's recorded steps (RegisterThenStepsKey).
		return KnownThenSteps(fields[1])
	case "tapcost":
		// #4264: a keyed pay-unless's waterbend clause kind.
		return KnownTapCost(fields[1])
	case "alias":
		effectRegistryMu.RLock()
		defer effectRegistryMu.RUnlock()
		to, ok := effectAliases[fields[1]]
		return ok && (len(fields) < 3 || to == fields[2])
	}
	return false
}

// appendEffectKeyLedger adds lines to the ledger at path and rewrites its
// body in sorted order, keeping the header comments and every existing
// line. The ledger is a set (readEffectKeyLedger), so order carries no
// meaning; keeping it sorted means two branches that each add a key insert
// at different places and merge cleanly, where appending at the end made
// every pair of card PRs conflict on the last line.
func appendEffectKeyLedger(path string, lines []string) error {
	header, body, err := splitEffectKeyLedger(path)
	if err != nil {
		return err
	}
	if header == "" {
		header = effectKeysHeader
	}
	seen := map[string]bool{}
	var all []string
	for _, line := range append(body, lines...) {
		if !seen[line] {
			seen[line] = true
			all = append(all, line)
		}
	}
	sort.Strings(all)
	return os.WriteFile(path, []byte(header+strings.Join(all, "\n")+"\n"), 0o644)
}

// splitEffectKeyLedger returns the ledger's leading comment block (with
// its trailing blank line) and its key lines in file order.
func splitEffectKeyLedger(path string) (header string, body []string, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, nil
		}
		return "", nil, err
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	i := 0
	for i < len(lines) && (strings.HasPrefix(lines[i], "#") || strings.TrimSpace(lines[i]) == "") {
		i++
	}
	if i > 0 {
		header = strings.Join(lines[:i], "\n") + "\n"
		if strings.TrimSpace(lines[i-1]) != "" {
			header += "\n"
		}
	}
	for _, line := range lines[i:] {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			body = append(body, line)
		}
	}
	return header, body, nil
}

// EffectKeyLedgerIsSorted reports whether the ledger's key lines are in
// sorted order, the order appendEffectKeyLedger writes.
func EffectKeyLedgerIsSorted(path string) (bool, error) {
	_, body, err := splitEffectKeyLedger(path)
	if err != nil {
		return false, err
	}
	return sort.StringsAreSorted(body), nil
}

const effectKeysHeader = `# effect_keys.txt — ADR 0041 phase 3's effect-key ledger (tier 2, #1497).
#
# Every delayed-trigger body, event condition and alias key a production
# binary registers, one per line: "body <key>", "condition <key>" or
# "alias <old> <new>". A restore point can name any of them, so they are
# on-disk identities: never delete or edit a line; rename a
# key with game.EffectAlias(old, new), which keeps the old one resolving.
# "pick <key>" is a revealed-hand pick's continuation (#2115), which
# has no alias: never rename one. "option <key>" is a keyed option
# pick's continuation (#2854), on the same terms, and so is "choose
# <key>", a keyed choose_cards pick's (#2040).
#
# Add new keys (kept in sorted order, so branches merge cleanly) with:
#   go test ./internal/cards/effects -run TestEveryPersistedEffectKeyResolves -args -update-effect-keys

`

// ledgerLineKey is everything after a ledger line's kind.
func ledgerLineKey(line string) string {
	_, key, _ := strings.Cut(strings.TrimSpace(line), " ")
	return key
}
