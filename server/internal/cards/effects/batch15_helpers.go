package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch15_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 15 (#308, `edhrec_rank` 1626–1726). Own file per
// the #231 convention; every package-level name carries the b15
// prefix.
//
// What is NOT here, because main already had it: "whenever you cast
// a noncreature spell" is b10NoncreatureSpellCastByYou, "an instant
// or sorcery" is b12InstantOrSorceryCastByYou, "another creature
// you control enters" is b13AnotherCreatureYouControlEntered, "each
// opponent loses N life" is eachOpponentLosesLife, "N damage to each
// opponent" is damageToEachOpponent, "draw, then discard" is
// lootOne, "creature cards in your graveyard" is
// b11CreatureCardsInGraveyard, the tapped-Treasure payout is
// b13CreateTappedTreasures, the Overlook fetch-land is
// b08OverlookLand, "this permanent enters" is b06SelfETB, and the
// mana value of a card off the stack is manaValueOf.

// --- token templates ---------------------------------------------

// b15GreenInsectToken is Hornet Queen's 1/1 green Insect with flying
// and deathtouch.
func b15GreenInsectToken() game.Card {
	return game.Card{
		Name:      "Insect",
		TypeLine:  "Token Creature — Insect",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"G"},
		Keywords:  []string{"flying", "deathtouch"},
	}
}

// b15BlueRedInsectToken is The Locust God's 1/1 blue and red Insect
// with flying and haste.
func b15BlueRedInsectToken() game.Card {
	return game.Card{
		Name:      "Insect",
		TypeLine:  "Token Creature — Insect",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"U", "R"},
		Keywords:  []string{"flying", "haste"},
	}
}

// b15KoboldToken is Kher Keep's 0/1 red Kobold named Kobolds of Kher
// Keep — the name is printed, and it is what Rohgahh and the other
// Kobold cards read.
func b15KoboldToken() game.Card {
	return game.Card{
		Name:      "Kobolds of Kher Keep",
		TypeLine:  "Token Creature — Kobold",
		Power:     0,
		Toughness: 1,
		Colors:    []string{"R"},
	}
}

// b15ConstructToken is Simulacrum Synthesizer's 0/0 colorless
// Construct artifact creature. The printed token also carries "This
// token gets +1/+1 for each artifact you control"; a token template
// has no static-ability slot and no oracle ID for the catalog to key
// one on, so the Synthesizer carries that static on the token's
// behalf — see simulacrum_synthesizer.go for what that costs.
func b15ConstructToken() game.Card {
	return game.Card{
		Name:      "Construct",
		TypeLine:  "Token Artifact Creature — Construct",
		Power:     0,
		Toughness: 0,
	}
}

// --- trigger conditions ------------------------------------------

// b15OpponentCastSpell is "whenever an opponent casts a spell" —
// Sunscorch Regent. Any spell, any opponent.
func b15OpponentCastSpell(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventCast && ev.Actor != uuid.Nil && ev.Actor != source.Controller
}

// b15RedSpellCastByYou is Runaway Steam-Kin's condition: the
// controller cast a red spell. Colours are read off the card on the
// stack — Scryfall's stamped list, or the mana cost for a fixture —
// so a red-and-green Boros Charm cast off a hybrid cost still counts
// as red, as printed.
func b15RedSpellCastByYou(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && spell.HasColor("R")
}

// b15AnotherNontokenCreatureYouControlEntered is Soul of the
// Harvest's condition — b13AnotherCreatureYouControlEntered with the
// token exclusion.
func b15AnotherNontokenCreatureYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsCreature() && !IsToken(c)
}

// b15AnotherCreatureDied is "whenever another creature dies" —
// Poison-Tip Archer. Anyone's creature, the controller's own
// included; the source's own death is excluded by ID.
func b15AnotherCreatureDied(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == source.InstanceID {
		return false
	}
	_, died := diedCreature(ev, g)
	return died
}

