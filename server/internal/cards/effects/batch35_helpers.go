package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch35_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 35 (#398, `edhrec_rank` 3650–3752). Own file per
// the #231 convention; every package-level name carries the b35
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "a land you control entered" is
// b33LandYouControlEntered, "an opponent cast a spell" is
// b15OpponentCastSpell, "a creature you control dealt combat damage
// to a player" is combatDamageToPlayerBy, "another creature died" is
// diedCreature, the per-event
// counter delta is b33CountersPlacedDelta, "lands with different
// names" is b04LandNamesControlled, "exile all graveyards" is
// b02ExileAllGraveyards, "each opponent loses N life" is
// eachOpponentLosesLife, "N damage to each opponent" is
// damageToEachOpponent, "draw then discard" is lootOne, the
// graveyard-to-hand body is b34ReturnChosenGraveyardCardToHand, the
// first legal target read is b16FirstLegalTargetCard, the Zombie is
// BlackZombieToken and its tapped entry is createTappedZombie's
// Token(...).EntersTapped().

// --- token templates ---------------------------------------------

// --- predicates --------------------------------------------------

// b35Battle passes for a battle — Final Act's "destroy all battles".
// Not in targets.go because that file is off limits to a batch.
func b35Battle() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsBattle() }
}

// b35TargetAnyOther is Screaming Nemesis's "any other target":
// TargetAny with the label the card prints. The "other" is enforced
// at resolution (Aang, Swift Savior's posture — a target clause is
// built once per card and cannot name the instance it hangs off),
// so a Nemesis chosen as its own target is skipped rather than
// damaged.
func b35TargetAnyOther() *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any other target"
	return spec
}

// --- card reads --------------------------------------------------

// b35CardsInAllGraveyards counts the cards in every seat's graveyard
// — Lord of Extinction's size. Eliminated players' graveyards are
// still graveyards, so they count.
func b35CardsInAllGraveyards(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p != nil && p.Graveyard != nil {
			n += p.Graveyard.Size()
		}
	}
	return n
}

// b35CountersOnArtifactsAndCreaturesYouControl sums every counter of
// every kind on the artifacts and creatures `controller` controls —
// Lux Artillery's thirty. A permanent that is both is counted once.
func b35CountersOnArtifactsAndCreaturesYouControl(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || (!c.IsArtifact() && !c.IsCreature()) {
			continue
		}
		for _, v := range c.Counters {
			if v > 0 {
				n += v
			}
		}
	}
	return n
}

// b35WasAttackingWhenItLeft reports whether `cardID` was an attacking
// creature at the moment of the event at `seq` — Garna's "if it was
// attacking". The battlefield-leave choke point clears
// AttackingTarget before the dies event fires and the LKI
// characteristic carries no combat state, so the answer is read off
// the log: walking back from the event, an EventAttack naming the
// card means it was declared this combat, while reaching a step
// boundary that is not one of the steps an attacker stays in combat
// through (declare attackers, declare blockers, combat damage, end
// of combat — CR 511.3 removes attackers as that last step ENDS)
// means either no attack was declared or the combat it attacked in
// has ended. A creature that attacked and left combat by leaving the
// battlefield came back as a new object with no attack of its own, so
// it reads as not attacking, as printed.
//
// The walk is bounded by g.EventsThisTurn() — combat does not span
// turns, and the slice starts at the real turn boundary (#1009). The
// separate EventBeginUpkeep case it used to carry was redundant once
// bounded: EventStepBegan announces the upkeep too, and "upkeep" is
// not one of the steps above.
func b35WasAttackingWhenItLeft(g *game.Game, cardID uuid.UUID, seq uint64) bool {
	inCombat := map[string]bool{
		string(game.StepDeclareAttackers):  true,
		string(game.StepDeclareBlockers):   true,
		string(game.StepFirstStrikeDamage): true,
		string(game.StepCombatDamage):      true,
		string(game.StepEndCombat):         true,
	}
	turn := g.EventsThisTurn()
	for i := len(turn) - 1; i >= 0; i-- {
		ev := turn[i]
		if ev.Seq > seq {
			continue
		}
		switch ev.Kind {
		case game.EventAttack:
			if ev.CardID == cardID {
				return true
			}
		case game.EventStepBegan:
			if !inCombat[ev.Label] {
				return false
			}
		}
	}
	return false
}

