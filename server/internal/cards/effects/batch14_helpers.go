package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch14_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 14 (#307, `edhrec_rank` 1525–1625). Own file per
// the #231 convention; every package-level name carries the b14
// prefix because batch 13 is landing beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "another permanent you control entered" is
// enteredUnderYourControl, "whenever you cast an instant or sorcery"
// is instantOrSorceryCastByYou, "sacrifice a Goblin" as a cost is
// b12SacrificeAGoblin, "a +1/+1 counter on each creature you
// control" is b11PutCounterOnEachCreatureYouControl, "historic" is
// b09IsHistoric, "lands you control" is b10LandsControlled, "an
// Equipment card" is b09IsEquipmentCard, "legendary" is
// b05Legendary, "N damage to each opponent" is damageToEachOpponent,
// the counters a permanent had when it left are b13LastKnownCounters
// (and its walk, b13LastKnownCounterWalk), and the per-ability "one
// or more" dedup is OncePerBatch.

// --- token templates ---------------------------------------------

// b14RedDragonToken is the red Dragon with flying two cards in this
// batch make at different sizes: Utvara Hellkite's 6/6 and
// Dragonmaster Outcast's 5/5.
func b14RedDragonToken(size int) game.Card {
	return game.Card{
		Name:      "Dragon",
		TypeLine:  "Token Creature — Dragon",
		Power:     size,
		Toughness: size,
		Colors:    []string{"R"},
		Keywords:  []string{"flying"},
	}
}

// b14ReplicatedRingToken is Replicating Ring's payout: a colorless
// snow artifact named Replicated Ring with "{T}: Add one mana of any
// color". The mana ability rides the template the way Treasure's
// does, because a token has no oracle ID for the catalog to key one
// on. "Any color" is the printed text, so the pipe keeps its five
// options rather than narrowing to the commander's identity.
func b14ReplicatedRingToken() game.Card {
	return game.Card{
		Name:     "Replicated Ring",
		TypeLine: "Token Snow Artifact",
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
	}
}

// --- trigger conditions ------------------------------------------

// b14DragonYouControlAttacked is Utvara Hellkite's condition: a
// creature the source's controller controls was declared as an
// attacker and it is a Dragon. The Hellkite itself qualifies — the
// printed text is "a Dragon you control", not "another". Effective
// subtypes, so a changeling counts.
func b14DragonYouControlAttacked(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Dragon")
}

// b14SelfOrDragonYouControlEntered is Ganax's condition: the source
// itself entered, or another Dragon entered under the source's
// controller's control.
func b14SelfOrDragonYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventETB {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype("Dragon")
}

// b14OpponentsEndStepBegan is "at the beginning of each opponent's
// end step" — Archfiend of Depravity. The event's Actor is the
// active player, whose end step it is.
func b14OpponentsEndStepBegan(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventBeginEndStep && ev.Actor != uuid.Nil && ev.Actor != source.Controller
}

// b14PermanentYouControlLeft is Resourceful Defense's condition: a
// permanent the source's controller controlled left the battlefield,
// for anywhere. The card is read post-move — the Controller field
// survives the move — so a token that ceased to exist on leaving is
// not seen, which is the diedCreature posture and weaker than
// printed. Returns the last-known counters it carried, keyed by
// kind, so the "if it had counters on it" intervening-if and the
// effect read one snapshot.
func b14PermanentYouControlLeft(ev game.Event, source *game.Card, g *game.Game) (map[string]int, bool) {
	if ev.Kind != game.EventLTB || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return nil, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller {
		return nil, false
	}
	counters := b14LastKnownCounterKinds(g, ev.CardID)
	if len(counters) == 0 {
		return nil, false
	}
	return counters, true
}

// --- counters read back off the log ------------------------------

// b14LastKnownCounterKinds is every kind of counter `cardID` had when
// it last left the battlefield, with its count — Resourceful
// Defense's "those counters". Batch 13's b13LastKnownCounterWalk
// (the walk that stops at the card's arrival on the battlefield, but
// still reads the "enters with" placements a beat before it), taking
// the most recent total per kind and dropping kinds that were removed
// to nothing. The single-kind reader is b13LastKnownCounters.
func b14LastKnownCounterKinds(g *game.Game, cardID uuid.UUID) map[string]int {
	out := map[string]int{}
	seen := map[string]bool{}
	b13LastKnownCounterWalk(g, cardID, func(ev game.Event) bool {
		if seen[ev.Label] {
			return true
		}
		seen[ev.Label] = true
		if ev.Amount > 0 {
			out[ev.Label] = ev.Amount
		}
		return true
	})
	return out
}

// --- predicates --------------------------------------------------

// b14SpellTargetsAPermanentYouControl is Rebuff the Wicked's clause:
// a spell on the stack at least one of whose targets is a permanent
// the caster controls ON THE BATTLEFIELD. The zone check is
// b10SpellTargetsYouOrACreatureYouControl's: a card in a graveyard
// keeps its Controller field, and without it a spell aimed at a
// creature that has already left could still be countered. A spell
// that targets the caster (a player) does not qualify — the printed
// clause is "a permanent", and Rebuff cannot save you from a burn
// spell to the face.
func b14SpellTargetsAPermanentYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			z := g.FindCardZoneForEffect(t.ID)
			if z == nil || z.Kind != game.ZoneBattlefield {
				continue
			}
			if tc, ok := g.LookupCardForEffect(t.ID); ok && tc.Controller == caster {
				return true
			}
		}
		return false
	}
}

