package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch17_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 17 (#310, `edhrec_rank` 1830–1929). Own file per
// the #231 convention; every package-level name carries the b17
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "enters with N counters" is
// b10EntersWithCounters, "a player attacks with N or more creatures"
// is b16PlayerAttackedWithAtLeast, the per-label "one or more" dedup
// is OncePerBatch, the once-per-turn tally is
// b11TriggeredThisTurn, "an opponent controls more lands than you"
// is b03OpponentControlsMoreLands, lands / creature cards you have
// are b03LandsControlled / b11CreatureCardsInGraveyard, a dead card's
// counters are b13LastKnownCounters, "another creature dies" is
// b15AnotherCreatureDied, a whole graveyard's exile is
// exileGraveyardForEffect, "each player draws" is b05EachPlayerDraws,
// and the whole-hand discard is discardWholeHand.

// --- token templates ---------------------------------------------

// --- board reads -------------------------------------------------

// b17LandCardsInGraveyard counts the land cards in `player`'s
// graveyard — Multani's second term. Printed type lines: a card in a
// graveyard has no layer cache.
func b17LandCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsLand() {
			n++
		}
	}
	return n
}

// b17CardTypesInOpponentsGraveyards counts the distinct card types
// among the cards in every graveyard an opponent of `controller`
// owns — Nighthawk Scavenger's power. ONE set across all of them, as
// printed ("among cards in your opponents' graveyards"): a creature
// in two graveyards is one type, not two.
func b17CardTypesInOpponentsGraveyards(g *game.Game, controller uuid.UUID) int {
	seen := map[string]bool{}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			_, types, _ := game.ParseTypeLine(c.TypeLine)
			for _, t := range types {
				seen[t] = true
			}
		}
	}
	return len(seen)
}

// b17DefendingPlayer resolves the player an EventAttack is aimed at.
// Since S27 an attack target may be a planeswalker or a battle, and
// "attacks you" / "you're the defending player" mean the player
// behind it.
func b17DefendingPlayer(g *game.Game, ev game.Event) uuid.UUID {
	if ev.Kind != game.EventAttack || ev.Target == uuid.Nil {
		return uuid.Nil
	}
	return g.DefendingPlayerForAttackForEffect(ev.Target)
}

// b17OpponentHasMoreLifeThanAnother is Breena's intervening if: `opp`
// (an opponent of `controller`) has more life than at least one
// OTHER opponent of `controller` still in the game.
func b17OpponentHasMoreLifeThanAnother(g *game.Game, controller, opp uuid.UUID) bool {
	target := g.PlayerByIDForEffect(opp)
	if target == nil || target.Eliminated || opp == controller {
		return false
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller || p.ID == opp {
			continue
		}
		if target.Life > p.Life {
			return true
		}
	}
	return false
}

// b17AttackersAllAvoid reports whether every creature `attacker`
// controls that is attacking right now is attacking someone other
// than `player` — Firemane Commando's "if none of those creatures
// attacked you". Read at resolution against the combat state, which
// AttackingTarget carries until combat ends.
func b17AttackersAllAvoid(g *game.Game, attacker, player uuid.UUID) bool {
	for _, id := range b13AttackingCreaturesYouControl(g, attacker) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if g.DefendingPlayerForAttackForEffect(c.AttackingTarget) == player {
			return false
		}
	}
	return true
}

