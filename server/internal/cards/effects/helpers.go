package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cardDied reports whether an EventLTB marks `source` going to the
// graveyard from the battlefield — i.e. it "died" (CR 700.4) — as
// opposed to being exiled, bounced, or tucked into the library.
// Dies-triggers gate their AppliesTo on this: a plain EventLTB fires
// for every battlefield exit, so matching on the event kind alone
// would mis-fire a "when ~ dies" ability on a bounce or a Path to
// Exile. The harvester stamps EventLTB.NewZone at every exit site.
// Added in S19 sub-PR 4.
func cardDied(ev game.Event, source *game.Card) bool {
	return ev.CardID == source.InstanceID && ev.NewZone == game.ZoneGraveyard
}

// IsBasicLand reports whether a card's type line contains the
// "basic land" supertype (case-insensitive substring). Used by
// tutor / fetch primitives that need to match Forest / Island /
// Mountain / Plains / Swamp / Wastes without enumerating each
// subtype. Path to Exile, Cultivate, and Solemn Simulacrum share
// this predicate; keep it here rather than duplicating the
// lowercase loop per-card.
func IsBasicLand(c game.Card) bool {
	return containsFoldASCII(c.TypeLine, "basic land")
}

// containsFoldASCII is a zero-alloc case-insensitive substring
// check for ASCII-only needles. The type lines we feed this come
// straight from Scryfall and are plain ASCII; if future text goes
// Unicode the caller should switch to strings.EqualFold + slicing.
func containsFoldASCII(haystack, needle string) bool {
	if len(haystack) < len(needle) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			h := haystack[i+j]
			if h >= 'A' && h <= 'Z' {
				h += 'a' - 'A'
			}
			if h != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// combatDamageToPlayerBy reports whether ev is combat damage dealt
// to a player by a creature that `controller` controls. The
// creature is looked up live: combat damage is dealt before SBAs
// run, so an attacker that traded with its blocker is still on the
// battlefield when its damage event fires. "Whenever a creature you
// control deals combat damage to a player" (Bident of Thassa,
// Coastal Piracy) gates on this. Added in S19 sub-PR 7.
//
// Caller must hold g.mu.
func combatDamageToPlayerBy(ev game.Event, controller uuid.UUID, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.Target); p == nil {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCreature() && src.Controller == controller
}

// --- Pirates deck helpers ----------------------------------------

// artifactEnteredUnderYourControl reports whether ev is an ETB for
// an artifact controlled by source's controller — the trigger
// condition shared by Reckless Fireweaver, Ingenious Artillerist and
// Quicksmith Genius.
//
// Batching gap (CR 603.1): the real cards read "whenever one or more
// artifacts you control enter", one trigger for a simultaneous
// batch. The engine emits one EventETB per card, so a mass token
// creation fires the trigger once per artifact instead of once with
// a count. For every card here that means the same total damage —
// it only diverges for a card that cares about the batch size in a
// nonlinear way.
func artifactEnteredUnderYourControl(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact() && c.Controller == source.Controller
}

// discardedByYou reports whether ev is a discard by source's
// controller.
func discardedByYou(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventDiscardCard && ev.Actor == source.Controller
}

// eventCardHasType reports whether the card just discarded has
// any of the given type-line words ("Island", "Pirate", "Vehicle").
// The card is read from the graveyard, where its printed type line
// is intact.
// Needles must be lowercase — containsFoldASCII folds the haystack,
// not the needle.
func eventCardHasType(ev game.Event, g *game.Game, words ...string) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	for _, w := range words {
		if containsFoldASCII(c.TypeLine, w) {
			return true
		}
	}
	return false
}

// damageToEachOpponent deals n damage to every opponent of the
// source's controller. The shared body of the "pings the table"
// pirates.
func damageToEachOpponent(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if err := g.DealDamageToPlayerForEffect(ctx.Source(), opp, n); err != nil {
			return err
		}
	}
	return nil
}

// lootOne draws a card then queues the discard choice — "draw a
// card, then discard a card" in the printed order, so the drawn card
// is a legal discard exactly as it is in paper.
func lootOne(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
		return err
	}
	g.DiscardChoiceForEffect(item.Controller, n)
	return nil
}

// --- S21 sub-PR 3: aristocrats helpers ---------------------------

// diedCreature resolves the creature that just died from a dies
// (EventLTB → graveyard) event: the card as it now sits in the
// graveyard, plus ok=false when the event isn't a creature death.
//
// The card is read post-move, so Tapped / Counters are already
// cleared (CR 400.7) but TypeLine, Controller and Owner survive —
// which is what "another creature YOU CONTROL dies" needs. A
// creature that stopped being a creature before it died is a
// sandbox gap: we read the printed type line rather than the
// battlefield LKI, because the LKI map is keyed to the dying card's
// own triggers and isn't reachable from a watcher's AppliesTo.
func diedCreature(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() {
		return game.Card{}, false
	}
	return c, true
}

