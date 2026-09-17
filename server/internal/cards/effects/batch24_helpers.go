package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch24_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 24 (#386, `edhrec_rank` 2536–2637). Own file per
// the #231 convention; every package-level name carries the b24
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "you gained life"
// is b10YouGainedLife and the turn's total is b15LifeGainedThisTurn,
// "a player lost life this turn" is b18LifeLostThisTurn, "an
// opponent discarded" is b18OpponentDiscarded, "a creature an
// opponent controls died" is b18OpponentsCreatureDied, "you cast a
// creature / noncreature spell" is creatureSpellCastByYou /
// b10NoncreatureSpellCastByYou, "an opponent cast a spell" is
// b15OpponentCastSpell with game.(*Game).ManaValueForEffect for its
// mana value,
// "a creature entered under your control" is enteredUnderYourControl,
// the combat-damage-to-a-player read is combatDamageToPlayerBy with
// OncePerBatch as the "one or more" dedup, the
// each-opponent drain is b21DrainEachOpponentAndGainTheTotal, the
// Vampire counter body is b17PutCounterOnEachVampireYouControl, the
// tutor-to-hand body is b06TutorToHand, the "unless you control a
// basic land" entry is EntersTappedUnless(b11ControlsBasicLand), the
// ETB-destroy item is destroyChosenTargetTrigger, and the Thopter /
// Treasure templates live in tokens.go.

// --- trigger conditions ------------------------------------------

// b24SubtypeYouControlDealtDamage is Wrathful Raptors' condition —
// Wrathful Red Dragon's b21DragonYouControlDealtDamage with the
// creature type as a parameter: a creature of `subtype` the source's
// controller controls was dealt damage, combat or not, by anyone's
// source, the source itself included. Damage is marked before the
// state-based sweep, so a creature that took lethal is still on the
// battlefield when its event fires and is read live.
func b24SubtypeYouControlDealtDamage(ev game.Event, source *game.Card, g *game.Game, subtype string) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == uuid.Nil {
		return false
	}
	if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.Controller == source.Controller && c.HasSubtype(subtype)
}

// b24CombatDamageToPlayerByYourCreatureOfSubtypes is the condition of
// the ability Mari, the Killing Quill grants — "whenever this
// creature deals combat damage to a player" — hung on Mari and read
// against every Assassin, Mercenary and Rogue her controller
// controls (Mari included; she is an Assassin). combatDamageToPlayerBy
// already checks the combat flag, the player target and the dealing
// creature's controller; this adds the creature type, read
// post-layer so a changeling counts.
func b24CombatDamageToPlayerByYourCreatureOfSubtypes(ev game.Event, source *game.Card, g *game.Game, subtypes ...string) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	if !ok {
		return false
	}
	for _, s := range subtypes {
		if dealer.HasSubtype(s) {
			return true
		}
	}
	return false
}

// b24LegendaryCreatureYouControlDealtCombatDamageToPlayer is Vraska
// Joins Up's draw condition: combat damage to a player by a
// legendary creature the source's controller controls. The
// supertype is read post-layer, so a creature something else made
// legendary counts.
func b24LegendaryCreatureYouControlDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && isLegendary(&dealer)
}

// b24NaturesWillLabel is Nature's Will's stack label for one damaged
// player — the player's name is what makes two players' triggers two
// labels.
func b24NaturesWillLabel(g *game.Game, victim uuid.UUID) string {
	name := "that player"
	if p := g.PlayerByIDForEffect(victim); p != nil && p.Name != "" {
		name = p.Name
	}
	return "Nature's Will — tap " + name + "'s lands, untap your lands"
}

// b24YourEndStepAndYouGainedLifeThisTurn is Witch of the Moors'
// intervening-if at announce: the source's controller's end step
// began and they gained life this turn. Re-run at resolution by the
// effect body, as CR 603.4 asks.
func b24YourEndStepAndYouGainedLifeThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	return ev.Kind == game.EventBeginEndStep && ev.Actor == source.Controller &&
		b15LifeGainedThisTurn(g, source.Controller) > 0
}

// --- board reads -------------------------------------------------

// b24ControlsCreatureWithPowerAtLeast is Bugenhagen's upkeep
// intervening-if — "if you control a creature with power 7 or
// greater". Current power, so counters and anthems count.
func b24ControlsCreatureWithPowerAtLeast(g *game.Game, controller uuid.UUID, n int) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.CurrentPower() >= n {
			return true
		}
	}
	return false
}

// b24TappedCreaturesControlled counts the tapped creatures
// `controller` controls — Throne of the God-Pharaoh's amount, read as
// the trigger resolves.
func b24TappedCreaturesControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.Tapped {
			n++
		}
	}
	return n
}

