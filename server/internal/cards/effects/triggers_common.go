package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// triggers_common.go — constructors for the printed trigger shapes
// (#579, Discussion #558).
//
// A card file names the shape and the effect, and nothing else:
//
//	Triggered: []game.TriggeredAbility{
//	    WhenThisEnters("Adventurer's Inn — you gain 2 life", Do(GainLife{Amount: 2})),
//	    WheneverYouCast(Noncreature(), "Black Waltz No. 3 — 2 damage to each opponent",
//	        func(g *game.Game, item *game.StackItem) error { return damageToEachOpponent(g, item, 2) }),
//	    AtYourUpkeep("Awakening Zone — create an Eldrazi Spawn", Do(CreateToken{Template: EldraziSpawnToken(), N: 1})),
//	}
//
// Every constructor returns an ordinary game.TriggeredAbility —
// Watches, an AppliesTo, a Key that is the stack label, and the Effect
// DECLARED on the row (ADR 0041 P9, #1497, tier 4-2). The engine builds
// the item from the row itself (NewTriggeredItem(source, Key, Effect))
// and names the row on it, so a table with the trigger waiting on the
// stack is still a restore point. The S19 rules (ADR 0018) hold
// unchanged: nothing resolves before the stack says so, and the Effect
// reads the controller, source, targets and triggering event
// (item.Trigger) off the item it is handed and captures no *Card or
// *Game — and nothing that was only true when the ability triggered,
// because a restored item runs the row's Effect again.
//
// The two rules for using them:
//
//   - The label is the whole stack label, "<card> — <what happens>",
//     exactly as a hand-written NewTriggeredItem call would carry it.
//     Nothing is prefixed. A helper that counts resolutions by label
//     therefore keeps working when a card migrates.
//   - Do(...) takes primitive VALUES, so it cannot read the event or
//     the item. An effect that needs item.Targets, ev.Actor, or
//     anything else from resolution time is written as an Effect
//     closure, as before. Do is sugar for the common case, not the
//     only case.
//
// Why not derive the label from source.Name: two tallies in the
// catalog key on the exact label string, and a copied token carries
// the copied card's name, which is right for the rules but would
// silently rename the item on the stack.

// When is a trigger condition — the AppliesTo signature, named so a
// constructor can take one. Every named predicate below is a When,
// and any existing AppliesTo function or closure is one too.
type When = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool

// Effect is what a trigger does when its item resolves — the
// StackItem.Effect signature. It receives the live Game and the item;
// it must not capture a *Card or the *Game from Build (undo restores
// a cloned game and the closure has to resolve against that one).
type Effect = func(g *game.Game, item *game.StackItem) error

// Applier is any primitive value: every effects.X{…} has Apply(ctx).
type Applier interface {
	Apply(ctx *Context) error
}

// Do sequences primitive values into an Effect, applied in order
// against one Context built from the resolving item. A primitive's
// Player / Controller field left zero defaults to the item's
// controller (GainLife, DrawCards, MillCards, DiscardCards,
// CreateToken, Scry, Surveil, LookAtTop, AddMana), so the common
// "you gain 2 life" needs no ID at all.
//
// The first error stops the sequence and is reported as the item's
// effect error; the item has still resolved (CR 608.2m).
func Do(steps ...Applier) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, s := range steps {
			if err := s.Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// On is the general constructor: the ability watches one event kind,
// fires when `when` says so, and goes on the stack as `label` with
// `effect`. The named constructors below are all On (or OnAny) with a
// predicate filled in; reach for On directly when the printed
// condition has no name yet, and pass any AppliesTo you would have
// written by hand.
func On(kind game.EventKind, when When, label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{kind}, when, label, effect)
}

