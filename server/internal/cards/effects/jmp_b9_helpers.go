package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// jmp_b9_helpers.go — the shared pieces of the Jumpstart 2020 deck
// request batch 9 (#4894, Jumpstart Azorius).

// jmpB9AFlierAttacked is "whenever a creature with flying attacks": a
// creature, under anyone's control, declared as an attacker while it has
// flying (Windreader Sphinx). One event per attacking creature.
func jmpB9AFlierAttacked(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
}

// jmpB9YouAttackedWithAFlier is one attack event counted by "whenever you
// attack with two or more creatures with flying" (Tide Skimmer): a
// creature with flying the source's controller declared as an attacker.
// The count of two is the ability's AtBatchEnd condition.
func jmpB9YouAttackedWithAFlier(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && game.HasKeyword(&c, "flying")
}

// jmpB9FreeCastTargetExilingInstantsAndSorceries is Torrential
// Gearhulk's free cast of the trigger's target card from your graveyard
// (freeCastTargetFromYourGraveyardExilingIt), with Bilbo, Thief in the
// Night's narrower replacement: "If an instant or sorcery spell cast this
// way would be put into your graveyard, exile it instead"
// (ExileInsteadOnlyInstantOrSorcery, #4505). An artifact spell cast this
// way that is countered goes to the graveyard, as printed (Scholar of the
// Lost Trove).
func jmpB9FreeCastTargetExilingInstantsAndSorceries(sourceName string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			z := g.FindCardZoneForEffect(t.ID)
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok || z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:                           item.Controller,
				Zone:                             game.ZoneGraveyard,
				Duration:                         g.UntilEndOfTurnDuration(),
				Cost:                             "{0}",
				AltCostKey:                       game.AltCostKeyGrantedCost,
				Timing:                           game.TimingFlash,
				CastOnly:                         true,
				ExileInsteadOfGraveyard:          true,
				ExileInsteadOnlyInstantOrSorcery: true,
				// A pass closes the window; nothing moves, the card
				// stays in the graveyard.
				LapseOnPass: game.LapseStaysInHand,
				Source:      item.SourceCardID,
				SourceName:  sourceName,
				Label:       sourceName + " — cast " + c.Name + " without paying its mana cost",
			}, []game.Card{c})
			// #4250: cast during the resolution (CR 608.2g).
			offerResolutionCastOver(g, item.Controller, item.SourceCardID, sourceName, game.MayCastKeywordFreeCast, declineLeaveInGraveyard, c.InstanceID)
			return nil
		}
		return nil
	}
}
