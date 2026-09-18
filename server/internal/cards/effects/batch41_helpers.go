package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch41_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 41 (#448, `edhrec_rank` 4254–4353). Own file per
// the #231 convention; every package-level name carries the b41
// prefix because other batches land beside this one.
//
// What is NOT here, because the package already had it: the lord
// statics are TribalAnthem / TribalKeywordGrant over a TribeFilter,
// "this permanent enters" is b06SelfETB, "this land enters untapped"
// is b27SelfEnteredUntapped, "a creature you control died" is
// ACreatureYouControlDied, "life you gained this turn" is
// b15LifeGainedThisTurn, "lands of a subtype you control" is
// b08LandsWithSubtypeControlled, "a creature entered under your
// control" is enteredUnderYourControl, "the cards exiled with this
// permanent" is b27ExiledWith, the sweep verbs are
// DestroyAllMatching / BounceAllMatching, the equip ability is
// EquipAbility and the attachment statics are PumpAttached /
// GrantToAttached / RestrictAttached.

// --- predicates ----------------------------------------------------

// b41Monocolored is Vanishing Verse's "monocolored permanent": a
// permanent with EXACTLY one colour. Reads EffectiveColors, so a
// layer-5 colour change counts — a Song of the Dryads'd commander is
// colourless and is not a legal target, which is the printed answer.
//
// Colourless is not monocolored (CR 105.2a: an object with no colour
// is colorless, and colorless is not a colour), so an Eldrazi and a
// Sol Ring both survive.
func b41Monocolored() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return len(c.EffectiveColors()) == 1
	}
}

// b41NonHuman is Grumgully's "non-Human creature": effective
// subtypes, so a changeling IS a Human and is excluded, and a
// creature something turned into a Human stops qualifying.
func b41NonHuman(c game.Card) bool { return !c.HasSubtype("Human") }

// b41BasicLandYouControl is Ossification's enchant clause: a basic
// land its controller controls. "Basic" is the supertype off the type
// line, the same read every fetch predicate uses.
func b41BasicLandYouControl() CardPredicate {
	return And(Land(), YouControl(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return IsBasicLand(c)
	})
}

// b41CreatureOrPlaneswalkerAnOpponentControls is Ossification's exile
// clause.
func b41CreatureOrPlaneswalkerAnOpponentControls() CardPredicate {
	return And(Or(Creature(), Planeswalker()), OpponentControls())
}

// b41HasACounter is Chocobo Knights' "creatures you control with
// counters on them" — ANY counter, not only +1/+1: a permanent with a
// charge, loyalty, stun or oil counter is a creature with counters on
// it. Zero-valued entries do not count; the counter map keeps a key
// after the last one is removed.
func b41HasACounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, n := range c.Counters {
			if n > 0 {
				return true
			}
		}
		return false
	}
}

// --- state reads ---------------------------------------------------

// b41CreatureCardsInAllGraveyards is Nighthowler's X: every creature
// card in every player's graveyard, the whole table's. Printed types,
// which is all a card in a graveyard has.
func b41CreatureCardsInAllGraveyards(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				n++
			}
		}
	}
	return n
}

// b41SourceIsEquipped is Merry's "as long as it's equipped": some
// Equipment is attached to the source. Read live on every recompute,
// so the first strike appears and disappears with the sword.
func b41SourceIsEquipped(g *game.Game, source *game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.HasSubtype("Equipment") && c.IsAttachedTo(source.InstanceID) {
			return true
		}
	}
	return false
}

// --- tokens --------------------------------------------------------

// b41PhyrexianHorrorToken is Phyrexian Rebirth's "an X/X colorless
// Phyrexian Horror artifact creature token".
//
// Built by hand rather than as a tokens_table row because its size is
// decided at resolution: a table keyed on printed characteristics has
// no key for a token whose P/T is a count. Negative counts cannot
// happen (the sweep returns how many landed) but are clamped anyway,
// because a negative-toughness token would never reach the
// state-based-action sweep that is supposed to kill a 0/0.
func b41PhyrexianHorrorToken(n int) game.Card {
	if n < 0 {
		n = 0
	}
	return game.Card{
		Name:      "Phyrexian Horror",
		TypeLine:  "Token Artifact Creature — Phyrexian Horror",
		Power:     n,
		Toughness: n,
	}
}

