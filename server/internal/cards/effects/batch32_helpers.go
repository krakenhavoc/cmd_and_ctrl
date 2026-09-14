package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch32_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 32 (#395, `edhrec_rank` 3346–3446). Own file per
// the #231 convention; every package-level name carries the b32
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "another creature you control enters" is
// b13AnotherCreatureYouControlEntered, "a permanent you control
// entered" is enteredUnderYourControl, "this creature dealt combat
// damage to a player" is combatDamageToPlayerBy, the per-label
// "one or more" dedup is b12TriggerPendingOrOnStack and its
// pick-prompt third leg b17PickTargetPendingFrom, "you gained life
// this turn" is b15LifeGainedThisTurn, "each opponent loses N" is
// eachOpponentLosesLife, "untap each X you control" is
// b16UntapAllYouControlMatching, the bounded mill is
// b31MillAtMost, the first-legal-target read is
// b16FirstLegalTargetCard and its destroy body
// b17DestroyFirstLegalTarget, the Mirage fetch body is fetchDual,
// the painless {C} half is painlessColorless, the 4/4 Angel is
// b12WhiteAngelFlyingToken, the Blood / Clue / Food tokens are in
// tokens.go, and the lord keyword grant is TribalKeywordGrant.

// --- mana-ability riders -----------------------------------------

// b32EachOpponentGainsLifeRider is the post-production half of
// Grove of the Burnwillows' coloured ability: "Each opponent gains
// 1 life." A rider, not a cost, and life GAIN through
// ChangePlayerLifeForEffect so every opponent's "whenever you gain
// life" payoff sees it — which is the whole drawback of the card.
func b32EachOpponentGainsLifeRider(n int) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == controller {
				continue
			}
			if err := g.ChangePlayerLifeForEffect(source, p.ID, n); err != nil {
				return err
			}
		}
		return nil
	}
}

// --- card reads --------------------------------------------------

// b32ElvesControlled counts the Elves `controller` controls —
// Immaculate Magistrate's "for each Elf you control". Effective
// subtypes, so a changeling is an Elf; the Magistrate herself is
// one and counts, as printed.
func b32ElvesControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype("Elf") {
			n++
		}
	}
	return n
}

// b32AttackingTokenYouControl is Neyali's "attacking tokens you
// control": a creature token the caster controls that is currently
// declared as an attacker (of anything — a player, a planeswalker
// or a battle).
func b32AttackingTokenYouControl(_ *game.Game, caster uuid.UUID, c game.Card) bool {
	return c.Controller == caster && c.IsCreature() && IsToken(c) && c.AttackingTarget != uuid.Nil
}

// b32TokenYouControlAttacksAPlayer reports whether a creature token
// `controller` controls is attacking a PLAYER right now — Neyali's
// second sentence, read as the trigger resolves rather than at the
// event, so a token declared against a planeswalker first and one
// against a player second still count as "attack a player".
func b32TokenYouControlAttacksAPlayer(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if !b32AttackingTokenYouControl(g, controller, c) {
			continue
		}
		if g.ClassifyAttackTargetForEffect(c.AttackingTarget) == game.AttackTargetPlayer {
			return true
		}
	}
	return false
}

// b32PlayersDealtCombatDamageThisTurnByYourCreatureNamed is the set
// of players a creature named `name` under `controller`'s control
// dealt combat damage to this turn — walked off the event log back
// to the turn's upkeep (b06EnteredThisTurn's boundary). What Trygon
// Predator's target clause reads to narrow "target artifact or
// enchantment THAT PLAYER controls" to the players its Predators
// actually connected with: a target predicate is not handed the
// trigger's event, so the log stands in for it. The dealing creature
// is read live, and one that has since left the battlefield is
// matched by the LKI its death recorded.
func b32PlayersDealtCombatDamageThisTurnByYourCreatureNamed(g *game.Game, controller uuid.UUID, name string) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventBeginUpkeep {
			break
		}
		if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Actor != controller {
			continue
		}
		if p := g.PlayerByIDForEffect(ev.Target); p == nil {
			continue
		}
		src, ok := g.LookupCardForEffect(ev.Source)
		if !ok || src.Name != name {
			continue
		}
		out[ev.Target] = true
	}
	return out
}

