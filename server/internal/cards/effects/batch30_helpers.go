package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch30_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 30 (#393, `edhrec_rank` 3145–3244). Own file per
// the #231 convention; every package-level name carries the b30
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "you gained life
// this turn" is b15LifeGainedThisTurn and "you lost life this turn"
// b18LifeLostThisTurn (b24YourEndStepAndYouGainedLifeThisTurn is the
// end-step gate on the first), "if you cast it" is
// b16EnteredFromStack, the whole-hand discard is discardWholeHand, a
// graveyard's exile is exileGraveyardForEffect, the fight is
// b10Fight, "sacrifice a Goblin" is b25SacrificeAGoblin, "each
// opponent loses N" is eachOpponentLosesLife, "untap each X you
// control" is b16UntapAllYouControlMatching, "another permanent you
// control enters" is enteredUnderYourControl, the discard trigger
// reads are discardedByYou, the one-or-more dedup is
// OncePerBatch, the Desert read is b02IsDesert, the
// Plant is b02PlantToken, the Zombie is BlackZombieToken, the
// fetch-a-basic-tapped body is the shape of fetchBasicTapped, the
// lord builders are TribalAnthem / TribalKeywordGrant, and the
// first-legal-target read is b16FirstLegalTargetCard.

// --- tokens ------------------------------------------------------

// --- costs -------------------------------------------------------

// b30SacrificeAnArtifactOrCreature is Dockside Chef's "Sacrifice an
// artifact or creature" — any artifact or creature the activator
// controls, the Chef itself included (it is a creature).
func b30SacrificeAnArtifactOrCreature() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact or creature", Or(Artifact(), Creature()))}
}

// --- card reads --------------------------------------------------

// b30IsBasicLandCard is "a basic land card" read off the Basic
// SUPERTYPE rather than the "basic land" type-line substring
// IsBasicLand keys on — a Snow-Covered Forest is "Basic Snow Land —
// Forest" and is basic. Off the battlefield the effective supertypes
// are the printed ones.
func b30IsBasicLandCard(c game.Card) bool {
	if !c.IsLand() {
		return false
	}
	for _, s := range c.Effective().Supertypes {
		if s == "Basic" {
			return true
		}
	}
	return false
}

// b30SquirrelOrFood reports whether a permanent is a Squirrel or a
// Food — Honored Dreyleader's count and trigger. Effective subtypes,
// so a changeling is a Squirrel.
func b30SquirrelOrFood(c game.Card) bool {
	return c.HasSubtype("Squirrel") || c.HasSubtype("Food")
}

// b30OtherSquirrelsAndFoodControlled counts the Squirrels and Foods
// `controller` controls other than `self` — Honored Dreyleader's
// entry count.
func b30OtherSquirrelsAndFoodControlled(g *game.Game, controller, self uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.Controller == controller && b30SquirrelOrFood(c) {
			n++
		}
	}
	return n
}

// b30TotalToughnessControlled is the total toughness of the
// creatures `controller` controls — Betor's three thresholds.
// Current toughness (layers plus counters), so an anthem and a
// -1/-1 counter both count; a negative toughness contributes as
// itself, since the sum is what the card asks for.
func b30TotalToughnessControlled(g *game.Game, controller uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			total += c.CurrentToughness()
		}
	}
	return total
}

// b30DesertsControlled counts the Deserts `controller` controls —
// Hour of Promise's "three or more Deserts", read after the fetched
// lands have entered.
func b30DesertsControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && b02IsDesert(g, controller, c) {
			n++
		}
	}
	return n
}

// b30CastFromHand is "if you cast it from your hand" for an
// entering permanent (Wakening Sun's Avatar): the card's most recent
// battlefield entry came from the stack (b16EnteredFromStack), and
// the cast that put it there left the caster's hand — the cast
// event records its origin zone. A cast from the command zone, a
// graveyard or exile is a cast, but not from hand.
func b30CastFromHand(g *game.Game, cardID uuid.UUID) bool {
	if !b16EnteredFromStack(g, cardID) {
		return false
	}
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventCast && ev.CardID == cardID {
			return ev.OldZone == game.ZoneHand
		}
	}
	return false
}