// --- trigger conditions --------------------------------------------

// b41AnotherLegendaryPermanentYouControlEntered is Yoshimaru's
// condition: a permanent other than the source, with the Legendary
// supertype, entered under the source's controller's control.
//
// "Permanent", not "creature": a legendary land, artifact or
// planeswalker feeds the Dog exactly as a legendary creature does,
// which is why the card is a Partner commander for legend-heavy
// decks rather than a creature-count payoff.
func b41AnotherLegendaryPermanentYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsLegendary()
}

// b41AnotherLegendaryCreatureYouControlIsAttacking is the second half
// of Merry's "whenever you attack with Merry AND ANOTHER LEGENDARY
// CREATURE".
//
// Reading it off the battlefield is safe at trigger-harvest time
// because the whole attack declaration is staged first — every
// declared attacker has its AttackingTarget stamped — and only then
// are the EventAttacks emitted, one per creature
// (commitAttackDeclarationLocked). So the first event of a batch
// already sees the complete attacking set.
//
// "Another": the source is excluded by instance. "Legendary creature":
// effective type line, so a legendary artifact that is not a creature
// does not count and one that has been animated does. They need not
// attack the same player.
func b41AnotherLegendaryCreatureYouControlIsAttacking(g *game.Game, source *game.Card) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == source.InstanceID || c.Controller != source.Controller {
			continue
		}
		if c.IsCreature() && c.IsLegendary() && c.AttackingTarget != uuid.Nil {
			return true
		}
	}
	return false
}

// b41DrewYourSecondCardThisTurn is Thopter Fabricator's and Minn's
// "whenever you draw your second card each turn".
//
// The tally is read AFTER the draw, because the harvester runs inside
// the same EmitEvent that records it — so the second draw of a turn
// reports exactly two and every later draw reports more. That is what
// makes this fire once a turn rather than on every draw from the
// second onwards.
func b41DrewYourSecondCardThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDrawCard || ev.Actor != source.Controller {
		return false
	}
	return g.TurnTallyFor(source.Controller).CardsDrawn == 2
}

// b41PermanentDealtDamageToYou is Dissipation Field's condition: a
// permanent on the battlefield dealt damage to the source's
// controller. Combat or otherwise, and a permanent of any type — the
// Field bounces the Ball Lightning that got through and the Rakdos
// Charm that pinged you is not a permanent and is not caught.
//
// The source is looked up live: combat damage is dealt before
// state-based actions run, so an attacker that traded with its
// blocker is still on the battlefield when its damage event fires.
func b41PermanentDealtDamageToYou(ev game.Event, source *game.Card, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target != source.Controller {
		return uuid.Nil, false
	}
	z := g.FindCardZoneForEffect(ev.Source)
	if z == nil || z.Kind != game.ZoneBattlefield {
		return uuid.Nil, false
	}
	return ev.Source, true
}

// b41ZombieYouControlDealtCombatDamageToAnOpponent is Hordewing
// Skaab's condition before the per-combat dedup: a Zombie its
// controller controls dealt combat damage to one of their opponents.
// Effective subtypes, so a changeling counts and the Skaab itself
// counts.
func b41ZombieYouControlDealtCombatDamageToAnOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	p := g.PlayerByIDForEffect(ev.Target)
	if p == nil || p.ID == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.Controller == source.Controller && c.HasSubtype("Zombie")
}