// OnAny is On for one printed ability with two or more trigger
// conditions — "Whenever ~ enters or attacks" is one ability watching
// two kinds, not two abilities (Sun Titan).
//
// The effect is declared on the row (ADR 0041 P9): the engine builds
// the item as NewTriggeredItem(source, label, effect) and stamps it with
// the row's catalog name. A nil effect keeps the old hand-built item
// with no Effect, for the one card that replaces the Build afterwards
// (Forum Familiar) — a row with neither would never trigger at all.
func OnAny(kinds []game.EventKind, when When, label string, effect Effect) game.TriggeredAbility {
	t := game.TriggeredAbility{
		Watches:   kinds,
		AppliesTo: when,
		Key:       label,
		Effect:    effect,
	}
	if effect == nil {
		t.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label, nil)
		}
	}
	return t
}

// Optional makes a trigger a CR 603.5 "you may": the harvester asks
// `question` before Build runs, and a "No" drops the trigger.
func Optional(t game.TriggeredAbility, question string) game.TriggeredAbility {
	t.OptionalPrompt = &game.TriggerOptionalPrompt{Question: question}
	return t
}

// OncePerBatch marks a "whenever ONE OR MORE …" ability (#587): the
// engine emits one event per creature that attacks, enters or deals
// damage, and this makes the harvester fire on the FIRST event of a
// batch and decline every later event of the same batch (#829). A
// batch is every event between two points where play moves on — a
// stack item beginning to resolve, or the turn cursor entering a new
// step — stamped on Event.Batch and checked by
// oncePerBatchAllowsLocked; see AGENTS.md §7. Matched by the
// ability's label, which the constructors stamp as its Key.
func OncePerBatch(t game.TriggeredAbility) game.TriggeredAbility {
	t.OncePerBatch = true
	return t
}

// OncePerBatchPerPlayer is OncePerBatch for a clause that names A
// PLAYER (#784, CR 603.2c): the batch collapses per player rather
// than per batch, so three creatures connecting with three opponents
// in one damage step are three triggers and two creatures hitting one
// opponent are one. Keeper of Fables' ruling (2019-10-04) states it
// for the damage wording; Horizon Explorer's and Neyali's state it
// for "attack a player".
//
// The extra dimension is PerPlayer below, appended to the ability's
// Key by the engine's own guard — one key, one guard, no second
// dedupe path.
func OncePerBatchPerPlayer(t game.TriggeredAbility) game.TriggeredAbility {
	t.OncePerBatch = true
	t.BatchKey = PerPlayer
	return t
}

// PerPlayer is the TriggeredAbility.BatchKey reading for "… to a
// player" / "… attack a player": the player the event names. That is
// the damaged player for a damage event, and the defending player an
// attack ends on for an attack declaration — read through
// b17DefendingPlayer, so an attack at that player's planeswalker or
// battle counts as the same player and not a second one.
func PerPlayer(ev game.Event, _ *game.Card, g *game.Game) string {
	if ev.Kind == game.EventAttack {
		return b17DefendingPlayer(g, ev).String()
	}
	return ev.Target.String()
}

// WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer is
// the printed shape of Professional Face-Breaker, Keeper of Fables,
// Grazilaxx, Rapacious Guest, Thopter Spy Network and Olivia: one
// trigger per player the controller's creatures connect with in a
// combat damage step (CR 603.2c), whatever their number.
//
// `creature` narrows which of the controller's creatures count —
// non-Human, artifact, Faerie, outlaw — read off the damage SOURCE
// post-layer. Nil counts every creature they control.
func WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(creature CardPredicate, label string, effect Effect) game.TriggeredAbility {
	return OncePerBatchPerPlayer(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if !combatDamageToPlayerBy(ev, source.Controller, g) {
			return false
		}
		if creature == nil {
			return true
		}
		c, ok := g.LookupCardForEffect(ev.Source)
		return ok && creature(g, source.Controller, c)
	}, label, effect))
}

// Targeting gives a trigger a target clause (CR 603.3d — chosen as
// the ability goes on the stack, re-checked on resolution). The
// Effect reads the pick from item.Targets[0]; the constructors'
// Do(...) form cannot, so a targeted trigger takes an Effect closure.
func Targeting(t game.TriggeredAbility, spec *game.TargetSpec) game.TriggeredAbility {
	t.Targets = spec
	return t
}

