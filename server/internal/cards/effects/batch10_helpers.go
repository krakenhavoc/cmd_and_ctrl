package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch10_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 10 (#303, `edhrec_rank` 1111–1213). Own file per the
// #231 convention; every package-level name carries the b10 prefix
// because batch 09 is landing beside this one.
//
// What is NOT here, because main already had it: the attack-trigger
// dedup is OncePerBatch, "double the +1/+1 counters on
// each creature you control" is b08DoubleCountersOnEachCreatureYouControl,
// "unless you control a legendary creature" is
// b08ControlsLegendaryCreature, the filter-land shape is
// b08FilterLand, "this permanent enters" is b06SelfETB, "target
// legendary permanent" is b05Legendary, "each opponent loses N life"
// is eachOpponentLosesLife, and the wheel body is discardWholeHand +
// tablePlayers.

// --- token templates ---------------------------------------------

// --- entry replacements ------------------------------------------

// b10EntersWithCounters is "this creature enters with N <kind>
// counters on it" — Kalonian Hydra's four +1/+1 — as a
// self-replacement on the card's own entry (Mossborn Hydra's shape,
// with the count as a parameter). The counters are on the card
// before any ETB trigger or state check sees the printed 0/0.
func b10EntersWithCounters(kind string, n int, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, n)
			return nil
		},
		Label: label,
	}
}

// --- trigger conditions ------------------------------------------

// b10YouGainedLife is "whenever you gain life" — Marauding
// Blight-Priest. Every lifegain path in the engine emits
// EventChangeLife with a positive Amount, lifelink included (Sanguine
// Bond's note), so the condition is the sign of the delta on the
// source's controller.
func b10YouGainedLife(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventChangeLife && ev.Target == source.Controller && ev.Amount > 0
}

// b10NoncreatureSpellCastByYou is "whenever you cast a noncreature
// spell" — Firebrand Archer. Flux Channeler's condition, shaped as a
// TriggeredAbility AppliesTo so the card can name it directly. The
// spell is read off the stack, where its type line is intact.
func b10NoncreatureSpellCastByYou(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !spell.IsCreature()
}

// b10CommanderYouControlDealtCombatDamageToOpponent is Kediss,
// Emberclaw Familiar's condition: a commander the source's controller
// controls dealt combat damage to an opponent. The dealing creature
// is looked up live (combat damage is dealt before SBAs run, so an
// attacker that traded is still on the battlefield), and the
// IsCommander flag is the deck importer's — the same field Bastion
// Protector reads.
func b10CommanderYouControlDealtCombatDamageToOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) || ev.Target == source.Controller {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.IsCommander
}

// b10LandYouControlDied is Titania, Protector of Argoth's second
// trigger: a land the source's controller controlled was put into a
// graveyard from the battlefield. Read post-move, like diedCreature —
// TypeLine and Controller survive the move, so "you control" is the
// controller it left under.
func b10LandYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && c.Controller == source.Controller
}

// b10AnotherLandPlayedByYou is City of Traitors' "when you play
// another land". The engine emits no land-play event, so the trigger
// watches the EventZoneMove that precedes every battlefield entry and
// reads where the land came FROM: a land played from hand (or from an
// impulse-exile grant, or a graveyard permission) arrives from that
// zone, while the "put onto the battlefield" effects — Cultivate,
// Rampant Growth, every fetchland — arrive from the LIBRARY, which is
// the one origin a land play can never have.
//
// The residual gap runs the weaker way for the City's controller: a
// land RETURNED from a graveyard or exile by an effect (Titania's ETB,
// Splendid Reclamation, a flickered land) reads as a play and costs
// the City too. Declared on the card.
func b10AnotherLandPlayedByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneBattlefield || ev.CardID == source.InstanceID {
		return false
	}
	if ev.OldZone == game.ZoneLibrary || ev.OldZone == game.ZoneBattlefield || ev.OldZone == game.ZoneStack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && c.Controller == source.Controller
}