// b17PermanentSacrificedToPay is the permanent sacrificed to pay for
// the activated ability `item` is resolving — Jarad's "the
// sacrificed creature's power". The stack item does not carry it,
// so it is read back off the event log: the ability's announce is
// the EventTrigger naming its source in both Source and CardID (the
// triggered-ability breadcrumb names the source only), and the
// sacrifice it paid is the nearest EventSacrifice by the activator
// before that announce. When the same source has been activated
// more than once and the older activations are still on the stack,
// the resolving one is the (older + 1)-th most recent announce —
// the resolving item has already left StackMeta by the time its
// Effect runs.
func b17PermanentSacrificedToPay(g *game.Game, item *game.StackItem) (uuid.UUID, bool) {
	older := 0
	for _, other := range g.StackMeta {
		if other != nil && other.Kind == game.StackItemActivated && other.SourceCardID == item.SourceCardID {
			older++
		}
	}
	announce := -1
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventTrigger && ev.Source == item.SourceCardID && ev.CardID == item.SourceCardID {
			if older == 0 {
				announce = i
				break
			}
			older--
		}
	}
	if announce < 0 {
		return uuid.Nil, false
	}
	for i := announce - 1; i >= 0; i-- {
		ev := g.Events[i]
		switch ev.Kind {
		case game.EventSacrifice:
			if ev.Actor == item.Controller {
				return ev.CardID, true
			}
		case game.EventResolve, game.EventCast, game.EventManaAbilityActivated:
			// Nothing this activation paid can be older than the
			// last thing the game did before it was announced.
			return uuid.Nil, false
		case game.EventTrigger:
			if ev.Source == item.SourceCardID && ev.CardID == item.SourceCardID {
				return uuid.Nil, false
			}
		}
	}
	return uuid.Nil, false
}

// b17LastKnownPowerOffBattlefield is the power a card had when it
// last left the battlefield, read without the harvester's LKI
// characteristic (which only a dies trigger receives): its printed
// power plus the +1/+1 and -1/-1 counters read back off the log,
// floored at zero the way CurrentPower floors it. A static bonus
// from another permanent is not in it — declared on the card that
// reads this.
func b17LastKnownPowerOffBattlefield(g *game.Game, cardID uuid.UUID) int {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return 0
	}
	p := c.Power + b13LastKnownCounters(g, cardID, "+1/+1") - b13LastKnownCounters(g, cardID, "-1/-1")
	if p < 0 {
		return 0
	}
	return p
}

// b17MilledCreatureCards is the batch of creature cards `owner`
// milled in the mill that fired the event at `seq` — Colossal
// Grave-Reaver's "one or more creature cards are put into your
// graveyard from your library". A mill emits one EventMill per card
// with nothing of its own in between; only the harvester's
// EventTrigger breadcrumbs (the first creature milled fires this
// very trigger) interleave. The batch is therefore every EventMill
// by `owner` from the firing event forward until the first event
// that is neither, read at resolution when every card of it is in
// the graveyard. Cards that have since left the graveyard are
// dropped.
func b17MilledCreatureCards(g *game.Game, owner uuid.UUID, seq uint64) []uuid.UUID {
	start := -1
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Seq == seq {
			start = i
			break
		}
	}
	if start < 0 {
		return nil
	}
	p := g.PlayerByIDForEffect(owner)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var out []uuid.UUID
	for i := start; i < len(g.Events); i++ {
		ev := g.Events[i]
		switch ev.Kind {
		case game.EventTrigger:
			continue
		case game.EventMill:
			if ev.Actor != owner || ev.NewZone != game.ZoneGraveyard {
				return out
			}
		default:
			return out
		}
		if !p.Graveyard.Contains(ev.CardID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.IsCreature() {
			out = append(out, ev.CardID)
		}
	}
	return out
}

// b17GreatestManaValue picks the card with the greatest mana value
// among `ids`, the first listed on a tie — the auto-pick Colossal
// Grave-Reaver declares for "put one of them onto the battlefield".
func b17GreatestManaValue(g *game.Game, ids []uuid.UUID) (uuid.UUID, bool) {
	best, bestMV := uuid.Nil, -1
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if mv := c.ManaValue(); mv > bestMV {
			best, bestMV = id, mv
		}
	}
	return best, best != uuid.Nil
}

// --- trigger conditions ------------------------------------------

// b17SelfOrZombieYouControlDied is Undead Augur's condition: the
// Augur itself died, or a Zombie its controller controlled did. The
// dead card is read post-move (its printed type line and its
// controller survive the move), so a changeling counts and a Zombie
// that was one only through a layer effect does not — weaker, never
// stronger.
func b17SelfOrZombieYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if cardDied(ev, source) {
		return true
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.HasSubtype("Zombie")
}