// --- conditions -----------------------------------------------------

// Self — the event names the source: an ETB, an attack declaration, a
// "becomes the target" for the permanent itself.
func Self(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.CardID == source.InstanceID
}

// ByYou — the event's actor is the source's controller: your upkeep,
// your end step, your draw step, a card you drew, a spell you cast.
func ByYou(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller
}

// ByAnOpponent — the event's actor is a seated player other than the
// source's controller. Excludes the actor-less admin / SBA events.
func ByAnOpponent(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor != uuid.Nil && ev.Actor != source.Controller
}

// AnyPlayer — no condition beyond the event kind: "at the beginning
// of each upkeep", "whenever a player draws a card".
func AnyPlayer(game.Event, *game.Card, game.Characteristic, *game.Game) bool { return true }

// ThisDied — the source went from the battlefield to a graveyard
// (CR 700.4). A bounce or an exile is not a death.
func ThisDied(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return cardDied(ev, source)
}

// ThisAttacked — the source was declared as an attacker.
func ThisAttacked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return attackDeclared(ev, source)
}

// ThisBecameTapped — "whenever this creature becomes tapped" (CR
// 701.26a). Two event kinds, for the reason b11DwarfYouControlBecameTapped
// gives: the engine taps an attacker without an EventTapCard, so an
// EventAttack naming the source while it is now tapped is "became
// tapped", and a vigilance attacker did not become tapped.
//
// Watch both game.EventTapCard and game.EventAttack. Written for a
// GRANTED trigger (ADR 0093 — Dionus's "this creature"), where
// `source` is the host that has the ability.
func ThisBecameTapped(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.CardID != source.InstanceID {
		return false
	}
	switch ev.Kind {
	case game.EventTapCard:
		return true
	case game.EventAttack:
		return source.Tapped
	}
	return false
}

// YouCastYourSecondSpellEachTurn — "Whenever you cast your second
// spell each turn" (Breeches, the Blastmaker; Avatar Yangchen). The
// cast path bumps the per-turn tally BEFORE it emits EventCast, so a
// total of exactly two means "this is the second" — the same clock
// Maelstrom Nexus reads for "first".
func YouCastYourSecondSpellEachTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	return g.CastTallyFor(ev.Actor).Total == 2
}

// YouCast — you cast a spell matching `spell` (nil for any spell).
// The predicate is the same CardPredicate the target clauses use, so
// "whenever you cast a noncreature spell" is YouCast(Noncreature()).
func YouCast(spell CardPredicate) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor != source.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (spell == nil || spell(g, source.Controller, c))
	}
}

// AnOpponentSearchesTheirOwnLibrary — "Whenever an opponent searches
// THEIR library" (Archivist of Oghma, Wan Shi Tong, Librarian). Watch
// game.EventSearchLibrary.
//
// Two things this rules out, both load-bearing: a plain shuffle
// (Label == "shuffle") is not a search, so a fetchland doesn't
// double-trigger it; and Target — #1335's library-owner field — must
// equal Actor, so Bribery-style search of a DIFFERENT player's
// library (Actor searching, Target's library) does not read as "the
// searcher's own", which is stronger than printed.
func AnOpponentSearchesTheirOwnLibrary(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.Label == "shuffle" {
		return false
	}
	return ev.Actor != uuid.Nil && ev.Actor != source.Controller && ev.Actor == ev.Target
}

// AnOpponentCast — an opponent cast a spell matching `spell` (nil
// for any spell).
func AnOpponentCast(spell CardPredicate) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && (spell == nil || spell(g, ev.Actor, c))
	}
}

// LandEnteredUnderYourControl — landfall.
func LandEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsLand()
}

// CreatureEnteredUnderYourControl — "whenever a creature you control
// enters", the source itself included (Impact Tremors).
func CreatureEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsCreature()
}

