package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi Akosa of Zhalfir — Legendary Planeswalker — Teferi, loyalty
// 4, the BACK FACE of Invasion of New Phyrexia (oracle 480ea052…,
// face 1):
//
//	"+1: Draw two cards. Then discard two cards unless you discard a
//	     creature card.
//	 −2: You get an emblem with 'Knights you control get +1/+0 and
//	     have ward {1}.'
//	 −3: Tap any number of untapped creatures you control. When you
//	     do, shuffle target nonland permanent an opponent controls
//	     with mana value X or less into its owner's library, where X
//	     is the number of creatures tapped this way."
//
// Registered under "<oracle_id>#1" — see deluge_of_the_dead.go for
// why the composite key exists.
//
// # Why this card was held out of the catalog until now
//
// S27 shipped three of four battles and named this one as the card it
// would not ship in pieces (#92's #574 status comment, #626): a 4-loyalty
// planeswalker with two of three abilities omitted is a blank, and
// ADR 0032 says leave it out rather than stub it. Each ability was
// waiting on a different seam, and all three have since landed:
//
//   - the +1 on a SET-LEVEL rule for a discard — "two cards, unless
//     one of them is a creature card" is not a count — which is
//     ChooseCardsPrompt.Validate (#624, reached here through
//     DiscardPrompt so the discard still takes the one discard path);
//   - the −2 on emblems (#623, ADR 0064);
//   - the −3 on a CR 603.12 reflexive trigger (#636), which is the
//     only thing that lets the target clause be evaluated knowing how
//     many creatures were actually tapped.
//
// # The +1
//
// "Draw two cards. Then discard two cards unless you discard a
// creature card." The discard's floor is ONE, not two, and the rule
// that decides whether one is enough is a question about the SET —
// which is what a per-pick candidate list and a count can never ask.
// It is a DiscardPrompt with Min 1, Max 2 and the Validate hook, so
// the submit path and the bot's enumerator judge the same predicate
// (ChooseCardsPickLegalLocked) and a bot is never offered a set the
// resolver will refuse.
//
// A hand smaller than two discards as many as it can (CR 701.8a), and
// the "two" the rule compares against is that smaller number, frozen
// when the prompt is queued — exactly the scalar
// ChooseCardsPrompt.Validate's doc says a card must capture rather
// than re-derive later.
//
// # The −2
//
// "Knights you control get +1/+0 and have ward {1}" is one emblem
// with two halves and ONE predicate between them, so the anthem and
// the ward can never disagree about what a Knight is. The +1/+0 is an
// ordinary layer-7c static (TribalAnthem); the ward is NOT a layer-6
// keyword grant, because CR 702.21a makes ward a triggered ability
// carrying a cost — see WardGranted in ward.go, which is the general
// "<these permanents> have ward <cost>" and of which Lavaspur Boots'
// equipped-creature ward is now the narrow case.
//
// The emblem grants ward to a Knight it does not own and never
// touches: the trigger is the EMBLEM's, harvested from the command
// zone (CR 114.3), and it fires when a Knight its controller controls
// becomes the target of an opponent's spell or ability.
//
// # The −3
//
// Two stack items, as printed. "Tap any number of untapped creatures
// you control" is a choose_cards over the controller's untapped
// creatures with a floor of ZERO — tapping nothing is a legal answer,
// and then nothing else happens, because "when you do" is conditional
// on the doing. Tapping one or more creates the reflexive trigger,
// which goes on the stack above the parent with a response window and
// picks its target THERE (CR 603.3d) — knowing X, which is the whole
// point: the clause is "mana value X or less" and X is not known
// until the taps are in.
//
// The tapped creatures ride the trigger's payload rather than a
// closure, so the Effect stays a package-level func that captures
// nothing and survives Clone / undo (StackItem.Effect's contract).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfNewPhyrexiaOracleID + "#1",
		Name:         "Teferi Akosa of Zhalfir",
		Completeness: CompletenessFull,
		// Printed loyalty reaches a card through deck import
		// (ADR 0032 §1); this is the fallback for fixtures and the
		// dev spawner.
		StartingLoyalty: 4,
		Emblem: &EmblemSpec{
			Label: "Teferi Akosa of Zhalfir emblem",
			Text:  "Knights you control get +1/+0 and have ward {1}.",
			Static: []game.StaticAbility{
				TribalAnthem(teferiAkosaKnights, 1, 0),
			},
			Triggered: []game.TriggeredAbility{
				WardGranted(WardMana("{1}"), "Teferi Akosa of Zhalfir emblem — ward {1}", teferiAkosaKnights.Matches),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Draw two cards. Then discard two cards unless you discard a creature card.",
				Cost:   LoyaltyCost(1),
				Effect: teferiAkosaDrawThenDiscard,
			},
			{
				Label: "−2: You get an emblem with \"Knights you control get +1/+0 and have ward {1}.\"",
				Cost:  LoyaltyCost(-2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−3: Tap any number of untapped creatures you control. When you do, shuffle target nonland permanent an opponent controls with mana value X or less into its owner's library, where X is the number of creatures tapped this way.",
				Cost:   LoyaltyCost(-3),
				Effect: teferiAkosaTapAnyNumber,
			},
		},
	})
}

