package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// returnLegalGraveyardTargetsToHand returns every still-legal graveyard
// target to its owner's hand, in announce order — the body of the
// "return N target cards from your graveyard to your hand" spells whose
// count an optional cost raises (Peerless Recycling's gift, Blood
// Beckoning's kicker, #1716). A pick that left the graveyard in
// response is dropped by the CR 608.2b re-check before this runs.
func returnLegalGraveyardTargetsToHand(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

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

// resumeClause runs a prompt's continuation against the game the
// answer arrived in.
//
// The two lines every asynchronous clause in the catalog ends with: a
// nil `then` is a prompt nobody was waiting on, and a non-nil one gets
// a FRESH Context bound to the same stack item rather than one captured
// when the prompt was queued. An undo restores the game's fields in
// place, so the *Game a closure captured can be the wrong object by the
// time the player answers (the contract massEffect.apply spells out).
//
// Named because it is one idea rather than two coincidences: the clone
// gate found the body in ChoosePlayer and SacrificeChoice and the third
// copy would have been written the same way.
func resumeClause(g *game.Game, item *game.StackItem, then func(ctx *Context) error) error {
	if then == nil {
		return nil
	}
	return then(NewContext(g, item))
}

// IsBasicLand reports whether a card carries the Basic supertype
// (CR 205.4a) — read off Effective().Supertypes, not a "basic land"
// type-line substring. A Snow-Covered basic's type line is "Basic
// Snow Land — Plains": the word "Basic" and the word "Land" both
// appear, but never adjacent, so a substring match on "basic land"
// silently misses every snow basic. That was this function's
// original body, and it shipped 16 tutor / fetch call sites weaker
// than printed against a snow-basic-playing deck (#1334). Off the
// battlefield the effective supertypes are the printed ones.
//
// Used by tutor / fetch primitives that need to match Forest /
// Island / Mountain / Plains / Swamp / Wastes (and their snow
// counterparts) without enumerating each subtype. Path to Exile,
// Cultivate, and Solemn Simulacrum share this predicate; keep it
// here rather than duplicating the loop per-card.
func IsBasicLand(c game.Card) bool {
	if !c.IsLand() {
		return false
	}
	for _, s := range c.Effective().Supertypes {
		if s == "Basic" {
			return true
		}
	}
	return false
}

// lookAtTargetPlayersHandThenDraw is "Look at target player's hand.
// Draw a card." — Gitaxian Probe and Peek, which print the same two
// sentences at different prices.
//
// "Look at" is not "reveal", and that is the whole body: only the
// CASTER learns the hand, so each card is marked with the caster as a
// knower rather than routed through RevealHandForEffect, which would
// show the hand to the whole table — information neither printed card
// gives the other players.
//
// The hand is read at resolution, so a discard or a draw in response
// changes what is seen. A target player with no hand zone, or one who
// has left, is skipped and the draw still happens — the draw is not
// conditional on the look.
func lookAtTargetPlayersHandThenDraw(item *game.StackItem, ctx *Context) error {
	lookAtTargetPlayersHand(item, ctx)
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
}

// lookAtTargetPlayersHand is "Look at target player's hand": the body
// of lookAtTargetPlayersHandThenDraw without the draw, and Glasses of
// Urza's whole ability. Only the resolving effect's controller learns
// the hand (see above).
func lookAtTargetPlayersHand(item *game.StackItem, ctx *Context) {
	if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetPlayer {
		if p := ctx.PlayerByID(item.Targets[0].ID); p != nil && p.Hand != nil {
			for i := range p.Hand.Cards {
				p.Hand.Cards[i].AddKnower(ctx.Controller())
			}
		}
	}
}

// returnTargetedCardToHand returns the resolving item's first
// targeted card to hand — "return target artifact card from your
// graveyard to your hand" (Buried Ruin) and every "return another
// target X card in your graveyard to your hand" trigger (Junk Diver,
// Myr Retriever) that has no clause beyond the return itself. The
// target's zone is whatever TargetCardInGraveyard already restricted
// it to; this is only the resolution half.
func returnTargetedCardToHand(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return ReturnFromGraveyard{
		Target: item.Targets[0].ID,
		Dest:   game.ZoneHand,
	}.Apply(NewContext(g, item))
}

// IsBasicLandOfAnySubtype is "a basic <A>, <B>, or <C> card" — the
// narrowed fetch predicate the multicolour fetch lands print. Shared
// by the New Capenna Overlook cycle (b08OverlookLand) and the Alara
// Panorama cycle (b43PanoramaFetch), which name the same clause with
// different costs around it.
//
// Post-layer subtypes are irrelevant here: the card being matched is
// in a LIBRARY, where nothing changes its type line.
func IsBasicLandOfAnySubtype(subtypes ...string) func(game.Card) bool {
	want := append([]string(nil), subtypes...)
	return func(c game.Card) bool {
		if !IsBasicLand(c) {
			return false
		}
		for _, s := range want {
			if c.HasSubtype(s) {
				return true
			}
		}
		return false
	}
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

// damageToPlayerBy is combatDamageToPlayerBy without the CR 603
// combat-damage restriction. It exists for the printed text that
// says "deal damage" with no "combat" in it — Breeches, Brazen
// Plunderer and Malcolm, Keen-Eyed Navigator both print "Whenever
// one or more Pirates you control deal damage to your opponents",
// unlike Bident of Thassa, Coastal Piracy and every other caller of
// combatDamageToPlayerBy, which all print "combat damage" and must
// keep reading it that way.
//
// A SIBLING, not a broadened combatDamageToPlayerBy: docs/adding-cards.md's
// shared-file rule is append a function, never change an existing
// one's behaviour, and every existing caller of
// combatDamageToPlayerBy would silently widen if the combat check
// were dropped from it instead.
//
// Caller must hold g.mu.
func damageToPlayerBy(ev game.Event, controller uuid.UUID, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 {
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
// Batching gap: the real cards read "whenever one or more
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

// exileFromGraveyardIfStillThere is "exile that [discarded] card from
// your graveyard": a no-op if the card has already left — answered in
// response with a reanimation, a madness cast, or anything else that
// moves it first — because "from your graveyard" names an object in
// one zone (CR 400.7). Bag of Holding's own trigger and Currency
// Converter's share this body verbatim (#1218); the linked-exile
// identity each reads back afterward is b27ExiledWith, keyed on the
// caller's own (source, label) pair, not on anything this function
// does.
func exileFromGraveyardIfStillThere(g *game.Game, item *game.StackItem, cardID uuid.UUID) error {
	if z := g.FindCardZoneForEffect(cardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{Target: cardID}.Apply(NewContext(g, item))
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

// damageEachOpponentThenGainLife is "~ deals N damage to each opponent
// and you gain N life" as a whole effect (Y'shtola, Night's Blessed;
// Quintorius Kand).
func damageEachOpponentThenGainLife(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if err := damageToEachOpponent(g, item, n); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
	}
}

// damageToEachOpponent deals n damage to every opponent of the
// source's controller. The shared body of the "pings the table"
// pirates.
func damageToEachOpponent(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if err := g.DealDamageToPlayerForEffect(ctx.Source(), opp, n); err != nil {
				return err
			}
		}
		return nil
	})
}

// lootOne draws a card then queues the discard choice — "draw a
// card, then discard a card" in the printed order, so the drawn card
// is a legal discard exactly as it is in paper.
//
// The ORDER is the whole helper (#651). DrawCards resolves
// synchronously and QueueDiscardChoiceForEffect only queues a prompt,
// so the draw has to be the statement above: the prompt is built from
// the post-draw hand, and nothing after it may assume the cards are
// already in the graveyard. A card that reads the other way round
// ("discard a card, then ...") puts its second half in
// DiscardPrompt.Then instead.
func lootOne(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
		return err
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: item.Controller,
		Source: item.SourceCardID,
		N:      n,
	})
	return nil
}