// AnotherCreatureEnteredUnderYourControl — the same with "another".
func AnotherCreatureEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature()
}

// ArtifactEnteredUnderYourControl — "whenever an artifact you control
// enters", the source itself included (Quicksmith Genius).
func ArtifactEnteredUnderYourControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsArtifact()
}

// ACreatureYouControlDied — a creature you controlled went to the
// graveyard, the source included (Zulaport Cutthroat).
func ACreatureYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller
}

// AnotherCreatureDied — any other creature, anyone's, died (Blood
// Artist's "whenever ~ or another creature dies" is OnAny of this and
// ThisDied — or just diedCreature with no exclusion).
func AnotherCreatureDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	_, ok := diedCreature(ev, g)
	return ok
}

// ThisDealtCombatDamageToAPlayer — the source dealt combat damage to
// a player.
func ThisDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
}

// YouGainedLife — your life total went up (an EventChangeLife with a
// positive amount). Fires once per life-gain event, as printed.
func YouGainedLife(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventChangeLife && ev.Target == source.Controller && ev.Amount > 0
}

// ThisChangedController — the source permanent changed controller
// (EventControlChanged, #930). ev.Target is the player who lost it
// and ev.Actor the player who gained it; the engine emits the event
// only on a real delta, so those two are never the same player.
func ThisChangedController(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventControlChanged && ev.CardID == source.InstanceID &&
		ev.Target != uuid.Nil && ev.Actor != uuid.Nil
}

// AnOpponentGainedControlOfAPermanentYouOwn — some permanent you OWN
// (CR 108.3 — ownership does not move) came under an opponent's
// control. The source is the watcher, not the permanent, so this
// reads the event's card rather than the source's instance ID.
func AnOpponentGainedControlOfAPermanentYouOwn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventControlChanged || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// StepBegan — the named step began (EventStepBegan, #588). With
// yours set, only on the source's controller's turn. This is the
// condition for every step the older per-step kinds do not cover:
// beginning of combat, postcombat main, end of combat, declare
// attackers / blockers.
func StepBegan(step game.Step, yours bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		if ev.Kind != game.EventStepBegan || ev.Step != step {
			return false
		}
		return !yours || ev.Actor == source.Controller
	}
}

// AllOf — every condition holds.
func AllOf(conds ...When) When {
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		for _, c := range conds {
			if !c(ev, source, lki, g) {
				return false
			}
		}
		return true
	}
}

// --- the printed shapes ---------------------------------------------

// WhenThisEnters — "When ~ enters".
func WhenThisEnters(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventETB, Self, label, effect)
}

// WhenThisDies — "When ~ dies" (graveyard only).
func WhenThisDies(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventLTB, ThisDied, label, effect)
}

// WhenThisEntersOrAttacks — "Whenever ~ enters or attacks": one
// ability, two conditions (Sun Titan).
func WhenThisEntersOrAttacks(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventETB, game.EventAttack}, Self, label, effect)
}

// WheneverThisAttacks — "Whenever ~ attacks".
func WheneverThisAttacks(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventAttack, ThisAttacked, label, effect)
}

// AtYourUpkeep — "At the beginning of your upkeep".
func AtYourUpkeep(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginUpkeep, ByYou, label, effect)
}

// AtEachUpkeep — "At the beginning of each upkeep" / "each player's
// upkeep".
func AtEachUpkeep(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginUpkeep, AnyPlayer, label, effect)
}

// AtYourEndStep — "At the beginning of your end step".
func AtYourEndStep(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginEndStep, ByYou, label, effect)
}

// AtYourPrecombatMain — "At the beginning of your precombat main
// phase" / "your first main phase".
func AtYourPrecombatMain(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventBeginPrecombatMain, ByYou, label, effect)
}

// WheneverYouCast — "Whenever you cast a <spell> spell"; nil matches
// every spell.
func WheneverYouCast(spell CardPredicate, label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCast, YouCast(spell), label, effect)
}

