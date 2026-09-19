package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch18_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 18 (#311, `edhrec_rank` 1930–2031). Own file per
// the #231 convention; every package-level name carries the b18
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "enters with N counters" is
// b10EntersWithCounters, the per-label "one or more" dedup is
// OncePerBatch, the once-per-turn tally is
// b11TriggeredThisTurn, "each player draws" is b05EachPlayerDraws,
// "another creature dies" is b15AnotherCreatureDied, an opponent's
// life loss is b04OpponentLostLife, the Shadowmoor filter land is
// b08FilterLand, the Guildgate is a row in guildgates.go, and the
// tapped Treasure is tappedTreasureToken.

// --- token templates ---------------------------------------------

// b18SpringleafShapeshifterToken is Springleaf Parade's 1/1
// colorless Shapeshifter with changeling. It carries the Parade's
// grant — "{T}: Add one mana of any color" — on the template,
// gated by a Condition that the token's controller still controls
// a Springleaf Parade: a mana ability cannot be GRANTED to another
// permanent by a static (ManaAbilitiesForCard reads the token's own
// list or the catalog by oracle ID, nothing in between), so the
// ability lives on the tokens the Parade makes and switches off when
// the Parade leaves. See springleaf_parade.go for what that does and
// does not cover.
func b18SpringleafShapeshifterToken() game.Card { return tokenFromCatalog("springleaf-shapeshifter") }

// printedB18SpringleafShapeshifterToken is the Springleaf Parade Shapeshifter as PRINTED —
// the ability included. It is the catalog's entry for this token
// (token_catalog.go, #521): the ability is registered from here at
// boot, and the template that reaches the battlefield carries the
// key that finds it rather than the closure itself.
func printedB18SpringleafShapeshifterToken() game.Card {
	return game.Card{
		Name:      "Shapeshifter",
		TypeLine:  "Token Creature — Shapeshifter",
		Power:     1,
		Toughness: 1,
		Keywords:  []string{game.KeywordChangeling},
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:   true,
			Produced:  "{W|U|B|R|G}",
			Label:     "{T}: Add one mana of any color (while you control Springleaf Parade)",
			Condition: b18ControlsNamed("Springleaf Parade"),
		}},
	}
}

// b18ControlsNamed is a mana-ability Condition: the activator
// controls a permanent with the given name.
func b18ControlsNamed(name string) func(g *game.Game, controller, source uuid.UUID) bool {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && c.Name == name {
				return true
			}
		}
		return false
	}
}

// --- per-turn reads off the tally --------------------------------

// b18AttackedThisTurn reports whether `player` declared at least one
// attacker this turn — Chart a Course's "unless you attacked this
// turn". PlayerTurnTally.AttacksDeclared counts every EventAttack the
// player announced, so this is one cell (#1009: it used to walk the
// log back to the turn's upkeep).
func b18AttackedThisTurn(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).AttacksDeclared > 0
}

// b18LifeLostThisTurn is the total life `player` lost this turn —
// Wound Reflection's amount. The same two event kinds
// b04OpponentLostLife reads: a negative EventChangeLife and an
// EventDealDamage to the player, which writes the life total
// directly and emits no EventChangeLife of its own. Both are folded
// into PlayerTurnTally.LifeLost as they happen.
func b18LifeLostThisTurn(g *game.Game, player uuid.UUID) int {
	return g.TurnTallyFor(player).LifeLost
}

// --- board reads -------------------------------------------------

// b18ControlsYourCommander is the lieutenant condition: `player`
// controls a permanent that is a commander THEY OWN. Owner matters —
// "your commander" is not an opponent's commander you stole.
func b18ControlsYourCommander(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.Owner == player && c.Controller == player {
			return true
		}
	}
	return false
}

// b18AurasAttachedTo counts the Auras attached to the permanent —
// Kor Spiritdancer's "+2/+2 for each Aura attached to it". Walks
// the battlefield for anything whose AttachedTo points at the card
// (#379's relation) and is an Aura right now; an Equipment does not
// count.
func b18AurasAttachedTo(g *game.Game, cardID uuid.UUID) int {
	n := 0
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(cardID) && a.IsAura() {
			n++
		}
	}
	return n
}

// b18ControlsPermanentOfColor is Kederekt Parasite's intervening if:
// `player` controls a permanent of the colour. Effective colours, so
// a Painter's Servant would count, and a colourless artifact does
// not.
func b18ControlsPermanentOfColor(g *game.Game, player uuid.UUID, color string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasColor(color) {
			return true
		}
	}
	return false
}

// --- trigger conditions ------------------------------------------

