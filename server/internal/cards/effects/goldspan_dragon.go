package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goldspan Dragon — Creature — Dragon {3}{R}{R}, 4/4:
//
//	"Flying, haste
//	 Whenever this creature attacks or becomes the target of a spell,
//	 create a Treasure token.
//	 Treasures you control have "{T}, Sacrifice this artifact: Add
//	 two mana of any one color.""
//
// Flying and haste ride PrintedKeywords. The trigger is one ability
// watching two conditions (Sun Titan's shape): attackDeclared for the
// attack half, SelfTargetedByASpell's inner check for the "spell"
// half — deliberately NOT EventBecomesTarget's bare "spell or
// ability" reading, because the printed clause says "a spell" only,
// and offering it for a targeting ABILITY too would be stronger than
// printed (#259's direction, the wrong one).
//
// The upgrade is ADR 0093's layer-6 ability grant: every Treasure the
// controller controls gains a bundle that replaces the printed
// tap-and-sacrifice-for-one with tap-and-sacrifice-for-two, via
// OneColorOfAmount(2) — the same one-pick-two-tokens shape Ilysian
// Caryatid's power-4 clause uses. A Treasure made before the Dragon
// or after it, by any source, upgrades the same way: the grant reads
// "Treasures you control" off the board, not off treasures this
// Dragon itself created.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "716b3ea2-45b7-4a8f-af72-de7f4e510eff",
		Name:            "Goldspan Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack, game.EventBecomesTarget},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind == game.EventAttack {
					return attackDeclared(ev, source)
				}
				return SelfTargetedByASpell(ev, source, game.Characteristic{}, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Goldspan Dragon — create a Treasure",
					Do(CreateToken{Template: TreasureToken(), N: 1}))
			},
		}},
		Grants: []AbilityGrant{{
			Key: "goldspan-dragon/treasure-two-mana",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true, Sacrifice: true},
				Produced: OneColorOfAmount(2),
				Label:    "Add two mana of any one color",
			}},
			Text: "{T}, Sacrifice this artifact: Add two mana of any one color.",
		}},
		Static: []game.StaticAbility{
			GrantAbilities(goldspanTreasuresYouControl, "goldspan-dragon/treasure-two-mana"),
		},
	})
}

// goldspanTreasuresYouControl is "Treasures you control" — the
// grantor's own controller, any object with the Treasure subtype.
func goldspanTreasuresYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Treasure")
}
