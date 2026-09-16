package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch12_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 12 (#305, `edhrec_rank` 1316–1420). Own file per
// the #231 convention; every package-level name carries the b12
// prefix because batch 11 is landing beside this one.
//
// What is NOT here, because main already had it: "whenever you gain
// life" is b10YouGainedLife, "whenever you cast a noncreature spell"
// is b10NoncreatureSpellCastByYou, "sacrifice an artifact" as a cost
// is b10SacrificeAnArtifact, the instance-ID exclusion predicate is
// b10OneOf, "a +1/+1 counter on each creature you control" is
// b11PutCounterOnEachCreatureYouControl, the white Soldier is
// b11WhiteSoldierToken, "a creature with power 4 or greater" is
// youControlPowerFourOrGreater, "target nonlegendary" is
// Not(b05Legendary()), "another permanent you control entered" is
// enteredUnderYourControl, and "exile a graveyard" is
// exileGraveyardForEffect. The counter-placement readers below are
// NOT batch 11's b11CountersWerePlaced / b11ResolvingController:
// All Will Be One needs the delta, not a yes/no, and an attribution
// that stops at an action boundary — see b12CountersPlacedByYou.

// --- token templates ---------------------------------------------

// --- mana-ability riders -----------------------------------------

// b12GainLifeRider is the post-production half of Pristine Talisman:
// "{T}: Add {C}. You gain 1 life." A rider, not a cost, and life
// GAIN rather than damage, so it goes through
// ChangePlayerLifeForEffect and every "whenever you gain life"
// payoff (Archangel of Thune, in this same batch) sees it.
func b12GainLifeRider(n int) func(g *game.Game, controller, source uuid.UUID) error {
	return func(g *game.Game, controller, source uuid.UUID) error {
		return g.ChangePlayerLifeForEffect(source, controller, n)
	}
}

// --- costs -------------------------------------------------------

// b12SacrificeAGoblin is Skirk Prospector's "Sacrifice a Goblin:"
// clause, built through the same spec the other sac-outlet costs use
// so the client opens the same picker. Effective subtypes, so a
// changeling counts; the Prospector itself qualifies, as printed.
func b12SacrificeAGoblin() *game.TargetSpec {
	return sacrificeSpec("a Goblin", Subtype("Goblin"))
}

// --- trigger dedup -----------------------------------------------

// --- trigger conditions ------------------------------------------

// b12CardExiledFromYourLibraryOrGraveyard is Laelia's second
// condition: a card owned by the source's controller moved from
// their library or their graveyard into exile. Every exile path
// emits EventZoneMove with the origin in OldZone — the impulse
// exiles, ExileCardForEffect on a graveyard card, a graveyard sweep
// — and "your" library / graveyard is an ownership question, so the
// card's Owner is what is compared. A flashed-back spell is exiled
// from the STACK and does not count, as printed.
func b12CardExiledFromYourLibraryOrGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneExile {
		return false
	}
	if ev.OldZone != game.ZoneLibrary && ev.OldZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// b12MountainYouControlEntered is Valakut's trigger condition: a
// permanent with the Mountain subtype entered under the source's
// controller's control. Effective subtypes, so a land Urborg-style
// effects have made a Mountain counts, as does a Dryad Arbor made
// one.
func b12MountainYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtype("Mountain")
}

// b12OtherMountainsControlled counts the Mountains `controller`
// controls other than `except` — Valakut's "at least five OTHER
// Mountains", where the exception is the Mountain that just entered.
func b12OtherMountainsControlled(g *game.Game, controller, except uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == except || c.Controller != controller {
			continue
		}
		if c.HasSubtype("Mountain") {
			n++
		}
	}
	return n
}

// creatureSpellCastByYou is Lifecrafter's Bestiary's second
// condition.
func creatureSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsCreature()
}

// b12YouSacrificedAnArtifact is Crime Novelist's condition. The
// sacrifice event fires BEFORE the zone move, so the permanent is
// still on the battlefield to be read — the Mirkwood Bats shape.
func b12YouSacrificedAnArtifact(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact()
}