// b17SelfOrAnotherCreatureDied is Cordial Vampire's condition: any
// creature death at the table, its own included.
func b17SelfOrAnotherCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if cardDied(ev, source) {
		return true
	}
	_, died := diedCreature(ev, g)
	return died
}

// b17AnotherNontokenCreatureDied is Harvester of Souls's condition.
func b17AnotherNontokenCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && !IsToken(dead)
}

// b17SourceDealtDamageToSelf is Phyrexian Obliterator's condition:
// any source dealt damage to the Obliterator. EventDealDamage names
// the damaged card in Target and the dealer in Source.
func b17SourceDealtDamageToSelf(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventDealDamage && ev.Target == source.InstanceID && ev.Amount > 0
}

// b17DamageSourceController is the controller of the source named by
// an EventDealDamage. The combat paths stamp it on Actor; the spell
// and ability paths do not, so the source is looked up wherever it
// is — a spell dealing damage is still on the stack while its
// effect runs, and its Controller is the caster.
func b17DamageSourceController(ev game.Event, g *game.Game) uuid.UUID {
	if c, ok := g.LookupCardForEffect(ev.Source); ok && c.Controller != uuid.Nil {
		return c.Controller
	}
	return ev.Actor
}

// b17OpponentsCreatureAttackedYou is Kazuul's condition: a creature
// an opponent controls was declared attacking, and the source's
// controller is the defending player.
func b17OpponentsCreatureAttackedYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	return b17DefendingPlayer(g, ev) == source.Controller
}

// b17ElfSpellCastByYou is Leaf-Crowned Visionary's "whenever you cast
// an Elf spell". The spell is read off the stack, where its printed
// type line is intact; a changeling counts.
func b17ElfSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Elf")
}

// b17HistoricSpellCastByYou is Teshar's "whenever you cast a historic
// spell" — an artifact, a legendary, or a Saga.
func b17HistoricSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && b09IsHistoric(c)
}

// b17BreenaLabel is the stack label of Breena's trigger for one
// attacked opponent. The attacked player's name is in it because the
// ability triggers once per opponent attacked per combat. The label
// is computed per event, so the dedup is TriggerInFlightForEffect —
// the per-event-key leftover, not the OncePerBatch batch guard,
// until #784 — and it is what keeps a five-creature attack on one
// opponent from drawing five cards while a split attack on two still
// triggers twice.
func b17BreenaLabel(g *game.Game, opp uuid.UUID) string {
	name := "an opponent"
	if p := g.PlayerByIDForEffect(opp); p != nil {
		name = p.Name
	}
	return "Breena, the Demagogue — " + name + " was attacked: the attacker draws, two +1/+1 counters"
}

// --- effect bodies -----------------------------------------------

// b17DoubleCountersOn puts as many counters of each kind on `target`
// as it already has — Deepglow Skate, per target. The counts are
// snapshotted first, so a doubler (Doubling Season) that fires on
// the first kind cannot change what the second receives.
func b17DoubleCountersOn(ctx *Context, target uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || len(c.Counters) == 0 {
		return nil
	}
	counters := make(map[string]int, len(c.Counters))
	for kind, n := range c.Counters {
		if n > 0 {
			counters[kind] = n
		}
	}
	return b14PutCounters(ctx, target, counters)
}

// b17PutCounterOnEachVampireYouControl is Cordial Vampire's body:
// one +1/+1 counter on every Vampire its controller controls, the
// set snapshotted before the first counter lands.
func b17PutCounterOnEachVampireYouControl(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype("Vampire") {
			ids = append(ids, c.InstanceID)
		}
	}
	return b13PutCounterOnEach(NewContext(g, item), ids)
}

// b17PlayerSacrificesN queues `n` sacrifice prompts for `player`, each
// over the permanents `match` admits — Phyrexian Obliterator's "that
// many permanents of their choice". The engine's prompt picks one
// permanent, so N of them are asked as N prompts, the option lists
// trimmed after every answer; a player with fewer permanents than
// owed sacrifices what they have and the rest are dropped.
func b17PlayerSacrificesN(g *game.Game, source, player uuid.UUID, n int, reason string) {
	for i := 0; i < n; i++ {
		if g.PlayerSacrificesForEffect(source, player, nil, reason) == 0 {
			return
		}
	}
}

