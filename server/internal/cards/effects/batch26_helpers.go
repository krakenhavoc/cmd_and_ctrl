package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch26_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 26 (#388, `edhrec_rank` 2738–2838). Own file per
// the #231 convention; every package-level name carries the b26
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "mills half their
// library" is b22MillHalf, the once-per-turn trigger tally is
// b11TriggeredThisTurn with b11CountersWerePlaced for the placement
// read, the attack-trigger dedup is OncePerBatch, the
// "becomes blocked" once-per-attacker dedup is
// b18AttackerAlreadyBlocked, "another <type> you control enters" is
// b10AnotherPermanentWithSubtypeEnteredUnderYourControl, "a green
// creature you control enters" is b23AnotherGreenCreatureYouControlEntered,
// "you gained N life this turn" is b15LifeGainedThisTurn, the end
// step is b15EndStepBegan, the land-count CDA is b10LandsControlled,
// the basic-land search is b07SearchBasicOntoBattlefield, the land
// search is b11FetchLandTapped, the anthem is b16Anthem, the
// first-legal-target destroy is destroyFirstLegalTarget, the
// two-slot "creature you control, then …" posture is Soul's Fire,
// the Elf Warrior token is b13GreenElfWarriorToken and the Angel
// with flying and vigilance is AngelVigilanceToken.

// --- tokens ------------------------------------------------------

// --- trigger conditions ------------------------------------------

// b26AnotherPlayerCastFromOutsideHand is Aerial Extortionist's draw
// condition: a player other than the source's controller cast a
// spell from somewhere other than their hand. The zone is read off
// the spell's announce record; a commander cast from the command
// zone, a flashback, an impulse-exile cast and a cast off the
// Extortionist's own exile grant all count, as printed.
func b26AnotherPlayerCastFromOutsideHand(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	item := g.StackItemForEffect(ev.CardID)
	return item != nil && item.CastFromZone != "" && item.CastFromZone != game.ZoneHand
}

// b26SelfEnteredOrDealtCombatDamageToPlayer is "whenever this
// creature enters or deals combat damage to a player" — one printed
// ability with two conditions (Aerial Extortionist).
func b26SelfEnteredOrDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind == game.EventETB {
		return ev.CardID == source.InstanceID
	}
	return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
}

// b26CreatureYouControlWithPowerAtLeastAttacked is The Earth King's
// condition before the "one or more" dedup: a creature the source's
// controller controls was declared as an attacker and its power,
// read live at declaration, is at least n.
func b26CreatureYouControlWithPowerAtLeastAttacked(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.CurrentPower() >= n
}

// b26NontokenCreatureEntered resolves the nontoken creature that just
// entered the battlefield under anyone's control — Genesis Chamber's
// trigger, which has no controller clause.
func b26NontokenCreatureEntered(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || IsToken(c) {
		return game.Card{}, false
	}
	return c, true
}

// b26ArtifactCreatureYouControlBecameBlocked is the afflict trigger
// Cyberman Patrol carries for every artifact creature its controller
// controls (itself included — it is one): the creature named by an
// EventBecomesBlocked is an artifact creature the source's controller
// controls.
//
// No dedupe. EventBecomesBlocked is emitted once per blocked attacker
// at the block declaration's lock-in (CR 506.4, #830), so a double
// block is already one event; before that the engine emitted one
// event per BLOCKER and every reader had to walk the log back to the
// attacker's EventAttack to tell the first from the rest.
func b26ArtifactCreatureYouControlBecameBlocked(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBecomesBlocked {
		return false
	}
	attacker, ok := g.LookupCardForEffect(ev.Target)
	return ok && attacker.IsCreature() && attacker.IsArtifact() && attacker.Controller == source.Controller
}

// b26SelfEnteredOrWasSacrificed is "when this enters and when you
// sacrifice it" — Heaped Harvest's one ability with two conditions.
// The sacrifice event fires while the permanent is still on the
// battlefield, so the harvester finds the source.
func b26SelfEnteredOrWasSacrificed(ev game.Event, source *game.Card) bool {
	return (ev.Kind == game.EventETB || ev.Kind == game.EventSacrifice) && ev.CardID == source.InstanceID
}

// b26SelfEnteredUntapped is Dwarven Mine's "when this land enters
// untapped": the source's own ETB, read after its enters-tapped
// replacement has decided how it entered.
func b26SelfEnteredUntapped(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventETB && ev.CardID == source.InstanceID && !source.Tapped
}