// --- per-turn tallies --------------------------------------------

// b12InstantsAndSorceriesCastBeforeThisTurn counts the instant and
// sorcery spells `controller` cast this turn BEFORE the spell
// `spell` — Thousand-Year Storm's copy count. The engine's per-turn
// cast tally keeps only Total and Noncreature, so this is the
// b06EnteredThisTurn walk over the event log: every EventCast by the
// controller since the current turn's upkeep began, kept when the
// card — looked up where it sits now, a graveyard for a resolved one
// — is an instant or sorcery. Spell copies are not cast and emit no
// EventCast, so a Storm chain counts only the real casts, as
// printed.
//
// Weaker, never stronger: a spell that has since left every tracked
// zone is not counted.
func b12InstantsAndSorceriesCastBeforeThisTurn(g *game.Game, controller, spell uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != controller || ev.CardID == spell {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}

// --- counter placement -------------------------------------------

// b12CountersPlacedByYou reports how many counters ev PLACED on a
// permanent, and whether the source's controller is the one who put
// them there — All Will Be One's "whenever you put one or more
// counters on a permanent".
//
// EventCounterPlaced carries neither, so both are read back off the
// event log:
//
//   - The DELTA. The event holds only the post-change total (and
//     fires for removals too), so the previous total is the most
//     recent EventCounterPlaced for the same card and kind, or zero
//     if the card's arrival on the battlefield is more recent than
//     that (CR 400.7 — it arrived with no counters). A removal, or a
//     placement of zero, is not a placement.
//   - WHO. Counters land during a resolution, and the player
//     resolving is the player putting them — the Actor of the most
//     recent EventResolve, which the engine emits before it runs an
//     item's effect (Exemplar of Light reads "you put" the same
//     way). A cast, a mana ability, an attack or a step beginning
//     between that resolution and the placement means the placement
//     was NOT part of it, and nobody is credited.
//
// Three placements happen outside any resolution and are attributed
// by a stricter rule, so the card can never fire on an opponent's
// action: a loyalty cost and a Saga's lore counter are put there by
// the permanent's controller (CR 606.4, 714.3), and a permanent
// "entering with" counters — whose counter event precedes its own
// arrival event, which is how it is recognised — likewise. Each of
// those counts only when the permanent is the source's controller's
// AND no other player's resolution is in progress. Weaker than
// printed where the two disagree, never stronger.
//
// One event is one permanent, so "one or more counters on A
// PERMANENT" needs no batching: Cathars' Crusade putting a counter
// on each of five creatures is five placements and five triggers,
// which is the printed outcome.
func b12CountersPlacedByYou(ev game.Event, source *game.Card, g *game.Game) (int, bool) {
	if ev.Kind != game.EventCounterPlaced || ev.Target == uuid.Nil || ev.Label == "" {
		return 0, false
	}
	if z := g.FindCardZoneForEffect(ev.Target); z == nil || z.Kind != game.ZoneBattlefield {
		return 0, false
	}
	target, ok := g.LookupCardForEffect(ev.Target)
	if !ok {
		return 0, false
	}
	before, arrivalLogged := b12CounterTotalBefore(ev, g)
	delta := ev.Amount - before
	if delta <= 0 {
		return 0, false
	}
	me := source.Controller
	resolving := b12ResolvingController(ev, g)
	switch {
	case !arrivalLogged, ev.Label == game.CounterLoyalty, ev.Label == game.CounterLore:
		if target.Controller != me || (resolving != uuid.Nil && resolving != me) {
			return 0, false
		}
	default:
		if resolving != me {
			return 0, false
		}
	}
	return delta, true
}

// b12CounterTotalBefore is the count of ev's counter kind on ev's
// target before ev changed it, and whether the target's arrival on
// the battlefield has been logged yet. It has not when the counters
// are "enters with" ones: every entry site places those before it
// emits the zone-move and ETB events.
func b12CounterTotalBefore(ev game.Event, g *game.Game) (before int, arrivalLogged bool) {
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		if prev.CardID == ev.Target {
			switch prev.Kind {
			case game.EventETB, game.EventTokenCreated:
				return 0, true
			case game.EventZoneMove:
				if prev.NewZone == game.ZoneBattlefield {
					return 0, true
				}
			}
		}
		if prev.Kind == game.EventCounterPlaced && prev.Target == ev.Target && prev.Label == ev.Label {
			return prev.Amount, true
		}
	}
	return 0, false
}

