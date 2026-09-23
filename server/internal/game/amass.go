package game

import (
	"strconv"

	"github.com/google/uuid"
)

// amass.go — the CR 701 keyword action Amass [subtype] N (#1236).
//
//	701.47a. To amass [subtype] N means "If you don't control an Army
//	creature, create a 0/0 black [subtype] Army creature token. Choose
//	an Army creature you control. Put N +1/+1 counters on that
//	creature. If it isn't a [subtype], it becomes a [subtype] in
//	addition to its other types."
//
// Seventy-six printed cards amass, from Orcish Bowmasters (EDHREC 257)
// down. Every one of them prints the same four sentences and varies
// only two things: the SUBTYPE and the COUNT. So amass is ONE verb
// here and each card is one clause, in the shape ADR 0013 §5s gave
// proliferate, scry and surveil and ADR 0081 gave earthbend: a
// `KeywordAction` constant, a count on the CR 614 event, and one arm
// of `applyResolvedKeywordActionLocked`.
//
// # The rules number
//
// Amass is **CR 701.47** in the pinned edition (MagicCompRules
// 20260819). #1236 and the roadmap's detector both say "CR 701.35",
// which is the number amass had before Alchemy's keyword actions were
// inserted ahead of it; 701.51 is Open an Attraction. Corrected here
// for the same reason #1260 corrected storm's.
//
// # Why it is a keyword action and not a card primitive
//
// The count is printed and the subtype is printed, and nothing else
// about the verb moves. "Amass Orcs 1", "amass Zombies X where X is
// the number of cards in your hand", "amass Goblins 2" — the number is
// exactly the thing a CR 614 window is for. Opening one costs nothing
// today (nothing printed replaces an amass) and is the difference
// between a seam and a rewrite the day one does: a "whenever you would
// amass, amass twice that much instead" becomes a card-side
// `ReplacementEffect` narrowing on the action, with no engine change.
//
// It is also what #1236's third bullet asked for. A future "double the
// counters from amass" has to see the placement AS AN AMASS rather
// than as a bare counter add, and `RepEventKeywordAction` carrying
// `KeywordActionAmass` is where it sees it.
//
// Amass is the second counted keyword action whose count is a number
// of COUNTERS (earthbend was the first), and the second that DOES
// SOMETHING AT ZERO — see `KeywordAction.actsAtZeroCount` and the
// "amass 0" note on `applyAmassLocked`.
//
// # The four parts, and where each one already lived
//
//  1. FIND-OR-CREATE is the part with no home. Every token-creation
//     primitive always creates; none of them skips because a matching
//     permanent is already out. `armiesControlledLocked` is the live
//     board query and `CreateTokensThenForEffect` is the creation, so
//     the new thing is only the branch between them — and it goes
//     through the ordinary token pipeline, so ADR 0061's CR 701.7b
//     window still opens and Doubling Season still doubles.
//  2. THE CHOICE is `QueueChooseCardsForEffect` over the battlefield
//     with Min 1 / Max 1 — the same door `effects.SacrificeChoice`
//     already uses for a resolution-time pick among your own
//     permanents. It is only asked when the controller has MORE THAN
//     ONE Army; one Army is not a choice, and the prompt queue is not
//     the place to make the player click through a forced answer 76
//     cards' worth of times.
//  3. THE COUNTERS go through `AddCounterByForEffect`, so the CR 614
//     placement window opens and Hardened Scales, Doubling Season and
//     Vorinclex all apply. The CR 614 ENTRY pipeline is not involved
//     and must not be: the Army is already on the battlefield, so
//     these are placed counters, not entry counters (#1120's
//     distinction) — including on the token amass just made, which
//     enters 0/0 and is counted up afterwards, exactly as printed.
//  4. THE SUBTYPE is a layer-4 continuous effect with no stated
//     duration, pinned to the object: `grantAmassSubtypeLocked`, which
//     is `animateEarthbentLandLocked`'s layer-4 arm one subtype over.
//
// # "The Army you amassed"
//
// CR 701.47c gives the chosen creature a name the rest of the sentence
// can use — "the Army you amassed", "the amassed Army" — and five
// printed cards use it: Widespread Brutality, Grishnákh, Brash
// Instigator, Foray of Orcs, Surrounded by Orcs and Goblin Plate Mail.
// So the continuation this verb hands back is not `func(*Game) error` like
// scry's: it is `func(*Game, uuid.UUID) error`, and the UUID is the
// Army. It is `uuid.Nil` when there was none to choose and none to
// make — CR 701.47b still says the player amassed, so the rest of the
// sentence runs either way, with nothing to point at.