// WheneverYouDraw — "Whenever you draw a card" (once per card).
func WheneverYouDraw(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDrawCard, ByYou, label, effect)
}

// WheneverAnOpponentDraws — "Whenever an opponent draws a card".
func WheneverAnOpponentDraws(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDrawCard, ByAnOpponent, label, effect)
}

// Landfall — "Whenever a land you control enters" / landfall.
func Landfall(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventETB, LandEnteredUnderYourControl, label, effect)
}

// WheneverAnotherCreatureEntersUnderYourControl — "Whenever another
// creature you control enters".
func WheneverAnotherCreatureEntersUnderYourControl(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventETB, AnotherCreatureEnteredUnderYourControl, label, effect)
}

// WheneverACreatureYouControlDies — "Whenever a creature you control
// dies".
func WheneverACreatureYouControlDies(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventLTB, ACreatureYouControlDied, label, effect)
}

// WheneverThisDealsCombatDamageToAPlayer — "Whenever ~ deals combat
// damage to a player".
func WheneverThisDealsCombatDamageToAPlayer(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer, label, effect)
}

// AtYourStep — "At the beginning of your <step>" for any step. The
// upkeep, draw, precombat-main and end-step shapes above use the
// older per-step event kinds, which fire after that step's turn-based
// action; this one fires as the step begins.
func AtYourStep(step game.Step, label string, effect Effect) game.TriggeredAbility {
	return On(game.EventStepBegan, StepBegan(step, true), label, effect)
}

// AtEachStep — "At the beginning of each <step>" / "each player's".
func AtEachStep(step game.Step, label string, effect Effect) game.TriggeredAbility {
	return On(game.EventStepBegan, StepBegan(step, false), label, effect)
}

// AtBeginningOfYourCombat — "At the beginning of combat on your turn".
func AtBeginningOfYourCombat(label string, effect Effect) game.TriggeredAbility {
	return AtYourStep(game.StepBeginCombat, label, effect)
}

// AtBeginningOfEachCombat — "At the beginning of each combat".
func AtBeginningOfEachCombat(label string, effect Effect) game.TriggeredAbility {
	return AtEachStep(game.StepBeginCombat, label, effect)
}

// AtYourPostcombatMain — "At the beginning of your postcombat main
// phase" / "your second main phase".
func AtYourPostcombatMain(label string, effect Effect) game.TriggeredAbility {
	return AtYourStep(game.StepPostcombatMain, label, effect)
}

// AtEndOfYourCombat — "At end of combat on your turn" (CR 511).
func AtEndOfYourCombat(label string, effect Effect) game.TriggeredAbility {
	return AtYourStep(game.StepEndCombat, label, effect)
}

// WheneverYouGainLife — "Whenever you gain life".
func WheneverYouGainLife(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventChangeLife, YouGainedLife, label, effect)
}

// AnExhaustAbility — the activation event is of an ability that
// prints the exhaust keyword. Reads the bit the announcement stamped
// (game.Event.Exhaust) rather than going back to the source's ability
// list, because a SacrificeSelf or DiscardSelf cost has already ended
// the object by the time a watcher runs (#1184).
func AnExhaustAbility(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Exhaust
}

// theTwoActivationKinds is every event kind that says "somebody
// activated an activated ability": the CR 602 announcement, and CR
// 605's mana abilities, which never touch the stack and so have their
// own kind. A card that says "an ability" means both.
var theTwoActivationKinds = []game.EventKind{
	game.EventActivateAbility,
	game.EventManaAbilityActivated,
}