// b32ArtifactOrEnchantmentOfPlayerHitByYourTrygonPredator is Trygon
// Predator's target predicate: an artifact or enchantment controlled
// by a player one of the caster's Trygon Predators dealt combat
// damage to this turn. The trigger's body re-checks the specific
// damaged player, so with two Predators connecting with two players
// in one combat each trigger's pick is still validated against its
// own victim.
func b32ArtifactOrEnchantmentOfPlayerHitByYourTrygonPredator(g *game.Game, caster uuid.UUID, c game.Card) bool {
	if !c.IsArtifact() && !c.IsEnchantment() {
		return false
	}
	return b32PlayersDealtCombatDamageThisTurnByYourCreatureNamed(g, caster, "Trygon Predator")[c.Controller]
}

// b32CardsInExileLastExiledDuring is the set of cards currently in
// exile whose most recent arrival there happened while an ability
// labelled `label` controlled by `controller` was resolving — the
// cards Neyali, Suns' Vanguard has exiled, however many turns ago.
//
// Walked off the event log newest-first: the first exile-entry seen
// for a card is its latest, and it counts only when the nearest
// earlier resolution is the labelled ability's (nothing else
// resolves between an ability's EventResolve and the moves its
// body makes). A card Neyali exiled that was later played and then
// exiled again by something else has that something else's
// resolution nearest its latest arrival, so it is not claimed —
// which is what keeps a stale grant off a card that is no longer
// Neyali's.
func b32CardsInExileLastExiledDuring(g *game.Game, controller uuid.UUID, label string) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	var out []uuid.UUID
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneExile || ev.CardID == uuid.Nil || seen[ev.CardID] {
			continue
		}
		seen[ev.CardID] = true
		if !b32NearestResolutionIs(g, i, controller, label) {
			continue
		}
		if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		out = append(out, ev.CardID)
	}
	return out
}

// b32NearestResolutionIs reports whether the nearest EventResolve
// before index `at` in the log is the resolution of an ability
// labelled `label` controlled by `controller`.
func b32NearestResolutionIs(g *game.Game, at int, controller uuid.UUID, label string) bool {
	for j := at - 1; j >= 0; j-- {
		prev := g.Events[j]
		if prev.Kind != game.EventResolve {
			continue
		}
		return prev.Actor == controller && prev.Label == label
	}
	return false
}

// --- trigger conditions ------------------------------------------

// b32CreatureYouControlBecameTapped is Quest for Renewal's
// condition — b11DwarfYouControlBecameTapped without the tribe. Two
// event kinds feed it for the reason that helper gives: the engine
// taps an attacker without an EventTapCard, so an EventAttack whose
// creature is now tapped is "became tapped" and a vigilance attacker
// is not.
func b32CreatureYouControlBecameTapped(ev game.Event, source *game.Card, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTapCard, game.EventAttack:
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller || !c.IsCreature() {
		return false
	}
	if ev.Kind == game.EventAttack && !c.Tapped {
		return false
	}
	return true
}

// b32EndStepAndYouGainedLifeThisTurnAtLeast is Angelic Accord's
// intervening-if at announce: any player's end step began and the
// source's controller gained `n` or more life this turn. Re-run at
// resolution by the body, as CR 603.4 asks.
func b32EndStepAndYouGainedLifeThisTurnAtLeast(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	return ev.Kind == game.EventBeginEndStep && b15LifeGainedThisTurn(g, source.Controller) >= n
}

// b32AnotherFaerieYouControlEntered is Obyra's condition: a Faerie
// other than the source entered under the source's controller's
// control. Effective subtypes, so a changeling counts; a Faerie
// token counts, since a token's entry emits the same EventETB.
func b32AnotherFaerieYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype("Faerie")
}

// b32TokenYouControlAttacked is Neyali's condition before the "one
// or more" dedup: a creature token the source's controller controls
// was declared as an attacker.
func b32TokenYouControlAttacked(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && IsToken(c)
}

// --- effect bodies -----------------------------------------------