// --- S21 sub-PR 3: aristocrats helpers ---------------------------

// diedCreature resolves the creature that just died from a dies
// (EventLTB → graveyard) event: the card as it now sits in the
// graveyard, plus ok=false when the event isn't a creature death.
//
// The card is read post-move, so Tapped / Counters are already
// cleared (CR 400.7) but TypeLine and Owner survive. Its controller is
// NOT read off it: "another creature YOU CONTROL dies" asks
// leftUnderControlOf(ev, card), the controller the exit stamped on the
// event (#1682), because a card in a graveyard has no controller
// (CR 108.4).
//
// Whether it WAS a creature is not read off that card (#1675). CR
// 603.10a: a leaves-the-battlefield ability looks back in time, and
// in the graveyard every effect that made the permanent a creature
// has stopped applying — a crewed Vehicle is an artifact again, an
// animated manland a land. So the test is the permanent's type as it
// last existed, which the exit stamps on the event (leftAsType).
func diedCreature(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !leftAsType(ev, c, "creature") {
		return game.Card{}, false
	}
	return c, true
}

// leftAsType reports whether the permanent an EventLTB names had card
// type cardType ("creature", "artifact", "land", …) as it last existed
// on the battlefield (#1675, CR 603.10a) — Event.LastKnownTypes, the
// post-layer types the exit read with the card still in play. c is
// the card as it sits now, and is consulted only for an event that
// carries no last-known types (one logged before the field existed,
// or built by hand in a test), which is the pre-#1675 reading.
//
// Every dies / leaves-the-battlefield condition that asks "was it a
// creature / an artifact / a land" asks here, never c.IsCreature():
// the card in the graveyard, hand or exile no longer has the types an
// effect gave it.
func leftAsType(ev game.Event, c game.Card, cardType string) bool {
	if was, known := ev.WasType(cardType); known {
		return was
	}
	return c.HasCardType(cardType)
}

