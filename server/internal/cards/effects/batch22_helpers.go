package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch22_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 22 (#384, `edhrec_rank` 2335–2435). Own file per
// the #231 convention; every package-level name carries the b22
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "another nontoken
// creature you control entered" is b15AnotherNontokenCreatureYouControlEntered,
// "you gained life" is b10YouGainedLife, "an opponent discarded" is
// b18OpponentDiscarded, "you cast a creature / instant or sorcery
// spell" is creatureSpellCastByYou / instantOrSorceryCastByYou,
// landfall is b13LandYouControlEntered, "an opponent cast a spell" is
// b15OpponentCastSpell, the dead-creature counter read is
// b13LastKnownCounters, the counter doubler is b17DoubleCountersOn,
// the mass sacrifice is b08SacrificeAllMatching, the damage doubler
// shape is Angrath's Marauders', the evoke-by-pitch cost is
// EvokePitch, and the Food / Soldier templates live in tokens.go.

// --- token templates ---------------------------------------------

// --- predicates --------------------------------------------------

// b22SpellManaValueGE is "target spell with mana value N or greater"
// — Disdainful Stroke. Read on the stack (CR 202.3e), so a Hydra cast
// for X=5 is a legal target and one cast for X=1 is not.
func b22SpellManaValueGE(n int) CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return manaValueOnStack(c, g.StackItemForEffect(c.InstanceID)) >= n
	}
}

// --- trigger conditions ------------------------------------------

// b22FirstSpellOnAnOpponentsTurn is Wavebreak Hippocamp's condition:
// the source's controller cast a spell, it is not their turn, and the
// engine's per-turn cast tally — bumped before EventCast fires — says
// it is their first this turn.
func b22FirstSpellOnAnOpponentsTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if active == nil || active.ID == source.Controller {
		return false
	}
	return g.CastTallyFor(source.Controller).Total == 1
}

// b22CreatureYouControlDied is "whenever a creature you control
// dies" on a NONcreature source (Cauldron of Essence): the dead
// creature, read post-move, was the source's controller's.
func b22CreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller
}

// b22SlimedCreatureYouDontControlDied is Toxrill's third ability: a
// creature the source's controller did not control died with a slime
// counter on it. The card's counters are cleared on the way out
// (CR 400.7), so the count is read back off the log.
func b22SlimedCreatureYouDontControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || dead.Controller == source.Controller {
		return false
	}
	return b13LastKnownCounters(g, dead.InstanceID, "slime") > 0
}

// b22SelfDiedOrWasExiledFromBattlefield is God-Eternal Oketra's
// "dies or is put into exile from the battlefield": the source's own
// EventLTB, bound for a graveyard or exile — a bounce or a library
// tuck does not count.
func b22SelfDiedOrWasExiledFromBattlefield(ev game.Event, source *game.Card) bool {
	if ev.Kind != game.EventLTB || ev.CardID != source.InstanceID {
		return false
	}
	return ev.NewZone == game.ZoneGraveyard || ev.NewZone == game.ZoneExile
}

// b22SelfEnteredOrDied is Mycosynth Wellspring's one ability with
// two trigger conditions — "enters or is put into a graveyard from
// the battlefield".
func b22SelfEnteredOrDied(ev game.Event, source *game.Card) bool {
	return (ev.Kind == game.EventETB && ev.CardID == source.InstanceID) || cardDied(ev, source)
}

// --- board reads -------------------------------------------------

// b22OtherCreaturesOfSubtypeControlled counts the creatures
// `controller` controls with the given subtype, the source excluded
// — Earthshaker Dreadmaw's "each other Dinosaur you control". Post-
// layer subtype, so a changeling counts.
func b22OtherCreaturesOfSubtypeControlled(g *game.Game, controller, except uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != except && c.Controller == controller && c.IsCreature() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// --- effect bodies -----------------------------------------------

// b22PutCounterOnEachCreatureYouDontControl is Toxrill's end-step
// trigger: a slime counter on each creature the source's controller
// does not control, the set snapshotted before the first counter
// lands.
func b22PutCounterOnEachCreatureYouDontControl(g *game.Game, item *game.StackItem, kind string) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	ctx := NewContext(g, item)
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: kind, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b22SlimeShrink is Toxrill's static: "Creatures you don't control
// get -1/-1 for each slime counter on them." Layer 7c, reading the
// live counter map; a counter placement bumps the layer version, so
// the end-step trigger's counters shrink the creature at once.
func b22SlimeShrink() game.StaticAbility {
	return game.StaticAbility{
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller != source.Controller && target.Counters["slime"] > 0
		},
		Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
			n := target.Counters["slime"]
			c.Power -= n
			c.Toughness -= n
		},
	}
}