// b35TwoNonlandCardsShareAColor is Sphinx's Tutelage's repeat test:
// exactly the two milled cards, both nonland, with at least one
// colour in common. Colours are the printed ones — Scryfall's
// stamped list, or the mana cost for a fixture — read wherever the
// cards now are.
func b35TwoNonlandCardsShareAColor(g *game.Game, milled []uuid.UUID) bool {
	if len(milled) != 2 {
		return false
	}
	a, okA := g.LookupCardForEffect(milled[0])
	b, okB := g.LookupCardForEffect(milled[1])
	if !okA || !okB || a.IsLand() || b.IsLand() {
		return false
	}
	for _, col := range a.EffectiveColors() {
		if b.HasColor(col) {
			return true
		}
	}
	return false
}

// --- trigger conditions ------------------------------------------

// anotherZombieYouControlDied is Plague Belcher's condition: a
// Zombie the source's controller controlled, other than the source,
// died. The dead card is read post-move, so a changeling counts and a
// Zombie that was one only through a layer effect does not — weaker,
// never stronger (Undead Augur's read).
func anotherZombieYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.HasSubtype("Zombie")
}

// anotherCreatureYouControlDied is Garna's condition: a creature
// the source's controller controlled, other than the source, died.
func anotherCreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller
}

// b35SelfWasDealtDamage is Screaming Nemesis's condition: the source
// itself was dealt damage — combat or not, by anyone's source.
// Damage is marked before the state-based sweep runs, so a Nemesis
// that took lethal is still on the battlefield when its event fires.
func b35SelfWasDealtDamage(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 && ev.Target == source.InstanceID
}

// b35PlusCountersPutOnAnotherNonHydraCreatureYouControl is Wildwood
// Scourge's condition: the event PUT one or more +1/+1 counters (a
// removal is not a placement) on a creature the source's controller
// controls that is not the source and not a Hydra. The engine emits
// one EventCounterPlaced per permanent per placement, so "one or
// more" is the event itself.
func b35PlusCountersPutOnAnotherNonHydraCreatureYouControl(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Target == source.InstanceID || b33CountersPlacedDelta(ev, game.CounterPlusOne, g) <= 0 {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !onBattlefield(g, ev.Target) {
		return false
	}
	return c.IsCreature() && c.Controller == source.Controller && !c.HasSubtype("Hydra")
}

// b35DrawStepBegan is Nekusar's first condition: any player's draw
// step began. The drawer is the event's Actor.
func b35DrawStepBegan(ev game.Event) bool {
	return ev.Kind == game.EventBeginDrawStep && ev.Actor != uuid.Nil
}

// b35OpponentDrewACard is Nekusar's second condition: a player other
// than the source's controller drew a card. Fires once per card.
func b35OpponentDrewACard(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventDrawCard && ev.Actor != uuid.Nil && ev.Actor != source.Controller
}

// b35EndStepAndThirtyCounters is Lux Artillery's intervening-if at
// announce: the source's controller's end step began and their
// artifacts and creatures carry thirty or more counters between
// them. Re-run at resolution by the body, as CR 603.4 asks.
func b35EndStepAndThirtyCounters(ev game.Event, source *game.Card, g *game.Game) bool {
	return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller &&
		b35CountersOnArtifactsAndCreaturesYouControl(g, source.Controller) >= 30
}

// --- effect bodies -----------------------------------------------

// b35DrawTwoThenEachPlayerLosesTwo is Risky Shortcut's body: the
// controller draws two, then every live player — the controller
// first, then the opponents in seat order — loses 2 life. A loss,
// not damage, so no prevention shield sees it.
func b35DrawTwoThenEachPlayerLosesTwo(item *game.StackItem, ctx *Context) error {
	if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range tablePlayers(ctx) {
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -2); err != nil {
			return err
		}
	}
	return nil
}

// b35CounterChosenSpellThenGainLife is Absorb's body: the announced
// spell, if still on the stack, is countered, and the controller
// gains `n` life whether or not the counter took (a spell that can't
// be countered still pays out the life, as printed). With no legal
// target the spell has already fizzled and nothing here runs.
func b35CounterChosenSpellThenGainLife(item *game.StackItem, ctx *Context, n int) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return GainLife{Player: item.Controller, Amount: n}.Apply(ctx)
	}
	return nil
}