// b17WheelToGreatestDiscard is Jace's Archivist's body: each player
// discards their hand, then draws cards equal to the greatest number
// of cards a player discarded this way. Every discard happens before
// any draw, and the count is taken from the discards themselves.
func b17WheelToGreatestDiscard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	players := tablePlayers(ctx)
	greatest := 0
	for _, id := range players {
		n, err := discardWholeHand(g, id)
		if err != nil {
			return err
		}
		if n > greatest {
			greatest = n
		}
	}
	if greatest == 0 {
		return nil
	}
	for _, id := range players {
		if err := (DrawCards{Player: id, N: greatest}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b17ExileFromGraveyardAndMayPlay is Containment Construct's body:
// exile `cardID` from its owner's graveyard — if it is still there —
// and let `player` play it this turn. The grant is "play", so a land
// discarded to a looter can be played off it.
func b17ExileFromGraveyardAndMayPlay(g *game.Game, player, cardID uuid.UUID) error {
	if z := g.FindCardZoneForEffect(cardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(cardID, game.ExilePlayPermission{
		Player:    player,
		UntilTurn: g.Turn.Number,
	})
}

// b17ClaimJumperSearch is one pass of Claim Jumper's "you may search
// your library for a Plains card and put it onto the battlefield
// tapped", with `again` deciding whether the printed "then if an
// opponent controls more lands than you, repeat this process once"
// follows. The repeat runs in the search's continuation, after the
// first Plains is on the battlefield, so the second land count sees
// it; a declined first search still offers the second when the
// condition holds, as the printed "repeat this process" does.
func b17ClaimJumperSearch(g *game.Game, item *game.StackItem, again bool) error {
	controller, source := item.Controller, item.SourceCardID
	return SearchLibrary{
		Player:        controller,
		Predicate:     IsLandWithSubtype("Plains"),
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Optional:      true,
		Reason:        "Claim Jumper — a Plains card, onto the battlefield tapped",
		Source:        source,
		Then: func(g *game.Game, _ []uuid.UUID) error {
			if !again || !b03OpponentControlsMoreLands(g, controller) {
				return nil
			}
			return b17ClaimJumperSearch(g, item, false)
		},
	}.Apply(NewContext(g, item))
}

// destroyFirstLegalTarget destroys the first announce-time target
// slot that is still a legal battlefield card — the body of a
// single-target "destroy target X" activated ability (Insidious
// Fungus's two removal modes).
func destroyFirstLegalTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return DestroyTarget{Target: id}.Apply(ctx)
}

// b17FoodPerCreatureYouControl is The Battle of Bywater's second
// sentence: a Food for each creature the controller controls, read
// after the destruction. `destroyed` is the sweep's set. A card from
// it that is still on the battlefield is a commander waiting on its
// owner's CR 903.9 answer: it was destroyed, so it does not count.
func b17FoodPerCreatureYouControl(ctx *Context, destroyed []game.Card) error {
	gone := make(map[uuid.UUID]bool, len(destroyed))
	for _, c := range destroyed {
		gone[c.InstanceID] = true
	}
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && !gone[c.InstanceID] {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return CreateToken{Controller: ctx.Controller(), Template: FoodToken(), N: n}.Apply(ctx)
}

// --- replacements ------------------------------------------------

// b17OtherCreaturesYouControlEnterWithCounters is Arwen's "each other
// creature you control enters with a number of additional +1/+1
// counters on it equal to Arwen's toughness": a CR 614 replacement
// on the entry of a creature under the source's controller's
// control, adding the source's toughness as it is at that moment
// (counters and anthems included).
func b17OtherCreaturesYouControlEnterWithCounters(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.Controller == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
			ev.AddCounterAtETB("+1/+1", src.CurrentToughness())
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
