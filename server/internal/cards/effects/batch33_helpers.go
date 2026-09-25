package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch33_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 33 (#396, `edhrec_rank` 3447–3548). Own file per
// the #231 convention; every package-level name carries the b33
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "enters or attacks" is b21SelfEnteredOrAttacked,
// "another creature you control enters" is
// b13AnotherCreatureYouControlEntered, "a land you control entered"
// is enteredUnderYourControl + IsLand, "you gained life this turn" is
// b15LifeGainedThisTurn, "your first spell on an opponent's turn" is
// b22FirstSpellOnAnOpponentsTurn, "a noncreature spell you cast" is
// b10NoncreatureSpellCastByYou, "an opponent's creature died" is
// diedCreature, "this died or was exiled from the battlefield" is
// b22SelfDiedOrWasExiledFromBattlefield and its tuck
// b22TuckThirdFromTop, the per-label "one or more" dedup is
// OncePerBatch and its pick-prompt third leg
// OncePerBatch, "which player is resolving" is
// b11ResolvingController, the last-known power of a creature that is
// no longer on the battlefield is b17LastKnownPowerOffBattlefield,
// the milled-creature batch is b17MilledCreatureCards, the first
// legal target read is b16FirstLegalTargetCard and its destroy body
// destroyFirstLegalTarget, "exile target players' graveyards, then
// draw" is b30ExileTargetGraveyardsThenDraw, the counter-on-each body is
// b13PutCounterOnEach, the delayed-trigger token cursor is
// b25LastEventSeq + b27TokensCreatedByAfter, "enters tapped unless
// you control a <type>" is SelfEntersTappedUnless(youControlLandTyped),
// the basic-land test is b30IsBasicLandCard, the Zombie is
// BlackZombieToken, the Faerie Rogue is FaerieRogueToken, the red
// Warrior is b21RedWarriorToken and the tapped-token path is
// Token(...).EntersTapped().

// --- token templates ---------------------------------------------

// b33BlueBirdVigilanceToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func b33BlueBirdVigilanceToken() game.Card {
	return TokenCard("1/1 blue Bird with flying and vigilance")
}

// b33WhiteHumanWarriorToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func b33WhiteHumanWarriorToken() game.Card { return TokenCard("1/1 white Human Warrior") }

// b33TappedAttackingRedWarrior is Dalkovan Encampment's 1/1 red
// Warrior, stamped tapped so CreateTokensAttackingForEffect — which
// copies the template and sets only the attack — puts it in tapped
// and attacking (b21TappedAttackingGoblin's shape).
func b33TappedAttackingRedWarrior() game.Card {
	tmpl := TokenCard("1/1 red Warrior")
	tmpl.Tapped = true
	return tmpl
}

// --- card reads --------------------------------------------------

// b33AnyGraveyardHasAtLeast reports whether some player's graveyard
// holds `n` or more cards — Visions of Beyond's threshold, read over
// every seat's graveyard, eliminated players included (their cards
// are still in a graveyard).
func b33AnyGraveyardHasAtLeast(g *game.Game, n int) bool {
	for _, p := range g.Seats {
		if p != nil && p.Graveyard != nil && p.Graveyard.Size() >= n {
			return true
		}
	}
	return false
}

// b33IslandsControlled counts the Islands `controller` controls —
// Flow of Knowledge's "for each Island you control". Effective
// subtypes, so an Urborg-style type grant counts and so does a
// Prismatic Omen.
func b33IslandsControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && IsLandWithSubtype("Island")(c) {
			n++
		}
	}
	return n
}

// b33ResolutionsThisTurn counts how many times an ability of `source`
// labelled `label` has resolved this turn — the per-turn tally's
// per-object resolution count (Turn.Round is no use here: it counts
// rounds). Dalkovan Encampment's "whenever you
// attack THIS TURN" is created once per activation, so the count is
// how many copies of the delayed trigger exist.
func b33ResolutionsThisTurn(g *game.Game, source uuid.UUID, label string) int {
	return g.ResolvedThisTurn(source, label)
}