// WheneverYouActivateAnExhaustAbility — "Whenever you activate an
// exhaust ability" (Rangers' Refueler, Afterburner Expert, #1184).
//
// Watches BOTH activation kinds, because a mana ability is an
// activated ability (CR 605.1a) and Loot, the Pathfinder prints an
// exhaust one; the exhaust bit and the ability's label ride both
// events, stamped at the announcement.
//
// "You" is the source's controller — or, under InGraveyard, its owner
// (CR 108.4), which is how Afterburner Expert reads its own clause
// from the graveyard. The trigger fires for the source's OWN exhaust
// ability too: "an exhaust ability" does not say "another".
func WheneverYouActivateAnExhaustAbility(label string, effect Effect) game.TriggeredAbility {
	return OnAny(theTwoActivationKinds, AllOf(ByYou, AnExhaustAbility), label, effect)
}

// WheneverAnOpponentActivates — "Whenever an opponent activates an
// ability of <thing> [that isn't a mana ability], …" (#1210,
// ADR 0018's note of 2026-09-22): Harsh Mentor, Runic Armasaur.
//
// ONE declaration shape for the whole clause family, and the three
// arguments are the three ways printed cards differ.
//
//   - WHOSE. Always ByAnOpponent, on Event.Actor — the ACTIVATOR, not
//     the source's controller, because that is what the cards print
//     and the two differ exactly where it matters (an ability
//     activated from a permanent somebody else controls).
//   - WHICH ABILITIES. `includeMana` picks the watched KINDS:
//     EventActivateAbility alone, or both it and
//     EventManaAbilityActivated. "…if it isn't a mana ability" is
//     therefore not a predicate a card file writes and can forget —
//     it is the absence of a kind from Watches, decided at the only
//     place that knows (CR 605.1a: a mana ability never reaches the
//     stack and carries its own event kind precisely so a watcher can
//     tell the difference).
//   - OF WHAT. `of` runs against the ability's SOURCE object, looked
//     up live on the battlefield: "an ability of an artifact,
//     creature, or land on the battlefield" (Harsh Mentor), "of a
//     creature or land" (Runic Armasaur). Nil is "any source", the
//     plain "whenever an opponent activates an ability" clause. A
//     source that is no longer on the battlefield — a Lotus Petal
//     that sacrificed itself to its own cost — matches nothing, which
//     is the printed reading of "on the battlefield" and the safe
//     direction for a clause that does not print it.
//
// `build` receives the ACTIVATOR (Event.Actor) and returns the
// effect: "this creature deals 2 damage to THAT PLAYER". It does not
// target — the clause says "that player", so hexproof and "can't be
// the target of" do nothing about it and CR 608.2b re-checks nothing.
// The activator is read at RESOLUTION off the item's triggering event
// (item.Trigger, #1223), not captured when the ability triggers, so the
// row is declarative and a restored item finds the same player
// (ADR 0041 P9).
//
// Compose Optional() over the result for a "you may" (Runic
// Armasaur), exactly as with any other trigger.
func WheneverAnOpponentActivates(label string, of CardPredicate, includeMana bool, build func(activator uuid.UUID) Effect) game.TriggeredAbility {
	kinds := []game.EventKind{game.EventActivateAbility}
	if includeMana {
		kinds = theTwoActivationKinds
	}
	return game.TriggeredAbility{
		Watches: kinds,
		Key:     label,
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
				return false
			}
			return activationSourceMatches(g, ev, of)
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			return build(triggeringActor(item))(g, item)
		},
	}
}

// triggeringActor is the Actor of the event that put a triggered item
// on the stack — the player who activated, cast, drew or attacked — or
// uuid.Nil for an item that carries no triggering event.
func triggeringActor(item *game.StackItem) uuid.UUID {
	if item == nil || item.Trigger == nil {
		return uuid.Nil
	}
	return item.Trigger.Event.Actor
}

