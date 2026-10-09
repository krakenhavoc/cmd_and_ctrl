package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_fra_creature_b_helpers.go — small helpers for the
// Reality Fracture creature slice fra-creature-b. Names are prefixed
// rfCreatureB so a parallel slice cannot collide with them.

// rfCreatureBGreatestToughnessYouControl is Ghalta the Immovable's X:
// the greatest current toughness among creatures the caster controls,
// the toughness twin of greatestPowerAmongCreaturesYouControl (the
// Henge's X, which Ghalta the Unstoppable reuses). The spell itself is
// on the stack, so it never counts.
func rfCreatureBGreatestToughnessYouControl() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		best := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.Controller != q.Controller || !c.IsCreature() {
				continue
			}
			if t := c.CurrentToughness(); t > best {
				best = t
			}
		}
		return best
	}
}

// rfCreatureBEntersWithCounterFromGraveyard is "this enters with N
// <kind> counters on it" for a card that only gets them when it
// returns from a graveyard (Gallia, Tragic Host) — b10EntersWithCounters
// gated on the move's origin. A cast from hand (or any other zone)
// enters bare.
func rfCreatureBEntersWithCounterFromGraveyard(kind string, n int, label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
				ev.OldZone == game.ZoneGraveyard && src != nil && ev.CardID == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, n)
			return nil
		},
		Label: label,
	}
}

// rfCreatureBEnteredThisTurn is the target predicate "creature that
// entered this turn" (Gallia, the Merrymaker): the per-object entry
// tally, so a creature that entered, left and came back counts only for
// its new object, and a creature that entered during another player's
// turn (still summoning-sick on yours) does not.
func rfCreatureBEnteredThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.EnteredThisTurn(c.InstanceID)
	}
}

// rfCreatureBYouCastNoSpellThisTurn is the When for "if you didn't cast
// a spell this turn" (Edgar, Moonlit Sovereign), for composing with
// AllOf. The tally is the source controller's, bumped as a spell is
// announced, so a spell still on the stack counts.
func rfCreatureBYouCastNoSpellThisTurn(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.CastTallyFor(source.Controller).Total == 0
}

// rfCreatureBScriedOrSurveilledThisTurn is "you've scried or surveilled
// this turn" (Desperate Futurescribe): this turn's EventScry or
// EventSurveil with `player` as the actor. Neither event is emitted for
// a scry or surveil that looked at nothing, because none happened.
// Caller holds g.mu.
func rfCreatureBScriedOrSurveilledThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if (ev.Kind == game.EventScry || ev.Kind == game.EventSurveil) && ev.Actor == player {
			return true
		}
	}
	return false
}

// GingerbruteToken is Ginger, Queen of Sweets' token: a 1/1 colourless
// Food Golem artifact creature with haste, "{1}: This token can't be
// blocked this turn except by creatures with haste." and "{2}, {T},
// Sacrifice this token: You gain 3 life." It is the Gingerbrute card
// (gingerbrute.go) as a token, so it keeps both activated abilities and
// is a Food as well as a Golem.
func GingerbruteToken() game.Card { return tokenFromCatalog(printedGingerbruteToken) }

// printedGingerbruteToken is the Gingerbrute token as PRINTED. Its
// abilities are registered under game.TokenKey("gingerbrute") from the
// tokenTemplates list, not carried on the instance.
func printedGingerbruteToken() tokenTemplate {
	return tokenTemplate{
		Slug: "gingerbrute",
		Card: game.Card{
			Name:      "Gingerbrute",
			TypeLine:  "Token Artifact Creature — Food Golem",
			Power:     1,
			Toughness: 1,
			Keywords:  []string{"haste"},
		},
		Text: "Haste\n{1}: This token can't be blocked this turn except by creatures with haste.\n{2}, {T}, Sacrifice this token: You gain 3 life.",
		Activated: []game.ActivatedAbilityShape{
			{
				Label:  "{1}: This token can't be blocked this turn except by creatures with haste.",
				Cost:   game.AbilityCost{Mana: "{1}"},
				Effect: rfCreatureBGingerbruteEvade,
			},
			{
				Label: "{2}, {T}, Sacrifice this token: You gain 3 life.",
				Cost:  game.AbilityCost{Mana: "{2}", Tap: true, SacrificeSelf: true},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(NewContext(g, item))
				},
			},
		},
	}
}