// b33CountersPlacedDelta is how many `kind` counters the
// EventCounterPlaced `ev` PUT on its target, or zero for a removal.
// b11CountersWerePlaced's walk with the delta kept: the event carries
// only the post-change total, so the previous total is the most
// recent EventCounterPlaced for the same card and kind, or zero if
// the card arrived on the battlefield more recently than that (CR
// 400.7 — it came with no counters).
func b33CountersPlacedDelta(ev game.Event, kind string, g *game.Game) int {
	if ev.Kind != game.EventCounterPlaced || ev.Label != kind {
		return 0
	}
	before := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		if (prev.Kind == game.EventETB || prev.Kind == game.EventTokenCreated) && prev.CardID == ev.Target {
			break
		}
		if prev.Kind == game.EventCounterPlaced && prev.Target == ev.Target && prev.Label == kind {
			before = prev.Amount
			break
		}
	}
	if ev.Amount <= before {
		return 0
	}
	return ev.Amount - before
}

// b33PlayersDealtCombatDamageThisTurnByYourFaeries is the set of
// players a Faerie under `controller`'s control dealt combat damage
// to this turn — b32PlayersDealtCombatDamageThisTurnByYourCreatureNamed
// with a subtype where that reads a name. Alela's goad clause narrows
// "target creature THAT PLAYER controls" through it, because a target
// predicate is not handed the trigger's event.
//
// The per-turn tally's damage-source cell (#1009), and this is the
// reader #596 was about: the event-log scan looked the dealing Faerie
// up wherever it had since landed, and a Faerie TOKEN that traded in
// combat is nowhere — CR 704.5d takes it out of the graveyard at the
// next state-based check, one check after the damage that killed it.
// The player it hit dropped silently out of the set, which is the
// commonest case Alela is printed for, since she makes the Faeries
// herself. The tally records the subtype AT THE DAMAGE, while the
// Faerie is still on the battlefield, so a traded token counts and a
// changeling counts.
func b33PlayersDealtCombatDamageThisTurnByYourFaeries(g *game.Game, controller uuid.UUID) map[uuid.UUID]bool {
	return g.PlayersDealtCombatDamageThisTurnBySubtype(controller, "Faerie")
}

// b33CreatureOfPlayerHitByYourFaeries is Alela's target predicate: a
// creature controlled by a player one of the caster's Faeries dealt
// combat damage to this turn. The trigger's body re-checks the
// specific damaged player, so with Faeries connecting with two
// players in one combat each trigger's pick is still validated
// against its own victim.
func b33CreatureOfPlayerHitByYourFaeries(g *game.Game, caster uuid.UUID, c game.Card) bool {
	if !c.IsCreature() {
		return false
	}
	return b33PlayersDealtCombatDamageThisTurnByYourFaeries(g, caster)[c.Controller]
}

// --- trigger conditions ------------------------------------------

// b33FaerieYouControlDealtCombatDamageToPlayer is Alela's condition
// before the "one or more" dedup: a creature the source's controller
// controls dealt combat damage to a player, and it is a Faerie.
func b33FaerieYouControlDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.HasSubtype("Faerie")
}

// b33SmallCreatureYouControlAttacked is Raid Bombardment's condition:
// a creature the source's controller controls was declared as an
// attacker and its power, read as it is declared, is 2 or less.
func b33SmallCreatureYouControlAttacked(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.CurrentPower() <= 2
}

// b33BirdYouControlAttacked is Hermes' condition before the "one or
// more" dedup: a Bird the source's controller controls was declared
// as an attacker. Effective subtypes, so a changeling counts.
func b33BirdYouControlAttacked(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.HasSubtype("Bird")
}

// b33LandYouControlEntered is Maja's landfall: a land entered under
// the source's controller's control.
func b33LandYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsLand()
}

// b33AnyCreatureEntered is Blasting Station's "whenever a creature
// enters" — any player's, token or card.
func b33AnyCreatureEntered(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && onBattlefield(g, ev.CardID)
}

// b33LegendaryCreatureYouControlDied is Rakdos Joins Up's condition:
// a legendary creature the source's controller controlled died. The
// dead card is read post-move, so the supertype is its printed one.
func b33LegendaryCreatureYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && isLegendary(&dead)
}

// b33OpponentsCreatureDied is "whenever a creature an opponent
// controls dies" (Patron of the Vein) — the dead creature, read
// post-move, was controlled by someone other than the source's
// controller.
func b33OpponentsCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller != source.Controller
}

// b33OpponentsNontokenCreatureDied is Overseer of the Damned's
// condition: b33OpponentsCreatureDied narrowed to a nontoken creature.
func b33OpponentsNontokenCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller != source.Controller && !IsToken(dead)
}