// b18CommittedCrime is Magda's "whenever you commit a crime": the
// source's controller targeted an opponent, a permanent or spell an
// opponent controls, or a card in an opponent's graveyard (CR
// 700.13). EventBecomesTarget fires once per target slot at
// announce, with the targeting player in Actor and the target in
// Target; CardID is uuid.Nil when the target is a player.
func b18CommittedCrime(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || ev.Actor != source.Controller {
		return false
	}
	if ev.CardID == uuid.Nil {
		p := g.PlayerByIDForEffect(ev.Target)
		return p != nil && p.ID != source.Controller
	}
	z := g.FindCardZoneForEffect(ev.CardID)
	c, ok := g.LookupCardForEffect(ev.CardID)
	if z == nil || !ok {
		return false
	}
	switch z.Kind {
	case game.ZoneBattlefield, game.ZoneStack:
		return c.Controller != source.Controller
	case game.ZoneGraveyard:
		return c.Owner != source.Controller
	}
	return false
}

// b18AnotherAngelOrClericYouControlEntered is Righteous Valkyrie's
// condition, reading effective subtypes so a changeling counts.
func b18AnotherAngelOrClericYouControlEntered(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	if !ok || !c.IsCreature() {
		return game.Card{}, false
	}
	if !c.HasSubtype("Angel") && !c.HasSubtype("Cleric") {
		return game.Card{}, false
	}
	return c, true
}

// b18SpellCastByYouIsAura is Kor Spiritdancer's "whenever you cast
// an Aura spell": the spell is on the stack when EventCast fires, so
// its subtype is read where it sits.
func b18SpellCastByYouIsAura(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsAura()
}

// b18OpponentDiscarded is "whenever an opponent discards a card"
// (Sangromancer) — the discarding player is the event's Actor.
func b18OpponentDiscarded(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventDiscardCard && ev.Actor != uuid.Nil && ev.Actor != source.Controller
}

// b18OpponentsCreatureDied is "whenever a creature an opponent
// controls dies" (Sangromancer): the dead creature, read post-move,
// was controlled by someone other than the source's controller.
func b18OpponentsCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	return ok && dead.Controller != source.Controller
}

// --- effects -----------------------------------------------------

// b18EachPlayerDrawsAndLosesLife is Stormfist Crusader's upkeep:
// each player draws a card and loses 1 life, APNAP from the active
// seat, eliminated seats skipped. Draw then loss per player, in the
// printed order.
func b18EachPlayerDrawsAndLosesLife(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	for i := 0; i < numSeats; i++ {
		p := g.Seats[(start+i)%numSeats]
		if p == nil || p.Eliminated {
			continue
		}
		if err := (DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, p.ID, -1); err != nil {
			return err
		}
	}
	return nil
}

// b18ExileCreaturesThenControllersFetch is Winds of Abandon's body
// for a set of creatures: exile them as ONE event, then each former
// controller searches for as many basic lands as they lost creatures,
// onto the battlefield tapped. One search per player rather than one
// per creature, so no two prompts are ever open over the same library
// (Cultivate's rule); a player with three creatures exiled gets one
// three-card search, which is the same three lands.
//
// #870: "for each creature exiled this way" is a CONTINUATION, and it
// counts what ARRIVED in exile (CR 400.7). The old body called the
// single-card exile per creature and paid a land for every call that
// returned no error — which includes a leg that only PAUSED on the
// CR 903.9 prompt and a leg the CR 614 window took away, so an
// opponent whose commander was merely ASKED about the command zone
// fetched a basic for a creature that is still on the battlefield.
// The grouping is done from the batch's landed list instead: the
// controllers are read BEFORE anything moves (afterwards the card is
// in exile and a creature that had changed hands reads stale), the
// exile is the shared batch form, and the counting happens when the
// last leg has settled.
//
// Searches are queued in the order the exiled creatures were swept,
// so the prompts arrive in a stable order; each shuffles as it
// finishes, which nothing can observe.
func b18ExileCreaturesThenControllersFetch(ctx *Context, creatures []game.Card, reason string) error {
	controllers := make(map[uuid.UUID]uuid.UUID, len(creatures))
	ids := make([]uuid.UUID, 0, len(creatures))
	for _, c := range creatures {
		if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		ids = append(ids, c.InstanceID)
		controllers[c.InstanceID] = c.Controller
	}
	// The context is rebuilt inside the continuation from the live
	// *Game, the contract massEffect.apply explains.
	item := ctx.Item
	return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		owed := map[uuid.UUID]int{}
		var order []uuid.UUID
		for _, id := range exiled {
			controller := controllers[id]
			if owed[controller] == 0 {
				order = append(order, controller)
			}
			owed[controller]++
		}
		ctx := NewContext(g, item)
		for _, player := range order {
			if err := (SearchLibrary{
				Player:        player,
				Predicate:     IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         owed[player],
				Reveal:        true,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        reason,
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
