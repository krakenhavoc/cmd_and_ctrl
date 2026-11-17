package game

// borrowed_triggers.go — "has all activated AND TRIGGERED abilities of …"
// (ADR 0155's 2026-11-05 amendment, #4459). Koh, the Face Stealer is the
// card: "Koh has all activated and triggered abilities of the last chosen
// card." CR 113.10 and CR 613.1f: an ability-adding effect in layer 6,
// whose lent set is whatever the lender has right now.
//
// # The shape
//
// It is ADR 0155's borrow with one more flag (AbilityBorrow.Triggered).
// The borrow step writes, beside the activated and mana rows, one
// BorrowedAbility per lent TRIGGERED row:
//
//   - a catalog trigger row (Slot AbilitySlotTriggered, LentCatalog): the
//     key whose Triggered list holds it and its index there — the
//     identity IdentifyCatalogRows stamped on the row, a card's own key
//     or a granted bundle's GrantKey;
//   - a triggered mana ability (CR 605.1b; Slot BorrowSlotManaTrigger,
//     LentCatalog): the lender's ability key and the row's index in its
//     declared ManaTriggers list;
//   - a triggered keyword the engine derives from the ability list rather
//     than from a catalog row (Slot AbilitySlotTriggered, LentKeyword):
//     the token itself, "prowess", "annihilator 2", "undying".
//
// triggersOf and manaTriggersOf read the entries as a second half after
// the object's own, so every harvest (the per-event walk, the
// simultaneous-exit walk, the state-trigger walk) and every reader sees a
// lent trigger as an ability of the borrower. harvestLTB reads them off
// the LAST-KNOWN characteristic (CR 603.10a), so a borrower that dies has
// the lent "when this creature dies" it had while it was there.
//
// # Identity on the stack and across a restore
//
// A lent catalog row IS the lender's row: CatalogTriggers(Key)[Index]
// carries the lender's catalog stamp, so the harvested item is stamped
// with AbilityRef{Key: <lender's key>, Slot: triggered, Ref: own:<i> or
// grant:<b>:<i>:0, Name: <row Key>} — the row resolveTriggeredAbilityRef
// already reads. Restore, a pending trigger frame and the boot report
// need no lender (CR 113.7a: once triggered, the ability exists apart
// from its source), and the item's source is the borrower, whose ID is
// item.SourceCardID, so "this creature" in the lent row is the borrower
// (ADR 0155 Decision 6, CR 201.5b). A keyword trigger is the engine's own
// row, keyed or not exactly as the lender's own instance would be.
//
// # When it stops
//
// The entries are layer-6 output, rebuilt on every pass. A lender that
// stops being one (Koh's choice changes, the chosen card leaves exile)
// lends nothing on the next pass, so the borrower stops having the
// trigger at once; an instance already triggered or on the stack is
// unaffected (CR 113.7a). CR 113.6: a lent row that declared another
// zone (a graveyard or exile trigger) is lent but never fires from the
// battlefield, because every battlefield harvest already skips such a row
// (TriggerWatchesFromZone).

// BorrowSlotManaTrigger is a BorrowedAbility's Slot for a lent triggered
// mana ability (CR 605.1b).
const BorrowSlotManaTrigger = "mana-trigger"

// lentTriggerRowsOf lists the triggered rows `lender` lends right now,
// with only the row identity filled in: its own catalog trigger rows
// (through its designation gate, its CR 613.1f removals and its grants,
// as triggersOf reads them, catalogTriggerRowsOf), its triggered mana abilities, the triggered
// keywords on its ability list, and the triggered rows it was lent
// itself, re-lent as they are. A face-down lender lends nothing (CR
// 708.2a); a row the catalog did not register (a test stub, an engine
// trigger that is not a keyword) has no identity to name and is not lent.
func lentTriggerRowsOf(lender *Card) []BorrowedAbility {
	if lender == nil || lender.FaceDown {
		return nil
	}
	var out []BorrowedAbility
	entry := func(slot string, kind LentKind, key string, index int) BorrowedAbility {
		return BorrowedAbility{Slot: slot, Kind: kind, Key: key, Index: index, Lender: lender.InstanceID, LenderName: lender.Name}
	}
	for _, t := range catalogTriggerRowsOf(lender) {
		if t.row.key == "" {
			continue
		}
		out = append(out, entry(AbilitySlotTriggered, LentCatalog, t.row.key, t.row.index))
	}
	if CatalogManaTriggers != nil {
		if key := catalogAbilityKeyOf(lender); key != "" {
			for i, t := range CatalogManaTriggers(key) {
				if t.ActiveWhen.IsGate() && !t.ActiveWhen.Active(*lender) {
					continue
				}
				out = append(out, entry(BorrowSlotManaTrigger, LentCatalog, key, i))
			}
		}
	}
	forEachAbilityToken(lender, func(a string) bool {
		if lendableTriggeredKeyword(a) {
			out = append(out, entry(AbilitySlotTriggered, LentKeyword, a, 0))
		}
		return true
	})
	for _, b := range layeredBorrows(lender) {
		if b.Slot == AbilitySlotTriggered || b.Slot == BorrowSlotManaTrigger {
			out = append(out, b)
		}
	}
	return out
}

