package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delayed_bodies.go is the catalog's half of ADR 0041 phase 3, tier 2
// (#1497): every delayed trigger a card schedules names what it does by
// KEY, registered here, so a table with a delayed trigger waiting is a
// restore point.
//
// One file rather than a line beside each function, on purpose: the
// keys are on-disk identities — every one is in the append-only ledger,
// server/internal/game/testdata/effect_keys.txt — and a reviewer adding
// or renaming one should see all of them at once.
//
// THE RULES for a key (game/effect_bodies.go has the rest):
//
//   - "<area>/<name>", lowercase and hyphenated. The area is the
//     mechanic or the card, never a batch number.
//   - Never renamed, never deleted. A rename keeps the old key alive
//     with game.EffectAlias(old, new).
//   - A body reads its data off the item (Targets, Controller) and off
//     its game.EffectParams. It captures nothing, and now it cannot.
//
// A new delayed trigger adds a line here and to the ledger
// (`go test ./internal/game -run TestEveryPersistedEffectKeyResolves
// -args -update-effect-keys`); ScheduleDelayedTrigger takes the
// game.BodyRef, so a func literal there does not compile.

var (
	// Return the listed cards from exile to the battlefield under their
	// owners' control (flicker's "at the beginning of the next end
	// step" half; Cosmic Intervention).
	returnExiledToOwnersBody = game.SimpleDelayedBody("flicker/return-exiled-to-owners", returnExiledCardsToOwners)

	// The same return, with "each of them enters with an additional
	// +1/+1 counter on it if it's a creature and an additional loyalty
	// counter on it if it's a planeswalker" (Semester's End, #4193).
	returnExiledWithAdditionalCountersBody = game.SimpleDelayedBody(
		"flicker/return-exiled-with-additional-counters", returnExiledCardsWithAdditionalCounters)

	// The same return, with "if a land enters this way, it enters
	// tapped" (Silver Surfer, Cosmic Voyager).
	returnExiledLandsTappedBody = game.SimpleDelayedBody(
		"flicker/return-exiled-lands-tapped", returnExiledCardsLandsTapped)

	// Exile the listed cards (unearth, Whip of Erebos, Saheeli Rai and
	// friends: "exile it at the beginning of the next end step").
	exileListedCardsBody = game.SimpleDelayedBody("exile/listed-cards", b06ExileListedCards)

	// Sacrifice the listed permanents (Lagomos, Reflection of
	// Kiki-Jiki's token).
	sacrificeListedCardsBody = game.SimpleDelayedBody("sacrifice/listed-cards", b33SacrificeListedCards)

	// "The goad ends" — a no-op sweep since #1598 (the engine ends each
	// goad itself), kept registered and scheduled for restore points
	// read by an older binary. See b33ClearListedGoads.
	clearListedGoadsBody = game.SimpleDelayedBody("goad/clear-listed", b33ClearListedGoads)

	// Return the listed cards from a graveyard to their owners' hands
	// (Liesa, The Locust God).
	returnListedGraveyardToHandBody = game.SimpleDelayedBody("graveyard/return-listed-to-hand",
		b15ReturnListedCardsFromGraveyardToHand)

	// Return the listed permanents to their owners' hands (Sakashima).
	returnListedPermanentsToHandBody = game.SimpleDelayedBody("sakashima/return-listed-to-hand",
		sakashimaReturnListedPermanentsToHand)

	// The trigger's controller loses the game (Final Fortune, Last
	// Chance, Warrior's Oath: "at the beginning of that turn's end
	// step, you lose the game"). Scheduled bound to the extra turn
	// (DelayedTrigger.OnExtraTurn, ADR 0059 Decision 8).
	loseTheGameAtThatTurnsEndBody = game.SimpleDelayedBody("extra-turn/lose-the-game", loseTheGameAtThatTurn)

	// Untap all creatures that attacked this turn (Full Throttle's
	// "at the beginning of each combat this turn", ADR 0059 sub-PR 2b).
	// Scheduled with ScheduleDelayedTrigger.EachThisTurn, so it fires
	// at every beginning of combat until cleanup.
	untapCreaturesThatAttackedBody = game.SimpleDelayedBody("extra-combat/untap-creatures-that-attacked", untapCreaturesThatAttackedThisTurn)

	// World at War: "at the beginning of that combat, untap all
	// creatures that attacked this turn". The same function as Full
	// Throttle's, under its own key: it is scheduled bound to the added
	// combat (DelayedTrigger.OnPhaseID, ADR 0188 Decision 5), and a new
	// key makes an older binary refuse a file that holds the binding
	// rather than fire it at the turn's regular combat.
	untapAttackersAtThatCombatBody = game.SimpleDelayedBody("extra-combat/untap-attackers-at-that-combat", untapCreaturesThatAttackedThisTurn)

	// Moraug, Fury of Akoum: "at the beginning of that combat, untap all
	// creatures you control" (ADR 0188 Decision 5).
	untapAllYouControlAtThatCombatBody = game.SimpleDelayedBody("extra-combat/untap-all-you-control", untapAllCreaturesYouControlAtThatCombat)

	// Draw a card (Urza's Bauble, Portent).
	drawOneBody     = game.SimpleDelayedBody("draw/one-card", b27DrawOne)
	portentDrawBody = game.SimpleDelayedBody("portent/draw", portentDraw)

	necropotenceDeliverBody = game.SimpleDelayedBody("necropotence/deliver-exiled", necropotenceDeliverExiled)
	pheliaReturnBody        = game.SimpleDelayedBody("phelia/return", pheliaReturn)
	teferiUntapLandsBody    = game.SimpleDelayedBody("teferi-hero/untap-two-lands", teferiHeroUntapTwoLands)

	// Zara, Renegade Recruiter: return the creature <Object> to its
	// owner's hand at the next end step, if it is still that object.
	zaraReturnToHandBody = game.DelayedBody("zara/return-to-hand", zaraReturnToHand)

	// Gift of Immortality: return the Aura the item names and attach it
	// to the creature <Object>, if that is still the same object.
	giftOfImmortalityReturnAuraBody = game.DelayedBody("gift-of-immortality/return-aura", giftOfImmortalityReturnAura)

	// Copy the spell the event-conditioned trigger fired on (Doublecast,
	// Galvanic Iteration).
	copyTheSpellBody = game.SimpleDelayedBody("copy/the-spell-you-just-cast", copyTheSpellYouJustCast)

	// Copy the ability the event-conditioned trigger fired on, with the
	// new-targets offer: the triggering EventActivateAbility names its
	// stack item (Jaya's Phoenix's "copy the next loyalty ability you
	// activate this turn", Leori, Sparktouched Hunter's "copy that
	// ability"). An ability gone from the stack is not copied.
	copyTheAbilityBody = game.SimpleDelayedBody("copy/the-ability-you-just-activated", copyTheTriggeringAbility)

	// "When you next activate a loyalty ability this turn": an
	// EventActivateAbility with the Loyalty bit (CR 606.2) by the
	// trigger's controller (Jaya's Phoenix).
	youNextActivateALoyaltyAbilityCondition = game.DelayedCondition("activate/you-next-activate-a-loyalty-ability",
		func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventActivateAbility && ev.Loyalty && ev.Actor == dt.Controller
		})

	// The three bodies that used to be FACTORIES — a closure built
	// around a captured argument. The argument is params now.

	// Pact of Negation: "pay <Cost> or lose the game", named <Name>.
	pactPaymentBody = game.DelayedBody("pact/pay-or-lose", pactPayment)

	// Arcane Denial: <Player> draws two, you draw one.
	arcaneDenialDrawsBody = game.DelayedBody("arcane-denial/draws", arcaneDenialDraws)

	// Mana Drain: add {C} × <Amount>.
	manaDrainRefundBody = game.DelayedBody("mana-drain/refund",
		func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
			if p.Amount <= 0 {
				return nil
			}
			// The amount is read off disk after a restore; a mana value
			// no card has is a corrupt file, not a refund (#1568 review).
			if p.Amount > maxManaDrainRefund {
				return fmt.Errorf("mana-drain/refund: implausible amount %d", p.Amount)
			}
			return AddMana{Amount: ManaN(p.Amount, "C")}.Apply(NewContext(g, item))
		})

	// Pygmy Hippo: add {C} × <Amount>, the mana the defending player lost
	// (#3952). No plausibility bound like Mana Drain's: a pool may hold
	// any amount, and AddMana applies the engine's MaxProducedMana ceiling.
	pygmyHippoManaBody = game.DelayedBody("pygmy-hippo/add-colorless",
		func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
			if p.Amount <= 0 {
				return nil
			}
			return AddMana{Amount: ManaN(p.Amount, "C")}.Apply(NewContext(g, item))
		})

	// The event condition of "when you NEXT CAST a <filter> spell this
	// turn": an EventCast by the trigger's controller whose spell
	// passes CondParams.Filter.
	youNextCastCondition = game.DelayedCondition("cast/you-next-cast", youNextCast)

	// The same, narrowed to a cast FROM YOUR HAND — "when you next cast
	// an instant or sorcery spell from your hand this turn" (Narset
	// Transcendent's −2, ADR 0109 §5): an EventCast whose OldZone is the
	// caster's hand.
	youNextCastFromHandCondition = game.DelayedCondition("cast/you-next-cast-from-hand", youNextCastFromHand)

	// "It gains rebound": the spell the event-conditioned trigger fired
	// on gains rebound while it is still that spell (Narset
	// Transcendent's −2, ADR 0109 §5).
	reboundTheSpellBody = game.SimpleDelayedBody("rebound/the-spell-you-just-cast", reboundTheSpellYouJustCast)

	// The event condition of "until an opponent becomes the monarch"
	// (Palace Jailer, #1722): an EventMonarchChanged whose new monarch
	// is not the controller of the ability that exiled the card. Read
	// by a CR 610.3 "until" return since #1729 (exile_until.go), and by
	// a delayed trigger a restore point written before it may hold.
	anOpponentBecameTheMonarchCondition = game.DelayedCondition("monarch/an-opponent-became",
		anOpponentBecameTheMonarchSince)

	// The event condition of "when <the creature> dies this turn"
	// (Whippoorwill, ADR 0108 §2): an EventLTB of the listed object into
	// a graveyard. The trigger's duration is pinned to the object, so a
	// flicker ends it (CR 400.7).
	theListedObjectDiedCondition = game.DelayedCondition("dies/the-listed-object", theListedObjectDied)

	// "Exile the creature" after it died (Whippoorwill): exile each
	// listed card that is still in a graveyard (CR 603.7c).
	exileListedFromGraveyardBody = game.SimpleDelayedBody("dies/exile-from-graveyard", exileListedCardsFromGraveyard)

	// The event condition of "when that creature leaves the battlefield
	// this turn" (Runesword, #1946): an EventLTB of the listed object, to
	// any zone. The trigger's duration is pinned to the object.
	theListedObjectLeftCondition = game.DelayedCondition("ltb/the-listed-object", theListedObjectLeft)

	// "Flip this creature" after the creature it damaged died (Initiate of
	// Blood, #3216): the scheduling source flips if it is still that
	// object.
	flipTheSourceBody = game.SimpleDelayedBody("flip/the-source", flipTheSource)

	// "Manifest dread" after the creature died (Turn Inside Out, #2570):
	// the trigger's controller manifests dread.
	manifestDreadAfterItDiedBody = game.SimpleDelayedBody("dies/manifest-dread", manifestDreadBody)

	// "That spell gains sunburst" for the spell an event-conditioned
	// trigger fired on (Solar Array, ADR 0109 §11).
	thatSpellGainsSunburstBody = game.SimpleDelayedBody("sunburst/that-spell-gains", thatSpellYouJustCastGainsSunburst)

	// Berserk: destroy the creature <Object> at the next end step if it
	// is still that object and it attacked this turn.
	berserkDestroyIfAttackedBody = game.DelayedBody("berserk/destroy-if-attacked", berserkDestroyIfAttacked)

	// Destroy the permanent <Object> names, if it is still that object
	// on the battlefield: "Destroy it at the beginning of the next end
	// step" (Glyph of Destruction, #2027).
	destroyTheObjectBody = game.DelayedBody("destroy/the-object", destroyTheObject)

	// Return the card <Object> names from a graveyard to the battlefield
	// under the delayed trigger's controller, if it is still that object
	// in a graveyard: "return that card to the battlefield under your
	// control at the beginning of the next end step" (Avacyn, Angel of
	// Horror).
	returnTheObjectFromGraveyardUnderYourControlBody = game.DelayedBody("graveyard/return-the-object-under-your-control",
		returnTheObjectFromGraveyardUnderYourControl)

	// Exile the listed cards unless the delayed trigger's controller has
	// the city's blessing as it resolves: "At the beginning of the next
	// end step, exile those tokens unless you have the city's blessing"
	// (Tilonalli's Summoner, #2727).
	exileListedUnlessCitysBlessingBody = game.SimpleDelayedBody("ascend/exile-listed-unless-citys-blessing",
		exileListedUnlessCitysBlessing)

	// Return the card <Object> names from a graveyard to the battlefield
	// if the source is still on the battlefield as the object <Amount>
	// names, and its owner says yes (Shirei, Shizo's Caretaker).
	shireiReturnBody = game.DelayedBody("shirei/return-if-still-on-the-battlefield", shireiReturn)

	// "You may choose to have it deal damage equal to its power to a
	// target creature. If you do, it assigns no combat damage this turn"
	// (Gaze of Pain, #3531): "it" is the instance's payload, and the
	// target clause is chosen as each instance goes on the stack (ADR
	// 0187 amendment 2026-10-16).
	gazeOfPainBody = game.TargetedDelayedBody("gaze-of-pain/damage-equal-to-its-power", simpleBody(gazeOfPain),
		constTargets(func() *game.TargetSpec { return TargetCreature("target creature") }))

	// "You may gain life equal to its power. If you do, it assigns no
	// combat damage this turn" (Delif's Cone, #3531).
	delifsConeBody = game.SimpleDelayedBody("delifs-cone/gain-life-equal-to-its-power", delifsCone)

	// "It assigns no combat damage this turn and you put a cube counter
	// on this artifact" (Delif's Cube, #3531).
	delifsCubeBody = game.SimpleDelayedBody("delifs-cube/no-combat-damage-and-a-cube-counter", delifsCube)

	// "Whenever a creature is dealt damage" (Shriveling Rot, #3997): the
	// creature a damage event reached, while it is a creature on the
	// battlefield. With OncePerBatch it is one occurrence for each
	// creature in each damage instance (CR 603.2c, 510.2).
	eachCreatureDealtDamageCondition = game.DelayedEachCondition("damage/each-creature-dealt-damage", eachCreatureDealtDamage)

	// "Whenever a creature dies" (Shriveling Rot, #3997): the creature an
	// EventLTB to a graveyard names, if it was a creature as it last
	// existed (CR 700.4, 603.10a).
	eachCreatureDiedCondition = game.DelayedEachCondition("dies/each-creature", eachCreatureDied)

	// "Destroy it" (Shriveling Rot, #3997): the payload's permanent, if
	// it is still that object on the battlefield (CR 603.7c).
	destroyThePayloadPermanentBody = game.SimpleDelayedBody("destroy/the-payload-permanent", destroyThePayloadPermanent)

	// "That creature's controller loses life equal to its toughness"
	// (Shriveling Rot, #3997): both read off the payload's object as it
	// last existed on the battlefield (CR 608.2h).
	diedControllerLosesLifeEqualToToughnessBody = game.SimpleDelayedBody("dies/controller-loses-life-equal-to-toughness",
		diedControllerLosesLifeEqualToToughness)

	// Return every object <Objects> names that is still that object in
	// a graveyard, exile or the command zone, together, under the
	// delayed trigger's controller: "return that card and the sacrificed
	// card to the battlefield under your control at the beginning of
	// your next upkeep" (Rescue from the Underworld, #2099).
	returnTheObjectsTogetherUnderYourControlBody = game.DelayedBody("return/the-objects-together-under-your-control",
		returnTheObjectsTogetherUnderYourControl)

	// "You may cast that card without paying its mana cost. If that
	// spell would be put into your graveyard, exile it instead." (Power
	// Pack, #1815): the payload card, if it is still in exile.
	powerPackCastExiledBody = game.SimpleDelayedBody("power-pack/cast-exiled-free", powerPackCastExiled)

	// "Whenever a creature you control enters this turn" (First Day of
	// Class, #3525): the creature an EventETB names, if the delayed
	// trigger's controller controls it.
	eachCreatureYouControlEntersCondition = game.DelayedEachCondition("etb/each-creature-you-control", eachCreatureYouControlEntered)

	// "Put a +1/+1 counter on it and it gains haste until end of turn"
	// (First Day of Class, #3525): the payload's permanent, if it is
	// still that object on the battlefield (CR 603.7c).
	counterAndHasteForThePayloadBody = game.SimpleDelayedBody("counter/plus-one-and-haste-for-the-payload", counterAndHasteForThePayload)
)

// maxManaDrainRefund bounds the refund a restored Mana Drain may add. No
// printed card has a mana value anywhere near it.
const maxManaDrainRefund = 1000