// b26EachEndStepAndYouGainedAtLeastThisTurn is Valkyrie Harbinger's
// intervening-if at announce: any player's end step began and the
// source's controller gained at least n life this turn. Re-run at
// resolution by the effect body, as CR 603.4 asks.
func b26EachEndStepAndYouGainedAtLeastThisTurn(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	return b15EndStepBegan(ev) && b15LifeGainedThisTurn(g, source.Controller) >= n
}

// --- board reads -------------------------------------------------

// b26AttackingCreaturesYouControlWithPowerAtLeast counts the
// attacking creatures `controller` controls whose power is at least
// n — The Earth King's "that many", read as the trigger resolves.
func b26AttackingCreaturesYouControlWithPowerAtLeast(g *game.Game, controller uuid.UUID, n int) int {
	count := 0
	for _, id := range b13AttackingCreaturesYouControl(g, controller) {
		if c, ok := g.LookupCardForEffect(id); ok && c.CurrentPower() >= n {
			count++
		}
	}
	return count
}

// b26AttackingCreatures snapshots every attacking creature on the
// battlefield, any controller — Aetherspouts' "each attacking
// creature". Copies, taken before anything moves.
func b26AttackingCreatures(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil {
			out = append(out, c)
		}
	}
	return out
}

// b26PermanentsOfSubtypeControlled counts the permanents `controller`
// controls with the given subtype — Elven Ambush's "for each Elf you
// control". Any permanent, as printed; effective subtypes, so a
// changeling counts.
func b26PermanentsOfSubtypeControlled(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// b26NonbasicLandsControlled counts the nonbasic lands `player`
// controls — Price of Progress's per-player amount. Nonbasic is the
// absence of the Basic SUPERTYPE (b10NonbasicLand), not a type-line
// substring: a Snow-Covered Island is "Basic Snow Land — Island" and
// is basic.
func b26NonbasicLandsControlled(g *game.Game, player uuid.UUID) int {
	nonbasic := b10NonbasicLand()
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && nonbasic(g, player, c) {
			n++
		}
	}
	return n
}

// b26GreatestPowerOfOracleYouControl is the current power of the
// biggest permanent with `oracleID` that `controller` controls —
// "Carmen's power" read from inside a target predicate, where the
// source is not to hand. -1 when they control none, so a mana-value
// comparison against it admits nothing.
func b26GreatestPowerOfOracleYouControl(g *game.Game, controller uuid.UUID, oracleID string) int {
	best := -1
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || c.OracleID != oracleID {
			continue
		}
		if p := c.CurrentPower(); p > best {
			best = p
		}
	}
	return best
}

// b26ControlsCommanderCreatureOwnedBy reports whether `controller`
// controls a creature that is a commander `owner` owns — the
// permanent Inspiring Leader hangs its granted static on. The two
// players differ when the commander has been stolen: the printed
// grant then works for the thief, whose tokens grow, and not for the
// owner. Walks the live slice: it runs inside a layer recompute.
func b26ControlsCommanderCreatureOwnedBy(g *game.Game, controller, owner uuid.UUID) bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.IsCommander && c.Owner == owner && c.Controller == controller && c.IsCreature() {
			return true
		}
	}
	return false
}

// --- target specs ------------------------------------------------

// b26TargetPermanentCardInYourGraveyardWithManaValueAtMostPowerOf is
// Carmen's attack target: "up to one target permanent card with mana
// value less than or equal to Carmen's power from your graveyard".
// The power is Carmen's current power, read when the legal set is
// computed and again at resolution, so a counter placed in response
// widens the set and a shrink narrows it, as CR 608.2b asks.
func b26TargetPermanentCardInYourGraveyardWithManaValueAtMostPowerOf(oracleID string) *game.TargetSpec {
	return TargetCardInGraveyard("up to one target permanent card with mana value less than or equal to Carmen's power from your graveyard",
		YouOwn(), Permanent(),
		func(g *game.Game, caster uuid.UUID, c game.Card) bool {
			return c.ManaValue() <= b26GreatestPowerOfOracleYouControl(g, caster, oracleID)
		},
	).WithCount(0, 1)
}

// --- effect bodies -----------------------------------------------