// lendableTriggeredKeyword reports whether an ability token is a
// triggered keyword the engine derives from the ability list (prowess.go,
// undying_persist.go), which a borrow lends as its token. Every other
// triggered keyword (ward, echo, cumulative upkeep, exploit, …) is a
// catalog trigger row and is lent as one. Decayed (CR 702.147a) is a
// static and a triggered ability; lending its token lends only the
// trigger, because only the trigger is derived from it here — the
// "can't block" half reads the borrower's own ability list.
func lendableTriggeredKeyword(a string) bool {
	switch a {
	case KeywordProwess, KeywordEvolve, KeywordExalted, KeywordDecayed, KeywordMelee, KeywordEnlist,
		KeywordUndying, KeywordPersist, KeywordFlanking:
		return true
	}
	if _, ok := AnnihilatorValue(a); ok {
		return true
	}
	if _, ok := RenownValue(a); ok {
		return true
	}
	if _, ok := RampageValue(a); ok {
		return true
	}
	_, ok := ModularValue(a)
	return ok
}

// borrowedTriggersOf is the lent half of triggersOf: each lent catalog
// trigger row, looked up in the list its entry names, then the triggers
// of the lent keywords. An entry whose row the catalog no longer has is
// skipped. The lender's designation gate was settled when the entry was
// written (ADR 0155 Decision 2), so it is not asked again of the
// borrower.
func borrowedTriggersOf(c *Card) []TriggeredAbility {
	list := layeredBorrows(c)
	if len(list) == 0 {
		return nil
	}
	var out []TriggeredAbility
	var keywords []string
	var lastKey string
	var rows []TriggeredAbility
	for i := range list {
		b := &list[i]
		if b.Slot != AbilitySlotTriggered {
			continue
		}
		switch b.Kind {
		case LentKeyword:
			keywords = append(keywords, b.Key)
		case LentCatalog:
			if CatalogTriggers == nil {
				continue
			}
			if b.Key != lastKey || rows == nil {
				lastKey, rows = b.Key, CatalogTriggers(b.Key)
			}
			if b.Index >= 0 && b.Index < len(rows) {
				out = append(out, rows[b.Index])
			}
		}
	}
	if len(keywords) > 0 {
		out = append(out, keywordTriggersOver(func(fn func(string) bool) {
			for _, k := range keywords {
				if !fn(k) {
					return
				}
			}
		})...)
	}
	return out
}

// borrowedLTBTriggersOf is the lent half of harvestLTB's list, read off
// the departed permanent's LAST-KNOWN characteristic (CR 603.10a): the
// lent catalog rows, and the dies keywords (undying, persist, modular)
// among the lent tokens. The non-dies keywords' triggers are left to
// triggersOf, exactly as the permanent's own are.
func borrowedLTBTriggersOf(lki Characteristic) []TriggeredAbility {
	if len(lki.BorrowedAbilities) == 0 {
		return nil
	}
	var out []TriggeredAbility
	var keywords []string
	for _, b := range lki.BorrowedAbilities {
		if b.Slot != AbilitySlotTriggered {
			continue
		}
		switch b.Kind {
		case LentKeyword:
			keywords = append(keywords, b.Key)
		case LentCatalog:
			if CatalogTriggers == nil {
				continue
			}
			if rows := CatalogTriggers(b.Key); b.Index >= 0 && b.Index < len(rows) {
				out = append(out, rows[b.Index])
			}
		}
	}
	return append(out, ltbKeywordTriggersForTokens(keywords)...)
}

// borrowedManaTriggersOf is the lent half of manaTriggersOf.
func borrowedManaTriggersOf(c *Card) []ManaTrigger {
	if CatalogManaTriggers == nil {
		return nil
	}
	var out []ManaTrigger
	for _, b := range layeredBorrows(c) {
		if b.Slot != BorrowSlotManaTrigger || b.Kind != LentCatalog {
			continue
		}
		if rows := CatalogManaTriggers(b.Key); b.Index >= 0 && b.Index < len(rows) {
			out = append(out, rows[b.Index])
		}
	}
	return out
}

// borrowedStateTriggerCandidate is mayHaveStateTrigger's question for the
// lent rows: does one of them come from a definition with a state
// trigger?
func borrowedStateTriggerCandidate(c *Card, has func(string) bool) bool {
	for _, b := range layeredBorrows(c) {
		if b.Slot == AbilitySlotTriggered && b.Kind == LentCatalog && has(b.Key) {
			return true
		}
	}
	return false
}