// leftAsSubtype is leftAsType for a SUBTYPE (#1679, CR 603.10a):
// whether the permanent an EventLTB names was a Zombie / an Elf / an
// Egg as it last existed on the battlefield — Event.LastKnownSubtypes
// plus the "every creature type" flag, answered with Card.HasSubtype's
// semantics, so a printed changeling counts for every tribe and so
// does a creature under Maskwood Nexus or a lord's type grant. c is
// the card as it sits now, consulted only for an event that carries
// no last-known information.
//
// Every tribal dies / leaves-the-battlefield condition asks here,
// never dead.HasSubtype: in the graveyard no grant applies any more,
// so a creature that was a Zombie only through one reads as whatever
// it prints — and one an effect STOPPED being a Zombie reads as one
// again.
func leftAsSubtype(ev game.Event, c game.Card, subtype string) bool {
	if was, known := ev.WasSubtype(subtype); known {
		return was
	}
	return c.HasSubtype(subtype)
}

// leftAsSupertype is leftAsType for a SUPERTYPE (#1682, CR 603.10a):
// whether the permanent an EventLTB names was legendary / snow as it
// last existed on the battlefield — Event.LastKnownSupertypes, the
// post-layer supertypes the exit read with the card still in play. A
// Clone that copied a legend was legendary, and is a plain Clone in
// the graveyard. c is the card as it sits now, consulted only for an
// event that carries no last-known information.
func leftAsSupertype(ev game.Event, c game.Card, supertype string) bool {
	if was, known := ev.WasSupertype(supertype); known {
		return was
	}
	return c.HasSupertype(supertype)
}

// leftUnderControlOf is the player who controlled the permanent an
// EventLTB names as it last existed on the battlefield (#1682, CR
// 603.10a) — Event.LastKnownController. c is the card as it sits now,
// and its Controller is consulted only for an event that carries no
// stamp.
//
// Every "you control" / "an opponent controls" dies or
// leaves-the-battlefield condition asks here, never dead.Controller:
// "you control" is a question about the permanent, and a card in a
// graveyard has no controller (CR 108.4). A stolen creature sacrificed
// by its thief died under the THIEF's control, so the thief's
// Zulaport Cutthroat drains and the owner's does not. OWNER clauses —
// "its owner's graveyard", "under its owner's control", "a creature
// card you own" — read c.Owner and are not this question.
func leftUnderControlOf(ev game.Event, c game.Card) uuid.UUID {
	if controller, known := ev.LeftUnderControlOf(); known {
		return controller
	}
	return c.Controller
}

// leftAsColor is leftAsType for a COLOUR (#1689, CR 603.10a): whether
// the permanent an EventLTB names was colour `color` ("B" for black)
// as it last existed on the battlefield — Event.LastKnownColors, the
// post-layer colours the exit read with the card still in play. A
// creature that was black only through an effect (Darkest Hour, a
// black-making static, an Aura) was black on the battlefield and is
// whatever it prints in the graveyard. c is the card as it sits now,
// consulted only for an event that carries no last-known information.
//
// Teysa, Orzhov Scion's "whenever another black creature you control
// dies" asks here, never dead.HasColor: in the graveyard no effect
// applies any more, so a creature painted black by Darkest Hour reads
// as whatever it prints, and one an effect painted another colour
// still reads as its printed colour.
func leftAsColor(ev game.Event, c game.Card, color string) bool {
	if was, known := ev.WasColor(color); known {
		return was
	}
	return c.HasColor(color)
}