// --- statics -----------------------------------------------------

// b30CreaturesYouControlWithFlying is Favorable Winds' "creatures you
// control with flying". Read in layer 7c, so the flying is whatever
// layer 6 left on the creature — printed, a lord's grant or an
// until-end-of-turn grant all count, as printed.
func b30CreaturesYouControlWithFlying(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsCreature() && game.HasKeyword(target, "flying")
}

// --- trigger conditions ------------------------------------------

// b30AnotherSquirrelOrFoodYouControlEntered is Honored Dreyleader's
// second trigger: a Squirrel or Food other than the source entered
// under the source's controller's control. A Food token counts — a
// token's entry emits the same EventETB a card's does.
func b30AnotherSquirrelOrFoodYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && b30SquirrelOrFood(c)
}

// b30SelfOrAnotherEnchantmentYouControlEntered is constellation —
// "whenever this creature or another enchantment you control
// enters" (Grim Guardian). Eidolon of Blossoms' shape: the source's
// own entry, or another enchantment entering under the same
// control.
func b30SelfOrAnotherEnchantmentYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind == game.EventETB && ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsEnchantment()
}

// b30NontokenCreatureYouControlEnteredUncast is Satoru's condition:
// the source itself or another nontoken creature entered under the
// source's controller's control, and it was NOT cast — it arrived
// from somewhere other than the stack (a reanimation, a flicker, a
// "put onto the battlefield"). "One or more" is the caller's dedup.
//
// The printed "or no mana was spent to cast them" half is not read:
// the engine records a mana spend only under strict-mode casts, so
// a permissive-mode cast paid on paper and a genuinely free cast
// look the same from the log, and counting both would ship the
// card stronger. Declared on the card.
func b30NontokenCreatureYouControlEnteredUncast(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	if !ok || !c.IsCreature() || IsToken(c) {
		return false
	}
	return !b16EnteredFromStack(g, c.InstanceID)
}

// b30YourEndStepAndYouGainedAndLostLifeThisTurn is Lunar
// Convocation's second intervening-if at announce: the source's
// controller's end step began and they both gained and lost life
// this turn. Paying life is losing life (CR 119.4), and the ability
// cost emits the same life-change event a drain does, so the card's
// own "Pay 2 life" counts, as printed. Re-run at resolution by the
// effect body.
func b30YourEndStepAndYouGainedAndLostLifeThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	return b24YourEndStepAndYouGainedLifeThisTurn(ev, source, g) && b18LifeLostThisTurn(g, source.Controller) > 0
}

// b30YouDiscardedCardWhere reports whether the source's controller
// discarded a card that passes `match` — Surly Badgersaur's three
// discard triggers, one predicate each. The card is read from where
// it landed (the graveyard, or exile under a Library of Leng), where
// its printed type line is intact.
func b30YouDiscardedCardWhere(ev game.Event, source *game.Card, g *game.Game, match func(game.Card) bool) bool {
	if !discardedByYou(ev, source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && match(c)
}

// b30NonHumanCreatureYouControlDealtCombatDamageToPlayer is Keeper
// of Fables' condition: combat damage to a player by a creature the
// source's controller controls that is not a Human. Effective
// subtypes, so a changeling is a Human and does not count.
func b30NonHumanCreatureYouControlDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && !c.HasSubtype("Human")
}

// b30EndStepAndTotalToughnessAtLeast is Betor's intervening-if at
// announce: the source's controller's end step began and their
// creatures' total toughness meets `n`. Re-run at resolution by the
// effect body, as CR 603.4 asks.
func b30EndStepAndTotalToughnessAtLeast(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller &&
		b30TotalToughnessControlled(g, source.Controller) >= n
}

