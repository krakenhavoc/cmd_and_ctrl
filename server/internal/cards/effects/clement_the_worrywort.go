package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clement, the Worrywort — Legendary Creature — Frog Druid {1}{G}{U}, 3/3
// (Bloomburrow, #4467):
//
//	"Vigilance
//	 Whenever Clement or another creature you control enters, return up
//	 to one target creature you control with lesser mana value to its
//	 owner's hand.
//	 Frogs you control have "{T}: Add {G} or {U}. Spend this mana only to
//	 cast a creature spell.""
//
// The trigger is Bogwater Lumaret's "this or another creature you control
// enters" (CreatureEnteredUnderYourControl). "Lesser" is lesser than the
// mana value of the creature that entered, so the clause is built from the
// trigger as it goes on the stack (TargetsFrom, Scrap Trawler's seam,
// CR 603.3d), off the entering creature's snapshot. The creature that
// entered never qualifies (its own mana value is not lesser), and "up to
// one" lets you choose nothing. The target is checked again as the ability
// resolves (CR 608.2b).
//
// The grant is Manaweft Sliver's over Frogs you control (Clement among
// them), with Beastcaller Savant's spend restriction on the mana: both
// tags, so it pays for a creature spell and not for a creature's ability.
// "Frog" reads through Clement's text (CR 612.2).
//
// No simplification.
const clementFrogMana = "clement-the-worrywort/frog-mana"

func init() {
	Register(Spec{
		OracleID:        "57ce5ebd-cfac-4bb9-b010-04d0c285be27",
		Name:            "Clement, the Worrywort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventETB},
			AppliesTo:   CreatureEnteredUnderYourControl,
			TargetsFrom: clementCheaperCreatureClause,
			Key:         "Clement, the Worrywort — return up to one target creature you control with lesser mana value to its owner's hand",
			Effect:      bounceEachLegalTarget,
		}},
		Grants: []AbilityGrant{{
			Key: clementFrogMana,
			Mana: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{G|U}",
				Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")},
				Label:        "Add {G} or {U}. Spend this mana only to cast a creature spell",
			}},
			Text: "{T}: Add {G} or {U}. Spend this mana only to cast a creature spell.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Frog"}, YoursOnly: true}, clementFrogMana),
		},
	})
}

// clementCheaperCreatureClause is "up to one target creature you control
// with lesser mana value", lesser than the creature that entered.
func clementCheaperCreatureClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	ceiling := 0
	if tc.Object != nil {
		ceiling = tc.Object.ManaValue
	}
	return UpToOneTargetCreature("up to one target creature you control with lesser mana value",
		YouControl(), ManaValueLE(ceiling-1))
}
