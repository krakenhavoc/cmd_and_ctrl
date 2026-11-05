package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nog, Ambitious Cadet — Legendary Creature — Ferengi Rogue {B}{G}, 2/2
// (Star Trek, #4973):
//
//	"Menace
//	 {1}: If Nog is a Rogue, he becomes a Ferengi Scout with base power
//	 and toughness 3/4 and gains "Whenever Nog becomes tapped, surveil 2."
//	 {3}: If Nog is a Scout, he becomes a Ferengi Officer and gains "You
//	 may play lands from your graveyard" and "You may play an additional
//	 land on each of your turns.""
//
// Frodo, Sauron's Bane's ladder. Each "becomes" is a ScopedEffect record
// with no stated duration, so it lasts as long as Nog is this object (CR
// 611.2a, 400.7), and its creature types are a SET (CR 205.1a): the Scout
// is no longer a Rogue and the Officer is no longer a Scout. The "if" is
// checked as the ability resolves (CR 608.2c), against the creature types
// Nog has then, so the second ability on a Rogue does nothing and the
// first on a Scout does nothing. A later record does not take the earlier
// one's abilities: the Officer keeps the Scout's base 3/4 and its trigger.
//
// The gained text rides ADR 0093 grants in the same record as the new
// types, one timestamp. "Whenever Nog becomes tapped" is ThisBecameTapped
// over both tap events (an attack taps without an EventTapCard). The
// Officer's two statics are facts about Nog's controller, carried on the
// granted bundle and read live off Nog's composite key, exactly as
// Exploration's and Crucible of Worlds' own are: play, so a land from the
// graveyard spends a land drop (CR 305.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a8ae1c0-1935-4d03-aa94-9e9088650821",
		Name:            "Nog, Ambitious Cadet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Grants: []AbilityGrant{
			{
				Key: nogAmbitiousCadetScoutGrant,
				Triggered: []game.TriggeredAbility{
					OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
						"Nog, Ambitious Cadet — surveil 2", Do(Surveil{N: 2})),
				},
				Text: "Whenever Nog becomes tapped, surveil 2.",
			},
			{
				Key:                 nogAmbitiousCadetOfficerGrant,
				AdditionalLandPlays: 1,
				CastPermissions: []game.CastPermission{{
					Zone:   game.ZoneGraveyard,
					Filter: game.PermissionFilter{LandsOnly: true},
					Label:  "Play a land from your graveyard (Nog, Ambitious Cadet)",
				}},
				Text: "You may play lands from your graveyard. You may play an additional land on each of your turns.",
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}: If Nog is a Rogue, he becomes a Ferengi Scout with base power and toughness 3/4 and gains \"Whenever Nog becomes tapped, surveil 2.\"",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !nogAmbitiousCadetIs(g, item, "Rogue") {
						return nil
					}
					self := item.SourceCardID
					return GrantAbilitiesFor{
						Target: self,
						Keys:   []string{nogAmbitiousCadetScoutGrant},
						Also: []game.Mod{
							game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Ferengi", "Scout")...),
							game.SetBasePowerMod(3),
							game.SetBaseToughnessMod(4),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Nog, Ambitious Cadet — a 3/4 Ferengi Scout",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{3}: If Nog is a Scout, he becomes a Ferengi Officer and gains \"You may play lands from your graveyard\" and \"You may play an additional land on each of your turns.\"",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{3}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !nogAmbitiousCadetIs(g, item, "Scout") {
						return nil
					}
					self := item.SourceCardID
					return GrantAbilitiesFor{
						Target:   self,
						Keys:     []string{nogAmbitiousCadetOfficerGrant},
						Also:     []game.Mod{game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Ferengi", "Officer")...)},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Nog, Ambitious Cadet — a Ferengi Officer",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

const (
	nogAmbitiousCadetScoutGrant   = "nog-ambitious-cadet/scout"
	nogAmbitiousCadetOfficerGrant = "nog-ambitious-cadet/officer"
)

// nogAmbitiousCadetIs is the abilities' "if Nog is a <type>", read as the
// ability resolves: Nog must still be the permanent that activated it (CR
// 400.7) and have that creature type now. The type is a word of Nog's text
// (CR 612.2).
func nogAmbitiousCadetIs(g *game.Game, item *game.StackItem, subtype string) bool {
	if !sourceIsStillThisPermanent(g, item) {
		return false
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	return ok && c.HasSubtype(g.TextReaderForEffect(item).TextSubtype(subtype))
}