// b41SliverYouControlBecameAnOpponentsTarget is Diffusion Sliver's
// condition. It is Ward's AppliesTo widened from "this permanent" to
// "a Sliver creature you control": the same CR 702.21a shape, with
// the tax charged by the same pay-unless prompt, but watching a whole
// board rather than one card.
//
// `ev.Actor != source.Controller` is the "an opponent controls" half,
// and it means your own Giant Growth on your own Sliver does not tax
// you. Per target INSTANCE, because EventBecomesTarget is emitted per
// target slot (CR 115.7) — a spell targeting two of your Slivers is
// taxed twice.
func b41SliverYouControlBecameAnOpponentsTarget(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.IsCreature() && c.HasSubtype("Sliver")
}

// --- shared bodies -------------------------------------------------

// b41OtherNonHumanCreaturesYouControlEnterWithACounter is Grumgully's
// "each other non-Human creature you control enters with an
// additional +1/+1 counter on it" — b27's Renata replacement with the
// Human exclusion added.
//
// Tokens get the counter too: a created token runs the same
// battlefield-entry pipeline every other permanent runs, so this
// replacement sees one exactly as it sees a cast creature. Hardened
// Scales and Doubling Season apply to it, because it really is a CR
// 614 entry replacement rather than a counter added a beat later.
func b41OtherNonHumanCreaturesYouControlEnterWithACounter(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.Controller == src.Controller && b41NonHuman(entering)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// b41RevealTopThenTakeMatching is the "reveal the top N cards of your
// library, put all [Match] cards revealed this way into your hand and
// the rest on the bottom of your library" family — Goblin Ringleader,
// and the shape Horn of the Mark's look-at half shares.
//
// The take is BounceToHand on a card that is still in a library,
// which is the move b21RevealUntilBasicLandToHand already uses: the
// zone router does not care where a card starts. It is deliberately
// not a library SEARCH — a search shuffles, prompts over the whole
// library, and would let the player pick a Goblin the reveal never
// showed them.
//
// The rest go to the bottom in a RANDOM order. Ringleader prints "in
// any order", which is the player's choice; the engine has no
// ordering prompt for a pile headed to the bottom of a library, and
// random is the strictly-less-informed version of that choice. Every
// card that uses this must declare the caveat.
func b41RevealTopThenTakeMatching(ctx *Context, player uuid.UUID, n int, match func(game.Card) bool, reason string) error {
	revealed := ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), n, reason)
	var rest []uuid.UUID
	for _, id := range revealed {
		c, ok := ctx.Game.LookupCardForEffect(id)
		// A token in a library is not a card (CR 108.2) and cannot be
		// put into a hand; it is left for the bottom sweep, which
		// leaves it where it is for the same reason.
		if !ok || c.IsToken() || !match(c) {
			rest = append(rest, id)
			continue
		}
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return ctx.Game.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, rest)
}

// b41OssificationExileLabel is the stack label Ossification's entry
// trigger carries. b27ExiledWith keys the "until this Aura leaves the
// battlefield" record on it, so the two must agree — a const rather
// than two string literals for exactly that reason.
const b41OssificationExileLabel = "Ossification — exile target creature or planeswalker an opponent controls"

// b41ExileFirstLegalTarget exiles the first still-legal card target
// of a single-target trigger. CR 608.2b: a target that left in
// response is skipped and the ability does as much as it can, which
// for one target is nothing.
func b41ExileFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ExileTarget{Target: t.ID}.Apply(ctx)
		}
	}
	return nil
}

// b41ReturnCardsExiledWithToTheBattlefield is the "until this permanent
// leaves the battlefield" half of an Oblivion Ring: every card this
// source exiled with the named clause that is STILL in exile comes
// back, under its OWNER's control (CR 610.3 — the card returns to the
// player who owned it, not to whoever exiled it).
//
// A card something else has since moved out of exile is not pulled
// back out of wherever it went, which is what b27ExiledWith's
// most-recent-move record buys.
func b41ReturnCardsExiledWithToTheBattlefield(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, id := range b27ExiledWith(g, item.SourceCardID, label) {
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			if err := (ReturnFromExile{Target: id, Controller: c.Owner}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