// b10AnotherPermanentWithSubtypeEnteredUnderYourControl is "whenever
// another <type> you control enters" — Marwyn's Elf. Any permanent
// with the subtype, as printed; effective subtypes, so a changeling
// counts.
func b10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev game.Event, source *game.Card, g *game.Game, subtype string) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(subtype)
}

// --- predicates --------------------------------------------------

// b10NoCounters is Damning Verdict's "creatures with no counters on
// them": no counter of any kind with a positive count. A card whose
// counter map holds a zero entry (a counter that was removed) has no
// counters, as printed.
func b10NoCounters() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, n := range c.Counters {
			if n > 0 {
				return false
			}
		}
		return true
	}
}

// b10NonbasicLand is Wasteland's "nonbasic land": a land without the
// Basic supertype. Supertypes come from the effective characteristic,
// which off the battlefield and for every land in the catalog is the
// printed type line — a nonbasic land granted a basic land TYPE by
// Urborg is still nonbasic, because the type is not the supertype.
func b10NonbasicLand() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		if !c.IsLand() {
			return false
		}
		for _, s := range c.Effective().Supertypes {
			if s == "Basic" {
				return false
			}
		}
		return true
	}
}

// b10CreatureOrPlaneswalker is the object class Soul Shatter and
// Sheoldred's Edict's third mode speak about.
func b10CreatureOrPlaneswalker() CardPredicate { return Or(Creature(), Planeswalker()) }

// b10GreatestManaValueCreatureOrPlaneswalkerYouControl is Soul
// Shatter's "a creature or planeswalker with the greatest mana value
// among creatures and planeswalkers they control". The sacrifice
// fan-out hands each affected player's own ID in as the `caster`, so
// the comparison set is that player's board; ties all match, and the
// player chooses among them (CR 700.3 — "the greatest" picks out a
// set). Mana value is read off the battlefield, where X is zero (CR
// 202.3e) and a token with no mana cost is zero.
func b10GreatestManaValueCreatureOrPlaneswalkerYouControl() CardPredicate {
	return func(g *game.Game, player uuid.UUID, c game.Card) bool {
		if c.Controller != player || !(c.IsCreature() || c.IsPlaneswalker()) {
			return false
		}
		best := c.ManaValue()
		for _, other := range g.BattlefieldCardsForEffect() {
			if other.Controller != player || !(other.IsCreature() || other.IsPlaneswalker()) {
				continue
			}
			if other.ManaValue() > best {
				return false
			}
		}
		return true
	}
}

// b10SpellTargetsYouOrACreatureYouControl is Siren Stormtamer's
// clause: the candidate spell on the stack has at least one target
// that is the caster or a creature the caster controls ON THE
// BATTLEFIELD. The zone check matters: a creature already in a
// graveyard keeps its Controller field, and without it the Stormtamer
// could protect a creature that has already left — which is also why
// sacrificing the Stormtamer to counter a spell aimed only at itself
// fizzles the ability at resolution, exactly as the printed card's
// ruling has it.
func b10SpellTargetsYouOrACreatureYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			switch t.Kind {
			case game.TargetPlayer:
				if t.ID == caster {
					return true
				}
			case game.TargetCard:
				z := g.FindCardZoneForEffect(t.ID)
				if z == nil || z.Kind != game.ZoneBattlefield {
					continue
				}
				if tc, ok := g.LookupCardForEffect(t.ID); ok && tc.IsCreature() && tc.Controller == caster {
					return true
				}
			}
		}
		return false
	}
}

// b10OneOf matches a permanent whose instance ID is in `keep` — the
// "choose up to two creatures, then destroy the rest" exclusion on
// Mount Doom.
func b10OneOf(keep []uuid.UUID) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, id := range keep {
			if id == c.InstanceID {
				return true
			}
		}
		return false
	}
}

// --- counts ------------------------------------------------------