// teferiAkosaKnights is "Knights you control", shared by the emblem's
// anthem and its ward so the layer half and the trigger half can
// never drift apart. The source is the emblem, whose controller is
// its owner and never changes (CR 114.5), so "you control" is the
// same controller test an anthem uses.
var teferiAkosaKnights = TribeFilter{Tribes: []string{"Knight"}, YoursOnly: true}

// --- +1 -----------------------------------------------------------

// teferiAkosaDrawThenDiscard is the +1: draw two, then the
// conditional discard.
//
// The draw happens first and unconditionally — the card reads "draw
// two cards. THEN discard" — so the two cards drawn are among the
// ones that can be pitched, which is most of what makes the ability
// worth a loyalty.
//
// Caller holds g.mu.
func teferiAkosaDrawThenDiscard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return teferiAkosaQueueDiscard(g, item)
}

// teferiAkosaQueueDiscard queues the "discard two cards unless you
// discard a creature card" pick.
//
// `need` is the printed two, or the whole hand when it is smaller
// (CR 701.8a discards as many as you can). It is captured HERE, when
// the prompt is queued, rather than read back later: the hand moves
// under an asynchronous prompt, and the number the rule compares
// against is the one the ability asked with.
//
// A player with an empty hand is asked nothing;
// QueueDiscardChoiceForEffect handles that itself.
//
// Caller holds g.mu.
func teferiAkosaQueueDiscard(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Hand == nil {
		return nil
	}
	need := min(2, p.Hand.Size())
	if need <= 0 {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        need,
		Min:      1,
		Question: "Teferi Akosa of Zhalfir — discard two cards, unless you discard a creature card",
		Validate: teferiAkosaDiscardPaysTheClause(need),
	})
	return nil
}

// teferiAkosaDiscardPaysTheClause is the set-level rule: a pick pays
// the clause when it is the full count, or when it is one card and
// that card is a creature card.
//
// A function of the picks and one captured int, with no *Game: it
// runs on the submit path under the write lock AND inside
// legal.EnumerateFor under the read lock, which is the contract
// ChooseCardsPrompt.Validate sets out.
//
// "A creature card" is the card's own type in the hand, so a creature
// card is one whether or not anything on the battlefield would have
// animated it.
func teferiAkosaDiscardPaysTheClause(need int) func([]game.Card) bool {
	return func(picked []game.Card) bool {
		if len(picked) == need {
			return true
		}
		return len(picked) == 1 && picked[0].IsCreature()
	}
}

// --- −3 -----------------------------------------------------------

