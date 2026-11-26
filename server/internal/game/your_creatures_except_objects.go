package game

import "github.com/google/uuid"

// your_creatures_except_objects.go — S82 OLD b411 (Mark for Death): "Other
// creatures that player controls can't block this turn."
//
// ScopeCreaturesExceptObjects (#6087) spares named objects from a rule on EVERY
// creature. This is the same record narrowed to one player's creatures: the
// record's Controller is that player (as it is for ScopeYourCreatures, whose
// "you" is whoever the record names), and the mods' Objects are the creatures
// it spares — Mark for Death's target, which "blocks this turn if able" and so
// must not also be barred from blocking.
//
// "Can't block" changes no characteristic, so CR 611.2c does not lock the set:
// the set is read as the rule is asked, a creature that player gains control
// of or puts onto the battlefield later this turn is covered, and a spared
// creature that left and came back is a new object (CR 400.7) that is
// covered too. Like ScopeCreaturesExceptObjects it carries only addRestrictions
// mods, each naming at least one object (exceptObjectsScopeProblem).

// ScopeYourCreaturesExceptObjects is "creatures the record's Controller
// controls, except the objects its mods name", read live. An on-disk identity:
// never renamed, never reused.
const ScopeYourCreaturesExceptObjects AffectedScope = "yourCreaturesExceptObjects"

// yourCreaturesExceptObjectsPredicate is the AppliesTo of a
// ScopeYourCreaturesExceptObjects record: a creature `controller` controls that
// is not one of the objects the mods name.
func yourCreaturesExceptObjectsPredicate(mods []Mod, controller uuid.UUID) func(*Card, *Game, *Card) bool {
	except := exceptObjectsPredicate(mods)
	return func(target *Card, g *Game, source *Card) bool {
		return target != nil && target.Controller == controller && except(target, g, source)
	}
}

// OtherCreaturesOfPlayerCantBlockForEffect is "other creatures <player>
// controls can't block this turn", sparing `spared` as the object it is now
// (CR 400.7): a live ScopeYourCreaturesExceptObjects "can't block" until end
// of turn. Returns false, writing nothing, when `spared` is not on the
// battlefield.
//
// Caller must hold g.mu (write).
func (g *Game) OtherCreaturesOfPlayerCantBlockForEffect(sourceID, player, spared uuid.UUID, d Duration, label string) bool {
	c := findBattlefieldCard(g, spared)
	if c == nil {
		return false
	}
	mod := AddRestrictionsMod(CantBlock)
	mod.Objects = cloneObjectRefs([]ObjectRef{{ID: spared, Epoch: c.ObjectEpoch}})
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeYourCreaturesExceptObjects, player, []Mod{mod}, d, label)
}
