package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch28_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 28 (#390, `edhrec_rank` 2940–3040). Own file per
// the #231 convention; every package-level name carries the b28
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "whenever you gain
// life" is b10YouGainedLife, the once-per-turn trigger tally is
// b11TriggeredThisTurn, "whenever you attack with N or more
// creatures" is b16PlayerAttackedWithAtLeast, "if you cast it" is
// b16EnteredFromStack, the fight is b10Fight, "another creature you
// control enters" is b13AnotherCreatureYouControlEntered, "whenever
// you cast a noncreature spell" is b10NoncreatureSpellCastByYou, the
// dies-trigger power read is b13LastKnownPower, devotion is
// devotionTo, the lord builders are TribalAnthem /
// TribalKeywordGrant, "each opponent loses N and you gain N" is
// eachOpponentLosesLife + GainLife, the graveyard probe for the
// Hazel's Brewmaster split is b24GraveyardHasCreatureCard's shape,
// and the Soldier with lifelink is b22WhiteSoldierLifelinkToken.

// --- tokens ------------------------------------------------------

// b28WhiteSpiritFlyingToken is kept as a function because a card passes it as a value; the data lives in tokens_table.go.
func b28WhiteSpiritFlyingToken() game.Card { return TokenCard("1/1 white Spirit with flying") }

// b28EldraziScionToken is Spawnbed Protector's 1/1 colorless Eldrazi
// Scion with "Sacrifice this token: Add {C}" — the Spawn's mana
// ability on a 1/1 body.
func b28EldraziScionToken() game.Card { return tokenFromCatalog(printedB28EldraziScionToken) }

// printedB28EldraziScionToken is the Eldrazi Scion as PRINTED —
// the ability included. It is the catalog's entry for this token
// (token_catalog.go, #521): the ability is registered from here at
// boot, and the template that reaches the battlefield carries the
// key that finds it rather than the closure itself.
func printedB28EldraziScionToken() tokenTemplate {
	return tokenTemplate{
		Slug: "eldrazi-scion",
		Card: game.Card{
			Name:      "Eldrazi Scion",
			TypeLine:  "Token Creature — Eldrazi Scion",
			Power:     1,
			Toughness: 1,
		},
		Mana: []game.ManaAbilityShape{{
			SacrificeCost: true,
			Produced:      "{C}",
			Label:         "Sacrifice this creature: Add {C}",
		}},
		Text: "Sacrifice this token: Add {C}.",
	}
}

// --- trigger conditions ------------------------------------------

// b28TriggerPromptPendingFrom reports whether a yes/no trigger prompt
// sourced from `source` is already waiting on an answer — the
// optional-trigger twin of OncePerBatch. A "whenever you
// gain life … do this only once each turn" ability asked twice in
// one mutation (two lifelinkers connecting at once) would otherwise
// open two prompts and, answered yes twice, run twice.
func b28TriggerPromptPendingFrom(g *game.Game, source *game.Card) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerPrompt && c.Source == source.InstanceID {
			return true
		}
	}
	return false
}

// b28CreatureWasDealtDamage is Repercussion's condition: a creature
// on the battlefield was dealt damage, by anything. The creature is
// still there when the event fires (state-based actions run after),
// so its controller is read live and handed back for the trigger to
// aim at.
func b28CreatureWasDealtDamage(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target == uuid.Nil {
		return game.Card{}, false
	}
	if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !c.IsCreature() {
		return game.Card{}, false
	}
	return c, true
}

// b28DragonYouControlTargetedByOpponent is Thunderbreak Regent's
// condition: a Dragon the source's controller controls became the
// target of a spell or ability whose controller is an opponent.
// EventBecomesTarget fires once per target slot (CR 115.3) at
// announce, with the targeting player in Actor and the targeted card
// in CardID (uuid.Nil for a player target, which is what keeps a
// player-targeting spell from matching). The Regent is a Dragon and
// counts for its own trigger.
func b28DragonYouControlTargetedByOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.CardID == uuid.Nil {
		return false
	}
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	if !onBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.HasSubtype("Dragon")
}

// b28YouSacrificedArtifactOrCreature is Ravenous Squirrel's condition
// — b12YouSacrificedAnArtifact widened to "an artifact or creature".
// The sacrifice event fires before the zone move, so the permanent is
// still on the battlefield to be read.
func b28YouSacrificedArtifactOrCreature(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (c.IsArtifact() || c.IsCreature())
}

// b28SelfBlocked is "whenever this creature blocks a creature" —
// Brimaz's second trigger. EventBlock names the blocker in CardID and
// the attacker in Target, and is emitted once per pair of the FINAL
// block declaration (#830): a blocker re-pointed before the lock-in
// blocks once, against the attacker it ends on.
func b28SelfBlocked(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventBlock && ev.CardID == source.InstanceID && ev.Target != uuid.Nil
}