// amassLabel is the attribution the subtype grant carries. One string
// so a stall dump, the layer census and a test failure all name the
// same verb.
const amassLabel = "amass"

// ArmySubtype is the creature type every Army token and every amass
// target shares (CR 701.47a). A constant because it is a RULES word,
// not a label: four printed cards make an "an Army" board query of
// their own (Sauron, the Dark Lord; Great Goblin, Foul-Hearted; March
// from the Black Gate; and Widespread Brutality's "each non-Army
// creature"), and a typo in any of them would silently read as "no
// Armies" rather than failing.
const ArmySubtype = "Army"

// IsArmy reports whether a card is an Army CREATURE — CR 701.47a's
// "an Army creature you control", and the read side #1236 asked for.
//
// Both halves are load-bearing and both read through the effective
// characteristic. An Army that has stopped being a creature (a
// Song of the Dryads on it, a Humility) is not amassable, and a
// creature that BECAME an Army in layer 4 is. `HasSubtype` also
// answers for changeling, which is right: a Mistform Ultimus really
// is an Army and really can be amassed onto.
func IsArmy(c Card) bool { return c.IsCreature() && c.HasSubtype(ArmySubtype) }

// ArmiesControlledForEffect lists every Army creature `controller`
// controls, in battlefield order — CR 701.47a's board query, exported
// because it is also the read side of "an Army you control" for the
// cards that only LOOK.
//
// Caller must hold g.mu (it is an effect-time read).
func (g *Game) ArmiesControlledForEffect(controller uuid.UUID) []uuid.UUID {
	return g.armiesControlledLocked(controller)
}

// armiesControlledLocked is the walk above. Battlefield order, so the
// same board always offers the same prompt in the same order.
func (g *Game) armiesControlledLocked(controller uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && IsArmy(*c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// AmassForEffect takes the amass keyword action on behalf of `actor`:
// it opens the CR 614 window on the action and, once the window
// settles, finds or creates an Army, has `actor` choose one if there
// is more than one, puts the settled count of +1/+1 counters on it and
// gives it the subtype.
//
// `source` is the card whose effect is amassing — the resolving spell,
// the trigger's permanent, the activated ability's source. It is the
// attribution on the subtype grant and the prompt's Source.
//
// `token` is the template the find-or-create branch mints, and
// `subtype` is the creature type the chosen Army ends up with. They
// are two arguments rather than one because `internal/game` does not
// build token templates — the catalog does (`effects.ArmyToken`), and
// this end only needs to know the type name for CR 701.47a's last
// sentence.
//
// `then` is CR 701.47c's "the Army you amassed", run once the action
// is over. It may be nil.
//
// It can PAUSE — twice. A CR 614 window with two count replacements in
// it queues the CR 616 ordering prompt and returns with NOTHING done;
// and a controller with two or more Armies is asked which, and returns
// with the counters not yet placed. Both resume through the same
// `finishAmassLocked` the unpaused path runs.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) AmassForEffect(actor, source uuid.UUID, token Card, subtype string, n int, then func(g *Game, army uuid.UUID) error) error {
	if subtype == "" {
		// CR 701.47d: the War of the Spark cards printed "amass N"
		// with no subtype and were errata'd to "amass Zombies N". A
		// card file that says nothing is saying that.
		subtype = "Zombie"
	}
	_, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              actor,
		Source:             source,
		KeywordAction:      KeywordActionAmass,
		KeywordActionCount: n,
		keywordAction: &keywordActionTail{
			armyToken:   token,
			armySubtype: subtype,
			amassed:     then,
		},
	})
	return err
}

