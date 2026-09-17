package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch31_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 31 (#394, `edhrec_rank` 3245–3345). Own file per
// the #231 convention; every package-level name carries the b31
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "this creature enters or attacks" is
// b21SelfEnteredOrAttacked, "whenever this creature attacks" is
// attackDeclared and the player it attacks is b17DefendingPlayer,
// "a permanent you control entered" is enteredUnderYourControl,
// "you played a land" is b20LandPlayed, "this creature dealt combat
// damage to a player" is combatDamageToPlayerBy, the per-label
// "one or more" dedup is OncePerBatch, a dead
// creature's power with its counters is b13LastKnownPower, "the
// graveyard holds a land card" is b25GraveyardHasLandCard, "an
// opponent controls a creature" is b25OpponentControlsACreature,
// "enters with N counters" is b10EntersWithCounters, "sacrifice an
// artifact" is b10SacrificeAnArtifact, "target player or
// planeswalker" is targetPlayerOrPlaneswalker (Boros Charm's), the
// Goblin is RedGoblinToken, the lifelink Cat is
// b28WhiteCatLifelinkToken, the Food is FoodToken, and the
// keyword grant over a predicate is b16GrantKeywords.

// --- token templates ---------------------------------------------

// b31TappedCatLifelinkToken is Leonin Warleader's 1/1 white Cat with
// lifelink, stamped tapped so CreateTokensAttackingForEffect — which
// copies the template and sets only the attack — puts it in tapped
// and attacking (General Kreat's posture).
func b31TappedCatLifelinkToken() game.Card {
	tmpl := TokenCard("1/1 white Cat with lifelink")
	tmpl.Tapped = true
	return tmpl
}

// --- replacements ------------------------------------------------

// b31OpponentsCreaturesEnterTapped is "Creatures your opponents
// control enter tapped" (Kinjalli's Sunwing) — Urabrask the Hidden's
// CR 614 replacement on the entering creature's move, applied when
// its controller is not the source's. A cast, reanimated, fetched or
// flickered creature runs the entry pipeline and arrives tapped;
// an opponent's creature TOKEN does not, because token creation
// skips the zone-move pipeline (the Kismet gap, declared on the
// card).
func b31OpponentsCreaturesEnterTapped(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller == src.Controller {
				return false
			}
			return entering.IsCreature()
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// --- statics -----------------------------------------------------

// b31LandCreaturesYouControl is Blossoming Tortoise's "Land creatures
// you control" — post-layer types, so an animated land counts and a
// creature that became a land counts.
func b31LandCreaturesYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.IsCreature() && target.IsLand()
}

// b31NotACreatureWhileGraveyardBelow is The Warring Triad's "As long
// as there are fewer than N cards in your graveyard, this isn't a
// creature": a layer 4 static on itself that drops the Creature type
// — and with it the creature subtypes, CR 205.1b — while the
// controller's graveyard is short. Every other type the card prints
// (Artifact) stays. Read on every recompute, so the god wakes the
// moment the layer cache next rebuilds.
func b31NotACreatureWhileGraveyardBelow(n int) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && b31GraveyardSize(g, source.Controller) < n
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			kept := make([]string, 0, len(c.Types))
			for _, t := range c.Types {
				if t != "Creature" {
					kept = append(kept, t)
				}
			}
			c.Types = kept
			c.Subtypes = nil
		},
	}
}

// --- state reads ---------------------------------------------------

// b31GraveyardSize is the number of cards in `player`'s graveyard.
func b31GraveyardSize(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	return p.Graveyard.Size()
}

// b31CurrentPowerOf is a battlefield creature's power right now —
// layers brought current first, then counters (CurrentPower) — for
// Soul's Majesty's "the power of target creature". Zero for a card
// that is not on the battlefield.
func b31CurrentPowerOf(g *game.Game, cardID uuid.UUID) int {
	if !onBattlefield(g, cardID) {
		return 0
	}
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.LookupCardForEffect(cardID)
	if !ok {
		return 0
	}
	return c.CurrentPower()
}

// b31OpponentControlsNonlandPermanent reports whether any nonland
// permanent on the battlefield is controlled by someone other than
// `controller` — the question Tiller Engine's targeted declaration
// asks before it is declared, so the untap is never dropped with an
// empty legal set.
func b31OpponentControlsNonlandPermanent(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller && !c.IsLand() {
			return true
		}
	}
	return false
}

// --- trigger conditions ------------------------------------------

// b31LandYouControlEnteredTapped is Tiller Engine's condition — Amulet
// of Vigor's read narrowed to lands: a land entered under the
// source's controller's control and was tapped when its entry was
// announced. Every entry path stamps Tapped before EventETB fires,
// so a land that is tapped at the event entered that way.
func b31LandYouControlEnteredTapped(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsLand() && c.Tapped
}

// b31NontokenArtifactYouControlEntered is Weapons Manufacturing's
// condition: a nontoken artifact entered under the source's
// controller's control. Post-layer types, so an animated artifact
// counts; the Munitions the Manufacturing itself makes are tokens
// and do not chain.
func b31NontokenArtifactYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && c.IsArtifact() && !IsToken(c)
}

// b31YouSacrificedAFood is Rapacious Guest's second condition: the
// source's controller sacrificed a Food. EventSacrifice fires before
// the zone move, so the Food is still on the battlefield to be read
// — effective subtypes, so a Food that is also a creature counts.
func b31YouSacrificedAFood(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.HasSubtype("Food")
}

