package game

import "strings"

// triggers_of_key.go — a token whose TEXT is "all triggered abilities of
// <other cards>" (#5683). The Book of Vile Darkness: "Create Vecna, a
// legendary 8/8 black Zombie God creature token with indestructible and
// all triggered abilities of the exiled cards."
//
// # Why it is not a borrow
//
// ADR 0155's borrow is a layer-6 effect: "has all activated abilities of
// …", re-read on every pass. Vecna's clause is the effect that CREATES
// the token defining its characteristics (CR 111.3), which are "the
// token's text" and its copiable values (CR 707.2). So the
// abilities are fixed as the ability resolves (CR 608.2h), stay when the
// cards leave exile, are copied by a Clone, and are removed with the
// rest of the token's own abilities by an effect with ANY timestamp (CR
// 613.1f), where a layer-6 record would survive an earlier Humility.
//
// # The shape
//
// It is copy_grants.go's: the object's own text key is a composite, and
// each "all triggered abilities of <card>" is one part of it, carried on
// Card.GrantedAbilities like a CR 707.9a copy grant. The part is a
// DERIVED key, TriggersOfKey(<card's catalog key>), which the catalog
// answers with a definition holding that card's triggered rows and
// triggered mana abilities and nothing else: no activated or mana
// ability, static, replacement or cost modifier comes along, which is
// the word "triggered" in the clause.
//
// The rows are the lender's own, stamped with ITS catalog identity
// (IdentifyCatalogRows), so a trigger harvested off the token is keyed
// own:<i> under the lender's key, exactly as a lent trigger is (ADR
// 0155's 2026-11-05 amendment): it resolves and restores with no lender
// and no new effect key (CR 113.7a). "This permanent" in a row is the
// token, because the row reads item.SourceCardID.
//
// A triggered KEYWORD the engine derives from the ability list rather
// than from a catalog row is carried as the keyword itself on the
// token's printed Keywords when the keyword is a triggered ability and
// nothing else (prowess, exalted, undying, …). Decayed and modular are a
// triggered ability AND something else (CR 702.147a, CR 702.43a), so
// only their trigger is carried, as a second derived part,
// KeywordTriggerKey(<token>).
//
// Nothing here is stored state of its own: the parts are strings on a
// field the snapshot and the copy already carry.

// TriggersOfKeyPrefix marks a granted-ability part as "all triggered
// abilities of <the card whose catalog key follows>". It carries
// GrantKeyPrefix, so GrantKey leaves it alone.
const TriggersOfKeyPrefix = GrantKeyPrefix + "triggers-of:"

// KeywordTriggerKeyPrefix marks a granted-ability part as "the triggered
// ability of <the keyword that follows>", for a keyword that is not only
// a triggered ability (decayed, modular N).
const KeywordTriggerKeyPrefix = GrantKeyPrefix + "trigger-of-keyword:"

// TriggersOfKey is the granted-ability part for "all triggered abilities
// of" the card whose catalog key is `key` (a copy grant riding on that
// key is not the card's own text and is dropped). Empty for an empty key.
func TriggersOfKey(key string) string {
	base := BaseCatalogKey(key)
	if base == "" {
		return ""
	}
	return TriggersOfKeyPrefix + base
}

// TriggersOfKeyLender is the catalog key a TriggersOfKey part names.
func TriggersOfKeyLender(part string) (string, bool) {
	if !strings.HasPrefix(part, TriggersOfKeyPrefix) {
		return "", false
	}
	return part[len(TriggersOfKeyPrefix):], true
}

// KeywordTriggerKey is the granted-ability part for the triggered half
// of a keyword ability token ("decayed", "modular 2").
func KeywordTriggerKey(token string) string { return KeywordTriggerKeyPrefix + token }

// TriggersOnlyDef is the definition a TriggersOfKey part resolves to:
// the lender definition's triggered rows and triggered mana abilities,
// with `text` as the line a player reads. Nil when the lender has no
// triggered ability, so a card with none adds no part. The rows keep the
// lender's catalog stamp (see the file comment), so the result is NOT
// passed to IdentifyCatalogRows.
func TriggersOnlyDef(lender *CardDef, text string) *CardDef {
	if lender == nil || len(lender.Triggered)+len(lender.ManaTriggers) == 0 {
		return nil
	}
	return &CardDef{Triggered: lender.Triggered, ManaTriggers: lender.ManaTriggers, GrantText: text}
}

// KeywordTriggerDef is the definition a KeywordTriggerKey part resolves
// to: the one engine trigger of decayed (CR 702.147a's second ability)
// or of modular N (CR 702.43a's second ability). Nil for every other
// part.
func KeywordTriggerDef(part string) *CardDef {
	if !strings.HasPrefix(part, KeywordTriggerKeyPrefix) {
		return nil
	}
	token := part[len(KeywordTriggerKeyPrefix):]
	if token == KeywordDecayed {
		return &CardDef{Triggered: []TriggeredAbility{decayedTrigger}}
	}
	if _, ok := ModularValue(token); ok {
		return &CardDef{Triggered: []TriggeredAbility{modularTrigger}}
	}
	return nil
}

// TriggeredTextOf reads "all triggered abilities of" a card as the text
// a token is created with: `parts` join the token template's
// GrantedAbilities and `keywords` its printed Keywords. The card is read
// as the object it is where it is (a card in exile has its printed
// abilities, ADR 0155 Decision 3): its catalog triggered rows and
// triggered mana abilities as one TriggersOfKey part, and its triggered
// keywords. A face-down card has no abilities (CR 708.2a) and one the
// catalog has no triggered row for adds no part.
func TriggeredTextOf(lender Card) (parts, keywords []string) {
	if lender.FaceDown {
		return nil, nil
	}
	if part := TriggersOfKey(catalogKeyOf(&lender)); part != "" && catalogDef(part) != nil {
		parts = append(parts, part)
	}
	forEachAbilityToken(&lender, func(a string) bool {
		if !lendableTriggeredKeyword(a) {
			return true
		}
		if _, modular := ModularValue(a); modular || a == KeywordDecayed {
			parts = append(parts, KeywordTriggerKey(a))
		} else {
			keywords = append(keywords, a)
		}
		return true
	})
	return parts, keywords
}
