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
// Every constructor returns an ordinary game.TriggeredAbility built
// exactly as a hand-written one would be — Watches, an AppliesTo, and
// a Build that puts the item on the stack via NewTriggeredItem — so
// the engine sees no difference and the S19 rules (ADR 0018) hold
// unchanged: Build builds and never resolves; the Effect reads the
// controller, source and targets off the item it is handed and
// captures no *Card or *Game.
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
func OnAny(kinds []game.EventKind, when When, label string, effect Effect) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   kinds,
		AppliesTo: when,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label, effect)
		},
	}
}

// Optional makes a trigger a CR 603.4 "you may": the harvester asks
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
