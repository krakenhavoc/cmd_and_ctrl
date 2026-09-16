package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch25_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 25 (#387, `edhrec_rank` 2638–2737). Own file per
// the #231 convention; every package-level name carries the b25
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "this creature died" is cardDied, "another
// creature died" is diedCreature, "a creature you control dealt
// combat damage to a player" is combatDamageToPlayerBy, "a card left
// your graveyard" is b16CardLeftYourGraveyard, the per-label "one or
// more" dedup is OncePerBatch, the loot is lootOne,
// the fight is b10Fight, the +1/+1 doubling is
// b23DoublePlusOneCountersOn, the nonbasic predicate is b03Nonbasic,
// and the Zombie Druid, Treasure and red Goblin tokens are
// b16BlackZombieDruidToken, TreasureToken and RedGoblinToken.

// --- token templates ---------------------------------------------

// b25BlueTentacleToken is Nadir Kraken's 1/1 blue Tentacle.
func b25BlueTentacleToken() game.Card {
	return game.Card{
		Name:      "Tentacle",
		TypeLine:  "Token Creature — Tentacle",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"U"},
	}
}

// --- costs ---------------------------------------------------------

// b25SacrificeAGoblin is Pashalik Mons' "Sacrifice a Goblin" — any
// Goblin creature the activator controls, the source included.
// Effective subtypes, so a changeling is a Goblin.
func b25SacrificeAGoblin() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a Goblin", OfCreatureType("Goblin"))}
}

// --- mana ----------------------------------------------------------

// b25BlackIfLandPlayedElseBlue is River of Tears' colour: {B} if the
// controller has played a land this turn, {U} otherwise. Read off
// the engine's per-turn land-play tally at activation, which both
// play paths bump before the land's entry is even announced, so the
// River itself played this turn counts.
func b25BlackIfLandPlayedElseBlue(g *game.Game, controller, _ uuid.UUID) string {
	if g.LandsPlayedThisTurnFor(controller) > 0 {
		return "{B}"
	}
	return "{U}"
}

// --- trigger conditions ------------------------------------------

// b25CastFromNotHand is Vega's condition: the source's controller
// cast a spell, and the zone it left was not their hand. The cast
// event carries the origin in OldZone — the command zone, a
// graveyard, exile. A cast with no recorded origin is treated as a
// hand cast, the answer that cannot over-fire.
func b25CastFromNotHand(ev game.Event, controller uuid.UUID) bool {
	if ev.Kind != game.EventCast || ev.Actor != controller {
		return false
	}
	return ev.OldZone != "" && ev.OldZone != game.ZoneHand
}

// b25CastIsMulticolored reports whether the spell a cast event
// names has two or more colours.
func b25CastIsMulticolored(ev game.Event, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && len(c.EffectiveColors()) >= 2
}

// b25FirstMulticoloredSpellThisTurn is Zenith Chronicler's
// condition: a player cast a multicolored spell, and no earlier cast
// by that player this turn was multicolored. The engine's per-turn
// cast tally counts creature and noncreature casts only, so the
// answer is walked off the event log back to the start of the turn.
// The boundary is the most recent upkeep-began event (the walk
// b20LandPlayed makes): every player's turn emits one, and nothing
// can be cast in the untap step before it. Turn.Number is NOT a
// per-turn boundary — it counts rounds of the table, so a walk
// bounded on it would reach back through every other player's turn
// in the round. The cast being asked about is itself excluded.
func b25FirstMulticoloredSpellThisTurn(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil || !b25CastIsMulticolored(ev, g) {
		return false
	}
	for _, prev := range g.EventsThisTurn() {
		if prev.Seq >= ev.Seq {
			break
		}
		if prev.Kind == game.EventCast && prev.Actor == ev.Actor && b25CastIsMulticolored(prev, g) {
			return false
		}
	}
	return true
}

// b25DamageDealtByCounteredCreature is Bred for the Hunt's narrowing:
// the creature that dealt the damage carries at least one +1/+1
// counter. Read live, like the rest of combatDamageToPlayerBy's
// checks — combat damage lands before the state-based sweep, so the
// creature is still on the battlefield with its counters.
func b25DamageDealtByCounteredCreature(ev game.Event, g *game.Game) bool {
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.Counters["+1/+1"] > 0
}

// b25AnotherGoblinYouControlDied is the second half of Pashalik
// Mons' condition: a creature the source's controller controlled
// went to a graveyard from the battlefield, and it is a Goblin. Read
// post-move (diedCreature), where the printed type line and the last
// controller survive; a changeling counts. The source's own death
// is cardDied's business, so this excludes it.
func b25AnotherGoblinYouControlDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.InstanceID != source.InstanceID && dead.Controller == source.Controller && dead.HasSubtype("Goblin")
}

