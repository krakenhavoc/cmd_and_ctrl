package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch27_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 27 (#389, `edhrec_rank` 2839–2939). Own file per
// the #231 convention; every package-level name carries the b27
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "this creature died" is cardDied, "another
// creature died" is diedCreature, "a creature you control dealt
// combat damage to a player" is combatDamageToPlayerBy, "an
// opponent's creature attacked you" is b17OpponentsCreatureAttackedYou,
// "you cast an instant or sorcery" is instantOrSorceryCastByYou,
// "you cast a colorless spell" is b11ColorlessSpellCastByYou, the
// devotion count is devotionTo, the wheel is b10EachPlayerWheels,
// the tapped Treasures are b13CreateTappedTreasures, the pod search's
// "what was sacrificed to pay" is b17PermanentSacrificedToPay, the
// "exile these at a later step" body is b06ExileListedCards, the
// tapped-and-attacking token path is CreateTokensAttackingForEffect,
// and the Dragon, Elf Warrior, Zombie, Food and Treasure tokens are
// b14RedDragonToken, b13GreenElfWarriorToken, BlackZombieToken,
// FoodToken and TreasureToken.

// --- token templates ---------------------------------------------

// --- costs ---------------------------------------------------------

// b27SacrificeAForest is Orcish Lumberjack's "Sacrifice a Forest" —
// any permanent the activator controls with the Forest subtype,
// effective subtypes so a Dryad Arbor or a Forest-typed dual counts.
func b27SacrificeAForest() *game.TargetSpec {
	return sacrificeSpec("a Forest", HasSubtype("Forest"))
}

// --- predicates --------------------------------------------------

// b27HasTapAbility is Magewright's Stone's clause: "creature that has
// an activated ability with {T} in its cost". Read off what the
// engine would actually offer the creature's controller — its CR 602
// activated abilities and its mana abilities, intrinsic land ones
// included — so the answer is exactly the set of creatures the Stone
// could usefully untap.
func b27HasTapAbility() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		for _, a := range game.ActivatedAbilitiesForCard(c) {
			if a.Cost.Tap {
				return true
			}
		}
		for _, m := range game.ManaAbilitiesForCard(c) {
			if m.TapCost {
				return true
			}
		}
		return false
	}
}

// --- trigger conditions ------------------------------------------

// b27SelfOrNontokenZombieYouControlDied is Headless Rider's
// condition: the Rider itself died, or a NONTOKEN Zombie its
// controller controlled did. b17SelfOrZombieYouControlDied without
// the token half — the Rider's own 2/2 Zombies dying must not chain.
func b27SelfOrNontokenZombieYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if cardDied(ev, source) {
		return true
	}
	if ev.CardID == source.InstanceID {
		return false
	}
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller == source.Controller && dead.HasSubtype("Zombie") && !IsToken(dead)
}

// b27ColorlessSpellWithManaValueAtLeastCastByYou is Sanctum of
// Ugin's condition: the controller cast a colorless spell whose mana
// value is at least n. The spell is read off the stack, X included
// (CR 202.3e), so a Walking Ballista cast with X=4 is mana value 8.
func b27ColorlessSpellWithManaValueAtLeastCastByYou(ev game.Event, source *game.Card, g *game.Game, n int) bool {
	if !b11ColorlessSpellCastByYou(ev, source, g) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	mv, ok := g.ManaValueForEffect(spell)
	return ok && mv >= n
}

// b27AnotherElfYouControlEntered is Wolverine Riders' second
// condition: another Elf entered under the source's controller's
// control. Effective subtypes, so a changeling counts.
func b27AnotherElfYouControlEntered(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !c.IsCreature() || !c.HasSubtype("Elf") {
		return game.Card{}, false
	}
	return c, true
}

// b27SelfEnteredUntapped is Gingerbread Cabin's "when this land
// enters untapped". The harvester runs with the source already on
// the battlefield, its enters-tapped replacement applied, so the
// tapped flag is the answer.
func b27SelfEnteredUntapped(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventETB && ev.CardID == source.InstanceID && !source.Tapped
}

