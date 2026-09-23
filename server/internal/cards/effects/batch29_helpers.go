package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch29_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 29 (#391, `edhrec_rank` 3041–3144). Own file per
// the #231 convention; every package-level name carries the b29
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "another creature you control entered" is
// b13AnotherCreatureYouControlEntered, landfall is
// b13LandYouControlEntered, "another <type> you control entered" is
// b10AnotherPermanentWithSubtypeEnteredUnderYourControl, "you cast
// a noncreature spell" is b10NoncreatureSpellCastByYou, "this
// creature dealt combat damage to a player" is
// combatDamageToPlayerBy, "if you cast it" is b16EnteredFromStack,
// the counters a permanent had when it left are b13LastKnownCounters,
// the permanent sacrificed to pay an activation is
// b17PermanentSacrificedToPay, "creatures you control have X" is
// b16GrantKeywords over b16CreaturesYouControl, the per-source
// dedup for a carried token static is b15IsFirstOfItsNameControlledBy,
// the active player is isActivePlayer, the fetchland search is
// fetchDual, the Elf Warrior token is b13GreenElfWarriorToken, and
// the Food, Treasure and flying Spirit tokens are FoodToken,
// TreasureToken and SpiritToken.

// --- token templates ---------------------------------------------

// --- costs ---------------------------------------------------------

// b29SacrificeALand is Hammer of Purphoros's "Sacrifice a land" —
// any land the activator controls, post-layer types.
func b29SacrificeALand() *game.TargetSpec {
	return sacrificeSpec("a land", Land())
}

// --- statics -----------------------------------------------------

// b29AllCreatures is "all creatures" — every player's, the source
// included when it is one (Mass Hysteria).
func b29AllCreatures(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsCreature()
}

// b29OtherCreaturesYouControlWithFlying is Empyrean Eagle's "other
// creatures you control with flying". Read in layer 7c, so the
// flying is whatever layer 6 left on the creature — a printed
// keyword, a lord's grant or an until-end-of-turn grant all count,
// as printed.
func b29OtherCreaturesYouControlWithFlying(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID &&
		target.Controller == source.Controller &&
		target.IsCreature() && game.HasKeyword(target, "flying")
}

// b29SquirrelsYouControlPerAcornCounter is Chitterspitter's
// "Squirrels you control get +1/+1 for each acorn counter on this
// artifact": the scaling lord shape, read off the source's counters
// on every recompute — a counter change bumps the layer version.
func b29SquirrelsYouControlPerAcornCounter() game.StaticAbility {
	return TribalScalingAnthem(
		TribeFilter{Tribes: []string{"Squirrel"}, YoursOnly: true},
		func(source *game.Card, _ *game.Game) int { return source.Counters["acorn"] },
	)
}

// b29CreaturesYouControlWithSevenEnchantments is Hallowed
// Haunting's gate: "as long as you control seven or more
// enchantments, creatures you control …". Counted per recompute,
// post-layer types, the Haunting itself included.
func b29CreaturesYouControlWithSevenEnchantments(target *game.Card, g *game.Game, source *game.Card) bool {
	return b16CreaturesYouControl(target, g, source) && b29EnchantmentsControlled(g, source.Controller) >= 7
}

// b29SpiritClericSizing is the Spirit Cleric token's printed
// "this token's power and toughness are each equal to the number of
// Spirits you control", carried by Hallowed Haunting on the tokens'
// behalf: a layer 7a set over the 0/0 Spirit Cleric tokens its
// controller controls, applied by the FIRST Haunting the controller
// controls only, so two Hauntings size a token once. Layer 7a so
// counters and anthems still stack on top, the way they do on a
// printed characteristic-defining ability.
func b29SpiritClericSizing() game.StaticAbility {
	return game.StaticAbility{
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7A_CDA,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.Controller == source.Controller && IsToken(*target) && target.Name == "Spirit Cleric" &&
				target.IsCreature() && target.Power == 0 && target.Toughness == 0 &&
				b15IsFirstOfItsNameControlledBy(g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := b29SpiritsControlled(g, source.Controller)
			c.Power = n
			c.Toughness = n
		},
	}
}

// --- state reads ---------------------------------------------------

// b29EnchantmentsControlled counts the enchantments `controller`
// controls, post-layer types.
func b29EnchantmentsControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsEnchantment() {
			n++
		}
	}
	return n
}