// diedWhileAttacking resolves the creature that just died from a dies
// event when it was ATTACKING as it died — "whenever an attacking
// creature dies" (Kardur, Doomscourge), Garna's "if it was attacking".
// The card is returned as it now sits in the graveyard, exactly as
// diedCreature returns it, so a "you control" clause reads its
// Controller the same way.
//
// The attack is read off the event, never the card: the exit took the
// creature out of combat and cleared Card.AttackingTarget before the
// event fired, and Event.AttackingTarget is the combat state as it
// last existed (#1661, CR 603.10a). So a creature that dies to combat
// damage, to a removal spell after blockers and to a sacrifice
// mid-combat all count, and one that was removed from combat first
// (a control change, CR 506.4) does not.
//
// No IsCreature check on the graveyard card, deliberately: only a
// creature attacks, and a crewed Vehicle that died attacking is an
// artifact again by the time it is in the graveyard. The event says
// what it was.
func diedWhileAttacking(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.AttackingTarget == uuid.Nil {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(ev.CardID)
}

// diedWhileBlocking is diedWhileAttacking for a creature that was
// BLOCKING as it died — "whenever a blocking creature dies" (Death
// Tyrant). Event.BlockingTarget names the attacker it was blocking.
func diedWhileBlocking(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.BlockingTarget == uuid.Nil {
		return game.Card{}, false
	}
	return g.LookupCardForEffect(ev.CardID)
}

// IsToken reports whether a card is a token. Token type lines are
// stamped "Token Creature — Spirit" by the templates in tokens.go;
// "Token" isn't one of the supertypes ParseTypeLine knows, so a
// substring check on the printed line is the reliable test. The check
// itself is game.Card.IsToken, so the engine and the catalog can never
// disagree about what a token is.
func IsToken(c game.Card) bool {
	return c.IsToken()
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

// IsLandWithAnySubtype is "a <A>, <B>, or <C> card" — any land
// carrying ANY of the named land subtypes, basic or not. Composes
// IsLandWithSubtype rather than re-scanning the type line, so it
// admits Hallowed Fountain and Indatha Triome (Plains Swamp Forest)
// exactly the way a fetchland's two-name landWithEitherSubtype does;
// this is that same shape generalised past two names.
//
// Farseek's "a Plains, Island, Swamp, or Mountain card" is this
// predicate over those four names — a land TYPE test, not a "basic"
// test, so it also excludes Wastes, which prints none of the four.
func IsLandWithAnySubtype(subtypes ...string) func(game.Card) bool {
	want := append([]string(nil), subtypes...)
	return func(c game.Card) bool {
		for _, s := range want {
			if IsLandWithSubtype(s)(c) {
				return true
			}
		}
		return false
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

// putTargetOnTopOfOwnersLibrary is "put target <permanent> on top of
// its owner's library" as a spell's whole effect — Submerge, Set
// Adrift. The tuck is the last instruction, so the fire-and-forget
// form is right: nothing reads where the card went.
func putTargetOnTopOfOwnersLibrary(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(item.Targets[0].ID, false)
}

func destroyChosenPermanent(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
}

// sourceDealsOneToFirstTarget is "{T}: this creature deals 1 damage to
// any target" — Goblin Sharpshooter's own ping and the ping Thornbite
// Staff GRANTS (ADR 0093), where the source is the equipped host. The
// ability's source deals 1 damage to the first chosen target.
func sourceDealsOneToFirstTarget(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 1}.Apply(NewContext(g, item))
}

// destroyFirstLegalCardTarget is the whole Effect of an activated
// ability whose printed text is "Destroy target [permanent]." — it
// re-checks legality through ctx.LegalTargets() (CR 608.2b) rather
// than trusting item.Targets[0] blindly, which matters for an
// activation whose target could leave the battlefield in response.
// Hopeful Initiate's and Staff of Compleation's destroy activations
// share this exact shape.
func destroyFirstLegalCardTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind == game.TargetCard {
			return DestroyTarget{Target: ref.ID}.Apply(ctx)
		}
	}
	return nil
}

// exileFirstLegalCardTarget is destroyFirstLegalCardTarget's sibling
// for "Exile target [permanent]." — the first card target still legal
// at resolution (CR 608.2b) is exiled, and nothing happens when none
// is. Teysa, Orzhov Scion's and Hanged Executioner's activations.
func exileFirstLegalCardTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, ref := range ctx.LegalTargets() {
		if ref.Kind == game.TargetCard {
			return ExileTarget{Target: ref.ID}.Apply(ctx)
		}
	}
	return nil
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

// returnFirstLegalGraveyardTarget returns the first still-legal card
// target of a single-target "return target … card from your graveyard
// to <zone>" ability to `dest`, under its owner's control. A target
// that left the graveyard in response is skipped (CR 608.2b). The two
// named forms below are what a card's Effect slot takes.
func returnFirstLegalGraveyardTarget(g *game.Game, item *game.StackItem, dest game.ZoneKind) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ReturnFromGraveyard{Target: t.ID, Dest: dest}.Apply(ctx)
		}
	}
	return nil
}

// returnFirstLegalGraveyardTargetToHand is "return target … card from
// your graveyard to your hand" — Codex Shredder, Samwise Gamgee.
func returnFirstLegalGraveyardTargetToHand(g *game.Game, item *game.StackItem) error {
	return returnFirstLegalGraveyardTarget(g, item, game.ZoneHand)
}

// returnFirstLegalGraveyardTargetToBattlefield is "return target …
// card from your graveyard to the battlefield" — Cauldron of Essence,
// Whisper, Blood Liturgist.
func returnFirstLegalGraveyardTargetToBattlefield(g *game.Game, item *game.StackItem) error {
	return returnFirstLegalGraveyardTarget(g, item, game.ZoneBattlefield)
}

// counterTheTargetSpell is the whole OnResolve of "Counter target
// spell." — Cancel's body, named so a new card calls it rather than
// adding another copy to that clone family.
//
// It is also the whole OnResolve of "Counter target spell, activated
// ability, or triggered ability" (#1211): CounterTarget takes a STACK
// ITEM id and has discriminated on item.Kind since S13.1, and a
// picked ability arrives in the same TargetRef a picked spell does.
// The name keeps the spell in it because the clause is what differs
// between the two families and the body is not — see Disallow,
// Voidslime, Stifle and Tale's End, which all call this.
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