// b33YouPutMinusCountersOnACreature is Nest of Scarabs' condition:
// the event placed one or more -1/-1 counters on a creature, and the
// player who PUT them is the source's controller —
// b11CounterPlacerOf's reading of "whenever YOU put", which is the
// placement's own actor when it named one and the resolving player
// otherwise. The count of counters placed rides the event's delta,
// read again by the body.
//
// ADR 0056 makes this load-bearing: -1/-1 counters from a wither or
// infect source are put by the SOURCE'S CONTROLLER (CR 120.3d), and
// combat damage happens inside no resolution at all, so the old
// fallback credited whoever had resolved anything most recently.
func b33YouPutMinusCountersOnACreature(ev game.Event, source *game.Card, g *game.Game) bool {
	if b33CountersPlacedDelta(ev, game.CounterMinusOne, g) <= 0 {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !c.IsCreature() || !onBattlefield(g, ev.Target) {
		return false
	}
	return b11CounterPlacerOf(ev, g) == source.Controller
}

// b33CreatureCardMilledIntoYourGraveyard is Sidisi's condition before
// the "one or more" dedup: a creature card the source's controller
// owns was milled into their graveyard (Colossal Grave-Reaver's
// read). The engine emits one EventMill per card, so the first
// creature card of a mill fires the trigger and the rest are
// declined as later events of the same batch (OncePerBatch; see
// AGENTS.md §7).
func b33CreatureCardMilledIntoYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventMill || ev.Actor != source.Controller || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Owner == source.Controller
}

// b33EndStepAndYouGainedLifeThisTurn is Lathiel's intervening-if at
// announce: any player's end step began and the source's controller
// gained life this turn. Re-run at resolution by the body, as CR
// 603.4 asks.
func b33EndStepAndYouGainedLifeThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	return ev.Kind == game.EventBeginEndStep && b15LifeGainedThisTurn(g, source.Controller) > 0
}

// --- effect bodies -----------------------------------------------

// b33DrawOneOrThreeIfAGraveyardIsFull is Visions of Beyond's body:
// one card, or three when any graveyard holds twenty or more cards,
// read as the spell resolves.
func b33DrawOneOrThreeIfAGraveyardIsFull(item *game.StackItem, ctx *Context) error {
	n := 1
	if b33AnyGraveyardHasAtLeast(ctx.Game, 20) {
		n = 3
	}
	return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
}