// activationSourceMatches runs `of` against the object whose ability
// was activated, as it stands on the battlefield NOW.
//
// The event carries the source's ID and not a snapshot of it, so a
// source that has left — a Lotus Petal sacrificed to its own cost, a
// creature that died in response — matches nothing. Both printed
// cards say "on the battlefield" or name permanent types, so that is
// the reading, and it is the one that cannot over-trigger.
//
// A nil predicate skips the lookup entirely: "whenever an opponent
// activates an ability" asks nothing about the source and must not
// start caring whether it is still there.
func activationSourceMatches(g *game.Game, ev game.Event, of CardPredicate) bool {
	if of == nil {
		return true
	}
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.InstanceID != ev.CardID {
			continue
		}
		return of(g, ev.Actor, c)
	}
	return false
}

// --- where the ability watches from (CR 113.6, #925) ----------------

// InGraveyard makes a trigger watch from its owner's GRAVEYARD
// instead of from the battlefield — "when you cycle this card"
// (CR 702.29c, the card is already in the graveyard when the ability
// triggers), Bloodghast's landfall, Narcomoeba's arrival.
//
// It replaces the zone list rather than adding to it, which is the
// rule and not a shortcut: an ability printed to work from the
// graveyard does not also work from play, and a permanent with a
// graveyard trigger firing on the battlefield is the bug this
// wrapper exists to make impossible to write.
//
// "You" inside the trigger is the card's OWNER while it sits there
// (CR 108.4): the harvest hands the predicate a source whose
// Controller is its Owner, so ByYou, Self and Landfall all read the
// way the card prints them.
func InGraveyard(t game.TriggeredAbility) game.TriggeredAbility {
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	return t
}

// InExile is InGraveyard for exile — suspend's "at the beginning of
// your upkeep, remove a time counter from this card" and "when the
// last is removed, cast it" (CR 702.62b/c, #659).
func InExile(t game.TriggeredAbility) game.TriggeredAbility {
	t.Zones = []game.ZoneKind{game.ZoneExile}
	return t
}

// WhenThisBecomesPlotted — "When this card becomes plotted, …"
// (CR 702.170c/d; Longhorn Sharpshooter, Aloe Alchemist; #1382).
//
// A card becomes plotted in exile, so the ability watches from exile —
// the plot special action from hand and an "it becomes plotted" effect
// (Aven Interrupter) both end in Game.PlotExiledCardForEffect, which is
// the one emitter of EventBecomesPlotted. A plain exile emits nothing,
// so an exiled-but-not-plotted card does not trigger. The trigger's
// controller is the card's owner (CR 108.4 — the harvest makes the
// exiled source's Controller its Owner), even when an opponent's Aven
// Interrupter did the plotting.
func WhenThisBecomesPlotted(label string, effect Effect) game.TriggeredAbility {
	return InExile(On(game.EventBecomesPlotted, Self, label, effect))
}

// ThisWasPutIntoYourGraveyardFromYourLibrary — "when this card is put
// into your graveyard from your library" (Narcomoeba, the dredge and
// mill recursion family).
//
// Watches both kinds the route emits: EventMill for a mill proper
// (CR 701.17a) and EventZoneMove for every other library-to-graveyard
// move, which is one ability with two conditions rather than two
// abilities — the card prints one sentence. The zone pair is read off
// the event rather than assumed, so a card milled from an OPPONENT's
// library does not trigger their copy.
func ThisWasPutIntoYourGraveyardFromYourLibrary(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.CardID == source.InstanceID &&
		ev.OldZone == game.ZoneLibrary &&
		ev.NewZone == game.ZoneGraveyard
}

// WhenThisIsPutIntoYourGraveyardFromYourLibrary is the printed shape
// over that condition, already scoped to the graveyard — the zone the
// card is in when it triggers.
func WhenThisIsPutIntoYourGraveyardFromYourLibrary(label string, effect Effect) game.TriggeredAbility {
	return InGraveyard(OnAny([]game.EventKind{game.EventMill, game.EventZoneMove},
		ThisWasPutIntoYourGraveyardFromYourLibrary, label, effect))
}