// payLifeThenDraw is the "you may pay N life. If you do, draw M
// cards" body — Crossway Troublemakers' and Erebos, Bleak-Hearted's.
// The "may" is answered before this runs, by the trigger's optional
// prompt; this is only the payment and the linked draw.
//
// Paying life is a COST (CR 118.3), so "if you do" has to know the
// payment finished before the draw happens (#793) — which is why
// this is one body and not two primitives in a Do().
//
// CR 119.4: a player cannot pay life they do not have, and life can
// move between the question and the answer. An unaffordable payment
// degrades to the decline rather than erroring, the same shape
// ResolveEntryPayLife uses — so neither the life nor the cards move.
//
// Caller holds g.mu.
func payLifeThenDraw(ctx *Context, life, cards int) error {
	controller := ctx.Controller()
	p := ctx.Game.PlayerByIDForEffect(controller)
	if p == nil || p.Eliminated || p.Life < life {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), controller, life); err != nil {
		return err
	}
	return DrawCards{Player: controller, N: cards}.Apply(ctx)
}

// tapChosenPermanent is the whole Effect of an activated ability whose
// printed text is "Tap target <permanent>." — Staff of Domination's
// fourth ability and Ring of the Lucii's second. Named because two
// cards shipping the same six lines is one helper waiting to exist
// (#583).
func tapChosenPermanent(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return TapTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
}

// plusOneCounterOnChosen is "put a +1/+1 counter on target creature" as
// a trigger or ability Effect — Triumph of Gerrard's first two chapters
// and Pridemalkin's enters trigger.
func plusOneCounterOnChosen(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return AddCounter{
		Target: item.Targets[0].ID,
		Kind:   game.CounterPlusOne,
		N:      1,
	}.Apply(NewContext(g, item))
}

// regenerateTheTargetPermanent is the whole Effect of "{cost}:
// Regenerate target [permanent]." — Asceticism's {1}{G} ability,
// Welding Jar's sacrifice, Goblin Chirurgeon's. What the three may
// point at differs and is declared on each ability's TargetSpec; what
// they DO is one line (CR 701.19a).
//
// A target that has left the battlefield by the time the ability
// resolves is a no-op (CR 701.19b), which the primitive handles;
// reading the target through LegalTargets is what makes an ability
// whose ONLY target is gone fizzle properly instead.
func regenerateTheTargetPermanent(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return Regenerate{Target: id}.Apply(ctx)
}

// destroyTheTargetPermanentNoRegen is the same sentence with the
// CR 701.19c rider on the end — "Destroy target [permanent]. It can't
// be regenerated." Terminate, Mortify and Putrefy are that sentence
// and nothing else; they differ only in what their TargetSpec lets
// them point at, which is declared on the Spec rather than written
// here.
//
// It exists because #667 made the rider real: the three bodies were
// already identical and became a new clone family the moment they all
// grew the same extra field. One helper, three callers.
func destroyTheTargetPermanentNoRegen(item *game.StackItem, ctx *Context) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: item.Targets[0].ID, CantBeRegenerated: true}.Apply(ctx)
}

// returnThisCardFromYourGraveyard is the whole effect of a graveyard
// trigger that brings its own card back — Bloodghast's landfall,
// Narcomoeba's arrival (#925). The card is its own payload, so it
// reads the ID off item.SourceCardID rather than out of a closure,
// which is what lets both cards share one package-level func.
func returnThisCardFromYourGraveyard(g *game.Game, item *game.StackItem) error {
	return ReturnFromGraveyard{
		Target: item.SourceCardID,
		Dest:   game.ZoneBattlefield,
	}.Apply(NewContext(g, item))
}

// damageToFirstTarget is "~ deals N damage to any target" — the whole
// body of Lightning Bolt and of Rift Bolt, which print the same
// sentence at different prices and with a keyword between them.
//
// It reads the FIRST target slot and nothing else, which is what a
// single "any target" clause fills, and does nothing when the slot is
// empty (CR 608.2b — every target became illegal, so the spell was
// countered by game rules before it got here; the guard is belt and
// braces).
func damageToFirstTarget(amount int) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 {
			return nil
		}
		return DealDamage{
			Source: ctx.Source(),
			Target: item.Targets[0].ID,
			Amount: amount,
		}.Apply(ctx)
	}
}

// EachLandIsAlso is "each land is a <basic land type> in addition to
// its other land types" — Urborg's sentence with the type as an
// argument, which is Yavimaya, Cradle of Growth.
//
// One layer-4 static, and an APPEND rather than a set: "in addition
// to" takes nothing away, so an Island keeps {U} and keeps its printed
// ability in slot 0. The intrinsic mana ability for the added type is
// not declared here — game.ManaAbilitiesForCard derives it from the
// EFFECTIVE subtypes (CR 305.6), which is the whole reason the
// sentence does anything.
//
// "Each land" is every land on the battlefield under every player's
// control, the source included. IsLand is read through the effective
// view on purpose: a permanent another layer-4 effect made a land is
// one, and CR 613.8a then makes this depend on that effect rather
// than race it by timestamp (ADR 0067).
//
// Contrast SetsBasicLandType, which is CR 305.7's REPLACEMENT (Magus
// of the Moon, Blood Moon) and takes the land's own rules text with
// it.
// putIndestructibleCounterOnSource is the ability effect body
// behind "…: Put an indestructible counter on this permanent" (Tekuthal,
// Inquiry Dominus; Solphim, Mayhem Dominus; Drivnod; Zopandrel): a no-op if something
// killed the source before the ability resolves, otherwise a counter
// on the source itself. The indestructible counter needs nothing
// else: the engine reads keyword counters itself (CR 122.1b, ADR 0101).
// plusOneCountersOnThis is "Put N +1/+1 counters on this creature" —
// the body most of the exhaust cards print (Prowcatcher Specialist,
// Greenbelt Guardian, Afterburner Expert, Elvish Refueler, Boom
// Scholar) and a common one outside them.
//
// Unlike putIndestructibleCounterOnSource beside it, it does NOT
// check that the source is still on the battlefield: AddCounter is a
// no-op on a card that has gone, and the two cards that read that
// check place a keyword counter on a source that must still be there.
// Keep them separate rather than merging them into one flagged helper.
func plusOneCountersOnThis(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return AddCounter{
			Target: item.SourceCardID,
			Kind:   game.CounterPlusOne,
			N:      n,
		}.Apply(NewContext(g, item))
	}
}

