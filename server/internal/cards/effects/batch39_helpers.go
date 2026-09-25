package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch39_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 39 (#402, `edhrec_rank` 4053–4153). Own file per
// the #231 convention; every package-level name carries the b39
// prefix because other batches land beside this one.
//
// What is NOT here, because the package already had it: "this
// permanent enters" is WhenThisEnters, "this creature dies" is
// WhenThisDies / cardDied, "whenever this attacks" is
// WheneverThisAttacks, "equipped creature deals combat damage to a
// player" is attachedCreatureDealtCombatDamageToPlayer, "a creature
// you control dealt combat damage to a player" is
// combatDamageToPlayerBy, "an opponent draws" is
// WheneverAnOpponentDraws, "another creature you control dies" is
// WheneverACreatureYouControlDies, the counters-you-placed read is
// b12CountersPlacedByYou, the last-known power read is
// b13LastKnownPower, the "enters with N counters counted off the
// board" replacement is b19EntersWithCountersCounted, the tapped
// entry is SelfEntersTapped, devotion is devotionTo, the tribal
// anthem is TribalAnthem, "can't be blocked this turn" is
// RestrictUntilEOT with game.CantBeBlocked, "N mana of any one
// colour" is ProducedOneColor, the life-total assignment is
// b31LifeBecomes, and the equip / enchant / attach vocabulary is
// attachments.go.

// --- search filters ------------------------------------------------

// b39IsPlainsCard is Kor Cartographer's search filter: a card with
// the Plains SUBTYPE on its printed type line. Deliberately not "a
// basic Plains" — the printed text says "a Plains card", so a
// shockland, a Snow-Covered Plains and a typed tapland all qualify.
func b39IsPlainsCard(c game.Card) bool { return c.HasSubtype("Plains") }

// b39IsVampireOrWizardCreatureCard is Bloodline Necromancer's
// reanimation filter, read off a card in a graveyard: a creature card
// that is a Vampire or a Wizard. Printed subtypes, because a card in
// a graveyard has no layer-applied characteristics to read.
func b39IsVampireOrWizardCreatureCard() CardPredicate {
	return And(Creature(), Or(HasSubtype("Vampire"), HasSubtype("Wizard")))
}

// --- trigger conditions --------------------------------------------

// b39LandEnteredController is Zo-Zu the Punisher's condition: some
// land entered the battlefield, and this is who controls it. ANY
// land, under ANY player's control, however it arrived — CR 603.2
// watches the entry, not the land drop, so a fetched or reanimated
// land counts and so does Zo-Zu's controller's own.
//
// The controller is read from the card as the trigger is BUILT, which
// is the last moment the entering permanent is guaranteed to be
// findable; the ID is then closed over, so a land that has left by
// the time the trigger resolves still damages the right player.
func b39LandEnteredController(ev game.Event, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventETB {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsLand() {
		return uuid.Nil, false
	}
	return c.Controller, true
}

// b39SeaMonsterYouControlDealtCombatDamageToPlayer is Spawning
// Kraken's condition: "whenever a Kraken, Leviathan, Octopus, or
// Serpent you control deals combat damage to a player". The Kraken
// itself is one of the four, so it triggers off its own connection.
//
// Effective subtypes, so a changeling and anything a type-granting
// lord has touched joins the club (CR 702.73a).
func b39SeaMonsterYouControlDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && (c.HasSubtype("Kraken") || c.HasSubtype("Leviathan") ||
		c.HasSubtype("Octopus") || c.HasSubtype("Serpent"))
}

// b39AnotherCreatureYouControlDied is Erebos, Bleak-Hearted's
// condition: a creature its controller controlled, other than Erebos
// himself, died. The printed "another" excludes only the source
// (CR 603.6c), so a second creature dying in the same batch is a
// second trigger.
func b39AnotherCreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.InstanceID != source.InstanceID && dead.Controller == source.Controller
}

// b39PlusOneCountersOnCreatureYouControl is Shalai and Hallar's
// condition: one or more +1/+1 counters were put on a creature its
// controller controls, and this is how many landed.
//
// Built on b12CountersPlacedByYou, which carries the attribution
// rules (and errs toward NOT firing) — see All Will Be One. The extra
// narrowing here is the printed one: the counters must be +1/+1 and
// the permanent they landed on must be a CREATURE the source's
// controller controls.
func b39PlusOneCountersOnCreatureYouControl(ev game.Event, source *game.Card, g *game.Game) (int, bool) {
	if ev.Label != game.CounterPlusOne {
		return 0, false
	}
	n, ok := b12CountersPlacedByYou(ev, source, g)
	if !ok {
		return 0, false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !target.IsCreature() || target.Controller != source.Controller {
		return 0, false
	}
	return n, true
}

// --- board reads ---------------------------------------------------

// b39CountersOn is the number of `kind` counters on `cardID` right
// now — Dawn of a New Age's hope counters, Empowered Autogenerator's
// charge counters. Zero for a card that has left the battlefield.
func b39CountersOn(g *game.Game, cardID uuid.UUID, kind string) int {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return 0
	}
	return c.Counters[kind]
}

// b39ChargeCountersOn is b39CountersOn fixed to charge counters.
func b39ChargeCountersOn(g *game.Game, cardID uuid.UUID) int {
	return b39CountersOn(g, cardID, game.CounterCharge)
}