// b35BounceChosenPermanents is Aether Gale's body: every announced
// nonland permanent that is still legal returns to its owner's hand,
// all at once through the simultaneous bounce path so an Aura on a
// bounced creature is handled as one event.
func b35BounceChosenPermanents(ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard && onBattlefield(ctx.Game, t.ID) {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	ctx.Game.BounceCardsToHandForEffect(ids)
	return nil
}

// b35PutMinusCountersOnChosen is Plague Belcher's entry body: two
// -1/-1 counters on the announced creature, if it is still legal.
func b35PutMinusCountersOnChosen(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok || !onBattlefield(g, id) {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
	}
}

// b35EachOpponentLosesOne is Plague Belcher's dies body.
func b35EachOpponentLosesOne(g *game.Game, item *game.StackItem) error {
	return eachOpponentLosesLife(g, item, 1)
}

// b35ThatPlayerMillsTwo is Memory Erosion's body: the caster of the
// spell, captured in Build, mills two — or the rest of their library
// when it holds fewer.
func b35ThatPlayerMillsTwo(caster uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if g.PlayerByIDForEffect(caster) == nil {
			return nil
		}
		return MillCards{Player: caster, N: 2}.Apply(NewContext(g, item))
	}
}

// b35TutelageLabel is the stack label of Sphinx's Tutelage's draw
// trigger.
const b35TutelageLabel = "Sphinx's Tutelage — target opponent mills two cards, repeating while two nonland cards share a color"

// b35TutelageMill is Sphinx's Tutelage's body: the announced
// opponent, if still legal, mills two; if both milled cards are
// nonland and share a colour, they mill two more, and so on. Each
// pass is one mill of two, so a mill-watcher sees the passes as
// separate batches — as printed, since each repeat is its own
// instruction. The loop ends when a pass mills fewer than two, when
// the pair fails the test, or when the library is empty.
func b35TutelageMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return b35TutelageMillPass(ctx, t.ID, 0)
	}
	return nil
}

// b35TutelageMillPass is one pass of that repeat, and the repeat is
// recursion through the mill's continuation rather than a loop.
//
// #893: the pair to test is the cards that were PUT INTO THE GRAVEYARD
// this way (CR 400.7), and a commander coming off the top stops to
// answer CR 903.9 — so the pass that decides whether to repeat cannot
// read its own result on the next line. `guard` is carried by value,
// which keeps the bound honest across an undo that rewinds into the
// prompt and replays the answer.
func b35TutelageMillPass(ctx *Context, victim uuid.UUID, guard int) error {
	if guard >= 200 {
		return nil
	}
	return MillToZone{
		Player: victim,
		N:      2,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			if !b35TwoNonlandCardsShareAColor(ctx.Game, milled) {
				return nil
			}
			return b35TutelageMillPass(ctx, victim, guard+1)
		},
	}.Apply(ctx)
}

// b35LootOne is "draw a card, then discard a card" as an activated
// body (Sphinx's Tutelage's {5}{U}).
func b35LootOne(g *game.Game, item *game.StackItem) error {
	return lootOne(g, item, 1)
}

// b35ReturnChosenToBattlefieldWithCounter is Smile at Death's body:
// every announced creature card that is still legal and still in the
// controller's graveyard returns to the battlefield under its
// owner's control ("from YOUR graveyard" — owner and controller
// coincide), and each one that arrived gets a +1/+1 counter. The
// counter lands a beat after the entry rather than as part of it,
// Rakdos Joins Up's posture: the reanimation path carries no counter
// option, and nothing in the catalog reads a creature's counters
// between its arrival and the next event.
func b35ReturnChosenToBattlefieldWithCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var returned []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		if onBattlefield(g, t.ID) {
			returned = append(returned, t.ID)
		}
	}
	for _, id := range returned {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b35NecrobloomLandfall is The Necrobloom's landfall body: a 0/1
// green Plant, or a 2/2 black Zombie instead when the controller
// controls seven or more lands with different names — read at
// resolution, so the land that entered is among them.
func b35NecrobloomLandfall(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	template := TokenCard("0/1 green Plant")
	if b04LandNamesControlled(g, item.Controller) >= 7 {
		template = BlackZombieToken()
	}
	return CreateToken{Controller: item.Controller, Template: template, N: 1}.Apply(ctx)
}

// b35DrawIfAttackingElsePingOpponents is Garna's body: a card if the
// dead creature was attacking (read in Build off the log, at the
// event), otherwise 1 damage from Garna to each opponent.
func b35DrawIfAttackingElsePingOpponents(attacking bool) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if attacking {
			return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
		}
		return damageToEachOpponent(g, item, 1)
	}
}