func putIndestructibleCounterOnSource() Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !b15OnBattlefield(g, item.SourceCardID) {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: game.CounterIndestructible, N: 1}.
			Apply(NewContext(g, item))
	}
}

func EachLandIsAlso(subtype string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsLand()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, st := range c.Subtypes {
				if strings.EqualFold(st, subtype) {
					return
				}
			}
			c.Subtypes = append(c.Subtypes, subtype)
		},
	}
}

// permanentsControlledByMatching lists the permanents `playerID`
// controls that satisfy `pred`, in battlefield order — the candidate
// set behind "sacrifice an artifact of your choice" and its
// relatives.
//
// The fixed-predicate versions that predate it (landsControlledByPlayer,
// creaturesControlledByPlayer, NonlandPermanentsControlledBy) stay as
// they are; this is the shape for a clause whose filter is an ordinary
// CardPredicate, so the card file writes Artifact() rather than another
// battlefield loop.
//
// Caller must hold g.mu — it is an effect-time read.
func permanentsControlledByMatching(g *game.Game, playerID uuid.UUID, pred CardPredicate) []uuid.UUID {
	if g == nil || playerID == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != playerID {
			continue
		}
		if pred != nil && !pred(g, playerID, c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// spellManaValueForEffect is "that spell's mana value" for a card that
// may or may not still be findable: zero when it is in no zone at all,
// and otherwise the CR 202.3e read that counts {X} only while the card
// is on the stack (#788).
//
// Caller must hold g.mu.
func spellManaValueForEffect(g *game.Game, cardID uuid.UUID) int {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return 0
	}
	mv, _ := g.ManaValueForEffect(c)
	return mv
}

// firstLegalPlayerTarget returns the first still-legal player slot on
// the item being resolved, or false when there is none. The CR 608.2b
// read for every single-player-target card: a target that became
// illegal in response is skipped, and a card whose only target is
// gone does nothing rather than erroring.
func firstLegalPlayerTarget(ctx *Context) (uuid.UUID, bool) {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return t.ID, true
		}
	}
	return uuid.Nil, false
}

// notACreature strips the Creature card type, and the creature
// subtypes that rode on it, from a characteristic being built in
// layer 4. Not a CR citation — the Comprehensive Rules don't mandate
// erasing a permanent's printed creature subtypes when it stops being
// a creature; this engine does it anyway, for the reason below.
//
// Shared by every "as long as <condition>, this isn't a creature"
// clause — The Warring Triad's graveyard gate, impending's time
// counters — because the second half is the half that gets forgotten.
// Dropping Creature and leaving the subtypes behind leaves a God or an
// Avatar Horror that is not a creature, which reads as a bug on the
// card and, under a Maskwood Nexus, is one: the permanent would still
// be every creature type while not being a creature at all (#670).
//
// It does NOT touch power and toughness. A characteristic with no
// Creature type has no P/T that anything reads, and layer 7 runs
// after layer 4 in any case.
func notACreature(c *game.Characteristic) {
	kept := make([]string, 0, len(c.Types))
	for _, t := range c.Types {
		if t != "Creature" {
			kept = append(kept, t)
		}
	}
	c.Types = kept
	c.SetSubtypes(nil)
}

// EnteringPermanentChooser is the `Controller` hook every entry
// replacement that asks its own question shares: WHO is asked.
//
// One copy rather than three (#1198). The clone gate found the same
// sixteen lines in `EntersAsCopyOf` (copy_effects.go, CR 707.2's
// "enter as a copy of…?"), `EntersTappedUnlessYouPayLife`
// (shocklands.go) and `EntersTappedUnlessYouRevealFromHand`
// (reveal_lands.go), and it is the same RULE in all three rather than
// merely the same text: the question belongs to whoever is putting
// the permanent onto the battlefield.
//
// `ev.Actor` first, because the battlefield-entry path stamps
// Card.Controller only AFTER the replacement pipeline has run — so at
// this point src.Controller is whatever the card carried in the zone
// it is leaving, and the Actor is the one field guaranteed correct.
// The card's own controller, then its owner, are the fallbacks for an
// entry driven by something that stamped no actor.
func EnteringPermanentChooser(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
	if ev != nil && ev.Actor != uuid.Nil {
		return ev.Actor
	}
	if src == nil {
		return uuid.Nil
	}
	if src.Controller != uuid.Nil {
		return src.Controller
	}
	return src.Owner
}