// b27AttackedAndAnotherOpponentRemains is Legion Loyalty's
// condition: a creature the source's controller controls attacked,
// and at least one opponent other than the defending player is
// still in the game — myriad has nothing to do at a two-player
// table, and a trigger that would create nothing is not queued.
func b27AttackedAndAnotherOpponentRemains(ev game.Event, source *game.Card, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	return len(b27OtherOpponents(g, source.Controller, b17DefendingPlayer(g, ev))) > 0
}

// --- state reads ---------------------------------------------------

// b27OtherOpponents is "each opponent other than defending player":
// every seated, non-eliminated player who is neither `controller`
// nor `except`, in seat order.
func b27OtherOpponents(g *game.Game, controller, except uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller || p.ID == except {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// b27GreatestManaValueAmongArtifactsControlled is One with the
// Machine's number: the highest mana value among artifacts the
// player controls, zero with none. Post-layer types, so an
// animated or artifact-ified permanent counts; printed mana cost,
// which is what mana value reads off a permanent (CR 202.3).
func b27GreatestManaValueAmongArtifactsControlled(g *game.Game, controller uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsArtifact() {
			continue
		}
		if mv := c.ManaValue(); mv > best {
			best = mv
		}
	}
	return best
}

// b27ExiledWith is the "card exiled with this permanent" record
// (Duplicant), read back off the event log because nothing on a Card
// carries it. A card counts as exiled with `source` when its most
// recent move into exile happened while an ability of `source`
// labelled `label` was resolving: the resolution opens with an
// EventResolve naming the source and the label, and stays open until
// the next thing that cannot happen mid-resolution — another
// resolution, a cast, a mana ability, an attack, a fizzle or a step
// beginning (b13ResolutionInProgressBy's boundary set). A card that
// later left exile and came back under some other effect is not
// counted: its newer move closes the older record. Only cards still
// in exile are returned, oldest exile first, so the last entry is
// "the last card exiled with it".
//
// CR 400.7: `source` surviving as the same InstanceID across a zone
// change is a CARD fact, not an OBJECT fact — a Duplicant that leaves
// the battlefield and returns is a new object with the same
// InstanceID and no memory of what an earlier incarnation exiled.
// Event carries no Card.ObjectEpoch of its own, so the walk reads the
// one place that fact is already visible in the log: `source` getting
// a NEW EventZoneMove onto the battlefield is exactly the object
// changing (card.go's MoveCard bumps ObjectEpoch on that same move),
// so `order` — the accumulated record — is dropped right there,
// before the returning object's own abilities get a chance to add or
// read anything.
//
// Deliberately keyed on ENTERING rather than on `source` leaving (or
// on the label resolving again): a "when this leaves the battlefield"
// reader — Ossification and Angel of Serenity's own return clause,
// Bag of Holding's sacrifice ability — reads this record for the
// SAME incarnation that just departed, after the departure event is
// already in the log, and must still see it (CR 603.10, last known
// information). And Bag of Holding's discard trigger can open this
// same label's window many times across one unbroken stay on the
// battlefield — each has to add to the one record, not start a fresh
// one, since nothing has re-entered in between. Only an actual
// re-entry says "new object, no memory" (CR 400.7); merely leaving,
// or merely re-resolving the label, says nothing of the kind on its
// own.
func b27ExiledWith(g *game.Game, source uuid.UUID, label string) []uuid.UUID {
	open := false
	order := map[uuid.UUID]int{}
	for _, ev := range g.Events {
		if ev.Kind == game.EventZoneMove && ev.CardID == source && ev.NewZone == game.ZoneBattlefield {
			open = false
			order = map[uuid.UUID]int{}
		}
		switch ev.Kind {
		case game.EventResolve:
			open = ev.Source == source && ev.Label == label
			continue
		case game.EventCast, game.EventManaAbilityActivated, game.EventAttack, game.EventFizzle,
			game.EventBeginUpkeep, game.EventBeginPrecombatMain, game.EventBeginEndStep, game.EventStepBegan:
			open = false
			continue
		case game.EventZoneMove:
		default:
			continue
		}
		if ev.NewZone == game.ZoneExile && open {
			order[ev.CardID] = len(order) + 1
			continue
		}
		delete(order, ev.CardID)
	}
	if len(order) == 0 || g.Exile == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(order))
	for _, c := range g.Exile.Cards {
		if _, ok := order[c.InstanceID]; ok {
			out = append(out, c.InstanceID)
		}
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && order[out[j]] < order[out[j-1]]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// b27LastCreatureCardExiledWith is Duplicant's "the last creature
// card exiled with it": the newest of the exiled-with cards that is
// a creature card, read where it sits in exile (printed types — no
// layer applies off the battlefield).
func b27LastCreatureCardExiledWith(g *game.Game, source uuid.UUID, label string) (game.Card, bool) {
	ids := b27ExiledWith(g, source, label)
	for i := len(ids) - 1; i >= 0; i-- {
		c, ok := g.LookupCardForEffect(ids[i])
		if ok && c.IsCreature() {
			return c, true
		}
	}
	return game.Card{}, false
}

// b27CreatureTypesPlusShapeshifter is the subtype list Duplicant
// takes on: the exiled card's creature types (its land or other
// subtypes stay behind), plus Shapeshifter — "It's still a
// Shapeshifter."
func b27CreatureTypesPlusShapeshifter(exiled game.Card) []string {
	_, _, subtypes := game.ParseTypeLine(exiled.TypeLine)
	out := make([]string, 0, len(subtypes)+1)
	for _, s := range subtypes {
		if game.IsCreatureType(s) && !hasFold(out, s) {
			out = append(out, s)
		}
	}
	if !hasFold(out, "Shapeshifter") {
		out = append(out, "Shapeshifter")
	}
	return out
}

// b27DamageDealtToAfter is the damage `target` actually took from
// `source` in events logged after `after` — post-prevention, which
// is what the event carries. Zero when nothing landed.
func b27DamageDealtToAfter(g *game.Game, source, target uuid.UUID, after uint64) int {
	total := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Seq <= after {
			break
		}
		if ev.Kind == game.EventDealDamage && ev.Source == source && ev.Target == target {
			total += ev.Amount
		}
	}
	return total
}

// b27TokensCreatedByAfter lists the tokens `controller` created in
// events logged after `after`, in creation order.
func b27TokensCreatedByAfter(g *game.Game, controller uuid.UUID, after uint64) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.Events {
		if ev.Seq > after && ev.Kind == game.EventTokenCreated && ev.Actor == controller {
			out = append(out, ev.CardID)
		}
	}
	return out
}

