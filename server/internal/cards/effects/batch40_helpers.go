package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch40_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 40 (#403, `edhrec_rank` 4154–4253). Own file per
// the #231 convention; every package-level name carries the b40
// prefix because other batches land beside this one.
//
// What is NOT here, because the package already had it: "this
// permanent enters" is b06SelfETB, the enters-tapped replacements are
// SelfEntersTapped / SelfEntersTappedUnless, "your opponents control
// eight or more lands" is b40CatchUpDualCondition (below, so the
// catch-up cycle's threshold is written once), "each player draws N"
// is b05EachPlayerDraws, "N damage to each opponent" is
// damageToEachOpponent, "this ability triggers only once each turn"
// is b11TriggeredThisTurn, "whenever you gain life" is
// WheneverYouGainLife, the pain rider is PainRider, the dual pipe is
// dualManaAbility, and "another" by name is b03NotNamed.

// --- land cycles ---------------------------------------------------

// b40CatchUpDualCondition is the untapped clause the whole "catch-up"
// typed-dual cycle prints, word for word: "unless your opponents
// control eight or more lands". Turbulent Fen (batch 34), Turbulent
// Springs (batch 36) and the two in this batch share it, and the
// remaining member will.
//
// It exists so the cycle's threshold is written once. A card file
// spelling the 8 itself is a card file that can get the cycle's one
// number wrong, and five lands whose thresholds disagreed would be
// invisible until someone played the odd one out.
func b40CatchUpDualCondition() func(g *game.Game, controller uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		total := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != controller && c.IsLand() {
				total++
			}
		}
		return total >= 8
	}
}

// --- predicates ----------------------------------------------------

// b40Outlaw is the Outlaws of Thunder Junction batch type, as the
// reminder text spells it out: "Assassins, Mercenaries, Pirates,
// Rogues, and Warlocks are outlaws." It is a batch, not a creature
// type — no permanent has the subtype "Outlaw" — so the test is
// membership in those five.
//
// Effective subtypes, so a changeling is an outlaw and a creature
// something has turned into a Pirate is one too. Shoot the Sheriff
// asks for the complement of this.
func b40Outlaw() CardPredicate {
	return AnySubtype("Assassin", "Mercenary", "Pirate", "Rogue", "Warlock")
}

// --- mana-ability riders -------------------------------------------

// b40PingEachOpponent is Zhur-Taa Druid's rider: "it deals N damage
// to each opponent", run immediately after the mana lands in the
// pool, inside the same atomic mana-ability resolution (CR 605.3b).
//
// It cannot use damageToEachOpponent, which reads the opponent list
// off a resolving stack item: a mana ability never uses the stack
// (CR 605.3), so there is no item and the seat walk has to be done
// here against the activator. Eliminated seats are skipped, as
// everywhere else.
func b40PingEachOpponent(n int) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == controller {
				continue
			}
			if err := g.DealDamageToPlayerForEffect(source, p.ID, n); err != nil {
				return err
			}
		}
		return nil
	}
}

// --- mana-ability conditions ---------------------------------------

// b40NotActivatedForManaThisTurn is "Activate only once each turn" on
// a MANA ability (Three Tree Mascot). The engine keeps no per-ability
// activation tally, so the tally is the event log: the activation
// path emits EventManaAbilityActivated with Source set to the
// permanent, and g.EventsThisTurn() is already bounded to the current
// turn.
//
// Keyed on the SOURCE only, which is exact for a permanent whose sole
// mana ability this is, and would be too strict for one with two —
// the safe direction, and no card in the catalog has the second
// shape.
func b40NotActivatedForManaThisTurn() func(g *game.Game, controller, source uuid.UUID) bool {
	return func(g *game.Game, _ uuid.UUID, source uuid.UUID) bool {
		for _, ev := range g.EventsThisTurn() {
			if ev.Kind == game.EventManaAbilityActivated && ev.Source == source {
				return false
			}
		}
		return true
	}
}

// --- damage replacement reads --------------------------------------

// b40RedOrArtifactSourceControlledBy is the first half of Mechanized
// Warfare's condition: "a red or artifact source you control".
//
// The source is looked up wherever it now is — a burn spell is in its
// owner's graveyard by the time its damage resolves — so its colour
// and its controller stay readable. Effective characteristics, so an
// artifact creature an effect has turned blue still qualifies as an
// artifact, and a creature an effect has turned red qualifies as red.
//
// A source the engine cannot look up at all is not boosted, which is
// weaker than printed and never stronger.
func b40RedOrArtifactSourceControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	if ev.DamageSource == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.DamageSource)
	if !ok || c.Controller != controller {
		return false
	}
	return c.HasColor("R") || c.IsArtifact()
}

// b40DamageHitsAnOpponentOf is the second half: "to an opponent or a
// permanent an opponent controls". A player target is an opponent
// when it is a seated, non-eliminated player other than `controller`;
// a permanent target is one when its controller is.
//
// Damage to the controller themself, to their own permanents, and to
// anything the engine cannot identify is left alone.
func b40DamageHitsAnOpponentOf(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	if ev.DamageTarget == uuid.Nil || ev.DamageTarget == controller {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.DamageTarget); p != nil {
		return !p.Eliminated
	}
	c, ok := g.LookupCardForEffect(ev.DamageTarget)
	return ok && c.Controller != controller
}