// untapTheTarget is the Effect body for an ability whose whole
// instruction is "Untap target <something>". Four cards print it
// behind four different costs — Magewright's Stone ({1}, {T}),
// Wirewood Lodge ({G}, {T}), Quirion Ranger (return a Forest) and
// Wirewood Symbiote (return an Elf) — and the clause that differs
// between them is the TARGET clause, which lives on the ability, not
// this body.
//
// It walks ctx.LegalTargets() rather than indexing item.Targets so a
// target that became illegal in response is skipped (CR 608.2b) and
// the ability resolves doing nothing, which is the printed outcome.
// Named in #1213, where the third and fourth copies would have been
// written.
func untapTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return UntapTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// destroyEachLegalTarget is "destroy target …" for a clause of any
// count: every card target still legal at resolution is destroyed, and
// one that left in response is skipped (CR 608.2b). Curtains' Call's
// "two target creatures" and Wear Down's gift-promised "two target
// artifacts and/or enchantments" — and the one-target unpromised Wear
// Down, which is the same loop over one slot.
func destroyEachLegalTarget(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// putPlusOneCounterOnEachLegalTarget places one +1/+1 counter on each
// still-legal target, skipping one that left or stopped qualifying in
// response (CR 608.2b). Shared by Phyrexian Scriptures ("put a +1/+1
// counter on each of X target creatures") and Rishkar, Peema Renegade
// ("put a +1/+1 counter on each of up to two target creatures") — the
// same loop as destroyEachLegalTarget, one primitive over.
func putPlusOneCounterOnEachLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// plusOneCounterPlacementOnYourCreature is Hardened Scales' and
// Branching Evolution's shared predicate: "if one or more +1/+1
// counters would be PUT ON a creature you control" (CR 122.6 / CR
// 614.1 — placement only, ev.CounterDelta > 0 excludes a removal,
// #1291). Named here because both cards' AppliesTo bodies were
// byte-identical and the clone gate found the third growing copy
// before it could land.
func plusOneCounterPlacementOnYourCreature(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventCounter {
		return false
	}
	if ev.CounterDelta <= 0 {
		return false
	}
	if ev.CounterName != "+1/+1" {
		return false
	}
	target, ok := g.LookupCardForEffect(ev.CounterTarget)
	if !ok {
		return false
	}
	if !target.IsCreature() {
		return false
	}
	return target.Controller == src.Controller
}

// drawOnePerPicked is the "if you do, draw that many cards" tail of a
// resolution-time sacrifice pick: `player` draws one card for each
// permanent chosen, and nothing when none was. God-Eternal Bontu's and
// Sephiroth's shared Then (the clone gate found the second copy).
func drawOnePerPicked(player uuid.UUID) func(ctx *Context, picked game.PromptedPicks) error {
	return func(ctx *Context, picked game.PromptedPicks) error {
		n := picked.Count()
		if n == 0 {
			return nil
		}
		return DrawCards{Player: player, N: n}.Apply(ctx)
	}
}

// drainTargetOpponentOne is "target opponent loses 1 life and you gain
// 1 life" as a trigger's resolution: the first still-legal player
// target loses 1, then the item's controller gains 1. A target that
// left the game has already countered the ability (CR 608.2b), so
// nothing happens. Shared by Vengeful Bloodwitch, Sephiroth, Fabled
// SOLDIER and Sephiroth's Super Nova emblem.
func drainTargetOpponentOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -1); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
	}
	return nil
}

// dealOneDamageToPlayerWhoDrew is "whenever an opponent draws a card,
// this [creature/enchantment] deals 1 damage to that player" —
// Underworld Dreams and Fate Unraveler. The drawer is read off the
// item's triggering event (item.Trigger.Event.Actor, ADR 0041 P9)
// rather than captured, so a table with this trigger waiting on the
// stack is a restore point. A drawer who has since left the game
// takes no damage.
func dealOneDamageToPlayerWhoDrew(g *game.Game, item *game.StackItem) error {
	drawer := item.Trigger.Event.Actor
	if g.PlayerByIDForEffect(drawer) == nil {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: drawer, Amount: 1}.Apply(NewContext(g, item))
}

// damagedCreatureReflectsChosenAmount is "whenever a [Dragon/Dinosaur]
// you control is dealt damage, it deals that much damage to any
// target that isn't a [Dragon/Dinosaur]" — Wrathful Red Dragon and
// Wrathful Raptors. The damaged permanent and the amount are read off
// the item's triggering event (item.Trigger.Event, ADR 0041 P9)
// rather than captured.
func damagedCreatureReflectsChosenAmount(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.Trigger.Event.Target, Target: item.Targets[0].ID, Amount: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
}