// --- effect bodies -----------------------------------------------

// b27DrawOne is the delayed-trigger body Urza's Bauble schedules:
// the item's controller draws a card. Package-level so the trigger
// captures nothing.
func b27DrawOne(g *game.Game, item *game.StackItem) error {
	return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
}

// b27LookAtRandomCardInHand marks `viewer` as a knower of one card
// in `player`'s hand, chosen by the engine's keyed random stream. "Look at" is
// not "reveal": only the viewer learns the card. An empty hand shows
// nothing.
func b27LookAtRandomCardInHand(ctx *Context, viewer, player uuid.UUID) {
	p := ctx.PlayerByID(player)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return
	}
	ids := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	pick := randomPick(ctx, ids, 1)
	if len(pick) == 0 {
		return
	}
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == pick[0] {
			p.Hand.Cards[i].AddKnower(viewer)
			return
		}
	}
}

// b27DealDamageWithExcess is Hell to Pay's damage: `amount` from the
// source to the creature `target`, returning how much of it was
// EXCESS — CR 120.4a's "more than lethal": what landed minus the
// damage the creature still needed to die (toughness less damage
// already marked, floored at zero). Read after the damage lands so a
// prevention shield reduces the excess along with the damage.
func b27DealDamageWithExcess(ctx *Context, target uuid.UUID, amount int) (int, error) {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || amount <= 0 {
		return 0, nil
	}
	lethal := c.CurrentToughness() - c.DamageMarked
	if lethal < 0 {
		lethal = 0
	}
	cursor := b25LastEventSeq(ctx.Game)
	if err := (DealDamage{Source: ctx.Source(), Target: target, Amount: amount}).Apply(ctx); err != nil {
		return 0, err
	}
	excess := b27DamageDealtToAfter(ctx.Game, ctx.Source(), target, cursor) - lethal
	if excess < 0 {
		excess = 0
	}
	return excess, nil
}