// b12ResolvingController is the controller of the spell or ability
// whose resolution ev happened inside — the Actor of the most recent
// EventResolve before ev — or uuid.Nil when an action that cannot
// happen mid-resolution sits between the two: a cast, a mana ability,
// an attack declaration, a fizzle, or a step beginning.
func b12ResolvingController(ev game.Event, g *game.Game) uuid.UUID {
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		switch prev.Kind {
		case game.EventResolve:
			return prev.Actor
		case game.EventCast, game.EventManaAbilityActivated, game.EventAttack, game.EventFizzle,
			game.EventBeginUpkeep, game.EventBeginPrecombatMain, game.EventBeginEndStep:
			return uuid.Nil
		}
	}
	return uuid.Nil
}

// --- target clauses ----------------------------------------------

// b12TargetOpponentOrTheirCreatureOrPlaneswalker is All Will Be
// One's clause: "target opponent, creature an opponent controls, or
// planeswalker an opponent controls" — a player who is not the
// caster, or a battlefield creature / planeswalker the caster does
// not control.
func b12TargetOpponentOrTheirCreatureOrPlaneswalker() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   "target opponent, creature an opponent controls, or planeswalker an opponent controls",
		Players: true,
		PlayerOK: func(_ *game.Game, caster uuid.UUID, p *game.Player) bool {
			return p.ID != caster
		},
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.Controller != caster && (c.IsCreature() || c.IsPlaneswalker())
		},
		Min: 1, Max: 1,
	}
}

// --- effect bodies -----------------------------------------------

// b12EachOpponentMills is Altar of the Brood's body: every opponent
// of the item's controller mills n.
func b12EachOpponentMills(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if err := (MillCards{Player: opp, N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b12ImpulseExileForTurn is "exile the top N cards of your library.
// You may play them this turn" — Laelia's one and Bonehoard
// Dracosaur's two — returning the exiled IDs so a rider can read
// what came off. "Play", not "cast", so a land exiled this way can be
// played as the turn's land drop.
func b12ImpulseExileForTurn(g *game.Game, item *game.StackItem, n int) ([]uuid.UUID, error) {
	return g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, n, game.ExilePlayPermission{})
}

// b12Victim pairs a permanent Terastodon is about to destroy with
// the player who controlled it at that moment.
type b12Victim struct {
	ID         uuid.UUID
	Controller uuid.UUID
}

// b12ElephantsForTheDestroyed is Terastodon's rider: for every
// victim that is now in a graveyard, the player who controlled it
// gets a 3/3 Elephant. The controllers were read BEFORE the
// destruction (a card in a graveyard keeps its Controller field, but
// reading it there is a habit the reanimation path warns against),
// and the zone is checked AFTER, so an indestructible permanent —
// still on the battlefield — earns nothing, and a commander that
// went to the command zone instead earns nothing, as printed ("put
// into a graveyard this way"). Announce order, so the tokens are
// made in a deterministic order.
func b12ElephantsForTheDestroyed(ctx *Context, victims []b12Victim) error {
	for _, v := range victims {
		z := ctx.Game.FindCardZoneForEffect(v.ID)
		if z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if ctx.PlayerByID(v.Controller) == nil {
			continue
		}
		if err := (CreateToken{Controller: v.Controller, Template: TokenCard("3/3 green Elephant"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