// --- trigger conditions --------------------------------------------

// b40IsYourSecondDrawThisTurn is Gixian Puppeteer's "whenever you
// draw your SECOND card each turn": `ev` is a draw by the source's
// controller and exactly one earlier draw by that same player has
// happened this turn.
//
// The tally is the event log rather than TurnTally.CardsDrawn,
// because a trigger's AppliesTo runs while the event is being emitted
// and must not depend on whether the counter has been bumped yet.
// g.EventsThisTurn() is already bounded to the current turn, and the
// walk stops at `ev` itself, so the third and later draws of a turn
// find two earlier ones and do not fire.
//
// A draw the player makes during an OPPONENT's turn counts — the
// clause is "each turn", not "each of your turns", so a Windfall on
// someone else's turn triggers this.
func b40IsYourSecondDrawThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDrawCard || ev.Actor != source.Controller {
		return false
	}
	earlier := 0
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventDrawCard && prev.Actor == source.Controller {
			earlier++
			if earlier > 1 {
				return false
			}
		}
	}
	return earlier == 1
}

// b40CreatureYouControlTargetedByASpell is Gargos, Vicious Watcher's
// condition: "whenever a creature you control becomes the target of A
// SPELL". Gargos himself counts — he is a creature his controller
// controls.
//
// EventBecomesTarget is emitted for abilities as well as spells and
// carries no discriminator, so the spell half is read off the stack:
// a SPELL's stack item shares its ID with the source card, while an
// activated or triggered ability's ID is freshly minted and its
// source card stays in its zone. Looking the source ID up in
// StackMeta and requiring a spell item is therefore exact.
//
// Firing on abilities too would be a strictly better Gargos, which is
// the direction this catalog does not ship (#259).
func b40CreatureYouControlTargetedByASpell(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil {
		return false
	}
	if !b40TargetedByASpell(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller
}

// b40TargetedByASpell reports whether the object that did the
// targeting is a spell on the stack rather than an ability.
func b40TargetedByASpell(ev game.Event, g *game.Game) bool {
	if ev.Source == uuid.Nil {
		return false
	}
	item := g.StackItemForEffect(ev.Source)
	return item != nil && item.Kind == game.StackItemSpell
}

// --- trigger bodies ------------------------------------------------

// b40DrainEachOpponent is "each opponent loses N life and you gain N
// life" for an N the package's drainEachOpponent does not cover (it
// is fixed at 1). Two cards in this batch drain for two — Vraan,
// Executioner Thane and Gixian Puppeteer — which is what earns it a
// name rather than two copies of the same four lines.
//
// Life LOSS, not damage: nothing prevents it, no damage replacement
// doubles it, and lifelink does not see it.
func b40DrainEachOpponent(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if err := eachOpponentLosesLife(g, item, n); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
	}
}

// b40ChaosSpillover is Vincent, Vengeful Atoner's Chaos ability: the
// source deals `amount` damage to each opponent OTHER than `hit`, but
// only if its power is at least `minPower`.
//
// The power check happens here, at RESOLUTION, because that is where
// the printed text puts it — after the effect clause rather than as
// an intervening-if. So Vincent can be pumped in response and the
// spillover happens; a Vincent who has shrunk, or who has left the
// battlefield and has no power to read at all, does nothing.
//
// CurrentPower(), not Effective().Power (#1281): the gate is "if its
// power is 7 or greater", and Effective().Power excludes +1/+1 / -1/-1
// counters, so a Vincent pumped by counters rather than by an anthem
// never reached the threshold.
func b40ChaosSpillover(g *game.Game, item *game.StackItem, hit uuid.UUID, amount, minPower int) error {
	if amount <= 0 {
		return nil
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || src.CurrentPower() < minPower {
		return nil
	}
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if opp == hit {
			continue
		}
		if err := (DealDamage{Source: item.SourceCardID, Target: opp, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// --- cost predicates -----------------------------------------------

// b40HydraSpell passes on a Hydra spell — Gargos, Vicious Watcher's
// "Hydra spells you cast cost {4} less". The subtype is read off the
// card being cast, which on the stack is its printed type line.
func b40HydraSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasSubtype("Hydra") }
}

// --- card reads ----------------------------------------------------

// b40ControlsA reports whether `controller` controls a battlefield
// permanent matching `match` — Dazzling Denial's "if you control a
// Bird", read at resolution rather than at cast.
//
// Post-layer characteristics, so a changeling is a Bird and an
// animated land that something has made a Bird is one too.
func b40ControlsA(g *game.Game, controller uuid.UUID, match CardPredicate) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && match(g, controller, c) {
			return true
		}
	}
	return false
}

// The "lands you control" count Lumbering Worldwagon's
// characteristic-defining power needs is b10LandsControlled, which
// already walks g.Battlefield directly — the form a layer recompute
// requires, since that pass may hold only the read lock.