// b22KeepLowestPowerWithinTotal is Slaughter the Strong's auto-pick
// for one player: the creatures they keep, chosen lowest power first
// (battlefield order on a tie) while the running total stays at or
// under `limit`, so as many creatures as possible survive. A
// creature whose power alone exceeds the limit is never kept; a
// zero-power creature always is. Returns the kept instance IDs.
func b22KeepLowestPowerWithinTotal(g *game.Game, player uuid.UUID, limit int) []uuid.UUID {
	type entry struct {
		id    uuid.UUID
		power int
		order int
	}
	var mine []entry
	for i, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			mine = append(mine, entry{id: c.InstanceID, power: c.CurrentPower(), order: i})
		}
	}
	sort.SliceStable(mine, func(i, j int) bool {
		if mine[i].power != mine[j].power {
			return mine[i].power < mine[j].power
		}
		return mine[i].order < mine[j].order
	})
	var keep []uuid.UUID
	total := 0
	for _, e := range mine {
		if total+e.power > limit {
			break
		}
		total += e.power
		keep = append(keep, e.id)
	}
	return keep
}

// b22EachPlayerKeepsWithinPowerAndSacrificesTheRest is Slaughter the
// Strong's body: every seated player keeps the b22KeepLowestPowerWithinTotal
// set and sacrifices every other creature they control. The keep
// sets are all chosen before anything is sacrificed (CR 608.2c —
// "then"), and the sacrifices go one creature at a time in
// battlefield order through the ordinary sacrifice path, so
// aristocrats payoffs see each one.
func b22EachPlayerKeepsWithinPowerAndSacrificesTheRest(g *game.Game, item *game.StackItem, limit int) error {
	keep := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		for _, id := range b22KeepLowestPowerWithinTotal(g, p.ID, limit) {
			keep[id] = true
		}
	}
	ctx := NewContext(g, item)
	return b08SacrificeAllMatching(ctx, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && !keep[c.InstanceID]
	})
}

// b22DamageDividedEvenly is Fury's "4 damage divided as you choose
// among any number of target creatures and/or planeswalkers", with
// the division made for the player: as evenly as possible across the
// legal targets in the order they were picked, the remainder going
// to the earliest picks. One target takes it all; four take one
// each. Declared on the card — the engine's pick_target prompt
// carries no distribution.
func b22DamageDividedEvenly(ctx *Context, total int) error {
	var targets []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			targets = append(targets, t.ID)
		}
	}
	if len(targets) == 0 || total <= 0 {
		return nil
	}
	share, extra := total/len(targets), total%len(targets)
	for i, id := range targets {
		amount := share
		if i < extra {
			amount++
		}
		if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b22DestroyFirstArtifactAndFirstEnchantment is Hull Breach's third
// mode: of the two targets picked, the first that is an artifact and
// the first that is an enchantment are destroyed. The picker cannot
// tie a predicate to a slot, so it accepts any two artifacts or
// enchantments; two of the same kind destroy only the first, which
// is weaker than the printed clause and never stronger.
func b22DestroyFirstArtifactAndFirstEnchantment(ctx *Context) error {
	artifact, enchantment := uuid.Nil, uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		if artifact == uuid.Nil && c.IsArtifact() {
			artifact = t.ID
			continue
		}
		if enchantment == uuid.Nil && c.IsEnchantment() {
			enchantment = t.ID
		}
	}
	for _, id := range []uuid.UUID{artifact, enchantment} {
		if id == uuid.Nil {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b22TuckThirdFromTop is God-Eternal Oketra's return: the card goes
// into its owner's library third from the top — under the top two
// cards, or on the bottom when the library holds fewer than two. The
// engine's tuck puts a card on top; the reorder is the slice edit
// scry's answer makes, done here.
func b22TuckThirdFromTop(g *game.Game, cardID uuid.UUID) error {
	if err := g.TuckToLibraryForEffect(cardID, false); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return nil
	}
	owner := g.PlayerByIDForEffect(c.Owner)
	if owner == nil || owner.Library == nil {
		return nil
	}
	lib := owner.Library
	card, err := lib.Remove(cardID)
	if err != nil {
		return err
	}
	// Top is the last element; "third from the top" is two below it.
	at := len(lib.Cards) - 2
	if at < 0 {
		at = 0
	}
	lib.Cards = append(lib.Cards, game.Card{})
	copy(lib.Cards[at+1:], lib.Cards[at:])
	lib.Cards[at] = card
	return nil
}

// b22UntapEachLegalTarget untaps every card target that is still
// legal when the ability resolves — both of Ioreth's activations,
// one slot or two. A target that left or turned illegal in response
// is skipped rather than erroring.
func b22UntapEachLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b22MillHalf is "mills half their library, rounded down" for one
// player — Singularity Rupture.
func b22MillHalf(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return nil
	}
	return MillCards{Player: player, N: p.Library.Size() / 2}.Apply(ctx)
}
