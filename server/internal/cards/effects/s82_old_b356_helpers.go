package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b356_helpers.go — shared pieces of the S82 OLD batch 356 cards
// (Magic 2015, #5556). The slugs and the body key are on-disk identities:
// never renamed, never reused.

// OldB356DragonToken is Brood Keeper's Dragon for CreateToken.
func OldB356DragonToken() game.Card { return TokenFromCatalog(PrintedOldB356Dragon) }

// PrintedOldB356Dragon is the Dragon as PRINTED (Magic 2014 token): a 2/2 red
// Dragon creature with flying and "{R}: This creature gets +1/+0 until end of
// turn."
func PrintedOldB356Dragon() TokenTemplate {
	return TokenTemplate{
		Slug: "dragon-red-pump",
		Card: game.Card{
			Name:      "Dragon",
			TypeLine:  "Token Creature — Dragon",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R"},
			Keywords:  []string{"flying"},
		},
		Activated: []game.ActivatedAbilityShape{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Cost:    game.AbilityCost{Mana: "{R}"},
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect:  ThisGetsUntilEndOfTurn(1, 0, "Dragon — +1/+0 until end of turn"),
		}},
		Text: "Flying\n{R}: This creature gets +1/+0 until end of turn.",
	}
}

// OldB356LandMineToken is Goblin Kaboomist's Land Mine for CreateToken.
func OldB356LandMineToken() game.Card { return TokenFromCatalog(PrintedOldB356LandMine) }

// PrintedOldB356LandMine is the Land Mine as PRINTED (Magic 2015 token): a
// colorless artifact token with "{R}, Sacrifice this artifact: This artifact
// deals 2 damage to target attacking creature without flying." The damage is
// dealt by the Land Mine as it last existed (CR 608.2h).
func PrintedOldB356LandMine() TokenTemplate {
	return TokenTemplate{
		Slug: "land-mine",
		Card: game.Card{
			Name:     "Land Mine",
			TypeLine: "Token Artifact",
		},
		Activated: []game.ActivatedAbilityShape{{
			Label:   "{R}, Sacrifice this artifact: This artifact deals 2 damage to target attacking creature without flying.",
			Cost:    game.AbilityCost{Mana: "{R}", SacrificeSelf: true},
			Targets: TargetCreature("target attacking creature without flying", IsAttackingCreature(), WithoutKeyword("flying")),
			Purpose: game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(DamageToTarget(0, 2))},
			Effect:  DamageToFirstTarget(2),
		}},
		Text: "{R}, Sacrifice this artifact: This artifact deals 2 damage to target attacking creature without flying.",
	}
}

// OldB356TreefolkToken is Kalonian Twingrove's Treefolk Warrior for CreateToken.
func OldB356TreefolkToken() game.Card { return TokenFromCatalog(PrintedOldB356Treefolk) }

// PrintedOldB356Treefolk is the Treefolk Warrior as PRINTED (Magic 2015
// token): a green Treefolk Warrior creature token whose power and toughness
// are each equal to the number of Forests you control, a characteristic-
// defining ability (CR 604.3).
func PrintedOldB356Treefolk() TokenTemplate {
	return TokenTemplate{
		Slug: "treefolk-warrior-forests",
		Card: game.Card{
			Name:     "Treefolk Warrior",
			TypeLine: "Token Creature — Treefolk Warrior",
			Colors:   []string{"G"},
		},
		Static: []game.StaticAbility{OldB356ForestsPT("This creature's power and toughness are each equal to the number of Forests you control.")},
		Text:   "This creature's power and toughness are each equal to the number of Forests you control.",
	}
}

// OldB356ForestsPT is "<this>'s power and toughness are each equal to the
// number of Forests you control" (Kalonian Twingrove and its token): a layer 7a
// characteristic-defining ability read live. "Forest" is a word of the
// source's text (CR 612.2).
func OldB356ForestsPT(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := B08LandsWithSubtypeControlled(g, source.Controller, source.TextSubtype("Forest"))
			c.Power, c.Toughness = n, n
		},
	}
}

// OldB356ReturnListedTappedBody is the delayed body that returns the listed
// cards from a graveyard to the battlefield tapped under their owners'
// control (Phytotitan). A card that is no longer in a graveyard, or that has
// left and come back, is a different object and is left alone (CR 400.7).
var OldB356ReturnListedTappedBody = game.SimpleDelayedBody("old-b356/return-listed-tapped", oldB356ReturnListedTapped)

func oldB356ReturnListedTapped(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard {
			continue
		}
		if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