// b27ExileTopUntilTotalManaValue is Tasha's Hideous Laughter for one
// player: exile cards from the top of their library until the exiled
// cards' total mana value reaches `threshold`. An unbounded run (N
// left 0 with an Until): a library that totals less than the
// threshold is exiled whole and the run just stops — running out is
// not a draw, so nobody loses for it (CR 701.17b, CR 704.5b).
//
// #1161: the run SEQUENCES, so a commander that pauses on CR 903.9
// holds the rest of this seat's run rather than being walked past, and
// the total counts the cards that really reached exile.
//
// #1176 makes each repetition its own instruction, which changes
// nothing here: exiling the top card of a library is not a mill
// (CR 701.17a), so no repetition opens a mill-amount window and
// nothing can double one.
func b27ExileTopUntilTotalManaValue(ctx *Context, player uuid.UUID, threshold int) error {
	if ctx.PlayerByID(player) == nil {
		return nil
	}
	return MillToZone{
		Player: player,
		To:     game.ZoneExile,
		Until:  UntilTotalManaValue(threshold),
	}.Apply(ctx)
}

// b27LegionLoyaltyLabel is the stack label of the myriad trigger
// Legion Loyalty grants.
const b27LegionLoyaltyLabel = "Legion Loyalty — myriad: token copies attacking each other opponent"

// b27MyriadCopies is the myriad body (CR 702.116) for one attacking
// creature: for each opponent other than the defending player, a
// token copy of the attacker enters tapped and attacking that
// player, and the copies are exiled at the beginning of the end of
// combat step through a delayed trigger. The attacker is copied
// wherever it now sits — a creature that died in response is copied
// from the graveyard, its last-known printed values (CR 702.116a
// makes the copies regardless).
//
// The attacker and the defending player are computed once, at trigger
// (Build) time, and carried on the item's Params (Object, Player)
// rather than baked into a per-instance closure (ADR 0041 P9): the
// defending player is fixed at declaration (CR 506.4) and a live
// re-derivation at resolution could answer differently if the
// attacked planeswalker or battle changed hands in response.
func b27MyriadCopies(g *game.Game, item *game.StackItem) error {
	attacker, defender := item.Params.Object.ID, item.Params.Player
	ctx := NewContext(g, item)
	tmpl, ok := TokenCopyTemplate(g, attacker)
	if !ok {
		return nil
	}
	tmpl.Tapped = true
	cursor := b25LastEventSeq(g)
	for _, opp := range b27OtherOpponents(g, item.Controller, defender) {
		if err := g.CreateTokensAttackingForEffect(item.Controller, tmpl, 1, opp); err != nil {
			return err
		}
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: "Legion Loyalty — exile the myriad tokens",
		Cards: tokens,
		Body:  exileListedCardsBody,
	}.Apply(ctx)
}

// b27GooseMotherAttackLabel is the stack label of The Goose Mother's
// attack trigger.
const b27GooseMotherAttackLabel = "The Goose Mother — sacrifice a Food to draw a card"

// b27SacrificeChosenThenDraw is The Goose Mother's "you may sacrifice
// a Food. If you do, draw a card": the Food chosen when the trigger
// went on the stack is sacrificed if it is still there, and the
// draw follows only from the sacrifice.
func b27SacrificeChosenThenDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (SacrificePermanent{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	return nil
}

// b27SacrificeSelfThenTutorColorlessCreature is Sanctum of Ugin's
// "you may sacrifice this land. If you do, search your library for a
// colorless creature card, reveal it, put it into your hand, then
// shuffle" — the "may" was answered when the trigger fired; here the
// Sanctum is sacrificed if it is still on the battlefield, and the
// search happens only if it was.
func b27SacrificeSelfThenTutorColorlessCreature(g *game.Game, item *game.StackItem) error {
	// #1432: "if you do" — a Sanctum that left and came back is not
	// sacrificed, so it tutors nothing.
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return SearchLibrary{
		Player:    item.Controller,
		Predicate: func(c game.Card) bool { return c.IsCreature() && c.IsColorless() },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Sanctum of Ugin — a colorless creature card",
	}.Apply(ctx)
}

// b27OswaldSearch is Oswald Fiddlebender's body — Birthing Pod for
// artifacts: the sacrificed artifact's mana value plus one is the
// only mana value the search accepts, and the find enters the
// battlefield. The sacrifice was paid at announce and is read back
// off the log; a token's mana value is zero, so a sacrificed
// Treasure fetches a one-drop, as printed.
func b27OswaldSearch(g *game.Game, item *game.StackItem) error {
	sacrificed, ok := b17PermanentSacrificedToPay(g, item)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(sacrificed)
	if !ok {
		return nil
	}
	want := c.ManaValue() + 1
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			return c.IsArtifact() && c.ManaValue() == want
		},
		Dest:    game.ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Reason:  "Oswald Fiddlebender — an artifact card with mana value " + strconv.Itoa(want),
	}.Apply(NewContext(g, item))
}