// b39CreaturesControlledBy counts the creatures `controller` controls
// — Dawn of a New Age's "a hope counter for each creature you
// control", read as the enchantment enters.
func b39CreaturesControlledBy(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsCreature() {
			n++
		}
	}
	return n
}

// b39AnOpponentHasMoreLands / …MoreLife / …MoreCreatures /
// …MoreCardsInHand are Beza, the Bounding Spring's four independent
// comparisons. Each is "an opponent" — ANY ONE opponent beating the
// controller is enough, which is why the four are separate reads and
// not one "the opponent who is ahead" question. Each is evaluated at
// resolution, in printed order, against the board as it then is.

func b39AnOpponentControlsMoreLands(g *game.Game, you uuid.UUID) bool {
	return b39AnOpponentBeatsYou(g, you, func(p uuid.UUID) int {
		return b39PermanentsControlledBy(g, p, func(c game.Card) bool { return c.IsLand() })
	})
}

func b39AnOpponentControlsMoreCreatures(g *game.Game, you uuid.UUID) bool {
	return b39AnOpponentBeatsYou(g, you, func(p uuid.UUID) int {
		return b39PermanentsControlledBy(g, p, func(c game.Card) bool { return c.IsCreature() })
	})
}

func b39AnOpponentHasMoreLife(g *game.Game, you uuid.UUID) bool {
	return b39AnOpponentBeatsYou(g, you, func(p uuid.UUID) int {
		if pl := g.PlayerByIDForEffect(p); pl != nil {
			return pl.Life
		}
		return 0
	})
}

func b39AnOpponentHasMoreCardsInHand(g *game.Game, you uuid.UUID) bool {
	return b39AnOpponentBeatsYou(g, you, func(p uuid.UUID) int {
		pl := g.PlayerByIDForEffect(p)
		if pl == nil || pl.Hand == nil {
			return 0
		}
		return len(pl.Hand.Cards)
	})
}

// b39AnOpponentBeatsYou is the shared body: does any opponent of
// `you` who is still in the game score strictly higher than `you` on
// `measure`? A player who has left the game is not an opponent
// (CR 800.4a), so an eliminated seat is skipped — the same seat walk
// Context.Opponents does, spelled out here because these run from a
// trigger body that holds only the game, not a Context.
func b39AnOpponentBeatsYou(g *game.Game, you uuid.UUID, measure func(uuid.UUID) int) bool {
	mine := measure(you)
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you {
			continue
		}
		if measure(p.ID) > mine {
			return true
		}
	}
	return false
}

// b39PermanentsControlledBy counts the permanents `controller`
// controls that match. Reads g.Battlefield.Cards directly, which is
// legal under the resolution write lock the effect bodies run in.
func b39PermanentsControlledBy(g *game.Game, controller uuid.UUID, match func(game.Card) bool) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && match(c) {
			n++
		}
	}
	return n
}

// --- effect bodies -------------------------------------------------

// b39DamageToFirstTargetPlayer is the resolution body shared by every
// "deals N damage to target opponent" trigger in the batch: the
// announced player, if the target is still legal at resolution
// (CR 608.2b), takes `amount` from the source.
func b39DamageToFirstTargetPlayer(g *game.Game, item *game.StackItem) error {
	amount := item.Params.Amount
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}.Apply(ctx)
	}
	return nil
}

// b39MayDiscardThenDraw is the "you may discard N cards. If you do,
// draw M cards" shape — Thrilling Discovery — and, with draw == the
// count actually pitched, "discard up to N cards, then draw that
// many" — Cathartic Pyre's second mode.
//
// The count is the RUN's (#1027): "you may discard N cards" is one
// printed instruction, and its continuation is told what was really
// discarded once the chosen cards have finished moving. `draw` turns
// that count into the number of cards to draw, so a card with a fixed
// reward ("draw three") ignores it and a card with a matching reward
// ("draw that many") returns it.
//
// It used to be measured by snapshotting the hand size before the
// prompt and reading it again in the prompt's own Then, because the
// discard prompt reported nothing back. That was right about the
// ordinary case and wrong about two: a hand that a discard TRIGGER
// refilled between the pitch and the measurement counted short, and
// a card the CR 614 window left in hand counted as pitched only
// because the subtraction could not see it. The run counts the cards,
// not the hand.
//
// UpTo is the printed ceiling. With `exact` set, Validate refuses any
// answer between 1 and N-1: "you may discard TWO cards" is a yes-or-no
// on the pair, not a licence to pitch one. Without it the clause is a
// genuine "up to".
func b39MayDiscardThenDraw(n int, exact bool, question string, draw func(discarded int) int) func(*Context) error {
	return func(ctx *Context) error {
		player := ctx.Controller()
		prompt := game.DiscardPrompt{
			Player:   player,
			Source:   ctx.Source(),
			N:        n,
			UpTo:     true,
			Question: question,
		}
		if exact {
			prompt.Validate = func(picked []game.Card) bool {
				return len(picked) == 0 || len(picked) == n
			}
		}
		return ctx.Game.PlayerDiscardsThenForEffect(prompt,
			func(g *game.Game, discarded game.PromptedDiscards) error {
				if k := draw(discarded.Count()); k > 0 {
					return g.DrawNForEffect(player, k)
				}
				return nil
			})
	}
}