// --- effect bodies -----------------------------------------------

// b30ExileTargetGraveyards exiles the graveyard of every announced
// player target that is still legal — "exile any number of target
// players' graveyards" (Stonespeaker Crystal, Thraben Charm). With
// no targets named it exiles nothing, which is what "any number"
// allows.
func b30ExileTargetGraveyards(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := exileGraveyardForEffect(ctx.Game, ctx.Item, t.ID); err != nil {
			return err
		}
	}
	return nil
}

// b30ExileTargetGraveyardsThenDraw is Stonespeaker Crystal's
// activated body: the graveyards, then a card, in printed order.
func b30ExileTargetGraveyardsThenDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := b30ExileTargetGraveyards(ctx); err != nil {
		return err
	}
	return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
}

// b30DoubleEachCounterKindOn puts as many counters of each kind on
// `target` as it already has — Vorel of the Hull Clade. Every kind
// is snapshotted before the first placement, so a doubled kind is
// not read again mid-loop, and each placement runs through
// AddCounter so a counter doubler applies to the doubling, as
// printed. A target that has left the battlefield is left alone.
func b30DoubleEachCounterKindOn(ctx *Context, target uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(target); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	kinds := make(map[string]int, len(c.Counters))
	for kind, n := range c.Counters {
		if n > 0 {
			kinds[kind] = n
		}
	}
	for kind, n := range kinds {
		if err := (AddCounter{Target: target, Kind: kind, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b30DoubleCountersOnFirstLegalTarget is Vorel's activated body:
// the announced target, if it is still legal, has each kind of
// counter on it doubled.
func b30DoubleCountersOnFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return b30DoubleEachCounterKindOn(ctx, id)
}

// b30PutCounterOnEachArtifactCreatureOrVehicleYouControl is Iron
// Spider's tap ability: a +1/+1 counter on every artifact creature
// and every Vehicle the controller controls, the Spider itself
// included. The set is snapshotted first so a counter payoff that
// makes an artifact mid-loop is not counted.
func b30PutCounterOnEachArtifactCreatureOrVehicleYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != item.Controller {
			continue
		}
		if (c.IsArtifact() && c.IsCreature()) || c.HasSubtype("Vehicle") {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b30DestroyOnePerController is Windgrace's Judgment: of the
// announced targets still legal, the first named for each
// controller is destroyed and any later one under the same
// controller is skipped — "for any number of opponents, destroy
// target nonland permanent THAT PLAYER controls" enforced at
// resolution, since the target clause cannot say "one per
// opponent" at announce. Controllers are read as the spell
// resolves.
func b30DestroyOnePerController(ctx *Context) error {
	var ids []uuid.UUID
	seen := map[uuid.UUID]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok || seen[c.Controller] {
			continue
		}
		seen[c.Controller] = true
		ids = append(ids, t.ID)
	}
	for _, id := range ids {
		if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b30SourceFightsFirstLegalTarget is "this creature fights up to one
// target creature you don't control" (Surly Badgersaur): the source,
// if it is still on the battlefield, fights the announced target if
// one was chosen and it is still legal.
func b30SourceFightsFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return b10Fight(ctx, item.SourceCardID, id)
}

// b30PutCounterOnSelf puts one +1/+1 counter on the item's source if
// it is still on the battlefield — Honored Dreyleader's second
// trigger, Surly Badgersaur's first.
func b30PutCounterOnSelf(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
}

// b30CountersForOtherSquirrelsAndFood is Honored Dreyleader's entry
// body: one +1/+1 counter per other Squirrel or Food the controller
// controls, counted as the trigger resolves.
func b30CountersForOtherSquirrelsAndFood(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	n := b30OtherSquirrelsAndFoodControlled(g, item.Controller, item.SourceCardID)
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: n}.Apply(NewContext(g, item))
}

// b30EachOpponentLosesHalfTheirLife is Betor's last clause: each
// opponent loses half their life, rounded up. An opponent at 0 or
// less loses nothing more.
func b30EachOpponentLosesHalfTheirLife(ctx *Context) error {
	for _, opp := range ctx.Opponents() {
		p := ctx.PlayerByID(opp)
		if p == nil || p.Life <= 0 {
			continue
		}
		if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), opp, -((p.Life + 1) / 2)); err != nil {
			return err
		}
	}
	return nil
}

// b30BetorEndStep is Betor, Kin to All's end-step body: the three
// thresholds in printed order, each re-read as its clause resolves
// (the untap and the draw cannot change toughness, but a counter
// payoff on the draw could in principle) — draw at 10, untap each
// creature you control at 20, halve each opponent's life at 40.
// The whole trigger was gated on 10 at announce; below it now, the
// intervening-if fails and nothing happens.
func b30BetorEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if b30TotalToughnessControlled(g, item.Controller) < 10 {
		return nil
	}
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if b30TotalToughnessControlled(g, item.Controller) < 20 {
		return nil
	}
	if err := b16UntapAllYouControlMatching(ctx, item.Controller, func(c game.Card) bool { return c.IsCreature() }); err != nil {
		return err
	}
	if b30TotalToughnessControlled(g, item.Controller) < 40 {
		return nil
	}
	return b30EachOpponentLosesHalfTheirLife(ctx)
}

