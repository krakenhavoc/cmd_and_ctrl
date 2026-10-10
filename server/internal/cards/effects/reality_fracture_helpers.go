package effects

import (
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_helpers.go — small helpers shared by the Reality
// Fracture (FRA / FRC) cards. Append-only.

// rfCardsMilledThisTurn is the number of cards that were put into
// `owner`'s graveyard from their library this turn (Cruel
// Calculations). It reads this turn's events, which are bounded at the
// real turn boundary, and counts one per move: a card that left the
// graveyard and was milled again counts twice.
//
// A mill announces itself as an EventMill; other library-to-graveyard
// routes announce a plain EventZoneMove. A route that emitted both for
// one card would count it twice, so an event naming the same card as
// the one counted immediately before it is the same move.
//
// Caller holds g.mu.
func rfCardsMilledThisTurn(g *game.Game, owner uuid.UUID) int {
	n := 0
	var lastCounted uuid.UUID
	var lastIdx = -2
	for i, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventMill && ev.Kind != game.EventZoneMove {
			continue
		}
		if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard {
			continue
		}
		if ev.CardID == lastCounted && i-lastIdx <= 2 {
			lastIdx = i
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && c.Owner == owner {
			n++
			lastCounted, lastIdx = ev.CardID, i
		}
	}
	return n
}

// AttackedThisTurn is the target predicate "creature that attacked this
// turn" (Hexhaven Dueling Arena): the per-object attack tally the turn
// keeps, so a creature that attacked and then left combat still counts
// and a creature that left and came back (a new object) does not.
func AttackedThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.TimesAttackedThisTurn(c.InstanceID) > 0
	}
}

// BecomeUnprepared is "[it] becomes unprepared" (CR 722.3b): the
// permanent loses the prepared designation and the copy of its prepare
// spell kept in exile. Not an error when nothing happens — a permanent
// that isn't prepared, or has left the battlefield, simply stays as it is.
type BecomeUnprepared struct {
	Target uuid.UUID
}

func (b BecomeUnprepared) Apply(ctx *Context) error {
	if ctx.isNewSourceObject(b.Target) { // #1432
		return nil
	}
	return nothingIfGone(ctx.Game.UnprepareForEffect(b.Target))
}

// seedSutureResolve is Seed Suture, the prepare spell of both
// Blossom-Blessed Angel and Emergency Phytomedic: "Put a +1/+1 counter
// on target creature. You gain 1 life."
func seedSutureResolve(_ *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
}

// soulTetherResolve is Soul Tether, the prepare spell of both Heartwood
// Crafter and Konstrari Improviser: "Create a Heartwood token."
func soulTetherResolve(_ *game.StackItem, ctx *Context) error {
	return CreateToken{Template: HeartwoodToken(), N: 1}.Apply(ctx)
}

// castAPreparedSpell reports whether the spell the cast event named is a
// "prepared spell": the copy of a prepare spell its controller cast out
// of exile (CR 722.3c). Those are the only copies the cast path ever
// puts on the stack, so a cast spell that is a copy is a prepared spell.
func castAPreparedSpell(g *game.Game, spellID uuid.UUID) bool {
	if c, ok := g.LookupCardForEffect(spellID); ok && c.PrepareCopy {
		return true
	}
	item := g.StackItemForEffect(spellID)
	return item != nil && item.IsCopy
}

// returnTargetGraveyardCardToHand is the body of "return target <card>
// from your graveyard to your hand": the first legal card target the
// trigger or spell announced goes to its owner's hand. Evolution
// Witness and Carnivorous Cultivator both end this way.
func returnTargetGraveyardCardToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
		}
	}
	return nil
}

// frJaceSubtype is the planeswalker type "behold a Jace" and "Jace
// planeswalkers" name.
const frJaceSubtype = "Jace"

// frYouControlAJace reports whether `player` controls a permanent with the
// Jace subtype — a Jace planeswalker card or the Jace token. It is the
// "choose a Jace you control" half of behold. Caller holds g.mu.
func frYouControlAJace(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype(frJaceSubtype) {
			return true
		}
	}
	return false
}