// b32AngelIfYouGainedLifeThisTurnAtLeast is Angelic Accord's
// end-step body: the intervening-if re-checked at resolution, then
// a 4/4 white Angel with flying.
func b32AngelIfYouGainedLifeThisTurnAtLeast(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if b15LifeGainedThisTurn(g, item.Controller) < n {
			return nil
		}
		return CreateToken{Controller: item.Controller, Template: b12WhiteAngelFlyingToken(), N: 1}.Apply(NewContext(g, item))
	}
}

// b32GainLifeThenDraw is Sphinx's Revelation's body: gain X life,
// then draw X, in printed order.
func b32GainLifeThenDraw(item *game.StackItem, ctx *Context, n int) error {
	if n <= 0 {
		return nil
	}
	if err := (GainLife{Player: item.Controller, Amount: n}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
}

// b32TargetPlayerDrawsAndLosesLife is Blood Pact's body: the
// announced player, if still legal, draws `draw` cards and loses
// `life` life — a loss, not damage, so no prevention shield sees
// it.
func b32TargetPlayerDrawsAndLosesLife(item *game.StackItem, ctx *Context, draw, life int) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		if err := (DrawCards{Player: t.ID, N: draw}).Apply(ctx); err != nil {
			return err
		}
		return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -life)
	}
	return nil
}

// b32CounterTargetThenControllerMills is Didn't Say Please's body:
// the announced spell, if still on the stack, is countered and its
// controller — read before the counter moves it — mills `n`,
// bounded by their library.
func b32CounterTargetThenControllerMills(ctx *Context, n int) error {
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
		return b31MillAtMost(ctx, controller, n)
	}
	return nil
}

// b32ErtaiLabel is the stack label of Ertai Resurrected's entry
// trigger.
const b32ErtaiLabel = "Ertai Resurrected — counter a spell, or destroy a creature or planeswalker; its controller draws"

// b32CounterOrDestroyChosenThenControllerDraws is Ertai
// Resurrected's body. The mode is whatever was chosen: a spell on
// the stack is countered and its controller draws; a permanent on
// the battlefield is destroyed and its controller draws. Nothing
// chosen ("up to one") does nothing. The controller is read before
// the counter or destruction moves the card, and draws even when an
// indestructible creature survives the destroy — the draw is not
// conditional on the removal, as printed. "Another" is enforced by
// the target clause's name filter.
func b32CounterOrDestroyChosenThenControllerDraws(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if spell := g.StackItemForEffect(t.ID); spell != nil {
			controller := spell.Controller
			if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: controller, N: 1}.Apply(ctx)
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok || !onBattlefield(g, t.ID) {
			return nil
		}
		controller := c.Controller
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return DrawCards{Player: controller, N: 1}.Apply(ctx)
	}
	return nil
}

// b32TargetSpellOrAnotherCreatureOrPlaneswalker is Ertai
// Resurrected's target clause: "up to one" of a spell on the stack
// or a creature or planeswalker on the battlefield that is not Ertai
// himself. Two zones in one clause, Aang, Swift Savior's shape, and
// Mode "any" for the reason that card gives — the client's Mode
// string decides which surfaces enter targeting, and legality still
// comes from CardOK. Players is left false, so "any" cannot point at
// a player.
func b32TargetSpellOrAnotherCreatureOrPlaneswalker(label string) *game.TargetSpec {
	permanent := And(Or(Creature(), Planeswalker()), b03NotNamed("Ertai Resurrected"))
	return &game.TargetSpec{
		Mode:  "any",
		Label: label,
		Zones: []game.ZoneKind{game.ZoneBattlefield, game.ZoneStack},
		CardOK: func(g *game.Game, caster uuid.UUID, c game.Card, zone game.ZoneKind) bool {
			if zone == game.ZoneStack {
				// Every card on the stack is a spell; abilities
				// live in StackMeta with no card and cannot be
				// picked — the Disallow gap, declared on the card.
				return true
			}
			return permanent(g, caster, c)
		},
		Min: 0, Max: 1,
	}
}