// rfCreatureBBasicLandTypesAmong is domain (the ability word): the
// number of distinct basic land types among the lands `controller`
// controls, read from post-layer subtypes so a land that gained a type
// counts. Five at most.
func rfCreatureBBasicLandTypesAmong(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, sub := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if b08LandsWithSubtypeControlled(g, controller, sub) > 0 {
			n++
		}
	}
	return n
}

// rfCreatureBOpponentDealtCombatDamageOnYourTurn is the When for
// "one or more of your opponents are dealt combat damage during your
// turn" (Fblthp, Impossibly Lost): combat damage above zero dealt to a
// player who is not the source controller, while it is the source
// controller's turn. The damage source is unrestricted.
func rfCreatureBOpponentDealtCombatDamageOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if ev.Target == source.Controller || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	return IsYourTurn(g, source.Controller)
}

// rfCreatureBFblthpImpossiblyLostResolve is the resolution of Fblthp,
// Impossibly Lost's trigger: draw two, win on an empty library, then
// the owner shuffles Fblthp into their library.
func rfCreatureBFblthpImpossiblyLostResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (DrawCards{Player: item.Controller, N: 2}).Apply(ctx); err != nil {
		return err
	}
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Library.Size() == 0 {
		if err := (WinTheGame{Player: item.Controller}).Apply(ctx); err != nil {
			return err
		}
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	owner := c.Owner
	return g.TuckToLibraryThenForEffect(item.SourceCardID, game.TuckOptions{}, func(g *game.Game, _ bool) error {
		return g.ShuffleLibraryForEffect(owner)
	})
}

// rfCreatureBFinalityCounter is the counter kind of a finality counter
// (CR 122.1): a creature with one that would die is exiled instead. The
// replacement lives on Grim Repriser, the only card that makes one.
const rfCreatureBFinalityCounter = "finality"

// rfCreatureBAnOpponentWasDealtNoncombatDamage is Grim Repriser's
// activation condition: this turn some opponent of `controller` was
// dealt damage above zero that was not combat damage. Reads the turn's
// EventDealDamage log, so damage dealt before the Repriser died counts.
// Caller holds g.mu.
func rfCreatureBAnOpponentWasDealtNoncombatDamage(g *game.Game, controller uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || ev.Combat || ev.Amount <= 0 {
			continue
		}
		if ev.Target != controller && g.PlayerByIDForEffect(ev.Target) != nil {
			return true
		}
	}
	return false
}

// rfCreatureBDistinctNames is the Validate for "cards with different
// names": no two picked cards share a name.
func rfCreatureBDistinctNames(picked []game.Card) bool {
	seen := make(map[string]bool, len(picked))
	for _, c := range picked {
		if seen[c.Name] {
			return false
		}
		seen[c.Name] = true
	}
	return true
}

// rfCreatureBModeTargetDoes lifts "do <f> to the chosen target" into a
// modal bullet's body: the bullet's own target group, first target
// still legal as it resolves (CR 608.2b). Divining Duelist's tap and
// untap bullets differ only in f.
func rfCreatureBModeTargetDoes(f func(ctx *Context, id uuid.UUID) error) func(*game.StackItem, *Context, int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		if t, ok := ModeTarget(ctx, occ); ok {
			return f(ctx, t.ID)
		}
		return nil
	}
}

// rfCreatureBGingerbruteEvade is the Gingerbrute's "{1}: This can't be
// blocked this turn except by creatures with haste", shared by the
// Gingerbrute card (gingerbrute.go) and the token Ginger makes.
func rfCreatureBGingerbruteEvade(g *game.Game, item *game.StackItem) error {
	return CantBeBlockedThisTurnExceptBy{
		Target:   item.SourceCardID,
		Keywords: []string{"haste"},
		Text:     "creatures with haste",
		Label:    "Gingerbrute — can't be blocked except by creatures with haste",
	}.Apply(NewContext(g, item))
}