// b35EachPlayerMillsXThenZombiesPerCreature is Dread Summons's body:
// every live player, the controller first and then the opponents in
// seat order, mills X — or the rest of their library when it holds
// fewer — and the controller creates one tapped 2/2 black Zombie for
// each creature card that landed in a graveyard that way, all the
// Zombies after all the mills as the printed "for each" reads.
func b35EachPlayerMillsXThenZombiesPerCreature(item *game.StackItem, ctx *Context) error {
	return b35DreadSummonsMillStep(ctx, item, tablePlayers(ctx), ctx.X(), 0)
}

// b35DreadSummonsMillStep mills the head of `players` and continues
// with the tail from that mill's continuation; the empty list is the
// base case, where the Zombies are created.
//
// #893: "each creature card put into a graveyard this way" is what
// LANDED in a graveyard (CR 400.7), and any seat's mill can stop to
// ask its owner about CR 903.9, so the tally cannot be read on the line
// after the mill and the seats cannot all be milled on one line either.
// The running count is carried forward BY VALUE, which is what makes an
// undo across the prompt replay identically rather than counting the
// first run's creatures twice.
func b35DreadSummonsMillStep(ctx *Context, item *game.StackItem, players []uuid.UUID, x, creatures int) error {
	if len(players) == 0 {
		if creatures == 0 {
			return nil
		}
		return CreateTokenAdvanced{
			Controller: item.Controller,
			Spec:       Token(BlackZombieToken()).EntersTapped(),
			N:          creatures,
		}.Apply(ctx)
	}
	next, rest := players[0], players[1:]
	return MillToZone{
		Player: next,
		N:      x,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			found := creatures
			for _, cardID := range milled {
				if c, ok := ctx.Game.LookupCardForEffect(cardID); ok && c.IsCreature() {
					found++
				}
			}
			return b35DreadSummonsMillStep(ctx, item, rest, x, found)
		},
	}.Apply(ctx)
}

// b35GainLifeEqualToToughness is Ikra Shidiqi's declared Effect: life
// equal to the dealing creature's toughness, read live at resolution
// when it is still on the battlefield (so a pump in response counts)
// and otherwise the toughness it had when the damage was dealt — the
// fallback, carried on item.Params.Amount by a fill-in Build, along
// with the dealer's ID on item.Params.Object — Righteous Valkyrie's
// fallback. Only Ikra Shidiqi uses this body, so it is declared as
// the row's Effect directly rather than as a per-instance closure
// (ADR 0041 P9).
func b35GainLifeEqualToToughness(g *game.Game, item *game.StackItem) error {
	dealer := item.Params.Object.ID
	toughness := item.Params.Amount
	if onBattlefield(g, dealer) {
		if c, ok := g.LookupCardForEffect(dealer); ok {
			toughness = c.CurrentToughness()
		}
	}
	if toughness <= 0 {
		return nil
	}
	return GainLife{Player: item.Controller, Amount: toughness}.Apply(NewContext(g, item))
}

// b35RedirectDamageToChosen is Screaming Nemesis's body: the Nemesis
// deals the event's amount, stamped onto item.Params.Amount by a
// fill-in Build, to the announced target if it is still legal and is
// not the Nemesis itself. The SOURCE is the Nemesis, as printed, and
// a Nemesis that died to the damage still deals it — the damage path
// does not need the source on the battlefield.
func b35RedirectDamageToChosen(g *game.Game, item *game.StackItem) error {
	amount := item.Params.Amount
	if amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 || (ts[0].Kind == game.TargetCard && ts[0].ID == item.SourceCardID) {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: amount}.Apply(ctx)
}

// b35PutCounterOnSelf is Wildwood Scourge's body: one +1/+1 counter
// on the Scourge, if it is still on the battlefield.
func b35PutCounterOnSelf(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}

// b35ThatPlayerDrawsOne is Nekusar's draw-step body: the player
// whose draw step it is, captured in Build, draws an additional
// card.
func b35ThatPlayerDrawsOne(drawer uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if g.PlayerByIDForEffect(drawer) == nil {
			return nil
		}
		return DrawCards{Player: drawer, N: 1}.Apply(NewContext(g, item))
	}
}

