package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch20_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 20 (#313, `edhrec_rank` 2133–2234). Own file per
// the #231 convention; every package-level name carries the b20
// prefix because other batches land beside this one.
//
// What is NOT here, because main already had it: "this permanent
// enters" is b06SelfETB, "a creature you control entered" is
// enteredUnderYourControl, the artifact count is
// b15ArtifactsControlled, the creature count is
// b04CreaturesControlled, "sacrifice an artifact" is
// b10SacrificeAnArtifact, the Zombie lord is TribalAnthem, the
// looter is lootOne, "target creature card in a graveyard" is
// targetCreatureInAnyGraveyard, the DMU dual and the Guildgate are
// rows in their cycle tables, and the 3/3 Beast / 2/2 Zombie /
// Treasure templates live in tokens.go.

// --- token templates ---------------------------------------------

// b20GolemToken is one of Triplicate Titan's three 3/3 colorless
// Golem artifact creatures, each carrying exactly one of the Titan's
// keywords.
func b20GolemToken(keyword string) game.Card {
	return game.Card{
		Name:      "Golem",
		TypeLine:  "Token Artifact Creature — Golem",
		Power:     3,
		Toughness: 3,
		Keywords:  []string{keyword},
	}
}

// --- trigger conditions ------------------------------------------

// b20LandPlayed reports whether ev is a land PLAY (CR 305.1) —
// Horn of Greed's "whenever a player plays a land" for any player,
// and the land half of Prosper's "whenever you play a card from
// exile". Since #1326 the engine stamps the distinction itself
// (Event.Played, CR 305.4) on the settled entry, from whichever zone
// it was played: hand, exile (an impulse grant) or a graveyard (a
// Crucible grant) all set it, and every "put" path — a fetchland, a
// reanimation, a flickered or returned land — leaves it false. No
// zone-origin guess and no land-drop-tally arithmetic needed any
// more; both are gone along with the ambiguity they could not always
// resolve (Horn of Greed's and Prosper's caveats).
func b20LandPlayed(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneBattlefield || !ev.Played {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand()
}

// b20PlayedACardFromExile is Prosper's Pact Boon: the source's
// controller cast a spell from exile (EventCast stamps the origin in
// OldZone, the Appa shape) or played a land from exile
// (b20LandPlayed with an exile origin).
func b20PlayedACardFromExile(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Actor != source.Controller || ev.OldZone != game.ZoneExile {
		return false
	}
	if ev.Kind == game.EventCast {
		return true
	}
	ok := b20LandPlayed(ev, g)
	return ok
}

// b20ArtifactEntered is Grinding Station's "whenever an artifact
// enters" — any artifact, under anyone's control, the Station itself
// included.
func b20ArtifactEntered(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsArtifact()
}

// b20LandTappedForMana is Manabarbs' "whenever a player taps a land
// for mana": a mana ability fired (EventManaAbilityActivated — the
// manual activation and the auto-tapper both emit it) whose source
// is a land that is now tapped. The tapped check is what "taps ...
// for mana" adds over "activates a mana ability of" — a land whose
// mana ability has no tap cost is not one.
func b20LandTappedForMana(ev game.Event, g *game.Game) bool {
	if ev.Kind != game.EventManaAbilityActivated || ev.Source == uuid.Nil || ev.Actor == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	return ok && c.IsLand() && c.Tapped
}

// --- statics -----------------------------------------------------

// b20JetmirClause is one line of Jetmir, Nexus of Revels: "Creatures
// you control get +1/+0 and have <keyword> as long as you control
// <threshold> or more creatures." Two statics — the layer 7c bonus
// and the layer 6 grant — sharing one condition, read live on every
// recompute so the third creature entering (or the seventh leaving)
// moves the line at once. Jetmir counts himself, as printed.
func b20JetmirClause(threshold int, keyword string) []game.StaticAbility {
	applies := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller &&
			b04CreaturesControlled(g, source.Controller) >= threshold
	}
	return []game.StaticAbility{
		{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: applies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
			},
		},
		{
			Layer:     game.Layer6Ability,
			AppliesTo: applies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, k := range c.Abilities {
					if k == keyword {
						return
					}
				}
				c.Abilities = append(c.Abilities, keyword)
			},
		},
	}
}

// --- effect bodies -----------------------------------------------

// b20TutorOnETB is the Trinket Mage shape — "When this creature
// enters, you may search your library for a <pred> card, reveal it,
// put it into your hand, then shuffle" — for Fierce Empath and
// Tribute Mage. "You may" is the search prompt's decline; the pick is
// revealed and goes to hand.
func b20TutorOnETB(label, reason string, pred func(game.Card) bool) game.TriggeredAbility {
	return WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:    item.Controller,
			Predicate: pred,
			Dest:      game.ZoneHand,
			Limit:     1,
			Reveal:    true,
			Shuffle:   true,
			Optional:  true,
			Reason:    reason,
		}.Apply(NewContext(g, item))
	})
}

// b20ExileTopUntilEndOfNextTurn is Prosper's Mystic Arcanum: "exile
// the top card of your library. Until the end of your next turn, you
// may play that card." — b19ExileTopTwoUntilEndOfNextTurn (Reckless
// Impulse) generalised to N cards, and the same one-line duration
// since #945: ADR 0063's `UntilEndOfYourNextTurnDuration` is the
// printed clause, keyed on the controller's seat-turn count rather
// than on the shared round number.
func b20ExileTopUntilEndOfNextTurn(g *game.Game, item *game.StackItem, n int) error {
	controller := item.Controller
	_, err := g.ExileTopWithPermissionForEffect(controller, controller, n, game.CastPermission{
		Duration: g.UntilEndOfYourNextTurnDuration(controller),
	})
	return err
}

// b20EachPlayerDrawsAndGainsOne is Kwain's activation: each player
// draws a card, then each player who drew gains 1 life — APNAP from
// the active seat, eliminated seats skipped. A player whose library
// is empty is treated as declining the draw (the one case where the
// printed "may" is a real decision) and gains nothing.
func b20EachPlayerDrawsAndGainsOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	var drew []uuid.UUID
	for i := 0; i < numSeats; i++ {
		p := g.Seats[(start+i)%numSeats]
		if p == nil || p.Eliminated || p.Library == nil || p.Library.Size() == 0 {
			continue
		}
		if err := (DrawCards{Player: p.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
		drew = append(drew, p.ID)
	}
	for _, id := range drew {
		if err := (GainLife{Player: id, Amount: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// b20GainLifeEqualToToughnessOf builds the Righteous Valkyrie body
// for an entering creature: "you gain life equal to that creature's
// toughness", read as the trigger resolves — current toughness,
// counters and anthems included — falling back to the toughness it
// had when the trigger fired if it has since left (CR 608.2h's
// last-known value).
func b20GainLifeEqualToToughnessOf(entered uuid.UUID, fallback int) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		toughness := fallback
		if z := g.FindCardZoneForEffect(entered); z != nil && z.Kind == game.ZoneBattlefield {
			if c, ok := g.LookupCardForEffect(entered); ok {
				toughness = c.CurrentToughness()
			}
		}
		return GainLife{Player: item.Controller, Amount: toughness}.Apply(NewContext(g, item))
	}
}