// b14Historic is b09IsHistoric as a target / sweep predicate —
// Desynchronization's "permanent that's not historic" is
// Not(b14Historic()).
func b14Historic() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return b09IsHistoric(c) }
}

// --- board reads -------------------------------------------------

// b14InYourMainPhase reports whether it is `player`'s turn and a main
// phase — Return to Dust's "if you cast this spell during your main
// phase". Read at resolution rather than recorded at announce, which
// is the same answer: the stack has to be empty for a step to end,
// so a spell resolves in the step it was cast in.
func b14InYourMainPhase(g *game.Game, player uuid.UUID) bool {
	if g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		return false
	}
	if g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	return active != nil && active.ID == player
}

// b14CreaturesControlled counts the creatures `player` controls.
func b14CreaturesControlled(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			n++
		}
	}
	return n
}

// b14CardsInOpponentsGraveyards counts the cards in every graveyard
// that is not `controller`'s — Consuming Aberration's
// characteristic-defining size. Walks the live seats because it runs
// inside a layer recompute (b10LandsControlled's shape). An
// eliminated player's graveyard still counts: they are still an
// opponent as far as the printed text goes, and their cards are
// still in it.
func b14CardsInOpponentsGraveyards(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.ID == controller || p.Graveyard == nil {
			continue
		}
		n += p.Graveyard.Size()
	}
	return n
}

// b14HandSize is the number of cards in `player`'s hand, or zero
// for a player who is not seated.
func b14HandSize(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return 0
	}
	return p.Hand.Size()
}

// --- effect bodies -----------------------------------------------

// b14DamageEachOpponentGainThatMuch is Creeping Bloodsucker's body:
// the source deals n damage to each opponent, then its controller
// gains life equal to the damage actually dealt — so a prevention
// shield that ate some of it shrinks the gain and a damage doubler
// grows it, as printed.
//
// #807: the total comes from DealDamageEachThenForEffect's
// continuation, which reports what really landed. This used to read
// each opponent's life total back on the line after damaging them, and
// a damage event runs the CR 614 window — so two DIFFERENT damage
// replacements on one opponent pause it on a CR 616 ordering prompt,
// the read-back happens before the prompt is answered, and that
// opponent counts as having taken nothing. Same bug #793 fixed on the
// life side, same shape of fix.
func b14DamageEachOpponentGainThatMuch(g *game.Game, item *game.StackItem, n int) error {
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	controller, source := item.Controller, ctx.Source()
	return g.DealDamageEachThenForEffect(source, ctx.Opponents(), n, func(g *game.Game, totalDealt int) error {
		if totalDealt <= 0 {
			return nil
		}
		return g.ChangePlayerLifeForEffect(source, controller, totalDealt)
	})
}

// b14EachOpponentMillsUntilLand is Consuming Aberration's rider:
// every opponent reveals cards from the top of their library until
// they reveal a land card, then puts all of those cards into their
// graveyard. Milled one card at a time so the type of each is read
// from the top of the library before it moves; an opponent whose
// library holds no land mills the whole thing, as printed (and takes
// the empty-library loss only if they then have to draw).
func b14EachOpponentMillsUntilLand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if err := b14MillUntilLand(ctx, opp); err != nil {
			return err
		}
	}
	return nil
}