// b28GraveyardHasCreatureCardOfSubtype reports whether `player`'s
// graveyard holds a creature card with the given subtype — the
// question Spawnbed Protector's targeted declaration asks before it
// fires (the Hazel's Brewmaster split).
func b28GraveyardHasCreatureCardOfSubtype(g *game.Game, player uuid.UUID, subtype string) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && c.HasSubtype(subtype) {
			return true
		}
	}
	return false
}

// --- replacements ------------------------------------------------

// b28PermanentsYouControlEnterWithACounter is "each <match> you
// control enters with an additional +1/+1 counter on it" — Dragonstorm
// Globe's Dragons, Uncivil Unrest's riot on nontoken creatures — as a
// CR 614 replacement on the entry of a permanent under the source's
// controller's control. Renata's shape (b27OtherCreaturesYouControlEnterWithACounter)
// with the "what enters" test as a parameter; `match` receives the
// entering card, still in its old zone.
func b28PermanentsYouControlEnterWithACounter(label string, match func(game.Card) bool) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.Controller == src.Controller && match(entering)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB("+1/+1", 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// b28OtherCreaturesYouControlGetAnExtraCounter is Benevolent Hydra's
// "if one or more +1/+1 counters would be put on another creature
// you control, that many plus one are put on it instead" — Conclave
// Mentor's replacement with the Hydra itself excluded.
func b28OtherCreaturesYouControlGetAnExtraCounter(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventCounterPlaced},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventCounter || ev.CounterName != "+1/+1" || ev.CounterDelta <= 0 {
				return false
			}
			if src == nil || ev.CounterTarget == src.InstanceID {
				return false
			}
			target, ok := g.LookupCardForEffect(ev.CounterTarget)
			return ok && target.IsCreature() && target.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.CounterDelta++
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// b28DamageSourceControlledBy resolves the card a damage replacement
// event names as its source and reports whether `controller`
// controls it — a permanent on the battlefield, or a spell on the
// stack whose Controller is its caster. The "controls it" test reads
// damageSourceCharacteristics (#1430), so a source that has already
// left the battlefield is judged by the controller it had there (CR
// 608.2h), not by whoever the card in its new zone belongs to now.
// The Card itself is still the current-zone lookup, for callers that
// need live-only fields (instance ID, counters); those are only ever
// read after also confirming the source is presently on the
// battlefield, so a departed source never reaches them under the
// wrong identity. A source that cannot be found (it left every
// tracked zone) is nobody's, which errs weaker.
func b28DamageSourceControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) (game.Card, bool) {
	if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.DamageSource == uuid.Nil {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.DamageSource)
	if !ok {
		return game.Card{}, false
	}
	ch, chOk := damageSourceCharacteristics(ev, g)
	if !chOk || ch.Controller != controller {
		return game.Card{}, false
	}
	return c, true
}

// b28DamageTargetIsOpponentOrTheirPermanent reports whether a damage
// replacement event's target is an opponent of `controller` or a
// permanent an opponent controls — Fated Firepower's recipient
// clause.
func b28DamageTargetIsOpponentOrTheirPermanent(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	if ev.DamageTarget == uuid.Nil {
		return false
	}
	if p := g.PlayerByIDForEffect(ev.DamageTarget); p != nil {
		return p.ID != controller
	}
	if !onBattlefield(g, ev.DamageTarget) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.DamageTarget)
	return ok && c.Controller != controller
}