// b15AnotherBigArtifactYouControlEntered is Simulacrum Synthesizer's
// condition: another artifact the controller controls entered with
// mana value 3 or greater. Mana value is read off the permanent
// (CR 202.3b — X is zero, a token with no cost is zero).
func b15AnotherBigArtifactYouControlEntered(ev game.Event, source *game.Card, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.IsArtifact() && manaValueOf(c) >= 3
}

// b15EndStepBegan is "at the beginning of each end step" — any
// player's. The Gaffer, Smuggler's Share.
func b15EndStepBegan(ev game.Event) bool {
	return ev.Kind == game.EventBeginEndStep
}

// --- per-turn tallies read back off the log ----------------------
//
// The engine keeps no per-turn counters for life gained, cards drawn
// or lands entered, so these are the b06EnteredThisTurn walk: back
// through the event log to the current turn's EventBeginUpkeep.
// Every turn passes through its upkeep and nothing below can happen
// during the untap step before it, so "since the upkeep began" is
// "this turn".

// b15LifeGainedThisTurn is the total life `player` gained this turn
// — The Gaffer's "3 or more". Positive EventChangeLife only: a
// lifelink hit, a GainLife primitive and a drain's gain half all
// emit one, and a loss (negative Amount) is not a gain.
func b15LifeGainedThisTurn(g *game.Game, player uuid.UUID) int {
	return g.TurnTallyFor(player).LifeGained
}

// b15CardsDrawnThisTurn is how many cards each player drew this turn
// — Smuggler's Share's first count. EventDrawCard fires once per
// card with the drawer in Actor.
func b15CardsDrawnThisTurn(g *game.Game) map[uuid.UUID]int {
	drawn := map[uuid.UUID]int{}
	for id, t := range g.TurnTally.Players {
		if t.CardsDrawn > 0 {
			drawn[id] = t.CardsDrawn
		}
	}
	return drawn
}

// b15LandsEnteredThisTurn is how many lands entered the battlefield
// under each player's control this turn — Smuggler's Share's second
// count. EventETB carries only the card, so the land is looked up
// where it sits now for its type and controller; the Controller
// field survives a zone move, so a fetchland that entered and
// sacrificed itself still counts, as printed. A land that has since
// changed control is counted for its new controller — weaker for
// its old one, never stronger.
func b15LandsEnteredThisTurn(g *game.Game) map[uuid.UUID]int {
	lands := map[uuid.UUID]int{}
	for id, t := range g.TurnTally.Players {
		if t.LandsEntered > 0 {
			lands[id] = t.LandsEntered
		}
	}
	return lands
}

// b15ResolvedThisTurn counts how many times an ability of `source`
// with the given stack label has RESOLVED this turn — Gala
// Greeters' "hasn't been chosen this turn". The engine emits
// EventResolve, carrying the source and the label, immediately
// before it runs an item's effect, so during the Nth resolution the
// count is N. Counting resolutions rather than triggers (which
// b11TriggeredThisTurn does) is what keeps two triggers queued
// together from both reading the same tally.
func b15ResolvedThisTurn(g *game.Game, source uuid.UUID, label string) int {
	return g.ResolvedThisTurn(source, label)
}

// --- predicates --------------------------------------------------

// b15IsEnchanted reports whether an Aura is attached to the
// permanent — "creatures that aren't enchanted" is its negation
// (Winds of Rath). Walks the battlefield for anything whose
// AttachedTo points at the card (#379's relation) and is an Aura
// right now; an Equipment does not enchant. Reads effective
// subtypes, so it is safe inside a mass-effect predicate.
func b15IsEnchanted(g *game.Game, c game.Card) bool {
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(c.InstanceID) && a.IsAura() {
			return true
		}
	}
	return false
}

// b15NotEnchanted is the predicate form.
func b15NotEnchanted() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return !b15IsEnchanted(g, c)
	}
}