// b31YouPlayedALandOrCastASpell is The Endstone's condition — one
// printed ability with two trigger conditions, watching the land
// play (b20LandPlayed, with the play-versus-return arithmetic that
// helper documents) and the cast on one declaration.
func b31YouPlayedALandOrCastASpell(ev game.Event, source *game.Card, g *game.Game) bool {
	switch ev.Kind {
	case game.EventCast:
		return ev.Actor == source.Controller
	case game.EventZoneMove:
		if ev.Actor != source.Controller {
			return false
		}
		ok := b20LandPlayed(ev, g)
		return ok
	}
	return false
}

// b31MunitionsYouControlLeft is the condition of the trigger Weapons
// Manufacturing carries for its tokens: a Munitions token the
// source's controller controls left the battlefield, by any route.
// The token is read post-move (it persists in the graveyard until
// the next state check, and in exile or a hand indefinitely).
func b31MunitionsYouControlLeft(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && IsToken(c) && c.Name == "Munitions" && c.Controller == source.Controller
}

// --- effect bodies -----------------------------------------------

// b31TillerEngineLabel is the stack label both declarations of
// Tiller Engine's trigger share.
const b31TillerEngineLabel = "Tiller Engine — untap that land, or tap a nonland permanent an opponent controls"

// b31TapChosenOrUntapLand is Tiller Engine's body for one land: with
// a nonland permanent chosen (and still legal) it is tapped; with
// none chosen, the land that entered is untapped — if it is still
// on the battlefield and still tapped. A chosen permanent that
// became illegal never reaches here — the engine counters the
// trigger first (CR 608.2b) — so the untap can only run for a
// trigger whose controller chose it.
func b31TapChosenOrUntapLand(entered uuid.UUID) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return TapTarget{Target: t.ID}.Apply(ctx)
		}
		if !onBattlefield(g, entered) {
			return nil
		}
		if c, ok := g.LookupCardForEffect(entered); !ok || !c.Tapped {
			return nil
		}
		return UntapTarget{Target: entered}.Apply(ctx)
	}
}

// b31BlossomingTortoiseLabel is the stack label both declarations
// of Blossoming Tortoise's trigger share.
const b31BlossomingTortoiseLabel = "Blossoming Tortoise — mill three, then return a land card tapped"

// b31LifeBecomes sets `player`'s life total to `total` the way CR
// 119.5 says to: the player gains or loses the difference, so a
// lifegain or life-loss trigger sees the change. A total already
// equal to the target changes nothing.
func b31LifeBecomes(ctx *Context, player uuid.UUID, total int) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Life == total {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, total-p.Life)
}

// b31ChosenOpponentLosesLife is Rapacious Guest's leave trigger: the
// opponent chosen when the trigger went on the stack loses `amount`
// life — the power the Guest had as it left, read in Build.
func b31ChosenOpponentLosesLife(amount int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if amount <= 0 {
			return nil
		}
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -amount)
		}
		return nil
	}
}

// b31DamageChosenTargetFrom is the body of the trigger Weapons
// Manufacturing carries for a Munitions token: `amount` damage to
// the target chosen when the trigger went on the stack, dealt by
// the token that left (`token`) rather than by the Manufacturing —
// a colorless source, as printed, so a red-damage payoff does not
// see it.
func b31DamageChosenTargetFrom(token uuid.UUID, amount int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			return DealDamage{Source: token, Target: t.ID, Amount: amount}.Apply(ctx)
		}
		return nil
	}
}

// b31CountersOnChosenCreature is Fain's first activation: two +1/+1
// counters on the creature chosen at announce, if it is still legal.
func b31CountersOnChosenCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
	}
	return nil
}

// b31UntapSelf untaps the item's source if it is still on the
// battlefield — Fain's "{3}{B}: Untap Fain".
func b31UntapSelf(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// b31MillChosenPlayer is Nephalia Drownyard's activation: the player
// chosen at announce mills `n` cards — the rest of their library when
// it holds fewer (CR 701.17b).
func b31MillChosenPlayer(n int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			return MillCards{Player: t.ID, N: n}.Apply(ctx)
		}
		return nil
	}
}

// b31FetchPlainsTapped is Flagstones of Trokair's search: a Plains
// card — any land with the Plains type — onto the battlefield
// tapped, then shuffle. The "may" was answered when the trigger
// fired; the S22 chooser still lets the searcher fail to find.
func b31FetchPlainsTapped(g *game.Game, item *game.StackItem) error {
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     IsLandWithSubtype("Plains"),
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Flagstones of Trokair — a Plains card, onto the battlefield tapped",
	}.Apply(NewContext(g, item))
}

// b31RemoveMiningCounterOrSacrifice is Gemstone Mine's rider: one
// mining counter comes off as the mana lands, and with none left the
// Mine is sacrificed — the printed "If there are no mining counters
// on this land, sacrifice it", which is why the third activation
// both taps for mana and loses the land.
func b31RemoveMiningCounterOrSacrifice(g *game.Game, _, source uuid.UUID) error {
	if err := g.AddCounterForEffect(source, "mining", -1); err != nil {
		return err
	}
	c, ok := g.LookupCardForEffect(source)
	if !ok || c.Counters["mining"] > 0 {
		return nil
	}
	return g.SacrificePermanentForEffect(source)
}

// b31HasMiningCounter gates Gemstone Mine's activation: at least one
// mining counter must be there to remove (CR 602.5a — a cost that
// cannot be paid cannot be announced).
func b31HasMiningCounter(g *game.Game, _, source uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(source)
	return ok && c.Counters["mining"] > 0
}