// b28DoubleDamageFromCounteredCreaturesYouControl is Uncivil Unrest's
// second ability: damage dealt by a creature the source's controller
// controls that has a +1/+1 counter on it is doubled, whatever it is
// dealt to — combat and noncombat alike (Gratuitous Violence's
// shape, gated on the counter).
func b28DoubleDamageFromCounteredCreaturesYouControl(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil {
				return false
			}
			c, ok := b28DamageSourceControlledBy(ev, g, src.Controller)
			return ok && c.IsCreature() && onBattlefield(g, c.InstanceID) && c.Counters["+1/+1"] > 0
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.DamageAmount *= 2
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// b28AddCountersToDamageAtOpponents is Fated Firepower's replacement:
// damage a source the controller controls would deal to an opponent
// or to a permanent an opponent controls is increased by the number
// of `kind` counters on the source enchantment, read as the damage
// is dealt.
func b28AddCountersToDamageAtOpponents(label, kind string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || src.Counters[kind] <= 0 {
				return false
			}
			if _, ok := b28DamageSourceControlledBy(ev, g, src.Controller); !ok {
				return false
			}
			return b28DamageTargetIsOpponentOrTheirPermanent(ev, g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
			ev.DamageAmount += src.Counters[kind]
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// --- mana ---------------------------------------------------------

// b28ProducedDevotion is "Add an amount of {COLOR} equal to your
// devotion to <color>" — Karametra's Acolyte. Devotion is read at
// activation (devotionTo, CR 700.5: hybrid symbols count for every
// colour they offer); zero devotion adds nothing and still taps.
func b28ProducedDevotion(color string) func(*game.Game, uuid.UUID, uuid.UUID) string {
	slot := "{" + strings.ToUpper(color) + "}"
	return func(g *game.Game, controller, _ uuid.UUID) string {
		return strings.Repeat(slot, devotionTo(g, controller, color))
	}
}

// --- effect bodies -----------------------------------------------

// b28CopyEachLegalTargetXTimes is Doppelgang: for each announced
// target that is still legal, X tokens that are copies of it. Every
// target is snapshotted before any token lands, so a copy of one
// target is never itself copied.
func b28CopyEachLegalTargetXTimes(ctx *Context) error {
	x := ctx.X()
	if x <= 0 {
		return nil
	}
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	for _, id := range ids {
		if err := (CreateTokenCopy{Controller: ctx.Controller(), Copy: id, N: x}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b28FightChosenTargets is Ulvenwald Tracker's body: slot 0 must be a
// creature the controller controls and slot 1 another creature; both
// still legal as the ability resolves, or nothing fights. The "you
// control" half of slot 0 is checked here rather than refused at
// announce (the Bite Down posture — one predicate over both slots).
func b28FightChosenTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) < 2 {
		return nil
	}
	mine, other := item.Targets[0], item.Targets[1]
	if mine.Kind != game.TargetCard || other.Kind != game.TargetCard || mine.ID == other.ID ||
		!ctx.IsTargetLegal(mine) || !ctx.IsTargetLegal(other) {
		return nil
	}
	c, ok := g.LookupCardForEffect(mine.ID)
	if !ok || !c.IsCreature() || c.Controller != item.Controller {
		return nil
	}
	return b10Fight(ctx, mine.ID, other.ID)
}

// b28PutCountersOnEachCreatureYouControl puts n +1/+1 counters on
// every creature the item's controller controls — Nykthos Paragon's
// "that many +1/+1 counters on each creature you control". The set
// is snapshotted first so a counter payoff that makes a creature
// mid-loop is not counted.
func b28PutCountersOnEachCreatureYouControl(g *game.Game, item *game.StackItem, n int) error {
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: n}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// b28ReturnSelfToOwnersHand bounces the item's source if it is still
// on the battlefield — Arcanis the Omnipotent's "{2}{U}{U}: Return
// Arcanis to its owner's hand".
func b28ReturnSelfToOwnersHand(g *game.Game, item *game.StackItem) error {
	if !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// b28DifferentNames is a SearchLibrary Validate for "cards that each
// have different names" — Tiamat. An empty pick is always legal.
func b28DifferentNames(cards []game.Card) bool {
	seen := map[string]bool{}
	for _, c := range cards {
		if seen[c.Name] {
			return false
		}
		seen[c.Name] = true
	}
	return true
}

// b28SearchDragonsNotNamed is Tiamat's search: up to five Dragon cards
// not named `except` with different names, revealed, to hand, then
// shuffle. Effective subtypes are not readable in a library, so the
// printed type line is what "Dragon card" means here, as it does for
// every other library search.
func b28SearchDragonsNotNamed(g *game.Game, item *game.StackItem, except string, reason string) error {
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			return c.HasSubtype("Dragon") && c.Name != except
		},
		Dest:     game.ZoneHand,
		Limit:    5,
		Reveal:   true,
		Shuffle:  true,
		Validate: b28DifferentNames,
		Reason:   reason,
	}.Apply(NewContext(g, item))
}

// b28GainLifeAndDraw is Lifeblood Hydra's death payout: n life and n
// cards, in printed order.
func b28GainLifeAndDraw(g *game.Game, item *game.StackItem, n int) error {
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (GainLife{Player: item.Controller, Amount: n}).Apply(ctx); err != nil {
		return err
	}
	return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
}

// b28ReturnChosenGraveyardCardToHandThenScions is Spawnbed
// Protector's end-step body: the chosen graveyard card, if one was
// chosen and it is still there, goes to hand, then two Eldrazi
// Scions either way.
func b28ReturnChosenGraveyardCardToHandThenScions(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return CreateToken{Controller: item.Controller, Template: b28EldraziScionToken(), N: 2}.Apply(ctx)
}
