package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch11_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 11 (#304, `edhrec_rank` 1214–1315). Own file per
// the #231 convention; every package-level name carries the b11
// prefix because batch 10 is landing beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, the reveal-and-tutor body is
// b06TutorToHand, the Overlook land shape is b08OverlookLand, the
// entered-this-turn event walk is b06EnteredThisTurn (the two
// per-turn walks below follow its shape), "sacrifice ANOTHER
// creature" by name is b03NotNamed, and the Treasure and Food are
// tokens.go's.

// --- token templates ---------------------------------------------

// b11WhiteWarriorToken is Secure the Wastes' 1/1 white Warrior.
func b11WhiteWarriorToken() game.Card {
	return game.Card{
		Name:      "Warrior",
		TypeLine:  "Token Creature — Warrior",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// b11WhiteSoldierToken is Keeper of the Accord's 1/1 white Soldier.
// The same token Elspeth hands out, with the printed colour stamped
// — tokens.go's SoldierToken predates token colours.
func b11WhiteSoldierToken() game.Card {
	return game.Card{
		Name:      "Soldier",
		TypeLine:  "Token Creature — Soldier",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

// b11GreenSaprolingToken is Tendershoot Dryad's 1/1 green Saproling.
func b11GreenSaprolingToken() game.Card {
	return game.Card{
		Name:      "Saproling",
		TypeLine:  "Token Creature — Saproling",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"G"},
	}
}

// --- conditions and counts ---------------------------------------

// b11ControlsBasicLand is the Avatar-cycle land condition — "unless
// you control a basic land" (Abandoned Air Temple). The entering land
// is `src` and is not on the battlefield yet, so the walk counts only
// the OTHER lands its controller controls, exactly as the checkland
// condition does. Supertype rather than name, so a Snow-Covered
// Plains and a Wastes both satisfy it, as printed.
func b11ControlsBasicLand(g *game.Game, src *game.Card) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == src.InstanceID || c.Controller != src.Controller {
			continue
		}
		if c.IsLand() && IsBasicLand(c) {
			return true
		}
	}
	return false
}

// b11PermanentsControlled counts the permanents `controller`
// controls — Tendershoot Dryad's "ten or more permanents". Walks the
// live slice because it runs inside a layer recompute.
func b11PermanentsControlled(g *game.Game, controller uuid.UUID) int {
	if g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Controller == controller {
			n++
		}
	}
	return n
}

// b11CreatureCardsInGraveyard counts the creature cards in
// `controller`'s graveyard — Wight of the Reliquary's bonus. A card in
// a graveyard has no layer cache, so this is the printed type line,
// which is what "creature card" means off the battlefield.
func b11CreatureCardsInGraveyard(g *game.Game, controller uuid.UUID) int {
	p := g.PlayerByIDForEffect(controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			n++
		}
	}
	return n
}

// b11CreaturesDiedThisTurn counts the creatures that died this turn
// — Mahadi, Emporium Master's Treasure count. The engine keeps no
// per-turn death tally, so this is the b06EnteredThisTurn walk over
// the event log: every EventLTB into a graveyard since the current
// turn's upkeep began, kept when the card — looked up where it sits
// now — is a creature. Tokens stay in the graveyard (no CR 704.5d
// sweep), so a dead Saproling counts as printed.
//
// Weaker, never stronger: a creature card that has since left every
// tracked zone, or a permanent that was a creature only through a
// layer effect when it died, is not counted.
func b11CreaturesDiedThisTurn(g *game.Game) int {
	return g.TurnTally.CreaturesDied
}

// b11TriggeredThisTurn reports whether an ability of `source` with
// the given stack label has already been put on the stack this turn
// — "This ability triggers only once each turn" (Exemplar of Light).
// The harvester emits EventTrigger, carrying the source and the
// label, the moment it queues an item, so a walk back to the current
// turn's upkeep is the tally; a trigger that was countered still
// counts, as printed.
func b11TriggeredThisTurn(g *game.Game, source uuid.UUID, label string) bool {
	return g.TriggeredThisTurn(source, label) > 0
}

// b11CountersWerePlaced reports whether ev is a PLACEMENT of `kind`
// counters on `target`, as opposed to a removal. EventCounterPlaced
// is emitted for both and carries only the post-change total, so the
// previous total is read back off the log: the most recent
// EventCounterPlaced for the same card and kind, or zero if the card
// entered the battlefield more recently than that (CR 400.7 — it
// arrived with no counters).
func b11CountersWerePlaced(ev game.Event, target uuid.UUID, kind string, g *game.Game) bool {
	if ev.Kind != game.EventCounterPlaced || ev.Target != target || ev.Label != kind {
		return false
	}
	before := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		prev := g.Events[i]
		if prev.Seq >= ev.Seq {
			continue
		}
		if (prev.Kind == game.EventETB || prev.Kind == game.EventTokenCreated) && prev.CardID == target {
			break
		}
		if prev.Kind == game.EventCounterPlaced && prev.Target == target && prev.Label == kind {
			before = prev.Amount
			break
		}
	}
	return ev.Amount > before
}