// teferiAkosaTapAnyNumber is the −3's first sentence: offer the
// controller's untapped creatures, with a floor of zero because "any
// number" includes none.
//
// Package-level rather than a closure, for the reason every
// continuation here is: the prompt outlives this call, and a closure
// over a *Card would be a pointer into a zone slice that reallocates.
//
// A controller with no untapped creatures is asked nothing and
// nothing else happens — there is no "when you do" without a doing.
// QueueChooseCardsForEffect deliberately does not short-circuit an
// empty candidate set (its doc says why), so the skip belongs here.
//
// Caller holds g.mu.
func teferiAkosaTapAnyNumber(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	candidates := untappedCreaturesControlledBy(g, controller)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Teferi Akosa of Zhalfir — tap any number of untapped creatures you control",
		Cards:    candidates,
		Min:      0,
		// Max 0 means "all of them", which is what "any number" is.
		Max:  0,
		Zone: game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return teferiAkosaTapThenShuffle(g, item, picked)
		},
	})
	return nil
}

// teferiAkosaTapThenShuffle taps the picks and, if at least one
// actually became tapped, creates the reflexive trigger.
//
// Every pick is re-read before it is tapped. The prompt is
// asynchronous and the board moves under it: a creature can be
// removed, tapped by something else, or stolen between the question
// and the answer, and a pick that is no longer an untapped creature
// its controller controls was not "tapped this way" and must not
// count towards X.
//
// Caller holds g.mu.
func teferiAkosaTapThenShuffle(g *game.Game, item *game.StackItem, picked []uuid.UUID) error {
	ctx := NewContext(g, item)
	var tapped []uuid.UUID
	for _, id := range picked {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != item.Controller {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		tapped = append(tapped, id)
	}
	// "When you tap one or more creatures this way" — the condition
	// is the tapping, and the `if` is the card's (CR 603.12).
	if len(tapped) == 0 {
		return nil
	}
	x := len(tapped)
	return ReflexiveTrigger{
		Label:   "Teferi Akosa of Zhalfir — shuffle a nonland permanent with mana value " + strconv.Itoa(x) + " or less into its owner's library",
		Targets: teferiAkosaShuffleTarget(x),
		Cards:   tapped,
		Effect:  teferiAkosaShuffleIntoLibrary,
	}.Apply(ctx)
}

// teferiAkosaShuffleTarget is the reflexive trigger's clause, built
// with X already known — which is the reason the follow-up is a
// reflexive trigger and not the rest of this line. The legal set is
// computed when the trigger goes on the stack (CR 603.3d), and an
// empty one drops the trigger with no prompt.
func teferiAkosaShuffleTarget(x int) *game.TargetSpec {
	return TargetPermanent(
		"target nonland permanent an opponent controls with mana value "+strconv.Itoa(x)+" or less",
		Nonland(), OpponentControls(), ManaValueLE(x))
}

// teferiAkosaShuffleIntoLibrary is the reflexive half: the chosen
// permanent goes into its OWNER's library and that library is
// shuffled.
//
// Chaos Warp's shape exactly, and for its reason (#783): the shuffle
// is the tuck's CONTINUATION, not the next line. The tuck can pause —
// a commander's owner is asked about the command zone (CR 903.9) —
// and shuffling on the next line would shuffle while the permanent
// was still on the battlefield.
//
// The owner is read before the move, because afterwards the card is
// in a library and the question is about where it came from.
//
// Caller holds g.mu.
func teferiAkosaShuffleIntoLibrary(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 {
		return nil
	}
	t := item.Targets[0]
	if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	c, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	owner := c.Owner
	return g.TuckToLibraryThenForEffect(t.ID, game.TuckOptions{}, func(g *game.Game, _ bool) error {
		return g.ShuffleLibraryForEffect(owner)
	})
}

// untappedCreaturesControlledBy returns the untapped creatures a
// player controls, in battlefield order. Read through IsCreature and
// the live Tapped flag, so an animated land counts and a creature
// something already tapped does not.
//
// Caller holds g.mu.
func untappedCreaturesControlledBy(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == playerID && c.IsCreature() && !c.Tapped {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