// frEntersTappedUnlessYouBeholdAJace is "As this land enters, you may
// behold a Jace. If you don't, this land enters tapped." (Theorist's
// Sanctum). Behold is "choose a Jace you control or reveal a Jace card
// from your hand". Choosing a Jace you control is free and strictly
// better than the tapped alternative, so a controlled Jace makes the
// replacement not apply at all (no prompt, enters untapped); otherwise
// the reveal-from-hand prompt of the reveal-lands is asked.
func frEntersTappedUnlessYouBeholdAJace(name string) game.ReplacementEffect {
	rep := EntersTappedUnlessYouRevealFromHand(name, "a Jace card",
		func(c game.Card) bool { return c.HasSubtype(frJaceSubtype) })
	applies := rep.AppliesTo
	rep.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !applies(ev, g, src) {
			return false
		}
		return !frYouControlAJace(g, ev.Actor)
	}
	return rep
}

// frPutLoyaltyCounterOnEachPlaneswalkerYouControl is "put a loyalty
// counter on each planeswalker you control" (Way of the Mentor, Way of
// the Necromancer). The set is fixed when the trigger resolves, and the
// walkers are the trigger controller's.
func frPutLoyaltyCounterOnEachPlaneswalkerYouControl(g *game.Game, item *game.StackItem) error {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsPlaneswalker() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := nothingIfGone(g.AddCounterForEffect(id, game.CounterLoyalty, 1)); err != nil {
			return err
		}
	}
	return nil
}

// frYouActivatedALoyaltyAbility is "whenever you activate a loyalty
// ability": a loyalty ability (CR 606.2), printed or granted, announced
// by the source's controller.
func frYouActivatedALoyaltyAbility(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventActivateAbility && ev.Loyalty && ev.Actor == source.Controller
}

// frLoyaltyAbilityRemovedAtLeast reports whether the loyalty ability an
// activation event announces has a loyalty cost of −n or lower, i.e. "you
// removed n or more loyalty counters to activate it". The cost is read off
// the planeswalker's current abilities by the event's label, because the
// loyalty cost path records no counters-removed fact. A planeswalker that
// has already left reads as false (weaker than printed, never stronger).
func frLoyaltyAbilityRemovedAtLeast(ev game.Event, g *game.Game, n int) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	// #1944: a −X cost removed the X its activation announced, which
	// the ability's stack item carries.
	x := 0
	if it := g.StackItemForEffect(ev.StackItemID); it != nil && it.XValue > 0 {
		x = it.XValue
	}
	abs, _ := game.ActivatedAbilitiesWithOrigins(c)
	for _, ab := range abs {
		if ab.Label == ev.Label && ab.Cost.Loyalty != nil && ab.Cost.LoyaltyDelta(x) <= -n {
			return true
		}
	}
	return false
}

// frDealDamageWithExcess deals `amount` damage from the resolving
// source to a creature or planeswalker and returns the EXCESS (CR
// 120.4a): what landed beyond the lethal amount. Lethal is the toughness
// less damage already marked for a creature, the loyalty for a
// planeswalker, and the larger of the two for something that is both.
// Read after the damage lands, so prevention lowers the excess too.
func frDealDamageWithExcess(ctx *Context, target uuid.UUID, amount int) (int, error) {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || amount <= 0 {
		return 0, nil
	}
	lethal := 0
	if c.IsCreature() {
		lethal = c.CurrentToughness() - c.DamageMarked
	}
	if c.IsPlaneswalker() && c.Counters[game.CounterLoyalty] > lethal {
		lethal = c.Counters[game.CounterLoyalty]
	}
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

// --- fra-planeswalker-a ---------------------------------------------

// youControlAPlaneswalker is the Reality Fracture land condition —
// "unless you control a planeswalker". Read once as the land enters
// (CR 614.12), against effective types so a planeswalker that is also
// something else still counts.
func youControlAPlaneswalker(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsPlaneswalker() {
			return true
		}
	}
	return false
}

// AjanisPridemateToken is Ajani Resolute's −4 token: a 2/2 white Cat
// Soldier named Ajani's Pridemate with "Whenever you gain life, put a
// +1/+1 counter on this token."
func AjanisPridemateToken() game.Card { return tokenFromCatalog(printedAjanisPridemateToken) }