// IsToken reports whether a card is a token. Token type lines are
// stamped "Token Creature — Spirit" by the templates in tokens.go;
// "Token" isn't one of the supertypes ParseTypeLine knows, so a
// substring check on the printed line is the reliable test.
func IsToken(c game.Card) bool {
	return containsFoldASCII(c.TypeLine, "token")
}

// --- S22: attack triggers ----------------------------------------

// attackDeclared reports whether ev is `source` itself being
// declared as an attacker — "whenever this creature attacks"
// (Krenko, Tin Street Kingpin).
//
// EventAttack carries the attacking creature in CardID, exactly as
// EventETB carries the entering permanent. That is deliberate: a
// card printed "whenever this creature enters or attacks" (Sun
// Titan) is ONE ability with two trigger conditions, and the shared
// field lets it watch both kinds with the bare
// `ev.CardID == source.InstanceID` rather than a kind switch.
func attackDeclared(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventAttack && ev.CardID == source.InstanceID
}

// attackDeclaredByYou reports whether ev is any creature controlled
// by `controller` being declared as an attacker — "whenever a
// creature you control attacks" (Hellrider). EventAttack stamps the
// attacking creature's controller in Actor at declaration time.
//
// This deliberately includes the source itself when the source is
// the creature that attacked: the printed text is "a creature you
// control", not "another". A card that wants the "another" reading
// adds `ev.CardID != source.InstanceID`.
func attackDeclaredByYou(ev game.Event, controller uuid.UUID) bool {
	return ev.Kind == game.EventAttack && ev.Actor == controller
}

// --- staples helpers (Commander staples pass) --------------------

// IsLandWithSubtype returns a SearchLibrary predicate matching any
// land whose type line carries `subtype`. Distinct from IsBasicLand:
// Nature's Lore fetches "a Forest card", which is any land with the
// Forest subtype — a Snow-Covered Forest or a Bayou both qualify,
// while IsBasicLand would admit an Island.
//
// Reads EFFECTIVE types and subtypes, which for the library-search
// callers (Nature's Lore, the fetchlands) is the printed type line
// verbatim — a card in a library has no layer cache. The difference
// shows up for the battlefield caller, the checkland condition
// youControlLandTyped: under Urborg, Tomb of Yawgmoth every land
// really is a Swamp, and Drowned Catacomb really does enter
// untapped off a Forest.
//
// Whole-token subtype match rather than a substring of the type
// line, so "Forest" no longer matches a card that merely has the
// word somewhere. Callers wanting "basic" semantics should compose
// with IsBasicLand instead.
func IsLandWithSubtype(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		return c.IsLand() && c.HasSubtype(subtype)
	}
}

// IsBasicLandExcept returns a predicate matching a basic land whose
// type line does NOT carry `subtype` — Farseek's "Plains, Island,
// Swamp, or Mountain" expressed as "a basic land that isn't a
// Forest". Cheaper and more future-proof than enumerating four
// subtypes, and it keeps the Wastes case out by accident rather
// than by omission (Wastes has no basic land type Farseek names,
// but it also isn't a Forest — a known sandbox over-permission,
// noted on the card).
func IsBasicLandExcept(subtype string) func(game.Card) bool {
	return func(c game.Card) bool {
		return IsBasicLand(c) && !containsFoldASCII(c.TypeLine, subtype)
	}
}

// controllerOfTarget resolves the controller of a targeted card, for
// the "its controller …" clause on Beast Within / Generous Gift /
// Nature's Claim. Returns ok=false when the target has left the
// battlefield between announce and resolution — the caller has
// already destroyed nothing in that case, so it just skips the
// rider rather than guessing whose token it is.
func controllerOfTarget(ctx *Context, target uuid.UUID) (uuid.UUID, bool) {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return uuid.Nil, false
	}
	return c.Controller, true
}

// --- shared effect bodies (folded by #583) -------------------------

func returnTargetCardToHand(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{Target: item.Targets[0].ID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}

func destroyChosenPermanent(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
}

// targetOpponentLosesAndYouGain is "target opponent loses n life and
// you gain n life" for a trigger whose target clause is a player. A
// target that is no longer legal is skipped, and the gain happens only
// when a target was drained. Vein Ripper.
func targetOpponentLosesAndYouGain(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
	}
	return nil
}

// counterTheTargetSpell is the whole OnResolve of "Counter target
// spell." — Cancel's body, named so a new card calls it rather than
// adding another copy to that clone family.
func counterTheTargetSpell(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
}

// destroyTheTargetPermanent is the whole OnResolve of "Destroy target
// [permanent]." — Bedevil's body, named for the same reason.
func destroyTheTargetPermanent(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
}
