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
// targetCreatureInAnyGraveyard, the "until the end of your next
// turn" re-stamp is b19EndImpulseGrantWithThisTurn, the DMU dual and
// the Guildgate are rows in their cycle tables, and the 3/3 Beast /
// 2/2 Zombie / Treasure templates live in tokens.go.

// --- token templates ---------------------------------------------

// b20WhiteGoatToken is Trading Post's 0/1 white Goat.
func b20WhiteGoatToken() game.Card {
	return game.Card{
		Name:      "Goat",
		TypeLine:  "Token Creature — Goat",
		Power:     0,
		Toughness: 1,
		Colors:    []string{"W"},
	}
}

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

// b20RedWhiteSoldierHasteToken is Assemble the Legion's 1/1 red and
// white Soldier with haste. Its own template rather than
// SoldierToken: the colours and the haste are printed.
func b20RedWhiteSoldierHasteToken() game.Card {
	return game.Card{
		Name:      "Soldier",
		TypeLine:  "Token Creature — Soldier",
		Power:     1,
		Toughness: 1,
		Colors:    []string{"R", "W"},
		Keywords:  []string{"haste"},
	}
}

// --- trigger conditions ------------------------------------------

// b20LandPlayed reports whether ev is a land PLAY (CR 305.1) and
// returns the land — Horn of Greed's "whenever a player plays a
// land" for any player, and the land half of Prosper's "whenever you
// play a card from exile". The engine emits no land-play event, so
// the trigger watches the EventZoneMove that precedes every
// battlefield entry and reads where the land came FROM:
//
//   - The library, the stack and the battlefield are origins a land
//     play can never have (Cultivate, a fetchland, a cast permanent,
//     a control change), so those are never a play.
//   - Nothing in the engine moves a land from a HAND to the
//     battlefield except playing it (the "put a land card from your
//     hand onto the battlefield" clause has no primitive — the
//     Spelunking / Eureka Moment gap), so a hand origin is always a
//     play.
//   - Exile and a graveyard are ambiguous: a land played through a
//     permission (Prosper's, Crucible of Worlds') and a land an
//     effect RETURNS (a flickered land, Splendid Reclamation) both
//     arrive from there under the same Actor. The land-drop tally
//     the engine keeps for the legal-move enumerator separates them:
//     both play paths bump LandsPlayedThisTurn BEFORE emitting the
//     move, and no return path bumps it. So this event is a play
//     exactly when the tally is one more than the land entries by
//     this player earlier this turn from the same three zones. When
//     an earlier return this turn makes that arithmetic ambiguous
//     the answer is "not a play" — weaker than printed for the one
//     turn, never stronger. Declared on both cards.
func b20LandPlayed(ev game.Event, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneBattlefield || ev.Actor == uuid.Nil {
		return game.Card{}, false
	}
	if !b20LandPlayOrigin(ev.OldZone) {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsLand() {
		return game.Card{}, false
	}
	if ev.OldZone == game.ZoneHand {
		return c, true
	}
	prior := 0
	for i := len(g.Events) - 1; i >= 0; i-- {
		e := g.Events[i]
		if e.Seq >= ev.Seq {
			continue
		}
		if e.Kind == game.EventBeginUpkeep {
			break
		}
		if e.Kind != game.EventZoneMove || e.NewZone != game.ZoneBattlefield || e.Actor != ev.Actor || !b20LandPlayOrigin(e.OldZone) {
			continue
		}
		if earlier, ok := g.LookupCardForEffect(e.CardID); ok && earlier.IsLand() {
			prior++
		}
	}
	return c, g.LandsPlayedThisTurnFor(ev.Actor)-prior == 1
}

// b20LandPlayOrigin is the set of zones a land can be PLAYED from:
// the hand, exile (an impulse grant) and a graveyard (a Crucible
// grant).
func b20LandPlayOrigin(z game.ZoneKind) bool {
	return z == game.ZoneHand || z == game.ZoneExile || z == game.ZoneGraveyard
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
	_, ok := b20LandPlayed(ev, g)
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
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: b06SelfETB,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			return game.NewTriggeredItem(source, label,
				func(g *game.Game, item *game.StackItem) error {
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
		},
	}
}

// b20ExileTopUntilEndOfNextTurn is Prosper's Mystic Arcanum: "exile
// the top card of your library. Until the end of your next turn, you
// may play that card." — b19ExileTopTwoUntilEndOfNextTurn (Reckless
// Impulse) generalised to N cards, with the same two-part duration:
// the grant is stamped two rounds out as a backstop and a CR 603.7
// delayed trigger at the beginning of the controller's next upkeep
// re-stamps it to end with that turn (b19EndImpulseGrantWithThisTurn,
// which touches only cards still in exile under this grant). See the
// b19 helper for why no single UntilTurn value means "the end of
// your next turn" for every seat.
func b20ExileTopUntilEndOfNextTurn(g *game.Game, item *game.StackItem, n int, label string) error {
	controller := item.Controller
	exiled, err := g.ExileTopWithPermissionForEffect(controller, controller, n, game.ExilePlayPermission{
		UntilTurn: g.Turn.Number + 2,
	})
	if err != nil {
		return err
	}
	if len(exiled) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              label,
		Cards:              exiled,
		Effect:             b19EndImpulseGrantWithThisTurn,
	}.Apply(NewContext(g, item))
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