// b33EachPlayerDrawsThenDiscardsAtRandom is Burning Inquiry's body:
// every player, in APNAP order, draws `n` cards and then discards
// `n` at random — every draw before any discard, as the printed
// "then" reads. Eliminated players are skipped.
func b33EachPlayerDrawsThenDiscardsAtRandom(ctx *Context, n int) error {
	players := tablePlayers(ctx)
	for _, id := range players {
		if err := (DrawCards{Player: id, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range players {
		if err := (DiscardCards{Player: id, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b33DrawPerIslandThenDiscardTwo is Flow of Knowledge's body: one
// card per Island the controller controls, then a two-card discard
// of the controller's choice — the discard prompt opens after the
// draws land, so the drawn cards are among the choices.
func b33DrawPerIslandThenDiscardTwo(item *game.StackItem, ctx *Context) error {
	if n := b33IslandsControlled(ctx.Game, item.Controller); n > 0 {
		if err := (DrawCards{Player: item.Controller, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player: item.Controller,
		Source: item.SourceCardID,
		N:      2,
	})
	return nil
}

// b33CounterTargetThenControllerLosesLife is Countersquall's body:
// the announced spell, if still on the stack, is countered and its
// controller — read before the counter moves it — loses `n` life. A
// loss, not damage, so no prevention shield sees it.
func b33CounterTargetThenControllerLosesLife(item *game.StackItem, ctx *Context, n int) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		spell := ctx.Game.StackItemForEffect(t.ID)
		if spell == nil {
			return nil
		}
		controller := spell.Controller
		if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, controller, -n)
	}
	return nil
}

// b33GainLifeThenDraw is Cloudblazer's body: gain `life`, then draw
// `draw`, in printed order.
func b33GainLifeThenDraw(life, draw int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (GainLife{Player: item.Controller, Amount: life}).Apply(ctx); err != nil {
			return err
		}
		return DrawCards{Player: item.Controller, N: draw}.Apply(ctx)
	}
}

// b33SacrificeLandsThenSearchBasicsTapped is Planar Engineering's
// body: two sacrifice prompts over the controller's lands (their own
// pick, one land per prompt — a player with one land sacrifices it
// and the second prompt is withdrawn), THEN a search for up to `n`
// basic land cards put onto the battlefield tapped, then a shuffle.
//
// The "then" is the printed one since #1019: the two prompts are one
// run and the search is its continuation, so the basics arrive after
// the lands have gone. The search used to be queued alongside the
// sacrifice prompts because the sacrifice prompt had no continuation
// to hang it on, which put a land-count trigger and a landfall watcher
// in the wrong order relative to each other.
func b33SacrificeLandsThenSearchBasicsTapped(item *game.StackItem, ctx *Context, lands, n int) error {
	return ctx.Game.PlayerSacrificesThenForEffect(
		item.SourceCardID, item.Controller,
		sacrificeSpec("a land", Land()),
		"Planar Engineering — sacrifice a land",
		lands,
		func(g *game.Game, _ game.PromptedSacrifices) error {
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     b30IsBasicLandCard,
				Dest:          game.ZoneBattlefield,
				Limit:         n,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Planar Engineering — choose up to four basic land cards to put onto the battlefield tapped",
			}.Apply(NewContext(g, item))
		})
}

// b33DoubleUnspentMana is Doubling Cube's ProducedFunc: one slot of
// the same colour for every token in the controller's pool, read
// AFTER the {3} was paid (CR 601.2h — the total cost is paid before
// the ability is considered activated, and CR 605.3b — a mana
// ability resolves immediately once activated). The mana it adds is
// unrestricted, as the printed card's is: a restricted token in the
// pool is doubled by a plain token of its colour.
func b33DoubleUnspentMana(g *game.Game, controller, _ uuid.UUID) string {
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return ""
	}
	out := ""
	for _, tok := range p.ManaPool {
		if tok.Color == "" {
			continue
		}
		out += "{" + tok.Color + "}"
	}
	return out
}

// b33CoalitionRelicLabel is the stack label of Coalition Relic's
// main-phase trigger.
const b33CoalitionRelicLabel = "Coalition Relic — remove the charge counters and add mana"

// b33RemoveChargeCountersForMana is Coalition Relic's main-phase
// body: every charge counter on the Relic comes off, and one mana of
// any colour is added for each — one colour pick per counter, the
// same prompt a Birds of Paradise activation opens. A Relic that has
// left the battlefield, or carries no counters, adds nothing.
func b33RemoveChargeCountersForMana(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return nil
	}
	n := src.Counters["charge"]
	if n <= 0 {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: "charge", N: -n}).Apply(ctx); err != nil {
		return err
	}
	produced := ""
	for i := 0; i < n; i++ {
		produced += "{W|U|B|R|G}"
	}
	return AddMana{Player: item.Controller, Produced: produced}.Apply(ctx)
}

// b33PutChargeCounterOnSelf is "put a charge counter on this
// permanent" — Coalition Relic's "{T}: Put a charge counter on this
// artifact" and Lost Jitte's "whenever equipped creature deals combat
// damage, put a charge counter on Lost Jitte" (batch 43).
//
// The zone guard is the load-bearing half for both: AddCounter does
// not check where its target is, and an artifact destroyed in response
// to its own trigger should not bank a counter in the graveyard.
func b33PutChargeCounterOnSelf(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: "charge", N: 1}.Apply(NewContext(g, item))
}

// b33DamageChosenTargetFromSource is Blasting Station's body: 1
// damage from the Station to the announced target, if it is still
// legal.
func b33DamageChosenTargetFromSource(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
		}
		return nil
	}
}

// b33UntapSelf is "untap this permanent" as a trigger body (Blasting
// Station's "you may untap this artifact").
func b33UntapSelf(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// b33DamageDefendingPlayerFromSource is Raid Bombardment's body: 1
// damage from the enchantment to the player the attacker was
// declared against — the event's Target, stamped onto
// item.Params.Player by a fill-in Build. The engine has no
// planeswalker defenders, so "the player or planeswalker" collapses
// to the player.
func b33DamageDefendingPlayerFromSource(g *game.Game, item *game.StackItem) error {
	defender := item.Params.Player
	if g.PlayerByIDForEffect(defender) == nil {
		return nil
	}
	return DealDamage{Source: item.SourceCardID, Target: defender, Amount: 1}.Apply(NewContext(g, item))
}

// b33CreateInsectsPerMinusCounterPlaced is Nest of Scarabs' body:
// `n` 1/1 black Insects, `n` being the number of -1/-1 counters the
// firing event placed, captured in Build.
func b33CreateInsectsPerMinusCounterPlaced(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if n <= 0 {
			return nil
		}
		return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 black Insect"), N: n}.Apply(NewContext(g, item))
	}
}