// b35DamageThatPlayer is Nekusar's draw body: `n` damage from
// Nekusar to the player who drew, captured in Build.
func b35DamageThatPlayer(victim uuid.UUID, n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if g.PlayerByIDForEffect(victim) == nil {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: victim, Amount: n}.Apply(NewContext(g, item))
	}
}

// b35EachOpponentLosesAllCounters is Final Act's fifth mode: every
// counter of every kind on every opponent — poison, energy,
// experience, rad, anything — comes off. One removal per kind per
// opponent through the effect-side adder at a negative delta, which
// since ADR 0056 goes through the CR 614 counter window like any other
// player-counter change — harmlessly, because a removal is not a
// placement and every counter replacement in the catalog guards on a
// positive delta (CR 614.1). What it does buy is the event and the
// layer bump, so a static reading an opponent's poison count sees the
// counters leave.
func b35EachOpponentLosesAllCounters(ctx *Context) error {
	for _, id := range ctx.Opponents() {
		p := ctx.PlayerByID(id)
		if p == nil {
			continue
		}
		kinds := make([]string, 0, len(p.Counters))
		for kind := range p.Counters {
			kinds = append(kinds, kind)
		}
		for _, kind := range kinds {
			if n := p.Counters[kind]; n > 0 {
				if err := ctx.Game.AddPlayerCounterForEffect(id, kind, -n); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// b35FinalAct is Final Act's body: the chosen modes in printed order
// (CR 608.2c) — destroy all creatures, destroy all planeswalkers,
// destroy all battles, exile all graveyards, each opponent loses all
// counters. The three sweeps go through the mass destroy, so an
// indestructible permanent survives its mode.
func b35FinalAct(item *game.StackItem, ctx *Context) error {
	sweeps := []CardPredicate{Creature(), Planeswalker(), b35Battle()}
	for i, match := range sweeps {
		if !ctx.HasMode(i) {
			continue
		}
		if err := (DestroyAllMatching{Match: match}).Apply(ctx); err != nil {
			return err
		}
	}
	if ctx.HasMode(3) {
		if err := b02ExileAllGraveyards(ctx.Game, item); err != nil {
			return err
		}
	}
	if ctx.HasMode(4) {
		return b35EachOpponentLosesAllCounters(ctx)
	}
	return nil
}

// b35ReturnChosenGraveyardCardToHand is Greenwarden of Murasa's entry
// body: the announced card, if it is still in the controller's
// graveyard, returns to their hand.
func b35ReturnChosenGraveyardCardToHand(g *game.Game, item *game.StackItem) error {
	return b34ReturnChosenGraveyardCardToHand(NewContext(g, item))
}

// b35ExileSelfFromGraveyardThenReturnChosen is Greenwarden's dies
// body: the Greenwarden, if it is still in a graveyard, is exiled
// ("you may exile it" — the controller already said yes to the
// trigger), and if it was, the announced card returns to hand if it
// is still in the controller's graveyard. A Greenwarden that has
// since left the graveyard — reanimated in response, or a commander
// sent to the command zone — cannot be exiled, so "if you do" fails
// and nothing returns. A Greenwarden chosen as its own target is
// exiled first and is then no longer in the graveyard, so nothing
// returns either — as printed.
//
// #870: the "if you do" reads the exile's CONTINUATION rather than the
// exile zone on the next line. The Greenwarden is a commander often
// enough to matter, and a graveyard is a CR 903.9 zone: the old
// read-back ran while the owner's prompt was still open, found the
// card not in exile and returned nothing — and then the owner said
// "exile it" and the card sat in exile with the second half of its own
// trigger already skipped.
func b35ExileSelfFromGraveyardThenReturnChosen(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ExileTarget{
		Target: item.SourceCardID,
		Then: func(ctx *Context, exiled bool) error {
			if !exiled {
				return nil
			}
			return b34ReturnChosenGraveyardCardToHand(ctx)
		},
	}.Apply(ctx)
}

// b35TenDamageToEachOpponentIfThirtyCounters is Lux Artillery's body:
// the intervening if re-checked at resolution (CR 603.4), then 10
// damage from the Artillery to each opponent.
func b35TenDamageToEachOpponentIfThirtyCounters(g *game.Game, item *game.StackItem) error {
	if b35CountersOnArtifactsAndCreaturesYouControl(g, item.Controller) < 30 {
		return nil
	}
	return damageToEachOpponent(g, item, 10)
}