// b25AttackersHaveDoubleStrike is "attacking creatures you control
// have double strike" (Blade Historian) — Berserkers' Onslaught's
// trigger, shared: one trigger per attacking creature the
// controller controls, granting double strike until end of turn
// through the turn-scoped static registry. See the Onslaught's file
// for why the printed static is a trigger.
func b25AttackersHaveDoubleStrike(name string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
			return attackDeclaredByYou(ev, source.Controller)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			attacker := ev.CardID
			return game.NewTriggeredItem(source, name+" — the attacking creature has double strike",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Target:   attacker,
						Keywords: []string{"double strike"},
						Label:    name + " — double strike",
					}.Apply(NewContext(g, item))
				})
		},
	}
}

// --- state reads ---------------------------------------------------

// b25LifeAboveStarting is Cosmos Elixir's test: the player's life
// total is greater than the format's starting total.
func b25LifeAboveStarting(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Life > game.StartingLife
}

// b25GraveyardHasLandCard reports whether the player's graveyard
// holds a land card — the question Teval's targeted attack trigger
// asks before it is declared, so the mill is never dropped with an
// empty legal set.
func b25GraveyardHasLandCard(g *game.Game, player uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsLand() {
			return true
		}
	}
	return false
}

// b25OpponentControlsACreature reports whether any creature on the
// battlefield is controlled by someone other than `controller` —
// "creature you don't control", the question Voracious Hydra's
// targeted entry asks before it is declared.
func b25OpponentControlsACreature(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller && c.IsCreature() {
			return true
		}
	}
	return false
}

// b25LastEventSeq is the sequence number of the newest event in the
// log, or zero on an empty log — a cursor for "what happened after
// this point".
func b25LastEventSeq(g *game.Game) uint64 {
	if len(g.Events) == 0 {
		return 0
	}
	return g.Events[len(g.Events)-1].Seq
}

// b25DiscardedByAfter finds the card `player` discarded in the
// newest discard event logged after `after` — Reckless Handling's
// "if an artifact card was discarded this way", read back off the
// event the random discard just emitted. Bounded by the cursor so a
// discard from earlier in the game is never mistaken for this one:
// an empty hand discards nothing and reports nothing.
func b25DiscardedByAfter(g *game.Game, player uuid.UUID, after uint64) (game.Card, bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Seq <= after {
			break
		}
		if ev.Kind == game.EventDiscardCard && ev.Actor == player {
			return g.LookupCardForEffect(ev.CardID)
		}
	}
	return game.Card{}, false
}

// --- effect bodies -----------------------------------------------

// b25EachPlayerExceptDraws draws n cards for every seated player but
// `except` — Zenith Chronicler's "each other player", where the
// excluded player is the caster rather than the source's controller.
func b25EachPlayerExceptDraws(ctx *Context, except uuid.UUID, n int) error {
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Eliminated || p.ID == except {
			continue
		}
		if err := (DrawCards{Player: p.ID, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b25GrowAndSpawnTentacle is Nadir Kraken's "if you do": a +1/+1
// counter on the Kraken if it is still on the battlefield, and a
// 1/1 blue Tentacle either way.
func b25GrowAndSpawnTentacle(ctx *Context, kraken, controller uuid.UUID) error {
	if onBattlefield(ctx.Game, kraken) {
		if err := (AddCounter{Target: kraken, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return CreateToken{Controller: controller, Template: b25BlueTentacleToken(), N: 1}.Apply(ctx)
}

// b25TevalAttackLabel is the stack label both declarations of
// Teval's attack trigger share — one printed ability, one name.
const b25TevalAttackLabel = "Teval, the Balanced Scale — mill three, then return a land card tapped"

// b25TevalLeftGraveyardLabel is the label of Teval's token trigger,
// which the "one or more" dedup keys on.
const b25TevalLeftGraveyardLabel = "Teval, the Balanced Scale — cards left your graveyard"

// b25MillThreeThenReturnChosenLandTapped is Teval's attack trigger:
// mill three, then return the land chosen when the trigger went on
// the stack — if one was chosen and it is still in the graveyard —
// to the battlefield, tapped. The return is ReturnFromGraveyard
// followed by a tap, Lumra's declared gap.
func b25MillThreeThenReturnChosenLandTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (MillCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
		return err
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		return TapTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

// b25VoraciousHydraLabel is the stack label both declarations of
// Voracious Hydra's entry trigger share.
const b25VoraciousHydraLabel = "Voracious Hydra — double its +1/+1 counters, or fight a creature you don't control"

// b25DoubleCountersOrFightChosen is Voracious Hydra's entry trigger:
// with a creature chosen (and still legal), the Hydra fights it;
// with none chosen, its +1/+1 counters are doubled. A chosen
// creature that became illegal never reaches here — the engine
// counters the trigger first (CR 608.2b) — so the doubling can only
// ever be the "no target" answer.
func b25DoubleCountersOrFightChosen(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return b10Fight(ctx, item.SourceCardID, t.ID)
	}
	return b23DoublePlusOneCountersOn(ctx, item.SourceCardID)
}