// b32PutCountersPerElfOnChosenCreature is Immaculate Magistrate's
// body: one +1/+1 counter on the chosen creature, if it is still
// legal, for each Elf the controller controls as the ability
// resolves.
func b32PutCountersPerElfOnChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := b32ElvesControlled(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}

// b32TrygonPredatorLabel is the stack label of Trygon Predator's
// combat-damage trigger.
const b32TrygonPredatorLabel = "Trygon Predator — destroy an artifact or enchantment that player controls"

// b32DestroyChosenIfControlledBy is Trygon Predator's body: the
// chosen artifact or enchantment is destroyed if it is still legal
// and still controlled by `victim`, the player the Predator hit. A
// pick under some other player's control — possible when two
// Predators connected with two players in one combat and the
// clause offered both players' permanents — does nothing.
func b32DestroyChosenIfControlledBy(victim uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		c, found := g.LookupCardForEffect(id)
		if !found || c.Controller != victim {
			return nil
		}
		return DestroyTarget{Target: id}.Apply(ctx)
	}
}

// b32NeyaliLabel is the stack label of Neyali's attack trigger —
// the "one or more" dedup keys on it, and so does the exile-permission
// walk that finds the cards earlier resolutions of it exiled.
const b32NeyaliLabel = "Neyali, Suns' Vanguard — attacking tokens have double strike; exile the top card"

// b32NeyaliAttack is Neyali, Suns' Vanguard's attack body, both
// printed sentences on one stack item:
//
//   - every attacking creature token the controller controls gains
//     double strike until end of turn (the static's stand-in — an
//     attack declaration does not rebuild the layer cache, so a
//     static that read "attacking" would not see it);
//   - if a token is attacking a PLAYER, the top card of the
//     controller's library is exiled with permission to play it this
//     turn, and every card an earlier resolution of this trigger
//     exiled that is still in exile is re-granted for this turn —
//     which is "during any turn you attacked with a token", read as
//     "any turn this trigger resolved".
func b32NeyaliAttack(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GrantKeywordUntilEOT{
		Match:    b32AttackingTokenYouControl,
		Keywords: []string{"double strike"},
		Label:    "Neyali, Suns' Vanguard — attacking tokens have double strike",
	}).Apply(ctx); err != nil {
		return err
	}
	if !b32TokenYouControlAttacksAPlayer(g, item.Controller) {
		return nil
	}
	if _, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.ExilePlayPermission{}); err != nil {
		return err
	}
	for _, id := range b32CardsInExileLastExiledDuring(g, item.Controller, b32NeyaliLabel) {
		if err := g.ExileCardWithPermissionForEffect(id, game.ExilePlayPermission{Player: item.Controller, UntilTurn: g.Turn.Number}); err != nil {
			return err
		}
	}
	return nil
}

// b32CreateTokenBody is the body of one of Transmutation Font's
// three tap abilities: one token from `template` for the
// controller.
func b32CreateTokenBody(template func() game.Card) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return CreateToken{Controller: item.Controller, Template: template(), N: 1}.Apply(NewContext(g, item))
	}
}

// b32SearchAuraThenEquipmentToHand is Axgard Armory's body: "an Aura
// card and/or an Equipment card" as two searches in printed order —
// up to one Aura, then up to one Equipment — each a real prompt the
// searcher may decline, so two Auras cannot be taken. The library is
// shuffled once, after the second search; both finds are revealed.
// The second search runs from the first's continuation, which
// outlives the resolution, so it is built from IDs alone.
func b32SearchAuraThenEquipmentToHand(g *game.Game, item *game.StackItem) error {
	controller, source := item.Controller, item.SourceCardID
	return SearchLibrary{
		Player:    controller,
		Predicate: func(c game.Card) bool { return c.IsAura() },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Optional:  true,
		Reason:    "Axgard Armory — choose up to one Aura card to put into your hand",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player:   controller,
				Source:   source,
				Pred:     func(c game.Card) bool { return c.HasSubtype("Equipment") },
				Dest:     game.ZoneHand,
				Limit:    1,
				Reveal:   true,
				Shuffle:  true,
				Optional: true,
				Reason:   "Axgard Armory — choose up to one Equipment card to put into your hand",
			})
		},
	}.Apply(NewContext(g, item))
}