// b33ReanimateChosenWithCounters is Rakdos Joins Up's entry body: the
// announced creature card, if still in the controller's graveyard,
// returns to the battlefield under its owner's control ("from YOUR
// graveyard" — owner and controller coincide) and then gets `n`
// +1/+1 counters. The counters land a beat after the entry rather
// than as part of it: the reanimation path carries no counter
// option, and nothing in the catalog reads a creature's counters
// between its arrival and the next event.
func b33ReanimateChosenWithCounters(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
			return err
		}
		if !onBattlefield(g, id) {
			return nil
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
}

// b33DamageChosenOpponentByDeadCreaturesPower is Rakdos Joins Up's
// dies body: damage equal to the dead legend's last-known power to
// the announced opponent, if still legal. #1379's resolution-time LKI
// (ctx.TriggeringPermanent, backed by Game.lastKnownPermanents rather
// than the harvest-only Game.lastKnownBattlefield) answers this for a
// WATCHER, not just a card's own dies-trigger: PermanentInfo.Power is
// PowerForComparison as the legend last existed on the battlefield —
// layers (an anthem's bonus) and counters both included, uncapped.
// item.Params.Object.ID — the dead legend's instance ID, stamped by a
// fill-in Build — is kept only as the fallback for an item whose
// Trigger context did not carry the object (never true for a fresh
// trigger; the safe side for anything else).
func b33DamageChosenOpponentByDeadCreaturesPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 0
	if info, ok := ctx.TriggeringPermanent(); ok {
		n = info.Power
	} else {
		n = b17LastKnownPowerOffBattlefield(g, item.Params.Object.ID)
	}
	if n <= 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
	}
	return nil
}

// b33PutCounterOnEnteredCreature is Good-Fortune Unicorn's body: one
// +1/+1 counter on the creature that entered, if it is still on the
// battlefield.
func b33PutCounterOnEnteredCreature(entered uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, entered) {
			return nil
		}
		return AddCounter{Target: entered, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
	}
}

// createTappedZombie is Overseer of the Damned's dies body: one
// tapped 2/2 black Zombie.
func createTappedZombie(g *game.Game, item *game.StackItem) error {
	return CreateTokenAdvanced{
		Controller: item.Controller,
		Spec:       Token(BlackZombieToken()).EntersTapped(),
		N:          1,
	}.Apply(NewContext(g, item))
}

// b33ExileDeadThenCounterOnEachVampire is Patron of the Vein's dies
// body: the dead creature is exiled if it is still in a graveyard (a
// card that has since been reanimated, or an opponent's commander
// that went to the command zone, is left where it is), then every
// Vampire the controller controls gets a +1/+1 counter — the second
// half is not conditional on the first, as printed.
func b33ExileDeadThenCounterOnEachVampire(dead uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if z := g.FindCardZoneForEffect(dead); z != nil && z.Kind == game.ZoneGraveyard {
			if err := (ExileTarget{Target: dead}).Apply(NewContext(g, item)); err != nil {
				return err
			}
		}
		return b17PutCounterOnEachVampireYouControl(g, item)
	}
}

// b33ScryN is "scry N" as a trigger body (Hermes' Bird attack).
func b33ScryN(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return Scry{Player: item.Controller, N: n}.Apply(NewContext(g, item))
	}
}

// b33MillN is "mill N" as a trigger body (Sidisi's enters-or-attacks).
func b33MillN(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return MillCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
	}
}

// b33CreateTokenBody is a trigger body that makes one token from
// `template` for the controller.
func b33CreateTokenBody(template func() game.Card) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: template(), N: 1}.Apply(NewContext(g, item))
	}
}