// thatPlayerLosesOneLife is "you may have that player lose 1 life" —
// Blood Seeker's and Suture Priest's second ability, both triggered
// by a creature an opponent controls entering: the entering
// creature's controller, read off the trigger's last-known object
// (item.Trigger.Object, ADR 0041 P9), loses the life, and a departed
// player is asked nothing.
func thatPlayerLosesOneLife(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	victim := item.Trigger.Object.Controller
	if g.PlayerByIDForEffect(victim) == nil {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1)
}

// sourceDealsDamageToEachLegalTarget is the Effect "<this> deals N
// damage to <its target>": the ability's source deals `amount` to every
// target still legal at resolution (CR 608.2b). Niv-Mizzet, the
// Firemind's draw trigger and Balduvian Trading Post's "1 damage to
// target attacking creature" share it.
func sourceDealsDamageToEachLegalTarget(amount int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		return ctx.Game.DamageInstanceForEffect(func() error {
			for _, t := range ctx.LegalTargets() {
				if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
}

// bounceTheTarget is the OnResolve "Return target <permanent> to its
// owner's hand" as a spell's whole effect (Void Squall): the first
// target still legal as it resolves (CR 608.2b).
func bounceTheTarget(_ *game.StackItem, ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return BounceToHand{Target: id}.Apply(ctx)
}

// flickerTheTargetToItsOwner is the OnResolve "Exile target creature
// you control, then return it to the battlefield under its owner's
// control" (Momentary Blink, Ephemerate): the creature comes back as a
// new object (CR 400.7) under its OWNER's control, which is the
// difference from Cloudshift's "under your control".
func flickerTheTargetToItsOwner(_ *game.StackItem, ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return Flicker{Target: id}.Apply(ctx)
}

// boostTheTargetUntilEOT is the OnResolve "Target creature gets +P/+T
// until end of turn": the one target slot, pumped until cleanup (CR
// 514.2). A target that became illegal never gets here (CR 608.2b: a
// spell whose only target is illegal is countered by the game rules).
// Artful Maneuver and Prey's Vengeance print it.
func boostTheTargetUntilEOT(power, toughness int, label string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		return BoostUntilEOT{Target: item.Targets[0].ID, Power: power, Toughness: toughness, Label: label}.Apply(ctx)
	}
}

// boostTheTargetAndUnblockableThisTurn is the OnResolve "Target
// creature gets +P/+0 until end of turn and can't be blocked this
// turn" (Distortion Strike, Taigam's Strike): the pump and the
// restriction, both ending at cleanup.
func boostTheTargetAndUnblockableThisTurn(power int, name string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		target := item.Targets[0].ID
		if err := (BoostUntilEOT{Target: target, Power: power, Label: name + " — power boost"}).Apply(ctx); err != nil {
			return err
		}
		return RestrictUntilEOT{Target: target, Restrictions: game.CantBeBlocked, Label: name + " — can't be blocked"}.Apply(ctx)
	}
}

// protectionFromAChosenColorForTheTarget is the OnResolve "Target
// creature you control gains protection from the color of your choice
// until end of turn" (Center Soul, Emerge Unscathed): the colour is
// chosen as the spell resolves, and asked as the colour the creature is
// protected FROM (#780), so a bot names the biggest threat. A target
// that left in response is asked nothing.
func protectionFromAChosenColorForTheTarget(name string) func(item *game.StackItem, ctx *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		var target game.TargetRef
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				target = t
			}
		}
		if target.Kind != game.TargetCard {
			return nil
		}
		ChooseColorThen(game.ColorForProtection, ctx.Game, item.Controller, item.SourceCardID,
			name+" — choose a color",
			func(g *game.Game, color string) error {
				token := game.ProtectionFromColor(color)
				if token == "" {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   target.ID,
					Keywords: []string{token},
					Label:    name + " — " + token,
				}.Apply(NewContext(g, item))
			})
		return nil
	}
}

// tapTheFirstLegalTarget is "Tap target <permanent>" as an ability's
// whole effect: the first target still legal as it resolves (CR
// 608.2b). Citadel Siege's Dragons mode and Storm Elemental share it.
func tapTheFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return TapTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

// returnThisPermanentToOwnersHand is "Return this <permanent> to its
// owner's hand" as an ability's whole effect. It reads the source as
// the object the ability came from (CR 400.7): one that left and came
// back in response is a new object and stays where it is. Quicksilver
// Wall and Leashling share it.
func returnThisPermanentToOwnersHand(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// returnThisCardFromYourGraveyardToHand is "return this card from your
// graveyard to your hand" (Ringwraiths' trigger, Gollum, Patient
// Plotter's ability) and Spine of Ish Sah's "return it to its owner's
// hand" from the graveyard it died into. The card may have left the
// graveyard before the ability resolves; then it does nothing, since
// the card it names is gone.
func returnThisCardFromYourGraveyardToHand(g *game.Game, item *game.StackItem) error {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