// b24AttackingCreaturesOfSubtype counts the attacking creatures of
// `subtype` that `controller` controls — Dwynen's "each attacking Elf
// you control", read as the trigger resolves so an Elf declared
// after Dwynen still counts. Post-layer subtype, so a changeling
// counts.
func b24AttackingCreaturesOfSubtype(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, id := range b13AttackingCreaturesYouControl(g, controller) {
		if c, ok := g.LookupCardForEffect(id); ok && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// b24AnyPlayerLostAtLeastThisTurn is Y'shtola's end-step
// intervening-if — "if a player lost 4 or more life this turn" — any
// seated player, the source's controller included.
func b24AnyPlayerLostAtLeastThisTurn(g *game.Game, n int) bool {
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		if b18LifeLostThisTurn(g, p.ID) >= n {
			return true
		}
	}
	return false
}

// b24GraveyardHasCreatureCard reports whether `player`'s graveyard
// holds a creature card — the question Witch of the Moors' targeted
// declaration asks before it fires (the Hazel's Brewmaster split).
func b24GraveyardHasCreatureCard(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			return true
		}
	}
	return false
}

// b24ExiledCardWithCounterOwnedBy finds a card in exile owned by
// `owner` that carries at least one `kind` counter — the card Mari's
// granted ability removes a hit counter from. The first in exile
// order; every such card is interchangeable for the effect.
func b24ExiledCardWithCounterOwnedBy(g *game.Game, owner uuid.UUID, kind string) (uuid.UUID, bool) {
	if g.Exile == nil {
		return uuid.Nil, false
	}
	for _, c := range g.Exile.Cards {
		if c.Owner == owner && c.Counters[kind] > 0 {
			return c.InstanceID, true
		}
	}
	return uuid.Nil, false
}

// --- statics -----------------------------------------------------

// b24KeywordCounterGrant is CR 122.1e for one keyword: a creature
// with a `keyword` counter on it has that keyword. The engine reads
// no keyword counters of its own, so Vraska Joins Up carries the
// rule for the counters it places, while it is on the battlefield —
// any creature's, anyone's, which is exactly what the rule says and
// never more.
func b24KeywordCounterGrant(keyword string) game.StaticAbility {
	return b16GrantKeywords(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
		return target.IsCreature() && target.Counters[keyword] > 0
	}, keyword)
}

// --- target specs ------------------------------------------------

// b24TargetAnyNotSubtype is "any target that isn't a <subtype>" —
// Wrathful Raptors' reflected damage, b21TargetAnyNonDragon with the
// creature type as a parameter: TargetAny's set minus permanents
// with the subtype; players are never Dinosaurs, so the player half
// is untouched.
func b24TargetAnyNotSubtype(subtype string) *game.TargetSpec {
	spec := TargetAny()
	spec.Label = "any target that isn't a " + subtype
	base := spec.CardOK
	spec.CardOK = func(g *game.Game, caster uuid.UUID, c game.Card, z game.ZoneKind) bool {
		return base(g, caster, c, z) && !c.HasSubtype(subtype)
	}
	return spec
}

// --- effect bodies -----------------------------------------------

// eachOpponentDiscardsOne is Burglar Rat's body: every opponent
// picks a card from their own hand to discard, one prompt per
// opponent, addressed to that opponent. A player with an empty hand
// is skipped by the discard prompt itself.
func eachOpponentDiscardsOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player: opp,
			Source: item.SourceCardID,
			N:      1,
		})
	}
	return nil
}

// b24TapAllLandsControlledBy taps every untapped land `player`
// controls — Nature's Will's first half. Snapshot then tap, so the
// tap events don't disturb the walk.
func b24TapAllLandsControlledBy(ctx *Context, player uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsLand() && !c.Tapped {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b24PlayerSacrificesGreatestPowerCreatureAndLosesLife is one
// opponent's share of Will of the Abzan's first mode: they choose a
// creature with the greatest power among the creatures they control
// (GreatestPowerYouControl is evaluated for the CHOOSER — the
// sacrifice prompt passes them as the caster — so ties are theirs to
// break) and lose `life`. The loss is immediate; the sacrifice is
// their prompt.
func b24PlayerSacrificesGreatestPowerCreatureAndLosesLife(g *game.Game, item *game.StackItem, player uuid.UUID, life int) error {
	g.PlayerSacrificesForEffect(item.SourceCardID, player,
		sacrificeSpec("a creature with the greatest power among creatures you control", GreatestPowerYouControl()),
		"Will of the Abzan — sacrifice a creature with the greatest power")
	return g.ChangePlayerLifeForEffect(item.SourceCardID, player, -life)
}

// b24ReturnGraveyardTargetsToBattlefield puts up to `n` of the
// spell's still-legal graveyard targets onto the battlefield under
// their owner's control, in announce order — Lich-Knights' Conquest's
// return half.
func b24ReturnGraveyardTargetsToBattlefield(ctx *Context, n int) error {
	returned := 0
	for _, t := range ctx.LegalTargets() {
		if returned >= n {
			break
		}
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		returned++
	}
	return nil
}

// b24RemoveHitCounterDrawAndTreasures is the body of the ability
// Mari grants: remove a hit counter from a card `victim` owns in
// exile, and if one was removed, draw a card and create two
// Treasures. With no such card the ability does nothing — "if you
// do" is the whole gate.
func b24RemoveHitCounterDrawAndTreasures(g *game.Game, item *game.StackItem, victim uuid.UUID) error {
	card, ok := b24ExiledCardWithCounterOwnedBy(g, victim, "hit")
	if !ok {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: card, Kind: "hit", N: -1}).Apply(ctx); err != nil {
		return err
	}
	if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 2}.Apply(ctx)
}
