package game

import "fmt"

// granted_entry_trigger.go — #5428: a keyword a permanent gives the
// spells its controller casts, whose cost is an optional additional cost
// and whose payoff is an ability of the PERMANENT the spell becomes.
// Zinnia, Valley's Voice: "Creature spells you cast gain offspring {2} as
// you cast them."
//
//	702.175a Offspring represents two abilities. "Offspring [cost]"
//	         means "You may pay an additional [cost] as you cast this
//	         spell" and "When this permanent enters, if its offspring
//	         cost was paid, create a token that's a copy of it, except
//	         it's 1/1."
//
// The cost half is a GrantedOptionalCost (granted_optional_cost.go). The
// trigger half cannot be a row of the card, which does not print it, and
// cannot belong to the permanent that granted it, which may have left:
// the spell gained the ability as it was cast (CR 610.5: "Some static
// abilities create one-shot effects that cause spells a player casts to
// gain an ability as that player casts them"), so the permanent it
// becomes has it. So, like casualty's copy and prowess, the trigger is
// derived: the paid key rides the permanent's cast record
// (CastProvenance.GrantedOptional, CR 400.7d) and ownTriggersOf adds the
// trigger registered for it, once per payment (CR 702.175b).
//
// A trigger registered here builds a keyed item (NewKeyedTriggeredItem),
// so it is a restore point with no catalog row behind it. The registry is
// package state filled at init, not a route from Game.

// grantedOptionalEntryTriggers is the registry: granted optional cost
// key → the enters trigger of the permanent its spell becomes.
var grantedOptionalEntryTriggers = map[string]TriggeredAbility{}

// RegisterGrantedOptionalEntryTrigger declares that a spell cast with the
// granted optional cost `key` paid becomes a permanent with the triggered
// ability `t`: the "when this permanent enters, if its <keyword> cost was
// paid" half of a granted keyword. The ability's intervening "if" is the
// record itself, so `t` is derived only for a permanent whose spell paid
// the cost, once per payment. `t` must build a keyed item
// (NewKeyedTriggeredItem): the card has no catalog row for it.
//
// Call it at init. `key` is an on-disk identity (it is recorded on
// PaidCost.GrantedOptional and CastProvenance.GrantedOptional): never
// renamed, never reused. Panics on an empty or repeated key.
func RegisterGrantedOptionalEntryTrigger(key string, t TriggeredAbility) {
	if key == "" {
		panic("game: RegisterGrantedOptionalEntryTrigger with an empty key")
	}
	if _, dup := grantedOptionalEntryTriggers[key]; dup {
		panic(fmt.Sprintf("game: granted optional entry trigger %q registered twice", key))
	}
	if t.Build == nil {
		panic(fmt.Sprintf("game: granted optional entry trigger %q has no Build", key))
	}
	grantedOptionalEntryTriggers[key] = t
}

// grantedOptionalWithEntryTrigger narrows a payment record's granted
// optional keys to the ones with a registered entry trigger, as a slice
// of its own. Nil when there is none, which is every cast before #5428.
func grantedOptionalWithEntryTrigger(keys []string) []string {
	var out []string
	for _, k := range keys {
		if _, ok := grantedOptionalEntryTriggers[k]; ok {
			out = append(out, k)
		}
	}
	return out
}

// grantedEntryTriggersFor is the triggered abilities `c` has because its
// spell was cast with a granted keyword's cost paid: one per payment (CR
// 702.175b). None for a permanent that has lost all abilities (CR
// 613.1f), and none for a key this binary does not register.
func grantedEntryTriggersFor(c *Card) []TriggeredAbility {
	if len(c.Provenance.GrantedOptional) == 0 || c.lostAllAbilities() {
		return nil
	}
	var out []TriggeredAbility
	for _, k := range c.Provenance.GrantedOptional {
		if t, ok := grantedOptionalEntryTriggers[k]; ok {
			out = append(out, t)
		}
	}
	return out
}