// b33SacrificeListedCards is the delayed-trigger body Dalkovan
// Encampment schedules: sacrifice every card the item carries that
// is still on the battlefield under the controller's control.
// Package-level so the delayed trigger captures nothing.
func b33SacrificeListedCards(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !onBattlefield(g, t.ID) || c.Controller != item.Controller {
			continue
		}
		if err := (SacrificePermanent{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b33DalkovanEncampmentLabel is the stack label of Dalkovan
// Encampment's activated ability — what the attack trigger counts.
const b33DalkovanEncampmentLabel = "{2}{W}, {T}: Whenever you attack this turn, create two 1/1 red Warrior tokens tapped and attacking; sacrifice them at the next end step"

// b33DalkovanAttackLabel is the stack label of the attack trigger
// itself — the "one or more" dedup keys on it.
const b33DalkovanAttackLabel = "Dalkovan Encampment — create two 1/1 red Warriors tapped and attacking"

// b33DalkovanWarriors is Dalkovan Encampment's attack body: two 1/1
// red Warriors per activation of the land this turn, tapped and
// attacking the player the first declared attacker was declared
// against (item.Trigger.Event.Target, ADR 0041 P9), and a delayed
// trigger that sacrifices them at the beginning of the next end step.
func b33DalkovanWarriors(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := 2 * b33ResolutionsThisTurn(g, item.SourceCardID, b33DalkovanEncampmentLabel)
	if n <= 0 {
		return nil
	}
	cursor := b25LastEventSeq(g)
	if err := g.CreateTokensAttackingForEffect(item.Controller, b33TappedAttackingRedWarrior(), n, item.Trigger.Event.Target); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Dalkovan Encampment — sacrifice the Warriors",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(ctx)
}

// b33DistributeCountersRoundRobin is Lathiel's body: the life gained
// this turn, re-read at resolution (the intervening if), is dealt out
// as +1/+1 counters one at a time around the announced creatures in
// the order they were picked, skipping any that is no longer legal.
// With one creature chosen every counter lands on it; with three, the
// first gets the remainder.
func b33DistributeCountersRoundRobin(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := b15LifeGainedThisTurn(g, item.Controller)
	if n <= 0 {
		return nil
	}
	var targets []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || !g.TargetStillLegalForEffect(item, t) {
			continue
		}
		targets = append(targets, t.ID)
	}
	if len(targets) == 0 {
		return nil
	}
	share := make(map[uuid.UUID]int, len(targets))
	for i := 0; i < n; i++ {
		share[targets[i%len(targets)]]++
	}
	for _, id := range targets {
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: share[id]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// tuckSelfThirdFromTop is God-Eternal Bontu's return body —
// Oketra's: the card, if it is still in a graveyard or in exile, goes
// into its owner's library third from the top.
func tuckSelfThirdFromTop(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || (z.Kind != game.ZoneGraveyard && z.Kind != game.ZoneExile) {
		return nil
	}
	return b22TuckThirdFromTop(g, item.SourceCardID)
}

// --- goad --------------------------------------------------------

// b33AlelaGoadLabel is the stack label of Alela's combat-damage
// trigger — the "one or more" dedup keys on it.
const b33AlelaGoadLabel = "Alela, Cunning Conqueror — goad a creature that player controls"

// b33Goad stamps the engine's goad marker on a battlefield creature:
// `by` is the goading player. The engine surfaces the marker (the
// client badges the creature and the context menu offers to clear
// it) and, since #1571, enforces goad's two CR 701.15b requirements
// from it (game/attack_requirements.go) — so this is exactly the
// sandbox's goad, written from an effect under the lock the effect
// already holds. Nothing to stamp is not an error.
func b33Goad(g *game.Game, cardID, by uuid.UUID) {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			g.Battlefield.Cards[i].GoadedBy = by
			return
		}
	}
}

// b33ClearListedGoads is the delayed-trigger body that ends a goad
// "until your next turn": every card the item carries that is still
// on the battlefield and still goaded by the item's controller has
// its marker cleared. Package-level so the delayed trigger captures
// nothing.
func b33ClearListedGoads(g *game.Game, item *game.StackItem) error {
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !onBattlefield(g, t.ID) || c.GoadedBy != item.Controller {
			continue
		}
		b33Goad(g, t.ID, uuid.Nil)
	}
	return nil
}

// b33AlelaGoadChosen is Alela's body: the chosen creature is goaded
// by the controller if it is still legal and still controlled by the
// player the Faeries hit — read back off the item's carried trigger
// event (item.Trigger, #1379, ADR 0041 P9) rather than captured at
// trigger time — and a delayed trigger clears the marker at the
// beginning of the controller's next turn (CR 701.15a's "until your
// next turn"). A pick under some other player's control — possible
// when Faeries connected with two players in one combat and the
// clause offered both players' creatures — does nothing.
func b33AlelaGoadChosen(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	var victim uuid.UUID
	if item.Trigger != nil {
		victim = item.Trigger.Event.Target
	}
	c, found := g.LookupCardForEffect(id)
	if !found || c.Controller != victim || !onBattlefield(g, id) {
		return nil
	}
	b33Goad(g, id, item.Controller)
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Alela, Cunning Conqueror — the goad ends",
		Cards:              []uuid.UUID{id},
		Body:               clearListedGoadsBody,
	}.Apply(ctx)
}