// applyAmassLocked is the settled action's first half: find or create
// the Army, then choose one.
//
// # Amass 0
//
// Reachable with n <= 0, and it is not a no-op:
// `KeywordAction.actsAtZeroCount` is true for amass because the count
// is a number of COUNTERS and the other three sentences of the keyword
// do not depend on it. "If you're instructed to amass 0, you'll create
// an Army token if you don't control one, but you won't put any
// counters on it" is the War of the Spark release note, and a printed
// "amass Orcs X" with X = 0 (Summons of Saruman off an empty
// graveyard, Shagrat with no Equipment) is a real board state.
//
// Caller must hold g.mu.
func (g *Game) applyAmassLocked(ev *ReplacementEvent) error {
	tail := ev.keywordAction
	if tail == nil {
		// applyResolvedKeywordActionLocked tolerates a tail-less event
		// and every arm has to as well: an amass with no template
		// makes nothing and amasses onto whatever is already out,
		// which is the honest answer to an event nobody filled in.
		tail = &keywordActionTail{}
		ev.keywordAction = tail
	}
	armies := g.armiesControlledLocked(ev.Actor)
	if len(armies) > 0 {
		return g.chooseAmassArmyLocked(ev, armies)
	}
	// CR 701.47a's first sentence. Through the ordinary token
	// pipeline, so ADR 0061's CR 701.7b window opens and a Doubling
	// Season really does make two Armies — at which point the choice
	// below is a real one, which is the ruling.
	//
	// CreateTokensThenForEffect and not CreateTokensForEffect: the
	// creation can pause on a CR 616 ordering prompt, and the returned
	// slice is empty when it does. The rest of the amass is the
	// continuation, not the next line.
	return g.CreateTokensThenForEffect(TokenCreation{
		Controller: ev.Actor,
		Groups:     []TokenGroup{{Template: tail.armyToken, Count: 1}},
	}, func(g *Game, _ []uuid.UUID) error {
		// Re-read the board rather than trusting the created IDs: a
		// creation replaced away entirely leaves none, a doubled one
		// leaves two, and "choose an Army creature you control" is a
		// question about the board either way.
		return g.chooseAmassArmyLocked(ev, g.armiesControlledLocked(ev.Actor))
	})
}

// chooseAmassArmyLocked is CR 701.47a's "choose an Army creature you
// control": nothing to choose, one forced answer, or a prompt.
//
// The prompt is `PendingChoiceChooseCards` over the battlefield with
// Min 1 / Max 1 — `effects.SacrificeChoice`'s shape, and the only
// prompt kind in the engine that carries a CONTINUATION, which this
// needs because the counters and the subtype come after the answer.
// The legal enumerator already answers the kind (legal/choices.go),
// so the bot needs nothing new.
//
// Caller must hold g.mu.
func (g *Game) chooseAmassArmyLocked(ev *ReplacementEvent, armies []uuid.UUID) error {
	switch len(armies) {
	case 0:
		// Nothing to amass onto and nothing was made — a token
		// creation replaced to nothing. CR 701.47b: the player
		// amassed anyway, so the rest of the sentence still runs,
		// with no Army to point at.
		return g.runAmassThenLocked(ev.keywordAction, uuid.Nil)
	case 1:
		return g.finishAmassLocked(ev, armies[0])
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:    ev.Actor,
		FromPlayer: ev.Actor,
		Source:     ev.Source,
		Question:   amassQuestion(ev.KeywordActionCount, ev.keywordAction.armySubtype),
		Cards:      armies,
		Min:        1,
		Max:        1,
		// Re-checked against the live battlefield on submit: the
		// prompt is asynchronous and an Army can die between the
		// question and the answer.
		Zone: ZoneBattlefield,
		Then: func(g *Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return g.runAmassThenLocked(ev.keywordAction, uuid.Nil)
			}
			return g.finishAmassLocked(ev, picked[0])
		},
	})
	return nil
}

// amassQuestion is the prompt's header, written the way the card is.
func amassQuestion(n int, subtype string) string {
	return "Amass " + subtype + "s " + strconv.Itoa(n) + " — choose an Army you control"
}