// --- control changes (#930, CR 613.1b) ------------------------------
//
// One event, emitted from the one materialise step at the end of the
// layer pass, covers every way control moves: a spell or ability
// taking a permanent, an Aura's static, an exchange (CR 701.12), and
// the permanent going home when the effect ends. So a card that
// triggers on losing control fires on the revert as well as on the
// theft, with nothing written per card.

// WhenYouLoseControlOfThis — "When you lose control of ~" (Khârn the
// Betrayer's Sigil of Corruption, Coffin Queen, Gustha's Scepter).
//
// "You" is the player who LOST control, and that is the one printed
// shape whose ability is not controlled by the source's current
// controller: by the time the event is emitted the permanent is
// already the other player's, so an item built the ordinary way would
// hand the thief the trigger. The item is built for ev.Target instead
// — the only reading under which "you lose control of ~" can be true
// of its own controller.
//
// The Build here only fills in the controller (ADR 0041 P9): it leaves
// item.Effect nil and the engine installs the declared effect.
func WhenYouLoseControlOfThis(label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventControlChanged, ThisChangedController, label, effect)
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label, nil)
		item.Controller, item.Owner = ev.Target, ev.Target
		return item
	}
	return t
}

// WhenYouGainControlOfThis — "When you gain control of ~ from another
// player" (Risky Move). The gaining player is the permanent's
// controller by the time the event lands, so the ordinary item is
// already theirs.
func WhenYouGainControlOfThis(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventControlChanged, ThisChangedController, label, effect)
}

// WheneverAnOpponentGainsControlOfAPermanentYouOwn — the Zedruu
// shape: the watcher is one permanent and the permanent that moved is
// another, matched by OWNERSHIP (CR 108.3), which no zone change or
// theft alters. Fires once per permanent, as printed; two permanents
// donated by one resolution are two triggers.
func WheneverAnOpponentGainsControlOfAPermanentYouOwn(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventControlChanged, AnOpponentGainedControlOfAPermanentYouOwn, label, effect)
}

// selfEnteredUntapped — "when this land enters untapped" (Gingerbread
// Cabin, Idyllic Grange, Mystic Sanctuary). The `When` form of
// b27SelfEnteredUntapped, which takes only the event and the source:
// the harvester runs with the source already on the battlefield and
// its enters-tapped replacement applied, so the tapped flag is the
// answer.
func selfEnteredUntapped(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return b27SelfEnteredUntapped(ev, source)
}

// SelfTargetedByASpell — "when this creature becomes the target of a
// SPELL" (Departed Deckhand, Spiketail Drakeling's cousins).
//
// Not the same condition as `Self` on EventBecomesTarget, which is
// "a spell or ability": the event is emitted for both and carries no
// discriminator, so the spell half is read off the stack the way
// Gargos reads it. Firing on abilities too would make a card with
// this printed DRAWBACK strictly worse than printed, which is the
// same #259 rule pointed the other way.
func SelfTargetedByASpell(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID != source.InstanceID {
		return false
	}
	return b40TargetedByASpell(ev, g)
}

// DealAmountFromParamsToFirstTarget is the Effect body for "…deals
// that much damage to target X", where the amount was computed once
// at trigger time and carried on item.Params.Amount by a fill-in
// Build — Kaervek the Merciless's spell mana value, All Will Be One's
// counters placed — rather than baked into a per-instance closure. A
// target that left in response (CR 608.2b) leaves the ability to do
// nothing rather than error.
func DealAmountFromParamsToFirstTarget(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: item.Params.Amount}.Apply(NewContext(g, item))
}

// PutChosenTargetOnTopOfLibrary is the Effect body behind "put target
// <card> from your graveyard on top of your library" — Mystic
// Sanctuary's instant or sorcery, Mortuary Mire's creature. It tucks
// the chosen target to the top of its owner's library through the
// shared exit primitive, so a commander card gets the CR 903.9 offer
// on the way. A target that left in response (CR 608.2b) is skipped
// rather than errored.
func PutChosenTargetOnTopOfLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.TuckToLibraryForEffect(t.ID, false)
	}
	return nil
}