// b15ArtifactWithManaAbilityOrBasicLand is Moonsilver Key's search
// clause. "An artifact card with a mana ability" is answered by the
// same lookup the battlefield uses (game.ManaAbilitiesForCard reads
// the catalog by oracle ID), so an artifact the catalog knows a mana
// ability for qualifies and one it does not know is not found —
// weaker than printed, declared on the card.
func b15ArtifactWithManaAbilityOrBasicLand(c game.Card) bool {
	if IsBasicLand(c) {
		return true
	}
	return c.IsArtifact() && len(game.ManaAbilitiesForCard(c)) > 0
}

// --- board reads -------------------------------------------------

// b15OtherCreaturesControlledWithSameName counts the OTHER creatures
// `controller` controls whose name matches `name` — Mirror Box's
// third clause. Names are read post-layer, so a Clone or a token
// copy counts under the name it has now. Walks the live slice: it
// runs inside a layer recompute, at layer 7, by which point layer 1
// has settled every name and layer 4 every type.
func b15OtherCreaturesControlledWithSameName(g *game.Game, controller, except uuid.UUID, name string) int {
	if name == "" {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == except || c.Controller != controller || !c.IsCreature() {
			continue
		}
		if c.Effective().Name == name {
			n++
		}
	}
	return n
}

// b15IsFirstOfItsNameControlledBy reports whether `source` is the
// earliest-entered permanent with its oracle ID that its controller
// controls — the dedup a static that stands in for a TOKEN's own
// ability needs (Simulacrum Synthesizer's Constructs): two
// Synthesizers must size a Construct once, not twice. Ties on the
// timestamp (a fixture board) break on instance ID so exactly one
// wins.
func b15IsFirstOfItsNameControlledBy(g *game.Game, source *game.Card) bool {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == source.InstanceID || c.Controller != source.Controller || c.OracleID != source.OracleID {
			continue
		}
		if c.EnteredBattlefieldAt < source.EnteredBattlefieldAt {
			return false
		}
		if c.EnteredBattlefieldAt == source.EnteredBattlefieldAt && c.InstanceID.String() < source.InstanceID.String() {
			return false
		}
	}
	return true
}

// b15ArtifactsControlled counts the artifacts `controller` controls
// on the post-layer type line. Walks the live slice: it runs inside
// a layer recompute.
func b15ArtifactsControlled(g *game.Game, controller uuid.UUID) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsArtifact() {
			n++
		}
	}
	return n
}

// b15OnBattlefield reports whether the card is on the battlefield
// right now — the guard a trigger's Effect runs before putting a
// counter on its own source, since AddCounter does not gate on zone.
func b15OnBattlefield(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneBattlefield
}

// --- effect bodies -----------------------------------------------

// b15ReturnListedCardsFromGraveyardToHand is the delayed-trigger
// body The Locust God schedules: return every card the item carries
// that is in a graveyard to its owner's hand. Package-level so the
// delayed trigger captures nothing. A card that is not in a
// graveyard any more — reanimated, exiled, tucked into the command
// zone by CR 903.9 — is a different object and is left alone.
func b15ReturnListedCardsFromGraveyardToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b15CanopyLand is the Horizon Canopy cycle, the shape Fiery Islet
// established:
//
//	"{T}, Pay 1 life: Add {A} or {B}.
//	 {1}, {T}, Sacrifice this land: Draw a card."
//
// The life is a COST (Mana Confluence's shape — validated before the
// land taps, refused at zero life), not a painland's rider; the two
// printed colours are not narrowed to the commander's identity. The
// cash-in is Buried Ruin's three-component cost with a draw.
func b15CanopyLand(oracleID, name, a, b string) Spec {
	produced := "{" + a + "|" + b + "}"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true, Life: 1},
			Produced:                produced,
			Label:                   "{T}, Pay 1 life: Add {" + a + "} or {" + b + "}",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice this land: Draw a card.",
			Cost:  Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	}
}