// printedAjanisPridemateToken is that token as PRINTED, ability included.
func printedAjanisPridemateToken() tokenTemplate {
	return tokenTemplate{
		Slug: "ajanis-pridemate",
		Card: game.Card{
			Name:      "Ajani's Pridemate",
			TypeLine:  "Token Creature — Cat Soldier",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"W"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Ajani's Pridemate — put a +1/+1 counter on it", putCounterOnSelf),
		},
		Text: "Whenever you gain life, put a +1/+1 counter on this token.",
	}
}

// surveilKeepNoncreatureNonland is Chandra, Chill of Compliance's first
// +1: "Surveil 1. If you put a noncreature, nonland card into your
// graveyard this way, put that card into your hand."
//
// Surveil.Then is told which cards were put away only by what changed,
// so the graveyard is read before the prompt opens and again when the
// player has answered: every card that is new in it was put there by the
// surveil (nothing else moves a card into this graveyard while the
// prompt is open), and each noncreature, nonland one is returned to hand.
func surveilKeepNoncreatureNonland(g *game.Game, item *game.StackItem) error {
	me := item.Controller
	p := g.PlayerByIDForEffect(me)
	if p == nil {
		return nil
	}
	before := map[uuid.UUID]bool{}
	for _, c := range p.Graveyard.Cards {
		before[c.InstanceID] = true
	}
	g.SurveilThenForEffect(me, item.SourceCardID, 1, func(g *game.Game) error {
		q := g.PlayerByIDForEffect(me)
		if q == nil {
			return nil
		}
		var back []uuid.UUID
		for _, c := range q.Graveyard.Cards {
			if !before[c.InstanceID] && !c.IsCreature() && !c.IsLand() {
				back = append(back, c.InstanceID)
			}
		}
		ctx := NewContext(g, item)
		for _, id := range back {
			if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

// --- fra-prepare-b -------------------------------------------------

// fraThresholdMet is threshold (ability word): seven or more cards in
// the controller's graveyard.
func fraThresholdMet(g *game.Game, controller uuid.UUID) bool {
	return b31GraveyardSize(g, controller) >= 7
}

// fraThresholdSelfPT is "Threshold — This creature gets +P/+T as long
// as there are seven or more cards in your graveyard": a layer 7c
// self-modifier read live on every recompute.
func fraThresholdSelfPT(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && fraThresholdMet(g, source.Controller)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// fraThresholdSelfKeywords is "Threshold — This creature has <keywords>
// as long as there are seven or more cards in your graveyard".
func fraThresholdSelfKeywords(keywords ...string) game.StaticAbility {
	return b16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && fraThresholdMet(g, source.Controller)
	}, keywords...)
}

// fraBecomesPreparedAtUpkeep is "At the beginning of your upkeep, if
// this creature isn't prepared, it becomes prepared." The condition is
// an intervening if (CR 603.4): checked when the upkeep begins, so a
// creature that is already prepared never puts the ability on the
// stack, and BecomePrepared re-checks as it resolves.
func fraBecomesPreparedAtUpkeep(name string) game.TriggeredAbility {
	return On(game.EventBeginUpkeep,
		func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
			return ByYou(ev, source, lki, g) && !g.IsPreparedForEffect(source.InstanceID)
		},
		name+" — becomes prepared",
		func(g *game.Game, item *game.StackItem) error {
			return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
		})
}

// omitVariablesResolve is Omit Variables, the prepare spell of Paradox
// Shaper, Theorix Metamage and Void Extrapolator: "Mill three cards."
func omitVariablesResolve(_ *game.StackItem, ctx *Context) error {
	return MillCards{N: 3}.Apply(ctx)
}

// peerReviewResolve is Peer Review, the prepare spell of Prudent
// Fateseer and Semester Foreseer: "Create a 2/2 colorless Wizard Soldier
// creature token named Cadet. Surveil 1." The surveil runs after the
// token is made, in printed order.
func peerReviewResolve(_ *game.StackItem, ctx *Context) error {
	if err := (CreateToken{Template: TokenCard("2/2 colorless Wizard Soldier named Cadet"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	return Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
}

// viciousVerseResolve is Vicious Verse, the prepare spell of Stingerquill
// Voxmancer and Whiplash Wordsmith: "Vicious Verse deals 1 damage to
// target opponent."
func viciousVerseResolve(item *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
		}
	}
	return nil
}

// rfExileTargetThenRevealCreatureOrPlaneswalker is the shared body of
// Jace, Multiverse Architect's −3 and Identity Echo's ability:
//
//	"Exile [another] target creature or planeswalker you control. Reveal
//	 cards from the top of your library until you reveal a creature or
//	 planeswalker card. Put that card onto the battlefield and the rest
//	 on the bottom of your library in a random order."
//
// The reveal is the exile's continuation, not the next line: the exile
// opens the CR 614 window, and a commander's owner may be asked about
// the command zone first (CR 903.9). The reveal is not an "if you do" —
// it happens even if the exile was replaced.
func rfExileTargetThenRevealCreatureOrPlaneswalker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ExileTarget{
		Target: id,
		Then: func(next *Context, _ bool) error {
			return RevealUntilThenPutOntoBattlefield{
				Match:  func(c game.Card) bool { return c.IsCreature() || c.IsPlaneswalker() },
				Reason: "revealed until a creature or planeswalker card",
			}.Apply(next)
		},
	}.Apply(ctx)
}

// rfPutCardFromHandOnBottom is "put a card from your hand on the bottom
// of your library": the controller picks one card (the whole hand is the
// pick when it holds one card; an empty hand is not asked), and it is
// tucked on the bottom.
func rfPutCardFromHandOnBottom(ctx *Context, question string) error {
	player := ctx.Controller()
	hand := allHandCardIDs(ctx.Game, player)
	if len(hand) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: question,
		Cards:    hand,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneHand,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			for _, id := range picked {
				if err := g.TuckToLibraryThenForEffect(id, game.TuckOptions{ToBottom: true}, nil); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return nil
}

// rfRemoveUpToCounters is "remove up to N counters from <permanent>":
// the chooser removes them one at a time, naming the kind each time,
// and may stop early. One option-pick per counter, so a permanent with
// several kinds (a planeswalker with a stun counter) is the chooser's
// call, as the rules make it. A permanent that has left the battlefield,
// or has no counters left, ends the run.
func rfRemoveUpToCounters(g *game.Game, chooser, source, target uuid.UUID, n int, question string) error {
	if n <= 0 {
		return nil
	}
	if _, ok := g.PermanentRefForEffect(target); !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	var kinds []string
	for kind, count := range c.Counters {
		if count > 0 {
			kinds = append(kinds, kind)
		}
	}
	if len(kinds) == 0 {
		return nil
	}
	sort.Strings(kinds)
	options := []game.ChoiceOption{{Label: "Stop removing counters"}}
	for _, kind := range kinds {
		options = append(options, game.ChoiceOption{Label: fmt.Sprintf("Remove a %s counter", kind)})
	}
	g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		Options:  options,
		Then: func(g *game.Game, index int) error {
			if index <= 0 || index > len(kinds) {
				return nil
			}
			if err := g.AddCounterForEffect(target, kinds[index-1], -1); err != nil {
				return err
			}
			return rfRemoveUpToCounters(g, chooser, source, target, n-1, question)
		},
	})
	return nil
}

// rfOpponentCreatureWouldDie is the AppliesTo of "if a creature an
// opponent controls would die": a battlefield-to-graveyard move of a
// creature whose controller is not the source's controller.
func rfOpponentCreatureWouldDie(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if ev.Kind != game.RepEventMove || ev.OldZone != game.ZoneBattlefield || ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller != src.Controller
}

// rfDamageFirstLegalTarget is "it deals N damage to target …" as an
// activated ability's whole body: the still-legal first target takes the
// damage from the ability's source, and a target that left in response
// takes nothing.
func rfDamageFirstLegalTarget(amount int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		legal := ctx.LegalTargets()
		if len(legal) == 0 {
			return nil
		}
		return DealDamage{Source: item.SourceCardID, Target: legal[0].ID, Amount: amount}.Apply(ctx)
	}
}