// b26DamageEachPlayerTwiceTheirNonbasicLands is Price of Progress:
// every seated player, the caster included, takes damage equal to
// twice the nonbasic lands they control. Counts are snapshotted
// before any damage is dealt.
func b26DamageEachPlayerTwiceTheirNonbasicLands(ctx *Context) error {
	type hit struct {
		player uuid.UUID
		amount int
	}
	var hits []hit
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		hits = append(hits, hit{p.ID, 2 * b26NonbasicLandsControlled(ctx.Game, p.ID)})
	}
	for _, h := range hits {
		if err := (DealDamage{Source: ctx.Source(), Target: h.player, Amount: h.amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b26TuckAttackersTopOrBottomByOwnersChoice is Aetherspouts: every
// attacking creature goes onto its owner's library, and the owner
// chooses top or bottom. The choice is asked as a scry — each
// owner's attackers are put on top of their library and the owner
// then scries that many, which is exactly "for each of these cards,
// top or bottom, and the order of the ones on top" (CR 401.4 gives
// the owner the order anyway). The scry looks at nothing new: the
// cards were public permanents a moment ago. Owners are asked in
// seat order; a player with no attacker is not asked.
//
// #783: the scry is the tuck's CONTINUATION and it counts what LANDED.
// A library is a CR 903.9 destination, so an attacking commander stops
// to ask its owner about the command zone — and the old per-card loop
// counted the question rather than the answer. The owner scried a
// library card they had no right to look at, and the commander then
// landed on top of the library, after the scry, never ordered. A
// commander that takes the command zone was never put into a library
// (CR 400.7), so it is not among the cards its owner arranges.
func b26TuckAttackersTopOrBottomByOwnersChoice(ctx *Context) error {
	attackers := b26AttackingCreatures(ctx.Game)
	ids := make([]uuid.UUID, 0, len(attackers))
	for _, c := range attackers {
		ids = append(ids, c.InstanceID)
	}
	item := ctx.Item
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract every continuation in the engine follows: an
	// undo restores this game's fields in place, so a captured *Game
	// would be the wrong one.
	return ctx.Game.TuckCardsToLibraryThenForEffect(ids, game.TuckOptions{}, func(g *game.Game, tucked []uuid.UUID) error {
		ctx := NewContext(g, item)
		byOwner := map[uuid.UUID]int{}
		for _, id := range tucked {
			if c, ok := g.LookupCardForEffect(id); ok {
				byOwner[c.Owner]++
			}
		}
		for _, p := range g.Seats {
			if p == nil || p.Eliminated {
				continue
			}
			if n := byOwner[p.ID]; n > 0 {
				if err := (Scry{Player: p.ID, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// b26ReturnCreatureCardsWithManaValueAtMostFromGraveyard is Raise the
// Past: every creature card with mana value at most n in `player`'s
// graveyard comes back to the battlefield under its owner's control.
// Snapshot then move, so the reanimations don't disturb the walk.
func b26ReturnCreatureCardsWithManaValueAtMostFromGraveyard(ctx *Context, player uuid.UUID, n int) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && c.ManaValue() <= n {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b26ExileFirstLegalTargetOwnerMayCast is Aerial Extortionist's
// exile: the first still-legal target leaves for exile carrying a
// grant that lets its OWNER cast it for as long as it stays there.
// The permission is unbounded ("for as long as that card remains
// exiled") and cast-only, as printed — nothing the ability can
// target is a land, so "cast" and "play" coincide anyway.
func b26ExileFirstLegalTargetOwnerMayCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
		WhileInZone: true,
		CastOnly:    true,
	})
}

// b26SearchBasicsOntoBattlefieldTapped is The Earth King's search:
// up to n basic land cards, onto the battlefield tapped, then
// shuffle. Zero is a shuffle and nothing else, as printed.
func b26SearchBasicsOntoBattlefieldTapped(g *game.Game, item *game.StackItem, n int, reason string) error {
	if n <= 0 {
		return nil
	}
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     IsBasicLand,
		Dest:          game.ZoneBattlefield,
		Limit:         n,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        reason,
	}.Apply(NewContext(g, item))
}

// b26PutCounterOnSelfAndGainLife is Carmen's sacrifice trigger body:
// a +1/+1 counter on the source if it is still on the battlefield,
// and 1 life either way — the life is not conditional on the
// counter landing.
func b26PutCounterOnSelfAndGainLife(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if z := g.FindCardZoneForEffect(item.SourceCardID); z != nil && z.Kind == game.ZoneBattlefield {
		if err := (AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
}

// b26ReturnFirstLegalGraveyardTargetToBattlefield puts the first
// still-legal graveyard target onto the battlefield under its
// owner's control — Carmen's attack trigger, whose "up to one" may
// have been answered with nothing.
func b26ReturnFirstLegalGraveyardTargetToBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}.Apply(ctx)
	}
	return nil
}

// b26SourceOnBattlefieldUntapped reports whether the item's source is
// still on the battlefield and untapped — Genesis Chamber's
// intervening-if, re-run at resolution.
func b26SourceOnBattlefieldUntapped(g *game.Game, item *game.StackItem) bool {
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	return ok && !c.Tapped
}