// b29SpiritsControlled counts the Spirits `controller` controls —
// effective subtypes, so a changeling and a Spirit Cleric token both
// count.
func b29SpiritsControlled(g *game.Game, controller uuid.UUID) int {
	return b26PermanentsOfSubtypeControlled(g, controller, "Spirit")
}

// b29LastKnownToughnessOffBattlefield is the toughness a card had
// when it last left the battlefield, read without the harvester's
// LKI characteristic (which only a dies trigger receives): its
// printed toughness plus the +1/+1 and -1/-1 counters read back off
// the log — b17LastKnownPowerOffBattlefield's other half. A static
// bonus from another permanent is not in it; declared on the card
// that reads this.
func b29LastKnownToughnessOffBattlefield(g *game.Game, cardID uuid.UUID) int {
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return 0
	}
	return c.Toughness + b13LastKnownCounters(g, cardID, "+1/+1") - b13LastKnownCounters(g, cardID, "-1/-1")
}

// b29AuraNamesControlled is the set of names among the Auras
// `controller` controls — Light-Paws' "a different name than each
// Aura you control", read at resolution.
func b29AuraNamesControlled(g *game.Game, controller uuid.UUID) map[string]bool {
	out := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsAura() {
			out[c.Name] = true
		}
	}
	return out
}

// --- trigger conditions ------------------------------------------

// b29ArtifactSpellCastByYou is Vedalken Archmage's condition — Sai's
// shape, named. The spell is read off the stack, where its type
// line is intact.
func b29ArtifactSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsArtifact()
}

// b29AnotherLegendaryCreatureYouControlEntered is Gimli's first
// condition: another legendary creature entered under the source's
// controller's control. Effective supertypes, so a token copy of a
// legend counts.
func b29AnotherLegendaryCreatureYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature() && c.IsLegendary()
}

// b29AuraYouCastEntered is Light-Paws' condition: an Aura entered
// under the source's controller's control, and it was cast — its
// most recent move onto the battlefield came from the stack
// (b16EnteredFromStack). Returns the Aura as it now sits, for its
// mana value.
func b29AuraYouCastEntered(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !c.IsAura() || !b16EnteredFromStack(g, c.InstanceID) {
		return game.Card{}, false
	}
	return c, true
}

// b29SpellCastOffTurn is Scytheclaw Raptor's condition: a player
// cast a spell, and it is not that player's turn.
func b29SpellCastOffTurn(ev game.Event, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor != uuid.Nil && !isActivePlayer(g, ev.Actor)
}

// b29CreatureWithMinusCounterDied is Blowfly Infestation's
// intervening-if: a creature died, and it had a -1/-1 counter on it
// when it left — read back off the log, because MoveCard cleared
// its counters on the way out (b13LastKnownCounters).
func b29CreatureWithMinusCounterDied(ev game.Event, g *game.Game) bool {
	if _, ok := diedCreature(ev, g); !ok {
		return false
	}
	return b13LastKnownCounters(g, ev.CardID, "-1/-1") > 0
}

// b29PowerAsItLastStood is a creature's power for a "that creature's
// power" clause read after it may have left: its current power
// (counters and anthems included) while it is on the battlefield,
// and otherwise its printed power plus the counters it had when it
// left, read back off the log (b17LastKnownPowerOffBattlefield). A
// reflexive trigger resolves well after the parent ability created
// it, by which time the sacrificed creature is in the graveyard with
// its counters cleared — which is why the second branch exists. A
// static bonus from another permanent is not in that branch;
// declared on the card that reads this.
func b29PowerAsItLastStood(g *game.Game, cardID uuid.UUID) int {
	if onBattlefield(g, cardID) {
		if c, ok := g.LookupCardForEffect(cardID); ok {
			return c.CurrentPower()
		}
	}
	return b17LastKnownPowerOffBattlefield(g, cardID)
}

// --- effect bodies -----------------------------------------------