// b10LandsControlled counts the lands `controller` controls — Lumra's
// characteristic-defining power and toughness. Walks the live slice
// because it runs inside a layer recompute (Adeline's shape).
func b10LandsControlled(g *game.Game, controller uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsLand() {
			n++
		}
	}
	return n
}

// --- costs -------------------------------------------------------

// b10SacrificeAnArtifact is Goblin Engineer's "Sacrifice an
// artifact:". Built through the same clause the sac-outlet costs use,
// so the client opens the same picker.
func b10SacrificeAnArtifact() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact", Artifact())}
}

// b10SacrificeALegendaryArtifact is the "and a legendary artifact"
// half of Mount Doom's last cost; the "Sacrifice Mount Doom" half is
// SacrificeThis, and Plus merges the two.
func b10SacrificeALegendaryArtifact() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a legendary artifact", Artifact(), b05Legendary())}
}

// --- effect bodies -----------------------------------------------

// b10EachPlayerWheels is "each player discards their hand, then draws
// seven cards" — Wheel of Fortune's body, shared with Magus of the
// Wheel. Every discard happens before any draw, so discard payoffs
// queue while the effect is still resolving.
func b10EachPlayerWheels(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := tablePlayers(ctx)
	for _, id := range players {
		if _, err := discardWholeHand(g, id); err != nil {
			return err
		}
	}
	for _, id := range players {
		if err := (DrawCards{Player: id, N: 7}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b10ReturnAllLandCardsFromGraveyardTapped is Splendid Reclamation's
// body, shared with Lumra's ETB: every land card in `player`'s
// graveyard returns to the battlefield under its owner's control and
// is tapped. The IDs are snapshotted before the first move, because
// ReturnFromGraveyard mutates the pile being walked.
//
// Same declared gap as the Reclamation: ReturnFromGraveyard has no
// tapped flag, so each land enters untapped and is tapped a beat
// later inside the same resolution.
func b10ReturnAllLandCardsFromGraveyardTapped(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var lands []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsLand() {
			lands = append(lands, c.InstanceID)
		}
	}
	for _, id := range lands {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b10Fight is CR 701.14: each creature deals damage equal to its
// power to the other. Both amounts are read before either is dealt,
// so a first blow cannot change the second (the two are simultaneous
// in the rules), and a creature that has left the battlefield fights
// nothing. Damage goes through DealDamageToCreatureForEffect, so a
// prevention shield or a damage doubler sees it; lethal damage is the
// SBA's business at the next check, as for every other effect.
func b10Fight(ctx *Context, a, b uuid.UUID) error {
	ctx.Game.RecomputeLayersIfStaleLocked()
	ca, okA := ctx.Game.LookupCardForEffect(a)
	cb, okB := ctx.Game.LookupCardForEffect(b)
	if !okA || !okB {
		return nil
	}
	za, zb := ctx.Game.FindCardZoneForEffect(a), ctx.Game.FindCardZoneForEffect(b)
	if za == nil || zb == nil || za.Kind != game.ZoneBattlefield || zb.Kind != game.ZoneBattlefield {
		return nil
	}
	pa, pb := ca.CurrentPower(), cb.CurrentPower()
	if err := (DealDamage{Source: a, Target: b, Amount: pa}).Apply(ctx); err != nil {
		return err
	}
	return DealDamage{Source: b, Target: a, Amount: pb}.Apply(ctx)
}

// b10DestroyAllCreaturesExcept is Mount Doom's "choose up to two
// creatures, then destroy the rest": one simultaneous destruction of
// every creature not in `keep`.
func b10DestroyAllCreaturesExcept(ctx *Context, keep []uuid.UUID) error {
	return DestroyAllMatching{Match: Except(Creature(), b10OneOf(keep))}.Apply(ctx)
}

// b10DamageEachLegalTarget is Crackle with Power's body: `amount`
// damage from the spell to each announced target that is still legal
// (CR 608.2b — a target that left in response is skipped, the rest
// are still hit).
func b10DamageEachLegalTarget(ctx *Context, amount int) error {
	for _, t := range ctx.LegalTargets() {
		if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