// b14MillUntilLand reveals cards off the top of one player's library
// until a land has been revealed, then puts all of them into the
// graveyard.
//
// S22: the reveal is now a real reveal. This helper used to mill
// straight to the graveyard and declare the reveal unmodelled on the
// grounds that the graveyard is public, so the table sees the same
// cards a moment later — true, and not the same thing. What the table
// could not see was WHICH cards this trigger turned over as opposed
// to which arrived from something else resolving in the same window,
// and on an opponent's board a twenty-card run reads as a wall of
// graveyard motion with no cause attached to it.
//
// The run is measured first, read-only, so the whole thing is one
// announcement rather than one per card. The mill loop underneath is
// unchanged and still reads the top of the library on every pass, so
// anything that changes the library mid-mill is handled the way it
// always was; the reveal names what was on top when the trigger
// resolved, which is what the card reveals.
func b14MillUntilLand(ctx *Context, player uuid.UUID) error {
	if p := ctx.PlayerByID(player); p != nil && p.Library != nil {
		run := make([]uuid.UUID, 0, 8)
		for i := len(p.Library.Cards) - 1; i >= 0; i-- {
			run = append(run, p.Library.Cards[i].InstanceID)
			if p.Library.Cards[i].IsLand() {
				break
			}
		}
		if err := (RevealCards{
			Player: player,
			Cards:  run,
			Reason: "reveal from the top until a land card",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	for i := 0; i < 1000; i++ {
		p := ctx.PlayerByID(player)
		if p == nil || p.Library == nil || p.Library.Size() == 0 {
			return nil
		}
		top := p.Library.Cards[len(p.Library.Cards)-1]
		if err := (MillCards{Player: player, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if top.IsLand() {
			return nil
		}
	}
	return nil
}

// b14PlayerSacrificesAllButN is Archfiend of Depravity's body: the
// player keeps up to `keep` creatures and sacrifices the rest, as
// their own choice. The engine's sacrifice prompt picks ONE
// permanent, so "choose two to keep" is asked as (creatures − keep)
// prompts to sacrifice one each — the same decision from the other
// side, made by the same player, with the option list trimmed after
// every answer. Queued in one go, so a creature that leaves between
// two answers for some other reason (a lord's death shrinking a
// toughness to zero) is not deducted from what is still owed; that
// corner is declared on the card.
func b14PlayerSacrificesAllButN(g *game.Game, source, player uuid.UUID, keep int, reason string) {
	excess := b14CreaturesControlled(g, player) - keep
	for i := 0; i < excess; i++ {
		if g.PlayerSacrificesForEffect(source, player, sacrificeSpec("a creature", Creature()), reason) == 0 {
			return
		}
	}
}

// b14MoveAllCounters moves every counter on `from` onto `to`, kind
// by kind — Resourceful Defense's activated ability with "any
// number" read as all of them. Both must be on the battlefield; the
// removal goes through the same counter primitive as the placement,
// so a doubler sees the placement and nothing sees a removal it
// should not.
func b14MoveAllCounters(ctx *Context, from, to uuid.UUID) error {
	src, ok := ctx.Game.LookupCardForEffect(from)
	if !ok || len(src.Counters) == 0 {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(to); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	kinds := make([]string, 0, len(src.Counters))
	for kind := range src.Counters {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		n := src.Counters[kind]
		if n <= 0 {
			continue
		}
		if err := (AddCounter{Target: from, Kind: kind, N: -n}).Apply(ctx); err != nil {
			return err
		}
		if err := (AddCounter{Target: to, Kind: kind, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b14PutCounters puts a set of counters, keyed by kind, on a
// battlefield permanent — Resourceful Defense's trigger. Kinds are
// applied in sorted order so the events are deterministic.
func b14PutCounters(ctx *Context, target uuid.UUID, counters map[string]int) error {
	if z := ctx.Game.FindCardZoneForEffect(target); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	kinds := make([]string, 0, len(counters))
	for kind := range counters {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		if counters[kind] <= 0 {
			continue
		}
		if err := (AddCounter{Target: target, Kind: kind, N: counters[kind]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b14ReturnDistinctNamesFromGraveyard is Eerie Ultimatum's body: put
// every still-legal target onto the battlefield under its owner's
// control, skipping any card whose name has already come back this
// resolution — "with different names". Announce order, so which of
// two same-named picks returns is the one picked first.
func b14ReturnDistinctNamesFromGraveyard(ctx *Context) error {
	seen := map[string]bool{}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok || seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
