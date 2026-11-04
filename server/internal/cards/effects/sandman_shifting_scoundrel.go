package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sandman, Shifting Scoundrel — Legendary Creature — Sand Elemental Villain
// {1}{G}{G}, */* (Marvel's Spider-Man promos, #4460):
//
//	"Sandman's power and toughness are each equal to the number of lands
//	 you control.
//	 Sandman can't be blocked by creatures with power 2 or less.
//	 {3}{G}{G}: Return this card and target land card from your graveyard
//	 to the battlefield tapped."
//
// The power and toughness are Ashaya's characteristic-defining ability
// (CR 604.3, layer 7a), and the block bar is Legolas Greenleaf's, read as
// blockers are declared (CR 509.1b).
//
// The last ability works from the graveyard (CR 113.6) and targets a land
// card in the same graveyard. As it resolves, Sandman and the land are put
// onto the battlefield tapped as ONE simultaneous entry, under the
// activator's control (it is their graveyard, so they own both). If the
// land is no longer a legal target the ability does not resolve at all
// (CR 608.2b: its only target is illegal), so Sandman stays where he is; a
// Sandman that left the graveyard in response stays where he went while
// the land still returns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b205480-b1fb-490c-b1c9-cf5cc3c5acdb",
		Name:         "Sandman, Shifting Scoundrel",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			landsYouControlPT("Sandman's power and toughness are each equal to the number of lands you control."),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}{G}: Return this card and target land card from your graveyard to the battlefield tapped.",
			Cost:    ManaCost("{3}{G}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn()),
			Effect:  sandmanReturnBoth,
		}},
	})
}

// sandmanReturnBoth puts Sandman and the target land card, whichever are
// still in the graveyard, onto the battlefield tapped together.
func sandmanReturnBoth(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var entries []game.BatchEntry
	land, _ := b16FirstLegalTargetCard(ctx)
	for _, id := range []uuid.UUID{item.SourceCardID, land} {
		if id == uuid.Nil {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneGraveyard})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(entries,
		game.ZoneEntryOptions{Controller: item.Controller, Tapped: true}, game.KeyedTail{})
}
