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
func b18SpringleafShapeshifterToken() game.Card {
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

// --- per-turn reads off the event log ----------------------------

// b18AttackedThisTurn reports whether `player` declared at least one
// attacker this turn — Chart a Course's "unless you attacked this
// turn". The engine keeps no such flag, so this is the
// b06EnteredThisTurn walk: an EventAttack whose Actor is the player,
// more recent than the current turn's upkeep. Every turn passes
// through its upkeep before its combat, so "since the last upkeep
// began" is "this turn".
func b18AttackedThisTurn(g *game.Game, player uuid.UUID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		switch ev.Kind {
		case game.EventBeginUpkeep:
			return false
		case game.EventAttack:
			if ev.Actor == player {
				return true
			}
		}
	}
	return false
}

// b18LifeLostThisTurn is the total life `player` lost this turn —
// Wound Reflection's amount. The same two event kinds
// b04OpponentLostLife reads: a negative EventChangeLife and an
// EventDealDamage to the player, which writes the life total
// directly and emits no EventChangeLife of its own. Summed back to
// the current turn's upkeep.
func b18LifeLostThisTurn(g *game.Game, player uuid.UUID) int {
	return g.TurnTallyFor(player).LifeLost
}

// b18AttackerAlreadyBlocked reports whether the attacker named by an
// EventBlock had already been blocked earlier in this combat —
// Grazilaxx's "becomes blocked" fires once per attacker, not once
// per blocker (CR 509.1h), and the engine emits one EventBlock per
// blocker. The walk runs back from the event to the attacker's own
// EventAttack, which every attacker declared this combat has; a
// second EventBlock with the same Target inside that window means
// this one is not the first.
func b18AttackerAlreadyBlocked(g *game.Game, block game.Event) bool {
	seen := false
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if !seen {
			if ev.Seq == block.Seq {
				seen = true
			}
			continue
		}
		switch ev.Kind {
		case game.EventBlock:
			if ev.Target == block.Target {
				return true
			}
		case game.EventAttack:
			if ev.CardID == block.Target {
				return false
			}
		case game.EventBeginUpkeep:
			return false
		}
	}
	return false
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
// for a set of creatures: exile them, then each former controller
// searches for as many basic lands as they lost creatures, onto the
// battlefield tapped. One search per player rather than one per
// creature, so no two prompts are ever open over the same library
// (Cultivate's rule); a player with three creatures exiled gets one
// three-card search, which is the same three lands.
//
// Searches are queued in seat order so the prompts are answered in
// a stable order; each shuffles as it finishes, which nothing can
// observe.
func b18ExileCreaturesThenControllersFetch(ctx *Context, creatures []game.Card, reason string) error {
	owed := map[uuid.UUID]int{}
	var order []uuid.UUID
	for _, c := range creatures {
		if z := ctx.Game.FindCardZoneForEffect(c.InstanceID); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := (ExileTarget{Target: c.InstanceID}).Apply(ctx); err != nil {
			return err
		}
		if owed[c.Controller] == 0 {
			order = append(order, c.Controller)
		}
		owed[c.Controller]++
	}
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
}