// b27ReturnChosenGraveyardCardToHand is Memorial to Folly's
// activation: the creature card chosen at announce returns to its
// owner's hand if it is still in the graveyard.
func b27ReturnChosenGraveyardCardToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
	}
	return nil
}

// b27DuplicantLabel is the stack label of Duplicant's imprint
// trigger — the "exiled with" record keys on it.
const b27DuplicantLabel = "Duplicant — exile target nontoken creature"

// b27ExileChosenTarget exiles whatever the pick_target prompt
// stamped into the item, if it is still legal.
func b27ExileChosenTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ExileTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

// b27LoseOneAndYouGainOne is Revenge of Ravens' drain: the attacker's
// controller — stamped onto item.Params.Player by a fill-in Build —
// loses 1 life and the item's controller gains 1, in that order.
func b27LoseOneAndYouGainOne(g *game.Game, item *game.StackItem) error {
	victim := item.Params.Player
	ctx := NewContext(g, item)
	if p := g.PlayerByIDForEffect(victim); p != nil && !p.Eliminated {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1); err != nil {
			return err
		}
	}
	return GainLife{Player: item.Controller, Amount: 1}.Apply(ctx)
}

// --- replacements ------------------------------------------------

// b27OtherCreaturesYouControlEnterWithACounter is Renata's "each
// other creature you control enters with an additional +1/+1 counter
// on it" — Arwen's replacement (b17OtherCreaturesYouControlEnterWithCounters)
// with a fixed one counter instead of the source's toughness.
func b27OtherCreaturesYouControlEnterWithACounter(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.Controller == src.Controller
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

// --- statics -----------------------------------------------------

// b27LoseKeyword is a layer-6 "lose <keyword>" — Archetype of
// Aggression's "creatures your opponents control lose trample". It
// strips the keyword from whatever the layers had granted so far in
// timestamp order; a grant from a source newer than the Archetype
// lands after it and is not caught, which is the declared gap.
func b27LoseKeyword(applies func(target *game.Card, g *game.Game, source *game.Card) bool, keyword string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			kept := c.Abilities[:0]
			for _, a := range c.Abilities {
				if !equalFoldASCIIEffects(a, keyword) {
					kept = append(kept, a)
				}
			}
			c.Abilities = kept
		},
	}
}

// b27CreaturesOpponentsControl is "creatures your opponents control".
func b27CreaturesOpponentsControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller != source.Controller
}

// b27OtherArtifactCreaturesYouControl is Chief of the Foundry's
// "other artifact creatures you control" — post-layer types, so an
// animated artifact counts.
func b27OtherArtifactCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID &&
		target.Controller == source.Controller &&
		target.IsCreature() && target.IsArtifact()
}

// b27DuplicantStatics are Duplicant's two continuous effects, both
// gated on a creature card being exiled with it: layer 4 sets its
// creature types to the last such card's (plus Shapeshifter), layer
// 7b sets its power and toughness to that card's printed values.
// Not a CDA — the ability is conditional (CR 604.3a) — so counters
// and anthems still apply on top.
func b27DuplicantStatics() []game.StaticAbility {
	return []game.StaticAbility{
		{
			Layer:     game.Layer4Type,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if exiled, ok := b27LastCreatureCardExiledWith(g, source.InstanceID, b27DuplicantLabel); ok {
					c.SetSubtypes(b27CreatureTypesPlusShapeshifter(exiled))
				}
			},
		},
		{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if exiled, ok := b27LastCreatureCardExiledWith(g, source.InstanceID, b27DuplicantLabel); ok {
					c.Power = exiled.Power
					c.Toughness = exiled.Toughness
				}
			},
		},
	}
}