// b29DamageEachPlayerHalfTheirLife is Heartless Hidetsugu's
// activation: each player takes damage equal to half their life
// total, rounded down. Every amount is read before any is dealt —
// the damage is simultaneous — and a player at 1 or less takes
// nothing.
func b29DamageEachPlayerHalfTheirLife(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	type hit struct {
		player uuid.UUID
		amount int
	}
	var hits []hit
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.Life <= 1 {
			continue
		}
		hits = append(hits, hit{player: p.ID, amount: p.Life / 2})
	}
	for _, h := range hits {
		if err := (DealDamage{Source: item.SourceCardID, Target: h.player, Amount: h.amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b29ExileAttackersThenTheyFetchBasics is Settle the Wreckage: exile
// every attacking creature the chosen player controls, then that
// player may search their library for that many basic land cards
// and put them onto the battlefield tapped. The search is theirs —
// their library, their prompt — and "may" means they can decline
// the whole thing; zero exiled is no search at all.
//
// `exiled` is what LANDED in exile (#866), and the clause runs from
// the sweep's continuation, so it may run an action later than the
// sweep — see the card's file.
func b29ExileAttackersThenTheyFetchBasics(ctx *Context, player uuid.UUID) error {
	return ExileAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.Controller == player && c.IsCreature() && c.AttackingTarget != uuid.Nil
		},
		Then: func(ctx *Context, _ []game.Card, exiled int) error {
			if exiled <= 0 {
				return nil
			}
			return SearchLibrary{
				Player:        player,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         exiled,
				Shuffle:       true,
				TappedOnEntry: true,
				Optional:      true,
				Reason:        "Settle the Wreckage — basic land cards, onto the battlefield tapped",
			}.Apply(ctx)
		},
	}.Apply(ctx)
}

// b29ReturnZombieCardsTappedThenDestroyHumans is Zombie Apocalypse:
// every Zombie creature card in the controller's graveyard returns
// to the battlefield and is tapped (the Splendid Reclamation posture
// — ReturnFromGraveyard has no tapped flag, so each enters untapped
// and is tapped a beat later inside the same resolution), then every
// Human is destroyed. The IDs are snapshotted before the first
// move, because ReturnFromGraveyard mutates the pile being walked.
func b29ReturnZombieCardsTappedThenDestroyHumans(ctx *Context) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var zombies []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && c.HasSubtype("Zombie") {
			zombies = append(zombies, c.InstanceID)
		}
	}
	for _, id := range zombies {
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return DestroyAllMatching{Match: HasSubtype("Human")}.Apply(ctx)
}

// b29LightPawsLabel is the stack label of Light-Paws' Aura trigger.
const b29LightPawsLabel = "Light-Paws, Emperor's Voice — search for an Aura to attach to it"

// b29SearchAuraAttachedToSource is Light-Paws' search: an Aura card
// with mana value at most `maxMV` and a name no Aura the controller
// controls has, put onto the battlefield attached to the source. The
// "may" was answered when the trigger fired. Nothing is searched
// when the source has left the battlefield — there is nothing to
// attach the card to, and printed it would stay in the library. The
// names are read at resolution, as printed; the attach runs in the
// search's continuation, once the card is actually on the
// battlefield.
func b29SearchAuraAttachedToSource(g *game.Game, item *game.StackItem, maxMV int) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	taken := b29AuraNamesControlled(g, item.Controller)
	source := item.SourceCardID
	return SearchLibrary{
		Player: item.Controller,
		Predicate: func(c game.Card) bool {
			return c.IsAura() && c.ManaValue() <= maxMV && !taken[c.Name]
		},
		Dest:    game.ZoneBattlefield,
		Limit:   1,
		Shuffle: true,
		Reason:  "Light-Paws, Emperor's Voice — an Aura card with mana value " + strconv.Itoa(maxMV) + " or less",
		Then: func(g *game.Game, found []uuid.UUID) error {
			if !onBattlefield(g, source) {
				return nil
			}
			for _, id := range found {
				if !onBattlefield(g, id) {
					continue
				}
				if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: source}); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(NewContext(g, item))
}

// b29ZiatoraSacrificeLabel is the stack label of Ziatora's end-step
// trigger.
const b29ZiatoraSacrificeLabel = "Ziatora, the Incinerator — sacrifice another creature"