// finishAmassLocked is the settled action's second half: the subtype,
// then the counters, then the rest of the sentence.
//
// # Why the subtype goes before the counters, and CR 701.47a says after
//
// `AddCounterByForEffect` can PAUSE — a window with two counter
// replacements in it queues the CR 616 ordering prompt and returns nil
// with the placement owed to the resume — so anything the amass still
// owes has to be registered before then, or a paused Doubling Season
// prompt would leave an Army that never got its creature type. It is
// the ordering argument `applyEarthbendLocked` makes, and the swap is
// unobservable: counters are not subtypes, so "if it isn't a
// [subtype]" reads the same before and after, and no player receives
// priority in the middle of one resolution for a state-based action to
// notice (CR 704.3).
//
// Caller must hold g.mu.
func (g *Game) finishAmassLocked(ev *ReplacementEvent, army uuid.UUID) error {
	tail := ev.keywordAction
	g.grantAmassSubtypeLocked(ev.Source, army, tail.armySubtype)
	if n := ev.KeywordActionCount; n > 0 {
		if err := g.AddCounterByForEffect(ev.Actor, army, CounterPlusOne, n); err != nil {
			return err
		}
	}
	return g.runAmassThenLocked(tail, army)
}

// grantAmassSubtypeLocked is CR 701.47a's last sentence: "If it isn't
// a [subtype], it becomes a [subtype] in addition to its other types."
//
// Layer 4 (CR 613.1d), ADDING the subtype — Army is not touched, and
// neither is anything else the object already is, so an Orc-amassed
// Zombie Army is an Orc Zombie Army and both lords see it.
//
// # The guard is the rule AND the cost control
//
// "If it isn't" is a real conditional, and skipping the registration
// when the Army already has the type is what keeps the registry from
// growing without bound: Orcish Bowmasters amasses on EVERY opponent
// draw, and a game of four wheels would otherwise leave a hundred
// identical layer-4 effects in the snapshot census for the rest of the
// game. The read goes through the effective characteristic, so a grant
// that already applied answers the guard on the next amass.
//
// # The duration is indefinite, pinned to the object
//
// The sentence states no duration, so CR 611.2a says it lasts until
// the game ends — but CR 400.7 says an Army that dies and comes back
// is a NEW OBJECT and the effect named the old one. That is
// `IndefiniteDuration()` through `Game.PinnedTo`: the affected set is
// keyed on {instance, battlefield-entry stamp}, so the effect stops
// applying the instant the Army leaves and `durationExpiredLocked`
// drops the registry entry at the next sweep. Earthbend's animation
// takes the same shape for the same two rules.
//
// Caller must hold g.mu.
func (g *Game) grantAmassSubtypeLocked(source, army uuid.UUID, subtype string) {
	if subtype == "" {
		return
	}
	// The guard reads the EFFECTIVE subtype, so the layer cache has to
	// be current or a second Orc amass onto the same Army would not see
	// the first one's grant — and Orcish Bowmasters would register one
	// per opponent draw. A registration bumps layerVersion, so this is
	// a no-op on every amass that changed nothing.
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.battlefieldCardLocked(army)
	if !ok || c.HasSubtype(subtype) {
		return
	}
	stamp := c.EnteredBattlefieldAt
	g.registerScopedStaticLocked(StaticAbility{
		Layer: Layer4Type,
		AppliesTo: func(t *Card, _ *Game, _ *Card) bool {
			return t.InstanceID == army && t.EnteredBattlefieldAt == stamp
		},
		Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
			if !typeListHas(ch.Subtypes, subtype) {
				ch.Subtypes = append(ch.Subtypes, subtype)
			}
		},
	}, source, amassLabel+" — it's also a "+subtype, g.PinnedTo(IndefiniteDuration(), army), timeNowUnixNano())
}

// runAmassThenLocked runs CR 701.47c's "the Army you amassed" clause,
// exactly once.
//
// Cleared THROUGH the tail pointer for the reason
// `runKeywordActionThenLocked` clears `then` through it: a
// continuation that re-enters the pipeline on the same tail value must
// not run itself twice, and an undone-then-redone prompt answer must
// not replay it. It is the same door the abandoned path takes
// (`abandonKeywordActionLocked`), so a settled amass and one replaced
// away cannot drift apart about whether the rest of the card happened.
//
// Caller must hold g.mu.
func (g *Game) runAmassThenLocked(tail *keywordActionTail, army uuid.UUID) error {
	if tail == nil || tail.amassed == nil {
		return nil
	}
	fn := tail.amassed
	tail.amassed = nil
	return fn(g, army)
}
