package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kellan, Planar Trailblazer — Legendary Creature — Human Faerie Scout {R},
// 2/1 (Foundations, #4465):
//
//	"{1}{R}: If Kellan is a Scout, it becomes a Human Faerie Detective and
//	 gains "Whenever Kellan deals combat damage to a player, exile the top
//	 card of your library. You may play that card this turn."
//	 {2}{R}: If Kellan is a Detective, it becomes a 3/2 Human Faerie Rogue
//	 and gains double strike."
//
// Nog, Ambitious Cadet's ladder. Each "becomes" is a record with no stated
// duration, so it lasts as long as Kellan is this object (CR 611.2a, 400.7);
// its creature types are a SET (CR 205.1a), so the Detective is no longer a
// Scout and the Rogue no longer a Detective, and the types are words of
// Kellan's text (CR 612.2). The "if" is checked as the ability resolves (CR
// 608.2c), against the types Kellan has then, so either ability does nothing
// to a Kellan of the wrong type. The base 3/2 is a layer 7b set that starts
// at the ability's timestamp: it overwrites an earlier set and not a later
// one, and counters and pumps still apply (the ruling). The granted trigger
// is an ADR 0093 grant in the same record as the new types; its exile is
// Count on Luck's impulse, "play" so a land may be played as the land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e8fd428f-7521-4e6a-86b7-835f20ee3c5c",
		Name:         "Kellan, Planar Trailblazer",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: kellanDetectiveGrant,
			Triggered: []game.TriggeredAbility{
				WheneverThisDealsCombatDamageToAPlayer("Kellan — exile the top card of your library; you may play it this turn",
					func(g *game.Game, item *game.StackItem) error {
						_, err := b12ImpulseExileForTurn(g, item, 1)
						return err
					}),
			},
			Text: "Whenever Kellan deals combat damage to a player, exile the top card of your library. You may play that card this turn.",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{R}: If Kellan is a Scout, it becomes a Human Faerie Detective and gains \"Whenever Kellan deals combat damage to a player, exile the top card of your library. You may play that card this turn.\"",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{1}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanentWithTextSubtype(g, item, "Scout") {
						return nil
					}
					self := item.SourceCardID
					return GrantAbilitiesFor{
						Target:   self,
						Keys:     []string{kellanDetectiveGrant},
						Also:     []game.Mod{game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Human", "Faerie", "Detective")...)},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Kellan — a Human Faerie Detective",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{2}{R}: If Kellan is a Detective, it becomes a 3/2 Human Faerie Rogue and gains double strike.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{2}{R}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanentWithTextSubtype(g, item, "Detective") {
						return nil
					}
					self := item.SourceCardID
					return ScopedEffectFor{
						Target: self,
						Mods: []game.Mod{
							game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Human", "Faerie", "Rogue")...),
							game.SetBasePowerMod(3),
							game.SetBaseToughnessMod(2),
							game.AddKeywordsMod("double strike"),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), self),
						Label:    "Kellan — a 3/2 Human Faerie Rogue with double strike",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// kellanDetectiveGrant is the bundle the first ability grants: an on-disk
// identity, never renamed, never reused.
const kellanDetectiveGrant = "kellan-planar-trailblazer/detective"