// b29SacrificeChosenCreature is Ziatora's end-step body: the
// creature chosen when the trigger went on the stack is sacrificed
// if it is still there and still legal, and — because it was — the
// printed "when you do" goes on the stack as a CR 603.12 reflexive
// trigger with its own "any target" clause, carrying the creature
// that just died as its payload.
//
// Nothing follows when the chosen creature has already left: "when
// you do" is conditional on the doing, so a creature removed in
// response takes the damage and the Treasures with it.
func b29SacrificeChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == item.SourceCardID {
			continue
		}
		if err := (SacrificePermanent{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return ReflexiveTrigger{
			Label:   "Ziatora, the Incinerator — damage equal to the sacrificed creature's power to any target, and three Treasures",
			Targets: TargetAny(),
			Cards:   []uuid.UUID{t.ID},
			Effect:  b29ZiatoraFling,
		}.Apply(ctx)
	}
	return nil
}

// b29ZiatoraFling is Ziatora's reflexive body: damage equal to the
// power of the creature this trigger carries, as it last stood, to
// the target chosen when the trigger went on the stack — plus three
// Treasures either way, since they are not conditional on the damage
// landing.
func b29ZiatoraFling(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power := 0
	if sacrificed := ctx.PayloadCards(); len(sacrificed) > 0 {
		power = b29PowerAsItLastStood(g, sacrificed[0])
	}
	if ts := ctx.LegalTargets(); len(ts) > 0 {
		if err := (DealDamage{Source: item.SourceCardID, Target: ts[0].ID, Amount: power}).Apply(ctx); err != nil {
			return err
		}
	}
	return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 3}.Apply(ctx)
}

// b29SacrificeChosenTokenThenAcorn is Chitterspitter's upkeep body:
// the token chosen when the trigger went on the stack is sacrificed
// if it is still there, and only then does an acorn counter go on
// the Chitterspitter — if it is still on the battlefield.
func b29SacrificeChosenTokenThenAcorn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (SacrificePermanent{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		if onBattlefield(g, t.ID) || !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: "acorn", N: 1}.Apply(ctx)
	}
	return nil
}

// b29FoodForSacrificedCreature is Witch's Oven's activation: one
// Food, or two if the creature sacrificed to pay had toughness 4 or
// greater as it last stood — printed toughness plus its counters
// (b29LastKnownToughnessOffBattlefield).
func b29FoodForSacrificedCreature(g *game.Game, item *game.StackItem) error {
	n := 1
	if sacrificed, ok := b17PermanentSacrificedToPay(g, item); ok && b29LastKnownToughnessOffBattlefield(g, sacrificed) >= 4 {
		n = 2
	}
	return CreateToken{Controller: item.Controller, Template: FoodToken(), N: n}.Apply(NewContext(g, item))
}

// b29TargetOpponentSacrificesACreatureThenWizard is Cornered by
// Black Mages: the chosen opponent picks a creature of theirs to
// sacrifice (their prompt, not a target — hexproof is irrelevant),
// and the caster gets the Wizard.
//
// In the printed order since #1019. The token used to be created on
// the line after the prompt went up, so it arrived while the opponent
// was still choosing — the clause is about a PLAYER and happens
// either way (ADR 0013 §5m item 5), but the ORDER is observable and
// the run's continuation is what makes it the printed one. A target
// with no creature is never asked and the Wizard still comes, which
// is the continuation running with nothing sacrificed.
func b29TargetOpponentSacrificesACreatureThenWizard(item *game.StackItem, ctx *Context) error {
	var victims []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			victims = append(victims, t.ID)
			break
		}
	}
	return ctx.Game.PlayersSacrificeThenForEffect(item.SourceCardID, victims,
		sacrificeSpec("a creature", Creature()),
		"Cornered by Black Mages — sacrifice a creature",
		func(g *game.Game, _ game.PromptedSacrifices) error {
			return CreateToken{Controller: item.Controller, Template: NoncreatureCastWizardToken(), N: 1}.
				Apply(NewContext(g, item))
		})
}

// b29TapAllCreaturesControlledBy taps every untapped creature
// `player` controls — Naya Charm's third mode. Snapshot then tap, so
// the tap events don't disturb the walk.
func b29TapAllCreaturesControlledBy(ctx *Context, player uuid.UUID) error {
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() && !c.Tapped {
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