// b11ResolvingController is the controller of the spell or ability
// whose resolution is in progress — the Actor of the most recent
// EventResolve, which the engine emits BEFORE it runs an item's
// effect. It is how "whenever YOU put counters" is read: counters
// land during a resolution, and the player resolving is the player
// putting them. uuid.Nil when nothing has resolved yet.
func b11ResolvingController(g *game.Game) uuid.UUID {
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == game.EventResolve {
			return g.Events[i].Actor
		}
	}
	return uuid.Nil
}

// b11OpponentControlsMoreThanYou is Keeper of the Accord's
// intervening-if: `opp` controls more permanents matching `match`
// than `you` do.
func b11OpponentControlsMoreThanYou(g *game.Game, you, opp uuid.UUID, match func(game.Card) bool) bool {
	yours, theirs := 0, 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if !match(c) {
			continue
		}
		switch c.Controller {
		case you:
			yours++
		case opp:
			theirs++
		}
	}
	return theirs > yours
}

// b11OpponentControlsCreatures reports whether some opponent of
// `controller` controls at least n creatures — Defense of the
// Heart's intervening-if.
func b11OpponentControlsCreatures(g *game.Game, controller uuid.UUID, n int) bool {
	counts := map[uuid.UUID]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller || !c.IsCreature() {
			continue
		}
		counts[c.Controller]++
		if counts[c.Controller] >= n {
			return true
		}
	}
	return false
}

// --- trigger predicates ------------------------------------------

// b11ColorlessSpellCastByYou is Glaring Fleshraker's first
// condition: the controller cast a colorless spell. Colours are read
// off the card on the stack — Scryfall's stamped list, or the mana
// cost for a fixture — so an artifact creature and an Eldrazi both
// qualify and a Devoid card does when its printed colours say so.
func b11ColorlessSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.IsColorless()
}

// b11DwarfYouControlBecameTapped is Magda's trigger condition. Two
// event kinds feed it, and the second is the reason the helper
// exists: ActivateManaAbility, the auto-tapper, convoke and the
// manual tap all emit EventTapCard, but DeclareAttacker stamps
// Card.Tapped directly and emits only EventAttack. Attacking is how
// a Dwarf becomes tapped in nearly every game, so the attack event
// is read as "became tapped" exactly when the attacker is tapped
// after the declaration — a vigilance Dwarf is not. If the engine
// ever emits EventTapCard from the attack path too, this fires
// twice and the batch 11 test that counts one Treasure per
// attacking Dwarf catches it.
func b11DwarfYouControlBecameTapped(ev game.Event, source *game.Card, g *game.Game) bool {
	switch ev.Kind {
	case game.EventTapCard, game.EventAttack:
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || c.Controller != source.Controller || !c.HasSubtype("Dwarf") {
		return false
	}
	if ev.Kind == game.EventAttack && !c.Tapped {
		return false
	}
	return true
}

// b11SacrificedAFood is Nuka-Cola Vending Machine's condition: the
// controller sacrificed a Food. EventSacrifice fires before the zone
// move, so the token is still on the battlefield to be read.
func b11SacrificedAFood(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Food")
}

// --- effect bodies -----------------------------------------------

// b11PutCounterOnEachCreatureYouControl is Abandoned Air Temple's
// activated ability. Snapshots the set first so a creature made by a
// counter-placement trigger mid-loop does not receive one.
func b11PutCounterOnEachCreatureYouControl(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b11DamageEachCreatureControlledBy is Balefire Dragon's rider: the
// source deals `amount` damage to each creature `victim` controls.
// The set is snapshotted before any damage is dealt (CR 608.2).
func b11DamageEachCreatureControlledBy(g *game.Game, item *game.StackItem, victim uuid.UUID, amount int) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == victim && c.IsCreature() {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: amount}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b11FetchLandTapped is the "search your library for a land card,
// put it onto the battlefield tapped, then shuffle" body Wight of
// the Reliquary shares with Sakura-Tribe Elder's shape.
func b11FetchLandTapped(reason string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:        item.Controller,
			Predicate:     func(c game.Card) bool { return c.IsLand() },
			Dest:          game.ZoneBattlefield,
			Limit:         1,
			Reveal:        true,
			Shuffle:       true,
			TappedOnEntry: true,
			Reason:        reason,
		}.Apply(NewContext(g, item))
	}
}
