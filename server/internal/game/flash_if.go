package game

import "github.com/google/uuid"

// flash_if.go — the card's own conditional timing permission, "You may
// cast this spell as though it had flash if <condition>" (#5207).
//
// It is a TIMING PERMISSION and nothing else. It is not an alternative
// cost (CR 118.9), so a cast that uses it still chooses its cost freely:
// the printed mana cost, any one alternative cost (a free-cast permission,
// Jodah's {W}{U}{B}{R}{G}), and any additional or optional costs. Until
// #5207 the catalog built it as an alternative-cost offer at the printed
// cost (AlternativeCost.AsThoughFlash), and CR 118.9a ("Only one
// alternative cost can be applied to any one spell") then refused a
// flash cast that also wanted another alternative cost.
//
// THREE SHAPES OF CONDITION, and the rules give each its own moment:
//
//   - A condition on the game (Bard's Company's "if you control a
//     Human", Fated Clash's "if a creature is attacking and a creature is
//     blocking"): CR 601.3, "A player can begin to cast a spell only if a
//     rule or effect allows that player to cast it", so it is read as the
//     cast BEGINS, by step 1c of castTimingOpenLocked. FlashIf.Condition.
//   - A condition on a choice made while casting (Silver Scrutiny's "if X
//     is 3 or less", Timely Ward's "if it targets a commander"): CR
//     601.3b, "that player may consider any choices to be made during
//     that spell's proposal that may cause that spell's qualities to
//     change. If any such choices could cause that effect to apply, that
//     player may begin to cast that spell as though it had flash." So the
//     window opens (Condition is nil) and the announcement must then
//     satisfy the condition: FlashIf.XCeiling and FlashIf.Targets, held
//     by castSpellLocked, the legal-move enumerator and the view through
//     FlashIfReliedOnLocked.
//   - A condition on a cost ("if you pay {2} more", "if you behold a
//     Dragon as an additional cost"): CR 601.3c. That is not this type at
//     all: it is an optional additional cost that opens the window when
//     announced (AdditionalCost.AsThoughFlash), which composes with an
//     alternative cost already.
//
// The choice-shaped conditions bind only a cast that NEEDS the
// permission. A Silver Scrutiny cast in its controller's main phase, or
// under a Vedalken Orrery, announces any X: some other rule lets it be
// cast then, and the engine reads the cast under that one, the choice
// every player makes (the same reading castUsesOwnFlashLocked gives
// Necromancy's rider).
//
// Catalog data, read through catalogDef: nothing on Game holds it, so it
// adds no route from Game to a func (closure_fields.txt).
type FlashIf struct {
	// Condition is the game-state condition, read as the cast begins
	// (CR 601.3). Nil for a condition that is only about the cast's own
	// choices, which CR 601.3b lets the caster begin at instant speed.
	Condition func(g *Game, you uuid.UUID) bool

	// XCeiling is "if X is N or less": the largest X a cast that needs
	// the permission may announce (CR 601.3b). Zero is no bound.
	XCeiling int

	// Targets is "if it targets <what>": the target clause a cast that
	// needs the permission is announced under, in place of the card's
	// own (CR 601.3b). It binds the announcement only (CR 601.2c); on
	// resolution the spell's printed clause governs (CR 608.2b), since
	// the condition is about casting. Nil is no narrowing.
	Targets *TargetSpec

	// Label is the clause as printed, for the view.
	Label string
}

// flashIfFor returns the card's own FlashIf, or nil.
func flashIfFor(card Card) *FlashIf {
	d := catalogDef(CatalogKey(card))
	if d == nil {
		return nil
	}
	return d.FlashIf
}

// FlashIfFor is flashIfFor by catalog key, for the view and tests.
func FlashIfFor(key string) *FlashIf {
	d := catalogDef(key)
	if d == nil {
		return nil
	}
	return d.FlashIf
}

// flashIfOpensLocked is step 1c of castTimingOpenLocked: the card prints
// a FlashIf whose game-state condition holds for `playerID` now, as the
// cast would begin (CR 601.3). A FlashIf with no Condition opens the
// window unconditionally (CR 601.3b); its choice-shaped conditions are
// held later, by FlashIfReliedOnLocked's readers.
//
// Caller must hold g.mu (read or write).
func (g *Game) flashIfOpensLocked(playerID uuid.UUID, card Card) bool {
	fi := flashIfFor(card)
	return fi != nil && (fi.Condition == nil || fi.Condition(g, playerID))
}

// FlashIfReliedOnLocked returns the card's FlashIf when the cast that
// claims `offer` (nil: the printed cost) under `perm`, announcing an
// optional cost that opens the window when `optionalFlash`, is open only
// BECAUSE of it: the window is open with the card's own permission and
// shut without it. Nil otherwise — the card prints none, another rule
// opens the window (its own type, the permission, the claimed offer, an
// optional cost, a per-player grant, or a sorcery's timing), or the
// window is shut either way, which the timing gate refuses on its own.
//
// Its readers hold the cast to the permission's choice-shaped condition
// (CR 601.3b): castSpellLocked refuses an X above XCeiling and judges the
// targets under Targets, and the enumerator and the view offer only that.
//
// Caller must hold g.mu (read or write).
func (g *Game) FlashIfReliedOnLocked(playerID uuid.UUID, card Card, zone ZoneKind, perm *CastPermission, offer *AlternativeCost, optionalFlash bool) *FlashIf {
	fi := flashIfFor(card)
	if fi == nil {
		return nil
	}
	if !g.CastTimingForCastOpenLocked(playerID, card, zone, perm, offer, optionalFlash) {
		return nil
	}
	if g.castTimingOpenLocked(playerID, card, zone, perm, offer.opensOwnCastWindow() || optionalFlash, false) {
		return nil
	}
	return fi
}

// XCeilingUnder folds the permission's "if X is N or less" into a
// ceiling the caller already has (`ok` false: none), keeping the
// smaller. A nil FlashIf, or one with no bound, returns the input.
func (fi *FlashIf) XCeilingUnder(ceiling int, ok bool) (int, bool) {
	if fi == nil || fi.XCeiling <= 0 {
		return ceiling, ok
	}
	if !ok || fi.XCeiling < ceiling {
		return fi.XCeiling, true
	}
	return ceiling, true
}

// TargetSpecUnder is the target clause a cast that relies on the
// permission is announced under: its Targets, or `base` when it narrows
// nothing (or fi is nil). Applied after the alternative cost's and the
// optional costs' rewrites, at announce only.
func (fi *FlashIf) TargetSpecUnder(base *TargetSpec) *TargetSpec {
	if fi == nil || fi.Targets == nil || base == nil {
		return base
	}
	return fi.Targets
}