// b30LoseOneIfYouGainedLifeThisTurn is Lunar Convocation's first
// end-step body: the intervening-if re-checked at resolution, then
// each opponent loses 1.
func b30LoseOneIfYouGainedLifeThisTurn(g *game.Game, item *game.StackItem) error {
	if b15LifeGainedThisTurn(g, item.Controller) <= 0 {
		return nil
	}
	return eachOpponentLosesLife(g, item, 1)
}

// b30BatIfYouGainedAndLostLifeThisTurn is Lunar Convocation's second
// end-step body: the intervening-if re-checked at resolution, then a
// Bat.
func b30BatIfYouGainedAndLostLifeThisTurn(g *game.Game, item *game.StackItem) error {
	if b15LifeGainedThisTurn(g, item.Controller) <= 0 || b18LifeLostThisTurn(g, item.Controller) <= 0 {
		return nil
	}
	return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 black Bat with flying"), N: 1}.Apply(NewContext(g, item))
}

// b30FetchBasicTapped is Promising Vein's search: a basic land card
// (by supertype, so a snow basic qualifies), onto the battlefield
// tapped, then shuffle.
func b30FetchBasicTapped(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     b30IsBasicLandCard,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Promising Vein — choose a basic land to put onto the battlefield tapped",
	}.Apply(NewContext(g, item))
}

// b30ZombiesIfThreeDeserts is Hour of Promise's rider, run from the
// search's continuation once the fetched lands have entered: two 2/2
// black Zombies if `controller` controls three or more Deserts.
// Takes IDs only — the continuation outlives the resolution that
// started it.
func b30ZombiesIfThreeDeserts(g *game.Game, controller uuid.UUID) error {
	if b30DesertsControlled(g, controller) < 3 {
		return nil
	}
	return g.CreateTokenForEffect(controller, BlackZombieToken(), 2)
}

// b30SearchUpToTwoLandsTappedThenZombies is Hour of Promise's body:
// up to two land cards onto the battlefield tapped, then shuffle,
// then the Desert check — in the continuation, because the search
// may be waiting on the searcher's pick when this returns.
func b30SearchUpToTwoLandsTappedThenZombies(item *game.StackItem, ctx *Context) error {
	controller := item.Controller
	return SearchLibrary{
		Player:        controller,
		Predicate:     func(c game.Card) bool { return c.IsLand() },
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Hour of Promise — choose up to two lands to put onto the battlefield tapped",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return b30ZombiesIfThreeDeserts(g, controller)
		},
	}.Apply(ctx)
}
