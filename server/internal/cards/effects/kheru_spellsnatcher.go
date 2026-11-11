package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kheru Spellsnatcher — Creature — Snake Wizard {3}{U}, 3/3 (Duskmourn
// Commander):
//
//	"Morph {4}{U}{U}
//	 When this creature is turned face up, counter target spell. If that
//	 spell is countered this way, exile it instead of putting it into its
//	 owner's graveyard. You may cast that card without paying its mana
//	 cost for as long as it remains exiled."
//
// Transcendent Dragon's counter-and-exile (CounterTarget with an exile
// destination): a spell that can't be countered stays on the stack and is
// neither exiled nor offered, and a commander whose owner chose the
// command zone isn't in exile either. Unlike the Dragon's one-shot offer,
// the permission lasts "for as long as it remains exiled": a standing
// {0} cast-only grant over the one card (the plot grant's shape, without
// plot's timing), so the controller may cast it later, with the card's own
// timing (a sorcery waits for a main phase), and it ends when the card
// leaves exile. The turned-up Spellsnatcher's controller holds it, not the
// spell's owner.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c01411e0-77b2-4e65-a369-5dbe13745769",
		Name:             "Kheru Spellsnatcher",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{4}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisIsTurnedFaceUp("Kheru Spellsnatcher — counter target spell; if countered this way, exile it, and you may cast it without paying its mana cost for as long as it remains exiled",
					kheruSpellsnatcherCounter),
				TargetSpell("target spell")),
		},
	})
}

func kheruSpellsnatcherCounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	stackID := item.Targets[0].ID
	if g.StackItemForEffect(stackID) == nil {
		return nil
	}
	if err := (CounterTarget{StackID: stackID, Dest: game.ZoneRef{Kind: game.ZoneExile}}).Apply(ctx); err != nil {
		return err
	}
	// Only a spell that really reached exile was "countered this way".
	if z := g.FindCardZoneForEffect(stackID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	c, ok := g.LookupCardForEffect(stackID)
	if !ok {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     item.Controller,
		Zone:       game.ZoneExile,
		Cost:       "{0}",
		CastOnly:   true,
		Duration:   game.WhileInZoneDuration(),
		Source:     item.SourceCardID,
		SourceName: "Kheru Spellsnatcher",
		Label:      "Kheru Spellsnatcher — you may cast it without paying its mana cost for as long as it remains exiled",
	}, []game.Card{c})
	return nil
}
